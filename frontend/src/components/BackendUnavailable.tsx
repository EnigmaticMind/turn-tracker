import { useEffect } from "react";
import { useNavigate } from "react-router";

export default function BackendUnavailable() {
  const navigate = useNavigate();

  const checkHealth = async () => {
    // Derive HTTP URL from WebSocket URL
    const wsUrl =
      import.meta.env.VITE_WS_URL ||
      (import.meta.env.PROD
        ? "wss://turn-tracker-backend.fly.dev/ws"
        : "ws://localhost:8080/ws");

    // Convert to HTTP health endpoint
    const healthUrl = wsUrl
      .replace(/^wss?:\/\//, import.meta.env.PROD ? "https://" : "http://")
      .replace(/\/ws$/, "/health");

    try {
      const response = await fetch(healthUrl, {
        method: "GET",
        cache: "no-cache",
        signal: AbortSignal.timeout(5000), // 5 second timeout
      });

      if (response.ok) {
        // Backend is available, redirect to home
        navigate("/");
      }
    } catch (error) {
      // Backend is still unavailable, stay on this page
      console.error("Health check failed:", error);
    }
  };

  useEffect(() => {
    // Auto-retry every 5 seconds
    const interval = setInterval(checkHealth, 5000);

    // Also check immediately
    checkHealth();

    return () => clearInterval(interval);
  }, [navigate]);

  return (
    <div className="flex flex-col items-center justify-center h-screen text-center space-y-8 p-6">
      <div className="space-y-4">
        <h1 className="text-6xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-red-400 via-red-300 to-red-400">
          Backend Unavailable
        </h1>
        <p className="text-slate-400 text-xl">
          Give the server a few seconds to start up...
        </p>
      </div>

      <div className="bg-slate-800/30 backdrop-blur-sm border border-slate-700/50 rounded-xl p-6 space-y-4 max-w-md">
        <p className="text-slate-300">
          The game server appears to be offline or unreachable. This page will
          automatically check for the server every few seconds and will redirect
          you to the home page when the server is available.
        </p>
        <button
          onClick={checkHealth}
          className="w-full bg-linear-to-r from-blue-600 to-cyan-600 text-white px-6 py-3 rounded-lg font-semibold hover:from-blue-700/50 hover:to-cyan-700/50 transition-all"
        >
          Check Again
        </button>
      </div>
    </div>
  );
}
