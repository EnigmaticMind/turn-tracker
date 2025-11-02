import { WebSocketConnection } from "./WebSocketConnection";
import type { Message, PeerInfo, PlayerJoinedData, PlayerLeftData, ProfileUpdatedData, RoomCreatedData, RoomJoinedData, TurnChangedData } from "./handlers/types";
import { saveClientID } from "../utils/gamePersistence";

export default class WebSocketManager {
  private _connection: WebSocketConnection;
  private _gameID?: string;
  private _clientID?: string;
  private _displayName?: string;
  private _color?: string;
  private _peers: PeerInfo[] = [];
  private _currentTurn: PeerInfo | null = null;
  private _turnStartTime: number | null = null; // Server timestamp in milliseconds
  private _lastTurnSequence: number = 0; // Last processed turn_changed sequence number
  private peerCallbacks: Set<(peers: PeerInfo[]) => void> = new Set();
  private turnCallbacks: Set<(turn: PeerInfo | null, turnStartTime: number | null) => void> = new Set();

  constructor() {
    this._connection = new WebSocketConnection();
    this._connection.setPersistent(true); // Make connection persistent
    this.setupMessageHandlers();
  }

  public get connection(): WebSocketConnection {
    return this._connection;
  }

  public get gameID(): string | undefined {
    return this._gameID;
  }

  public get displayName(): string | undefined {
    return this._displayName;
  }

  public get color(): string | undefined {
    return this._color;
  }

  public get clientID(): string | undefined {
    return this._clientID;
  }

  public setClientID(clientID: string): void {
    this._clientID = clientID;
    // Also set it on the connection so it can use it when connecting
    this._connection.setClientID(clientID);
  }

  public get peers(): PeerInfo[] {
    return [...this._peers];
  }

  public get currentTurn(): PeerInfo | null {
    return this._currentTurn;
  }

  public get turnStartTime(): number | null {
    return this._turnStartTime;
  }

  private set gameID(gameID: string) {
    this._gameID = gameID.toUpperCase();
  }

  private updatePeers(peers: PeerInfo[]) {
    this._peers = peers;
    this.peerCallbacks.forEach((cb) => cb(this._peers));
  }

  private updateCurrentTurn(turn: PeerInfo | null, turnStartTime: number | null = null) {
    this._currentTurn = turn;
    this._turnStartTime = turnStartTime;
    this.turnCallbacks.forEach((cb) => cb(this._currentTurn, this._turnStartTime));
  }

  public onPeersChange(callback: (peers: PeerInfo[]) => void): () => void {
    console.log("onPeersChange", this._peers);
    this.peerCallbacks.add(callback);
    // Immediately call with current peers
    callback(this._peers);
    return () => {
      this.peerCallbacks.delete(callback);
    };
  }

  public onTurnChange(callback: (turn: PeerInfo | null, turnStartTime: number | null) => void): () => void {
    this.turnCallbacks.add(callback);
    // Immediately call with current turn
    callback(this._currentTurn, this._turnStartTime);
    return () => {
      this.turnCallbacks.delete(callback);
    };
  }

  private setupMessageHandlers() {
    this._connection.onMessage((message: Message) => {
      switch (message.type) {
        case "room_created":
          if (message.data?.room_id) {
            this.gameID = message.data.room_id;
            const data = message.data as RoomCreatedData;
            const peers: PeerInfo[] = data.peers || [];
            const currentTurn: PeerInfo | null = data.current_turn || null;
            // Reset turn sequence when joining a new room
            this._lastTurnSequence = 0;
            // Use your_client_id from the message to identify ourselves
            if (data.your_client_id) {
              this._clientID = data.your_client_id;
              // Find our peer info to get display name and color
              const myPeer = peers.find((p) => p.client_id === data.your_client_id);
              if (myPeer) {
                this._displayName = myPeer.display_name;
                this._color = myPeer.color;
              }
              saveClientID(this._clientID);
            }
            this.updatePeers(peers);
            this.updateCurrentTurn(currentTurn, null);
          }
          break;

        case "room_joined":
          if (message.data?.room_id) {
            this.gameID = message.data.room_id;
            const data = message.data as RoomJoinedData;
            const peers: PeerInfo[] = data.peers || [];
            const currentTurn: PeerInfo | null = data.current_turn || null;
            // Reset turn sequence when joining a new room
            this._lastTurnSequence = 0;
            // Use your_client_id from the message to identify ourselves
            if (data.your_client_id) {
              this._clientID = data.your_client_id;
              saveClientID(this._clientID);
            }
            this.updatePeers(peers);
            // Note: room_joined doesn't include turn_start_time in current version
            this.updateCurrentTurn(currentTurn, null);
          }
          break;

        case "player_joined": {
          const data = message.data as PlayerJoinedData;
          // Add new peer to list
          const newPeer: PeerInfo = {
            client_id: data.peer_id,
            display_name: data.display_name,
            color: data.color,
            total_turn_time: 0, // New players start with 0 turn time
          };
          this.updatePeers([...this._peers, newPeer]);
          break;
        }

        case "player_left": {
          const data = message.data as PlayerLeftData;
          // Remove peer from list
          this.updatePeers(this._peers.filter((p) => p.client_id !== data.peer_id));
          // Clear turn if it was the current turn player
          if (this._currentTurn?.client_id === data.peer_id) {
            this.updateCurrentTurn(null, null);
          }
          break;
        }

        case "profile_updated": {
          const data = message.data as ProfileUpdatedData;
          // Check if this is our own profile update
          if (data.peer_id === this._clientID) {
            // Update local display name and color
            this._displayName = data.display_name;
            this._color = data.color;
          }
          // Update peer info in list (use total_turn_time from message)
          this.updatePeers(
            this._peers.map((p) =>
              p.client_id === data.peer_id
                ? {
                    client_id: data.peer_id,
                    display_name: data.display_name,
                    color: data.color,
                    total_turn_time: data.total_turn_time ?? p.total_turn_time, // Use from message, fallback to existing
                  }
                : p
            )
          );
          break;
        }

        case "turn_changed": {
          const data = message.data as TurnChangedData;
          // Ignore stale turn_changed messages (out of order or duplicate)
          if (data.sequence && data.sequence <= this._lastTurnSequence) {
            console.log(`Ignoring stale turn_changed message: sequence ${data.sequence} <= ${this._lastTurnSequence}`);
            break;
          }
          // Update sequence number and process turn change
          if (data.sequence) {
            this._lastTurnSequence = data.sequence;
          }
          this.updateCurrentTurn(data.current_turn, data.turn_start_time ?? null);
          break;
        }

        case "error": {
          const errorMessage = message.data?.message || "An error occurred";
          console.error("Server error:", errorMessage);
          // Dispatch error toast event
          window.dispatchEvent(
            new CustomEvent("toast", {
              detail: {
                type: "error",
                message: errorMessage,
              },
            })
          );
          break;
        }
      }
    });
  }

  public onMessage(callback: (message: Message) => void): () => void {
    return this._connection.onMessage(callback);
  }

  public destroy(): void {
    console.log("Destroying WebSocketManager");
    this._connection.destroy();
    this._gameID = undefined;
    // Don't clear _clientID - keep it for reconnection
    this._peers = [];
    this._currentTurn = null;
    this.peerCallbacks.clear();
    this.turnCallbacks.clear();
  }
}

