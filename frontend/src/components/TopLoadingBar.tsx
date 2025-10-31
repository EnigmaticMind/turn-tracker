import { useNavigation, useLocation } from "react-router";
import { motion } from "framer-motion";

export default function TopLoadingBar() {
  return (
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
  );
}
