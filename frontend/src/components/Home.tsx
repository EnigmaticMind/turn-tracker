import { useState } from "react";

import { NavLink, useNavigate } from "react-router";

export default function Home() {
  const navigate = useNavigate();

  const [code, setCode] = useState("");

  const onJoin = function () {
    // TODO: Check if room code exist
    navigate(`game/123`);
  };

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
          <NavLink to="/game/abc123" end>
            <button className="relative bg-linear-to-r from-blue-600 to-cyan-600 text-white">
              <div className="relative flex items-center space-x-2">
                <span>Start New Game</span>
              </div>
            </button>
          </NavLink>
        </div>

        {/* Join Room - Interactive Form */}
        <div className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-xl p-6 space-y-4">
          <div className="flex space-x-3">
            <input
              type="text"
              placeholder="Enter code"
              onChange={(e) => setCode(e.target.value)}
              className="flex-1 bg-slate-700/50 border border-slate-600 rounded-lg px-4 py-3 text-white placeholder-slate-400 focus:border-cyan-500 focus:outline-none focus:ring-2 focus:ring-cyan-500/20 transition-all duration-300"
            />
            <button
              onClick={onJoin}
              className="bg-linear-to-r from-teal-500 to-cyan-600 text-white font-semibold py-3 px-6 rounded-lg"
            >
              Join
            </button>
          </div>
        </div>

        {/* Options - Subtle Interactive */}
        <div className="space-y-6">
          <NavLink to="/options" end>
            <button className="relative border border-slate-800/50 bg-linear-to-r from-slate-600/30 to-slate-700/30 text-white">
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
