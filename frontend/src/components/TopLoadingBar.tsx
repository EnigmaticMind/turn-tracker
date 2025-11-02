import { useNavigation, useLocation } from "react-router";
import { motion } from "framer-motion";
import { useState, useEffect } from "react";

type LoadingState = {
  isActive: boolean;
  message?: string;
};

export function startLoading(message?: string): void {
  if (typeof window === "undefined") return;
  window.dispatchEvent(
    new CustomEvent("loading:start", {
      detail: { message },
    })
  );
}

export function stopLoading(): void {
  if (typeof window === "undefined") return;
  window.dispatchEvent(new CustomEvent("loading:stop"));
}

export default function TopLoadingBar() {
  const navigation = useNavigation();
  const location = useLocation();
  const [customLoading, setCustomLoading] = useState<LoadingState>({
    isActive: false,
  });

  // Listen for custom loading events (from async handlers)
  useEffect(() => {
    const handleLoadingStart = (event: Event) => {
      const customEvent = event as CustomEvent<{ message?: string }>;
      setCustomLoading({
        isActive: true,
        message: customEvent.detail?.message,
      });
    };

    const handleLoadingStop = () => {
      setCustomLoading({ isActive: false });
    };

    window.addEventListener("loading:start", handleLoadingStart);
    window.addEventListener("loading:stop", handleLoadingStop);

    return () => {
      window.removeEventListener("loading:start", handleLoadingStart);
      window.removeEventListener("loading:stop", handleLoadingStop);
    };
  }, []);

  // Check if navigation is in loading state
  const isNavigationLoading = navigation.state === "loading";

  // Also check for pending navigation (location mismatch indicates transition)
  const hasPendingNavigation =
    navigation.location && navigation.location.pathname !== location.pathname;

  // Show if navigation loading OR custom loading is active
  const shouldShow =
    isNavigationLoading || hasPendingNavigation || customLoading.isActive;

  // Determine if we're creating or joining a game
  const gameID =
    location.pathname.match(/\/game\/([^\/]+)/)?.[1] ||
    navigation.location?.pathname.match(/\/game\/([^\/]+)/)?.[1];
  const isCreating =
    location.pathname === "/game" ||
    location.pathname === "/game/" ||
    navigation.location?.pathname === "/game" ||
    navigation.location?.pathname === "/game/";

  if (!shouldShow) {
    return null;
  }

  // Determine message - prioritize custom message, then navigation-based
  const message =
    customLoading.message ||
    (isCreating
      ? "Creating game..."
      : gameID
        ? `Joining game ${gameID}...`
        : "Loading...");

  return (
    <>
      {/* Loading bar at top */}
      <div className="fixed top-0 left-0 right-0 h-1 z-50 bg-white/10 overflow-hidden">
        <motion.div
          className="h-full bg-sky-400/80 shadow-[0_0_8px_rgba(56,189,248,0.8)]"
          animate={{
            x: ["-100%", "100%"],
          }}
          transition={{
            duration: 1.5,
            repeat: Infinity,
            ease: "linear",
          }}
          style={{
            width: "40%",
          }}
        />
      </div>

      {/* Full-screen loading overlay with message */}
      <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-40">
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.9 }}
          transition={{ duration: 0.2 }}
          className="text-center space-y-6"
        >
          {/* Spinner */}
          <div className="w-16 h-16 border-4 border-sky-400 border-t-transparent rounded-full animate-spin mx-auto shadow-[0_0_20px_rgba(56,189,248,0.5)]"></div>

          {/* Message */}
          <p className="text-white text-xl font-semibold tracking-wide">
            {message}
          </p>
        </motion.div>
      </div>
    </>
  );
}
