import WebSocketManager from "../WebSocketManager";

export async function broadcast(ws: WebSocketManager, data: any): Promise<void> {
  if (!ws.gameID) {
    throw new Error("Cannot broadcast: not in a room");
  }

  const payload = typeof data === "string" ? data : JSON.stringify(data);
  
  // Just send - connection auto-connects if needed
  await ws.connection.send("broadcast", {
    room_id: ws.gameID,
    payload: payload,
  });
}

