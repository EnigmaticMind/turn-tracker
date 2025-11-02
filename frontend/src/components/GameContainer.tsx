import { useState, useEffect, useRef, useCallback, useMemo } from "react";
import {
  Outlet,
  redirect,
  useLoaderData,
  useBeforeUnload,
  useLocation,
} from "react-router";
import { useNavigate } from "react-router";

import { Home, Settings } from "lucide-react";
import Options from "./Options";
import WebSocketManager from "../lib/websocket/WebSocketManager";
import {
  getPersistentConnection,
  destroyPersistentConnection,
} from "../lib/websocket/persistentConnection";
import { createGame } from "../lib/websocket/handlers/createGame";
import { joinGame } from "../lib/websocket/handlers/joinGame";
import type { PeerInfo } from "../lib/websocket/handlers/types";
import { toast } from "./ToastProvider";

// React Router Loader
export async function clientLoader({
  params,
}: {
  params: { gameID?: string };
}) {
  console.log("GameContainer: Starting ClientLoader");
  // Init WebSocket server
  const ws = getPersistentConnection();

  const gameIDExists = !!(params.gameID && params.gameID.trim().length > 0);

  // Game ID does not exist, create a new game
  if (!gameIDExists) {
    console.log("GameContainer: Creating new game");
    try {
      await createGame(ws);
    } catch (error: any) {
      console.error("GameContainer: Error creating game:", error);
      handleLoaderError(error, "create");
    }

    // After creating, redirect to show the new gameID in URL
    // Since params.gameID doesn't exist, we always need to redirect
    if (ws.gameID) {
      console.log("GameContainer: Redirecting to game", ws.gameID);
      throw redirect(`/game/${ws.gameID}`);
    }
  }

  // Game ID exists, join the game
  if (gameIDExists && ws.gameID !== params.gameID?.toUpperCase()) {
    console.log("GameContainer: Joining game", params.gameID);

    try {
      await joinGame(ws, params.gameID!);
    } catch (error: any) {
      console.error("GameContainer: Error joining game:", error);
      handleLoaderError(error, "join", params.gameID);
    }
    console.log("GameContainer: Game joined successfully, gameID:", ws.gameID);
  }

  return ws;
}

// Helper function to handle errors and redirects
function handleLoaderError(
  error: unknown,
  operation: "create" | "join",
  gameID?: string
): never {
  console.error(
    `GameContainer: Error ${operation === "create" ? "creating" : "joining"} game:`,
    error
  );

  const errorMessage =
    error instanceof Error
      ? error.message
      : operation === "create"
        ? "Failed to create game"
        : "Failed to join game";

  // TODO: Expand back end to return more specific error messages
  toast.error(errorMessage);

  if (operation === "join" && gameID) {
    throw redirect(`/?code=${gameID}`);
  }
  throw redirect("/");
}

