import { useState } from "react";

interface OptionsProps {
  onClose: () => void;
}

export default function Options({ onClose }: OptionsProps) {
  const [playerName, setPlayerName] = useState<string>();
  const [playerColor, setPlayerColor] = useState<string>();

  const onSave = () => {
    // TODO: Save changes to local storage
    onClose();
  };

  const onCancel = () => {
    onClose();
  };

  return (
    <div className="flex flex-col items-center justify-center h-full space-y-4">
      <h2 className="text-2xl font-semibold mb-2">Settings</h2>

      <div className="flex flex-col items-center space-y-3">
        <input
          type="text"
          value={playerName}
          onChange={(e) => setPlayerName(e.target.value)}
          placeholder="Display name"
          className="flex-1 bg-slate-700/50 border border-slate-600 rounded-lg px-4 py-3 text-white placeholder-slate-400 focus:border-cyan-500 focus:outline-none focus:ring-2 focus:ring-cyan-500/20 transition-all duration-300"
        />

        <div className="flex items-center space-x-3">
          <label>Default Color:</label>
          <input
            type="color"
            value={playerColor}
            onChange={(e) => setPlayerColor(e.target.value)}
            className="w-10 h-10 border-none"
          />
        </div>
      </div>

      <div className="flex gap-3 mt-6">
        <button
          onClick={onCancel}
          className="px-6 py-2 bg-slate-700 text-white rounded-lg hover:bg-slate-600 transition-colors"
        >
          Cancel
        </button>
        <button
          onClick={onSave}
          className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
        >
          Save
        </button>
      </div>
    </div>
  );
}
