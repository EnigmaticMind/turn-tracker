import WebSocketManager from "../WebSocketManager";

export async function updateProfile(
  ws: WebSocketManager,
  displayName?: string,
  color?: string
): Promise<void> {
  // Build data object - only include fields that are provided
  const data: any = {};
  if (displayName !== undefined) {
    data.display_name = displayName;
  }
  if (color !== undefined) {
    data.color = color;
  }

  // Just send - connection auto-connects if needed
  await ws.connection.send("update_profile", data);
}

