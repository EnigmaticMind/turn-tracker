import WebSocketManager from "../WebSocketManager";

export async function startTurn(ws: WebSocketManager, newTurnClientID?: string): Promise<void> {
  // Get current turn from client's state (null becomes empty string)
  const currentTurn = ws.currentTurn?.client_id || "";

  // If newTurnClientID is not provided, send empty string to end turn
  const newTurn = newTurnClientID || "";

  // Just send - connection auto-connects if needed
  await ws.connection.send("start_turn", {
    current_turn: currentTurn,
    new_turn: newTurn,
  });
}

