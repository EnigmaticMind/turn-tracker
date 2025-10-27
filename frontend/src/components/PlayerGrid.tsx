import { LiquidAvatar } from "./LiquidAvatar";
import "./player-turn.css";

interface Player {
  id: number;
  name: string;
  color: string;
}

interface PlayerGridProps {
  players: Player[];
  onSelect?: (playerId: number) => void;
  rows?: number;
  cols?: number;
  small?: boolean;
  shouldScroll?: boolean;
}

export function PlayerGrid({
  players,
  onSelect,
  rows = 2,
  cols = 2,
  small = false,
  shouldScroll,
}: PlayerGridProps) {
  const needsScrolling = shouldScroll ?? players.length > rows * cols;

  if (players.length === 0) {
    return (
      <div className="w-full h-full flex items-center justify-center">
        <span className="text-sm text-slate-400">Waiting for players...</span>
      </div>
    );
  }

  return (
    <div
      className={`w-full h-full ${needsScrolling ? "player-grid-wrapper" : ""}`}
    >
      <div
        className={`w-full h-full ${needsScrolling ? "overflow-y-auto scrollbar-hide" : ""}`}
      >
        <div
          className={`player-grid justify-items-center items-center `}
          style={{
            gridTemplateColumns: `repeat(${cols}, 1fr)`,
            gridAutoRows: "auto",
          }}
        >
          {players.map((p) => (
            <div
              key={p.id}
              className="cursor-pointer flex flex-col items-center shrink-0"
              onClick={() => onSelect?.(p.id)}
            >
              <LiquidAvatar
                color={p.color}
                active={false}
                small={small}
                name={p.name}
              />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
