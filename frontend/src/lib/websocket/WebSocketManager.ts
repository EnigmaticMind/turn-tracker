import { WebSocketConnection } from "./WebSocketConnection";
import type { Message } from "./handlers/types";

export default class WebSocketManager {
  private _connection: WebSocketConnection;
  private _gameID?: string;
  private clientID?: string;

  constructor() {
    this._connection = new WebSocketConnection();
    this.setupMessageHandlers();
  }

  public get connection(): WebSocketConnection {
    return this._connection;
  }

  public get gameID(): string {
    if (!this._gameID) throw new Error("Game ID not set when trying to access it");
    return this._gameID;
  }

  private set gameID(gameID: string) {
    this._gameID = gameID.toUpperCase();
  }

  private setupMessageHandlers() {
    this._connection.onMessage((message: Message) => {
      switch (message.type) {
        case "room_created":
          if (message.data?.room_id) {
            this.gameID = message.data.room_id;
            this.clientID = message.data.peers?.[0]; // First peer is typically the creator
          }
          break;

        case "room_joined":
          if (message.data?.room_id) {
            this.gameID = message.data.room_id;
            // Find our client ID in the peers list (last one is typically the joiner)
            const peers = message.data.peers || [];
            if (peers.length > 0) {
              this.clientID = peers[peers.length - 1];
            }
          }
          break;

        case "error":
          console.error("Server error:", message.data?.message);
          break;
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
    this.clientID = undefined;
  }
}

