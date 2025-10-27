import { useNavigate } from "react-router";

export default function Home() {
  const navigate = useNavigate();

  const handleStartGame = () => {
    // Generate a random game ID
    const gameID = Math.random().toString(36).substring(2, 9);
    navigate(`/game/${gameID}`);
  };

  const handleJoinGame = () => {
    const gameID = prompt("Enter game code:");
    if (gameID) {
      navigate(`/game/${gameID}`);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center h-screen space-y-6 px-4">
      <h1 className="text-5xl font-bold mb-8">Turn Tracker</h1>

      <div className="flex flex-col gap-4 w-full max-w-md">
        <button
          onClick={handleStartGame}
          className="w-full bg-blue-500 hover:bg-blue-600 text-white font-semibold py-4 px-6 rounded-lg transition-colors shadow-lg"
        >
          Start New Game
        </button>

        <button
          onClick={handleJoinGame}
          className="w-full bg-slate-700 hover:bg-slate-600 text-white font-semibold py-4 px-6 rounded-lg transition-colors"
        >
          Join Game
        </button>
      </div>
    </div>
  );
}
