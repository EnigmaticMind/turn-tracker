import { useOutletContext } from "react-router";
import { LiquidAvatar } from "../components/LiquidAvatar";
import { PlayerGrid, type Player } from "../components/PlayerGrid";
import { startTurn } from "../lib/websocket/handlers/startTurn";
import { endTurn } from "../lib/websocket/handlers/endTurn";
import WebSocketManager from "../lib/websocket/WebSocketManager";
import type { PeerInfo } from "../lib/websocket/handlers/types";
import { toast } from "./ToastProvider";

interface GameOutletContext {
  ws: WebSocketManager | null;
  peers: PeerInfo[];
  currentTurn: PeerInfo | null;
}

export default function PlayerTurn() {
  const { ws, peers, currentTurn } = useOutletContext<GameOutletContext>();

  // Convert PeerInfo[] to Player[] for type compatibility
  const players: Player[] = peers;
  const currentPlayer = currentTurn;

  const otherPlayers = currentPlayer
    ? players.filter((p) => p.client_id !== currentPlayer.client_id)
    : players;

  const handleEndCurrentTurn = async () => {
    if (!ws) return;
    // Send end_turn message to backend to clear current turn
    try {
      await endTurn(ws);
      // Navigation will happen automatically when turn_changed message arrives (turn becomes null)
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to end turn";
      console.error("PlayerTurn: ", error);
      toast.error(errorMessage);
    }
  };

  const handleSelect = async (targetClientID: string) => {
    if (!ws) return;
    // Send start_turn request to backend
    try {
      await startTurn(ws, targetClientID);
      // Navigation will happen automatically when turn_changed message arrives
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : "Failed to start turn";
      console.error("PlayerTurn: ", error);
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

  if (!currentPlayer) {
    return (
      <div className="flex items-center justify-center h-full">
        <span className="text-slate-400">Player not found</span>
      </div>
    );
  }

  return (
    <div className="max-h-full h-full flex flex-col overflow-hidden">
      <div
        className="w-full flex flex-col items-center justify-center cursor-pointer shrink-0"
        onClick={handleEndCurrentTurn}
      >
        <h2 className="text-2xl font-semibold text-slate-200">Current Turn</h2>
        <span className="text-sm text-slate-400">Tap below to end turn</span>
        <LiquidAvatar
          color={currentPlayer.color}
          active
          name={currentPlayer.display_name}
        />
      </div>

      <div className="w-full flex-1 px-4 justify-center text-center overflow-hidden flex flex-col min-h-0">
        <span
          className="text-sm text-slate-400 shrink-0 pb-2"
          style={{ visibility: otherPlayers.length > 0 ? "visible" : "hidden" }}
        >
          Tap below to start the next player's turn
        </span>
        <div className="flex-1 min-h-0">
          <PlayerGrid
            players={otherPlayers}
            onSelect={handleSelect}
            rows={2}
            cols={3}
            small={true}
            shouldScroll={otherPlayers.length > 6}
          />
        </div>
      </div>
    </div>
  );
}
