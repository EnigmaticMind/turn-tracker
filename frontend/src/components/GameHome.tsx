import { useNavigate, useParams } from "react-router";
import { players } from "../lib/players";
import { PlayerGrid } from "../components/PlayerGrid";

export default function GameHome() {
  const navigate = useNavigate();
  const { gameID } = useParams();

  const handleSelect = (id: number) => {
    navigate(`/game/${gameID}/turn/${id}`);
  };

  return (
    <div className="flex flex-col items-center h-full w-full overflow-hidden">
      <span className="text-sm text-slate-400 shrink-0 py-6">
        Tap a player to start their turn
      </span>
      <div className="flex-1 w-full min-h-0">
        <PlayerGrid
          players={players}
          onSelect={handleSelect}
          rows={4}
          cols={2}
          shouldScroll={players.length > 8}
        />
      </div>
    </div>
  );
}
