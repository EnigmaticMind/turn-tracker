import WebSocketManager from "../WebSocketManager";
import { startTurn } from "./startTurn";

export async function endTurn(ws: WebSocketManager): Promise<void> {
  // End turn by calling startTurn with no new turn client ID
  await startTurn(ws);
}

