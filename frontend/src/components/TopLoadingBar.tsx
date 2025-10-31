import { useNavigation, useLocation } from "react-router";
import { motion } from "framer-motion";

export default function TopLoadingBar() {
  const navigation = useNavigation();
  const location = useLocation();

  // Check if navigation is in loading state
  const isLoading = navigation.state === "loading";

  // Also check for pending navigation (location mismatch indicates transition)
  const hasPendingNavigation =
    navigation.location && navigation.location.pathname !== location.pathname;

  const shouldShow = isLoading || hasPendingNavigation;

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

  const message = isCreating
    ? "Creating game..."
    : gameID
      ? `Joining game ${gameID}...`
      : "Loading...";

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
