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

  return (
    <div className="w-full px-4">
      <div
        className={`relative ${needsScrolling ? "player-grid-wrapper" : ""}`}
      >
        <div
          className={`player-grid gap-6 ${
            needsScrolling
              ? "justify-start overflow-x-auto scrollbar-hide"
              : "justify-items-center"
          }`}
          style={{
            ...(needsScrolling
              ? { gridTemplateRows: `repeat(${rows}, auto)` }
              : {
                  gridAutoFlow: "row",
                  gridTemplateColumns: `repeat(${cols}, 1fr)`,
                }),
            minHeight: `${rows * (small ? 120 : 120)}px`,
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