export default function GameContainer() {
  const location = useLocation();
  const ws = useLoaderData() as WebSocketManager | null;
  const navigate = useNavigate();

  const gameID = ws?.gameID;
  const [peers, setPeers] = useState<PeerInfo[]>(ws?.peers || []);
  const [currentTurn, setCurrentTurn] = useState<PeerInfo | null>(
    ws?.currentTurn || null
  );
  const [turnStartTime, setTurnStartTime] = useState<number | null>(
    ws?.turnStartTime || null
  );
  const [isOptionsOpen, setIsOptionsOpen] = useState(false);

  const wakeLockRef = useRef<WakeLockSentinel | null>(null);
  const wsRef = useRef<WebSocketManager | null>(ws);
  const locationRef = useRef(location.pathname);
  const gameIDRef = useRef(gameID);

  // Update refs when values change
  useEffect(() => {
    locationRef.current = location.pathname;
    gameIDRef.current = gameID;
  }, [location.pathname, gameID]);

  // Handle home button
  const onHome = () => {
    // Intentionally leaving - send leave_room message and destroy connection
    if (ws && ws.gameID) {
      // Send leave_room message (server will handle removing from room)
      ws.connection
        .send("leave_room", { room_id: ws.gameID })
        .catch((err) =>
          console.error("GameContainer: Error sending leave_room:", err)
        );
    }
    destroyPersistentConnection();
    navigate("/");
  };

  // // Handle game code button
  const onGameCode = async () => {
    if (!gameID) return;

    try {
      await navigator.clipboard.writeText(gameID);
      toast.success("Game code copied!");
    } catch (err) {
      console.error("GameContainer: Failed to copy game code:", err);
      toast.error("Failed to copy game code");
    }
  };

  // Handle settings button
  const onSettings = () => {
    setIsOptionsOpen(true);
  };

  // Handle closing options modal
  const closeOptions = () => {
    setIsOptionsOpen(false);
  };

  // Handle peers change - stable callback
  const handlePeersChange = useCallback((updatedPeers: PeerInfo[]) => {
    setPeers(updatedPeers);
  }, []);

  // Handle turn change - stable callback (no navigation dependencies)
  const handleTurnChange = useCallback(
    (currentTurnPeer: PeerInfo | null, serverTurnStartTime: number | null) => {
      setCurrentTurn(currentTurnPeer);
      setTurnStartTime(serverTurnStartTime);
    },
    [] // No dependencies - use refs instead
  );

  // Separate effect for navigation based on turn changes
  useEffect(() => {
    if (!gameID) return;

    if (!currentTurn) {
      // Turn ended - navigate back to home if we're on a turn page
      if (location.pathname.includes("/turn/")) {
        const homePath = `/game/${gameID}`;
        if (location.pathname !== homePath) {
          navigate(homePath);
        }
      }
      return;
    }

    // New turn started - navigate to that player's turn
    const targetPath = `/game/${gameID}/turn/${currentTurn.client_id}/`;
    if (location.pathname !== targetPath) {
      navigate(targetPath);
    }
  }, [currentTurn, gameID, navigate, location.pathname]);

  // Subscribe to WebSocket events - only runs when ws changes
  useEffect(() => {
    console.log("GameContainer: Subscribing to WebSocket events");
    if (!ws) {
      return;
    }

    const unsubscribePeers = ws.onPeersChange(handlePeersChange);
    const unsubscribeTurn = ws.onTurnChange(handleTurnChange);

    return () => {
      unsubscribePeers();
      unsubscribeTurn();
    };
  }, [ws, handlePeersChange, handleTurnChange]); // These callbacks are now stable

  // Update wsRef whenever ws changes
  useEffect(() => {
    wsRef.current = ws;
  }, [ws]);

  // Also cleanup on page unload (browser navigation, refresh, etc.)
  useBeforeUnload(() => {
    if (wsRef.current) {
      console.log(
        "GameContainer: Cleaning up WebSocket connection on page unload"
      );
      wsRef.current.destroy();
    }
  });

  // Request wake lock to prevent screen sleep
  useEffect(() => {
    console.log("GameContainer: Requesting wake lock");
    const requestWakeLock = async () => {
      if ("wakeLock" in navigator) {
        try {
          const wakeLock = await navigator.wakeLock.request("screen");
          wakeLockRef.current = wakeLock;
        } catch (err) {
          console.error("GameContainer: Failed to acquire wake lock:", err);
        }
      }
    };

    requestWakeLock();

    const handleVisibilityChange = () => {
      if (
        document.visibilityState === "visible" &&
        wakeLockRef.current === null
      ) {
        requestWakeLock();
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    // Release wake lock when component unmounts
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);

      if (wakeLockRef.current) {
        wakeLockRef.current.release().catch((err) => {
          console.error("GameContainer: Failed to release wake lock:", err);
        });
      }
    };
  }, []);

  // Memoize context to prevent unnecessary child re-renders
  const contextValue = useMemo(
    () => ({ ws, peers, currentTurn, turnStartTime }),
    [ws, peers, currentTurn, turnStartTime]
  );

  return (
    <div className="h-screen overflow-hidden">
      <div className="h-[calc(100vh-52px)] overflow-hidden pt-6">
        <Outlet context={contextValue} />
      </div>

      {/* Bottom Bar */}
      <div className="fixed bottom-0 left-0 w-full bg-[#0D111A]/90 backdrop-blur-md border-t border-gray-700 flex items-center justify-between px-3 py-3 z-50">
        <div className="flex items-center gap-5">
          {/* Home Button */}
          <button
            onClick={onHome}
            className="text-gray-300 hover:text-white transition-transform hover:scale-110 active:scale-95"
            aria-label="Home"
          >
            <Home size={26} />
          </button>
        </div>
        {/* Game Code */}
        <div
          className="text-white font-semibold tracking-widest text-lg select-none cursor-pointer"
          onClick={onGameCode}
          title="Click to copy"
        >
          {gameID}
        </div>

        {/* Right side home button */}
        <button
          onClick={onSettings}
          className="text-gray-300 hover:text-white transition-transform hover:scale-110 active:scale-95"
          aria-label="Settings"
        >
          <Settings size={26} />
        </button>
      </div>

      {/* Options Modal */}
      {isOptionsOpen && (
        <div
          data-testid="options-modal-backdrop"
          className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50"
          onClick={closeOptions}
        >
          <div
            className="bg-[#0D111A] border border-gray-700 rounded-lg p-8 max-w-md w-full mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <Options onClose={closeOptions} ws={ws || undefined} />
          </div>
        </div>
      )}
    </div>
  );
}
