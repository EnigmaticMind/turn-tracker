import { useNavigate, useParams } from "react-router";
import { LiquidAvatar } from "../components/LiquidAvatar";
import { players } from "../lib/players";
import { PlayerGrid } from "../components/PlayerGrid";

export default function PlayerTurn() {
  const navigate = useNavigate();
  const { gameID, playerID } = useParams();

  const currentPlayer = players.find((p) => String(p.id) === playerID);

  const otherPlayers = currentPlayer
    ? players.filter((p) => p.id !== currentPlayer.id)
    : players;

  const handleEndCurrentTurn = () => {
    navigate(`/game/${gameID}`);
  };

  const handleSelect = (id: number) => {
    navigate(`/game/${gameID}/turn/${id}`);
  };

  if (!currentPlayer) {
    return (
      <div className="flex items-center justify-center h-full">
        <span className="text-slate-400">Player not found</span>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full items-center justify-center text-center space-y-6">
      <div
        className="w-full h-1/2 flex flex-col items-center justify-center cursor-pointer"
        onClick={handleEndCurrentTurn}
      >
        <h2 className="text-2xl font-semibold text-slate-200">Current Turn</h2>
        <span className="text-sm text-slate-400">Tap below to end turn</span>
        <LiquidAvatar
          color={currentPlayer.color}
          active
          name={currentPlayer.name}
        />
      </div>

      <div className="w-full px-4">
        <span className="text-sm text-slate-400">
          Tap below to start the next player's turn
        </span>
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
  );
}
