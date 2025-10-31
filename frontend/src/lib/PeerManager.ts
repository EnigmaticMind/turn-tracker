import Peer from 'peerjs';
import type { DataConnection, PeerOptions } from 'peerjs';

const UNAVAILABLE_PEER_ID = "unavailable-id"
const PEER_ID_PREFIX = "TURN-TRACKER-"
const RECONNECT_DELAY = 2000;

const PEER_OPTIONS: PeerOptions = { 
  host: 'localhost',
  port: 9000,
  secure: false,
  key: 'peerjs',
  path: '/',
  // debug: 3,
  config: {
    iceServers: [
      { urls: 'stun:stun1.l.google.com:19302' },
    ],
  },
};

export default class PeerManager {
  // Our PeerJS connection
  private peer: Peer | null = null;

  // Store the game ID for our peer
  private _gameID?: string;
  reconnecting: any;
  connections: Map<string, DataConnection> = new Map();
  events: any;

  constructor() {}

  // Get the game ID, removing the prefix
  public get gameID() {
    if (!this._gameID) throw new Error("Game ID not set when trying to access it");

    return this._gameID;
  }

  // Set the game ID, removing the prefix
  private set gameID(gameID: string) {
    this._gameID = gameID.replace(PEER_ID_PREFIX, '');
  }

  // Generate a new game ID (6 chars, A-Z0-9)
  private generateNewGameID() {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    let id = "";
    for (let i = 0; i < 6; i++) {
      id += chars[Math.floor(Math.random() * chars.length)];
    }
    return id;
  }

  // Validate the game ID
  private isValidGameID(gameID: string) {
    return (gameID && gameID.length === 6 && /^[A-Z0-9]+$/.test(gameID));
  
  }

  // Generate a PeerJS ID for our peer
  // This includes our prefix
  private generatePeerID(gameID: string) {
    if (!this.isValidGameID(gameID)) throw new Error("Invalid game ID");
    const generatedID = PEER_ID_PREFIX + gameID
    console.log('Generated Peer ID:', generatedID);
    
    return generatedID
  }

  // Create a new game
  public createGame(gameID: string = this.generateNewGameID()) {
    console.log('Creating game:', gameID);

    // Create a new PeerJS connection and assign to this.peer
    this.peer = new Peer(this.generatePeerID(gameID), PEER_OPTIONS);

    // Wire up standard listeners and resolve once 'open' fires (with timeout)
    return this.setupPeerJS((id: string) => {
      this.gameID = id;
    });
  }

  // Join a game (connect to an existing host). Does NOT create a game.
  public joinGame(gameID: string) {
    console.log('Joining game:', gameID);
    const remotePeerId = this.generatePeerID(gameID);
    console.log('Remote Peer ID:', remotePeerId);

    // Create our own ephemeral peer identity
    this.peer = new Peer(PEER_OPTIONS);
    this._gameID = gameID;

    return this.setupPeerJS(() => {
      console.log('Joining peer game: ', remotePeerId);
      console.log('All peers:', this.peer?.listAllPeers());
      // Connect to the remote host peer for this game
      const conn = this.peer?.connect(remotePeerId, { reliable: true });
      return this.setupDataConnection(conn!);
    });
  }
  
  // Attempt to reconnect to the PeerJS server
  private reconnectPeerJS() {
    if (this.reconnecting) return;
    this.reconnecting = true;

    setTimeout(() => {
      if (this.peer && this.peer.disconnected) {
        console.log("Reconnecting to PeerJS server...");
        this.peer.reconnect();
        this.reconnecting = false;
      }
    }, RECONNECT_DELAY);
  }

  private setupPeerJS(onOpen: (id: string) => void) {
    if (!this.peer) throw new Error("Peer not initialized");

    return new Promise((resolve, reject) => {
      console.log('Setting up PeerJS listeners', this.peer);

      // Signaling connected; you have a PeerID
      this.peer?.on("open", async (id: string) => {
        console.warn("peer:open", id);

        await onOpen(id);
        resolve(id);
      });

      // Incoming data connection from a remote peer
      this.peer?.on("connection", (conn) => {
        console.warn("peer:connection from", conn.peer);
        this.setupDataConnection(conn); // ensure this adds open/data/close/error handlers
      });

      // Lost signaling connection (can attempt peer.reconnect())
      this.peer?.on("disconnected", () => {
        console.warn("peer:disconnected (signaling lost)");
        this.reconnectPeerJS();
      });

      // Peer destroyed; all conns closed
      this.peer?.on("close", (a) => {
        console.warn("peer:close", a);
        this.destroy();
        reject(new Error("Peer closed"));
      });

      // Any peer-level error (e.g. 'peer-unavailable', network issues)
      this.peer?.on("error", (err: any) => {
        console.error("peer:error", err?.type || err, err);
        this.destroy();
        reject(err instanceof Error ? err : new Error(String(err?.message || err)));
      });
    });
  }

