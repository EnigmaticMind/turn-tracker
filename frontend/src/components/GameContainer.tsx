import { useState } from "react";
import { Outlet, useParams } from "react-router";
import { useNavigate } from "react-router";

import { Home, Settings } from "lucide-react";
import Options from "./Options";

export default function GameContainer() {
  const { gameID } = useParams();

  const navigate = useNavigate();

  const [isOptionsOpen, setIsOptionsOpen] = useState(false);

  // Handle home button
  const onHome = () => {
    // TODO: Add confirmation dialog
    navigate("/");
  };

  // Handle game code button
  const onGameCode = () => {
    // TODO: Copy to clipboard
    navigate(`/game/${gameID}`);
  };

  // Handle settings button
  const onSettings = () => {
    setIsOptionsOpen(true);
  };

  const closeOptions = () => {
    setIsOptionsOpen(false);
  };

  return (
    <div className="h-screen flex flex-col">
      <div className="flex-1 pt-4 pb-[60px]">
        <Outlet />
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
          className="text-white font-semibold tracking-widest text-lg select-nonee"
          onClick={onGameCode}
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
          className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50"
          onClick={closeOptions}
        >
          <div
            className="bg-[#0D111A] border border-gray-700 rounded-lg p-8 max-w-md w-full mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            <Options onClose={closeOptions} />
          </div>
        </div>
      )}
    </div>
  );
}
