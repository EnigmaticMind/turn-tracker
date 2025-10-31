export interface Message {
  type: string;
  data: any;
}

export interface PeerInfo {
  client_id: string;
  display_name: string;
  color: string;
}

export interface RoomCreatedData {
  room_id: string;
  peers: PeerInfo[];
}

export interface RoomJoinedData {
  room_id: string;
  peers: PeerInfo[];
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
}