  private setupDataConnection(conn: DataConnection) {
    console.log("setting up data connection", conn);

    return new Promise((resolve, reject) => {
      conn.on("open", () => {
        console.warn("data:open opened to:", conn.peer);
        this.connections.set(conn.peer, conn);
        
        this.events?.onConnect?.(conn.peer);
        resolve(conn);
      });

      conn.on("data", (data) => {
        console.warn("data:data from:", conn.peer, data);

        this.events?.onMessage?.(conn.peer, data);
      });

      conn.on("close", () => {
        console.warn("data:closed closed:", conn.peer);
        this.connections.delete(conn.peer);

        this.events?.onDisconnect?.(conn.peer);
        reject();
      });

      conn.on("error", (err) => {
        console.warn("data:error :", err);
        this.events?.onError?.(err);
        reject(err);
      });
    });
  }

  public broadcast(data: any) {
    this.connections.forEach((conn) => {
      if (conn.open) conn.send(data);
    });
  }
      

  // Destroy the PeerManager events
  public destroy() {
    console.log('Destroying PeerManager');
    this.connections.forEach((conn) => conn.close());
    this.connections.clear();
    this.peer?.destroy();
    this.peer = null;
    
  }
}

// import Peer, { DataConnection } from "peerjs";

// export interface PeerManagerEvents {
//   onConnect?: (peerId: string) => void;
//   onDisconnect?: (peerId: string) => void;
//   onMessage?: (peerId: string, data: any) => void;
//   onError?: (err: any) => void;
// }

// export class PeerManager {
//   private peer: Peer | null = null;
//   private connections: Map<string, DataConnection> = new Map();
//   private events: PeerManagerEvents;
//   private reconnecting = false;

//   constructor(events: PeerManagerEvents = {}) {
//     this.events = events;
//   }

//   /**
//    * Generate or join a short room code like "ABC123"
//    */
//   public static generateRoomCode(): string {
//     const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
//     return Array.from({ length: 6 })
//       .map(() => chars[Math.floor(Math.random() * chars.length)])
//       .join("");
//   }

//   /**
//    * Initialize the PeerJS instance
//    */
//   async init(roomCode?: string): Promise<string> {
//     return new Promise((resolve, reject) => {
//       const peerId = roomCode || PeerManager.generateRoomCode();

//       this.peer = new Peer(peerId, {
//         host: "0.peerjs.com",
//         port: 443,
//         secure: true,
//         config: {
//           iceServers: [
//             { urls: "stun:stun.l.google.com:19302" },
//             { urls: "stun:stun1.l.google.com:19302" },
//           ],
//         },
//       });

//       this.peer.on("open", (id) => {
//         console.log("PeerJS connected as:", id);
//         resolve(id);
//       });

//       this.peer.on("connection", (conn) => {
//         this.setupConnection(conn);
//       });

//       this.peer.on("disconnected", () => {
//         console.warn("PeerJS disconnected. Attempting to reconnect...");
//         this.reconnect();
//       });

//       this.peer.on("error", (err) => {
//         console.error("PeerJS error:", err);
//         this.events.onError?.(err);
//         reject(err);
//       });
//     });
//   }

//   /**
//    * Connect to another peer
//    */
//   connectTo(roomCode: string) {
//     if (!this.peer) throw new Error("Peer not initialized");

//     const conn = this.peer.connect(roomCode, {
//       reliable: true,
//     });

//     this.setupConnection(conn);
//   }

//   /**
//    * Send data to all connected peers
//    */
//   broadcast(data: any) {
//     this.connections.forEach((conn) => {
//       if (conn.open) conn.send(data);
//     });
//   }

//   /**
//    * Disconnect everything gracefully
//    */
//   destroy() {
//     this.connections.forEach((conn) => conn.close());
//     this.connections.clear();
//     this.peer?.destroy();
//     this.peer = null;
//   }

//   /**
//    * Attempt automatic reconnect
//    */
//   private reconnect() {
//     if (this.reconnecting) return;
//     this.reconnecting = true;

//     setTimeout(() => {
//       if (this.peer && this.peer.disconnected) {
//         console.log("Reconnecting to PeerJS server...");
//         this.peer.reconnect();
//         this.reconnecting = false;
//       }
//     }, 2000);
//   }

//   /**
//    * Setup listeners for a DataConnection
//    */
//   private setupConnection(conn: DataConnection) {
//     conn.on("open", () => {
//       console.log("Connected to:", conn.peer);
//       this.connections.set(conn.peer, conn);
//       this.events.onConnect?.(conn.peer);
//     });

//     conn.on("data", (data) => {
//       this.events.onMessage?.(conn.peer, data);
//     });

//     conn.on("close", () => {
//       console.log("Connection closed:", conn.peer);
//       this.connections.delete(conn.peer);
//       this.events.onDisconnect?.(conn.peer);
//     });

//     conn.on("error", (err) => {
//       console.error("Connection error:", err);
//       this.events.onError?.(err);
//     });
//   }
// }
