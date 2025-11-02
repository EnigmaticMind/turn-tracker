import { useEffect, useState } from "react";
import { LiquidAvatar } from "./LiquidAvatar";
import "./player-turn.css";
import type { PeerInfo } from "../lib/websocket/handlers/types";

export interface Player {
  client_id: string;
  display_name: string;
  color: string;
}

interface PlayerGridProps {
  players: Player[];
  onSelect?: (clientId: string) => void;
  rows?: number;
  cols?: number;
  small?: boolean;
  shouldScroll?: boolean;
  currentTurn?: PeerInfo | null;
  turnStartTime?: number | null;
}

export function PlayerGrid({
  players,
  onSelect,
  rows = 2,
  cols = 2,
  small = false,
  shouldScroll,
  currentTurn = null,
  turnStartTime = null,
}: PlayerGridProps) {
  const needsScrolling = shouldScroll ?? players.length > rows * cols;
  const [elapsedSeconds, setElapsedSeconds] = useState<number>(0);

  // Update elapsed time every second for the current turn
  // TODO: Do I want to keep this timer?
  useEffect(() => {
    if (!currentTurn || !turnStartTime) {
      setElapsedSeconds(0);
      return;
    }

    // Only update time if there's a current turn
    if (currentTurn) {
      // Update immediately
      const updateElapsed = () => {
        if (!turnStartTime) {
          setElapsedSeconds(0);
          return;
        }
        const now = Date.now();
        const elapsed = Math.floor((now - turnStartTime) / 1000);
        setElapsedSeconds(elapsed);
      };

      updateElapsed();
      const interval = setInterval(updateElapsed, 1000);

      return () => clearInterval(interval);
    }
  }, [currentTurn, turnStartTime]);

  // Format seconds into MM:SS or HH:MM:SS
  const formatTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
    }
    return `${minutes}:${secs.toString().padStart(2, "0")}`;
  };

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
          {players.map((p) => {
            const isCurrentTurn = currentTurn?.client_id === p.client_id;
            return (
              <div
                key={p.client_id}
                className="cursor-pointer flex flex-col items-center shrink-0"
                onClick={() => onSelect?.(p.client_id)}
              >
                <LiquidAvatar
                  color={p.color}
                  active={false}
                  small={small}
                  name={p.display_name}
                />
                {isCurrentTurn && turnStartTime !== null && (
                  <div className="mt-2 text-sm text-slate-300 font-mono">
                    {formatTime(elapsedSeconds)}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
