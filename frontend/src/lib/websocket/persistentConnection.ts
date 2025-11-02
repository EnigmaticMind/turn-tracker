import WebSocketManager from "./WebSocketManager";
import { getPersistedClientID } from "../utils/gamePersistence";

let persistentWS: WebSocketManager | null = null;

export function getPersistentConnection(): WebSocketManager {
  if (!persistentWS) {
    console.log("getPersistentConnection: Creating new WebSocketManager");
    persistentWS = new WebSocketManager();
    
    // Get persisted clientID and set it for reconnection
    const persistedClientID = getPersistedClientID();
    if (persistedClientID) {
      persistentWS.setClientID(persistedClientID);
    }
  }
  return persistentWS;
}

export function destroyPersistentConnection(): void {
  if (persistentWS) {
    persistentWS.destroy();
    persistentWS = null;
    // Keep clientID in localStorage for reconnection
  }
}

