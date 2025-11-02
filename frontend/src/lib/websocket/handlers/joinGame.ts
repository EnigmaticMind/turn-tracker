import WebSocketManager from "../WebSocketManager";
import { isValidGameID } from "../utils/gameID";
import { getDefaultProfile } from "../../utils/userProfile";
import type { Message } from "./types";

export async function joinGame(ws: WebSocketManager, gameID: string): Promise<Message> {
  console.log("Joining game:", gameID);

  if (!isValidGameID(gameID)) {
    throw new Error("Invalid game ID");
  }

  const roomID = gameID.toUpperCase();

  // Get defaults from localStorage (only if they exist)
  const defaults = getDefaultProfile();

  // Build data object - only include fields that exist
  const data: any = { room_id: roomID };
  if (defaults.displayName) {
    data.display_name = defaults.displayName;
  }
  if (defaults.color) {
    data.color = defaults.color;
  }

  // Uses DEFAULT_MESSAGE_TIMEOUT from WebSocketConnection
  // gameID will be set automatically via setupMessageHandlers when room_joined arrives
  // Backend will generate random display_name and color if not provided
  return ws.connection.sendAndWait(
    {
      type: "join_room",
      data,
    },
    {
      type: "room_joined",
      validator: (data) => data?.room_id === roomID, // Ensure room_id matches
    }
  );
}

