export interface Message {
  type: string;
  data: any;
}

export interface PeerInfo {
  client_id: string;
  display_name: string;
  color: string;
  total_turn_time: number; // Total time spent in turns (in milliseconds)
}

export interface RoomCreatedData {
  room_id: string;
  your_client_id: string; // Client ID of the message recipient
  peers: PeerInfo[];
  current_turn?: PeerInfo | null;
}

export interface RoomJoinedData {
  room_id: string;
  your_client_id: string; // Client ID of the message recipient
  peers: PeerInfo[];
  current_turn?: PeerInfo | null;
}

export interface PlayerJoinedData {
  room_id: string;
  peer_id: string;
  display_name: string;
  color: string;
}

export interface PlayerLeftData {
  room_id: string;
  peer_id: string;
}

export interface ProfileUpdatedData {
  room_id: string;
  peer_id: string;
  display_name: string;
  color: string;
  total_turn_time?: number; // Total time spent in turns (optional for backward compatibility)
}

export interface TurnChangedData {
  room_id: string;
  current_turn: PeerInfo | null; // null if no turn active
  turn_start_time: number | null; // Unix timestamp in milliseconds when turn started (null if no turn active)
  sequence: number; // Sequence number to identify stale messages (higher = newer)
}

