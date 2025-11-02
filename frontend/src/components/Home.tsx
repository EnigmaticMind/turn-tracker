import { useEffect, useState } from "react";
import { useLocation } from "react-router";
import { NavLink, useNavigate } from "react-router";
import { isValidGameID } from "../lib/websocket/utils/gameID";

export async function clientLoader() {
  // Check backend health
  try {
    const wsUrl =
      import.meta.env.VITE_WS_URL ||
      (import.meta.env.PROD
        ? "wss://turn-tracker-backend.fly.dev/ws"
        : "ws://localhost:8080/ws");

    const healthUrl = wsUrl
      .replace(/^wss?:\/\//, import.meta.env.PROD ? "https://" : "http://")
      .replace(/\/ws$/, "/health");

    const response = await fetch(healthUrl, {
      method: "GET",
      cache: "no-cache",
      signal: AbortSignal.timeout(5000),
    });

    if (!response.ok) {
      throw new Response("Backend unavailable", {
        status: 503,
        statusText: "Service Unavailable",
      });
    }
  } catch (error) {
    console.error("Home: Backend unavailable:", error);
    // Redirect to offline page
    throw new Response(null, {
      status: 302,
      headers: {
        Location: "/offline",
      },
    });
  }

  return {};
}

export default function Home() {
  const [gameID, setGameID] = useState("");
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const gameIDFromParams = searchParams.get("code");

  useEffect(() => {
    if (gameIDFromParams) {
      // Normalize the game ID from URL params
      const normalized = gameIDFromParams
        .toUpperCase()
        .replace(/[^A-Z0-9]/g, "")
        .slice(0, 4);
      setGameID(normalized);
    }
  }, [gameIDFromParams]);

  const handleGameIDChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    // Auto-uppercase and filter to alphanumeric only
    const filtered = value.toUpperCase().replace(/[^A-Z0-9]/g, "");
    // Limit to 4 characters
    const limited = filtered.slice(0, 4);
    setGameID(limited);
  };

  const handleJoinClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    if (!isValidGameID(gameID)) {
      e.preventDefault();
      return false;
    }
  };

  const isGameIDValid = isValidGameID(gameID);

  return (
    <div>
      {/* Main Content */}
      <div className="relative z-10 flex flex-col items-center justify-center h-screen text-center space-y-8 p-6">
        {/* Title with Glow Effect */}
        <div className="space-y-4">
          <h1 className="text-6xl font-bold text-transparent bg-clip-text bg-linear-to-r from-blue-400 via-cyan-300 to-purple-400">
            Turn Tracker
          </h1>
        </div>

        {/* Interactive Buttons */}
        <div className="space-y-6">
          {/* Start New Game - Interactive Button */}
          <NavLink to="/game/" end>
            <button className="relative bg-linear-to-r from-blue-600 to-cyan-600 text-white px-6 py-3 rounded-lg font-semibold hover:from-blue-700/50 hover:to-cyan-700/50 transition-all">
              <div className="relative flex items-center space-x-2">
                <span>Start New Game</span>
              </div>
            </button>
          </NavLink>
        </div>

        {/* Join Room - Interactive Form */}
        <div className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-xl p-6 space-y-4">
          <div className="flex space-y-2">
            <div className="flex space-x-3">
              <input
                type="text"
                value={gameID}
                onChange={handleGameIDChange}
                placeholder="Enter code"
                maxLength={4}
                aria-label="Game code input"
                className={`flex-1 bg-slate-700/50 border border-slate-600 rounded-lg px-4 py-3 text-white placeholder-slate-400 focus:border-cyan-500 focus:outline-none focus:ring-2 focus:ring-cyan-500/20 transition-all duration-300 ${
                  gameID.length > 0 && !isGameIDValid
                    ? "border-red-500 focus:border-red-500 focus:ring-red-500/20"
                    : isGameIDValid
                      ? "border-green-500 focus:border-cyan-500 focus:ring-cyan-500/20"
                      : "border-slate-600 focus:border-cyan-500 focus:ring-cyan-500/20"
                }`}
              />
              <NavLink
                to={isGameIDValid ? `/game/${gameID}` : "#"}
                end
                onClick={handleJoinClick}
              >
                <button
                  disabled={!isGameIDValid}
                  className="bg-linear-to-r from-teal-500 to-cyan-600 text-white font-semibold py-3 px-6 rounded-lg hover:from-teal-600 hover:to-cyan-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:from-teal-500 disabled:hover:to-cyan-600"
                >
                  Join
                </button>
              </NavLink>
            </div>
          </div>
        </div>

        {/* Options - Subtle Interactive */}
        <div className="space-y-6">
          <NavLink to="/options" end>
            <button className="relative border border-slate-800/50 bg-linear-to-r from-slate-600/30 to-slate-700/30 text-white px-6 py-3 rounded-lg font-semibold hover:from-slate-600/50 hover:to-slate-700/50 transition-all">
              <div className="relative flex items-center space-x-2">
                <span>Options</span>
              </div>
            </button>
          </NavLink>
        </div>
      </div>
    </div>
  );
}
