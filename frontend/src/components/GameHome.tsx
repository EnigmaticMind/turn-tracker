import { useOutletContext } from "react-router";
import { PlayerGrid, type Player } from "../components/PlayerGrid";
import { startTurn } from "../lib/websocket/handlers/startTurn";
import WebSocketManager from "../lib/websocket/WebSocketManager";
import type { PeerInfo } from "../lib/websocket/handlers/types";
import { toast } from "./ToastProvider";

interface GameOutletContext {
  ws: WebSocketManager | null;
  peers: PeerInfo[];
}

export default function GameHome() {
  const { ws, peers } = useOutletContext<GameOutletContext>();

  const handleTurnStart = async (clientID: string) => {
    if (!ws) return;
    // Send start_turn request to backend
    try {
      await startTurn(ws, clientID);
      // Navigation will happen automatically when turn_changed message arrives
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to start turn";
      console.error("GameHome: ", error);
      toast.error(errorMessage);
    }
  };

  if (!ws) {
    return (
      <div className="flex items-center justify-center h-full">
        <span className="text-slate-400">Connecting...</span>
      </div>
    );
  }

  // Convert PeerInfo[] to Player[] for PlayerGrid
  const players: Player[] = peers;

  return (
    <div className="flex flex-col items-center h-full w-full overflow-hidden">
      <span className="text-sm text-slate-400 shrink-0 py-6">
        Tap a player to start their turn
      </span>
      <div className="flex-1 w-full min-h-0">
        <PlayerGrid
          players={players}
          onSelect={handleTurnStart}
          rows={4}
          cols={2}
          shouldScroll={peers.length > 8}
        />
      </div>
    </div>
  );
}
