import WebSocketManager from "../WebSocketManager";
import { getDefaultProfile } from "../../utils/userProfile";
import type { Message } from "./types";

export async function createGame(ws: WebSocketManager): Promise<Message> {
  console.log("Creating game (backend will generate ID)");

  // Get defaults from localStorage (only if they exist)
  const defaults = getDefaultProfile();

  // Build data object - only include fields that exist
  const data: any = {};
  if (defaults.displayName) {
    data.display_name = defaults.displayName;
  }
  if (defaults.color) {
    data.color = defaults.color;
  }

  // Uses DEFAULT_MESSAGE_TIMEOUT from WebSocketConnection
  // gameID will be set automatically via setupMessageHandlers when room_created arrives
  // Backend will generate random display_name and color if not provided
  return ws.connection.sendAndWait(
    {
      type: "create_room",
      data,
    },
    {
      type: "room_created",
      validator: (data) => !!data?.room_id, // Ensure room_id exists
    }
  );
}

