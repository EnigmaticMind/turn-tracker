import { motion } from "framer-motion";
import { useState } from "react";

interface Props {
  color: string;
  active?: boolean;
  small?: boolean;
  name?: string;
}

export function LiquidAvatar({
  color,
  active = false,
  small = false,
  name,
}: Props) {
  const [id] = useState(() => Math.random().toString(36).substring(2, 9));

  const size = small ? 100 : active ? 330 : 120;

  // Calculate text size based on avatar size
  const textSize = active ? "text-4xl" : small ? "text-md" : "text-md";

  // Calculate overflow size
  const overflowSize = size - (small ? 60 : active ? 80 : 80);

  // Calculate max width for text (80% of avatar size)
  const maxWidth = size * 0.8;

  return (
    <motion.div
      className="relative"
      animate={{
        scale: active ? 1.1 : 1,
        filter: active ? "drop-shadow(0 0 20px rgba(255,255,255,0.3))" : "none",
      }}
      transition={{
        duration: 0.6,
        ease: "easeInOut",
      }}
    >
      <svg
        width={size}
        height={size}
        viewBox="0 0 200 200"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <radialGradient id={`grad-${id}`} cx="50%" cy="50%" r="50%">
            <stop offset="0%" stopColor={color} stopOpacity="1" />
          </radialGradient>
        </defs>
        <motion.circle
          cx="100"
          cy="100"
          r="80"
          fill={`url(#grad-${id})`}
          filter={`url(#liquid-${id})`}
          animate={{
            r: [78, 82, 80],
          }}
          transition={{
            duration: 3,
            repeat: Infinity,
            repeatType: "mirror",
          }}
        />
      </svg>

      {/* Add name label */}
      {name && (
        <div
          className={`absolute inset-0 flex items-center justify-center pointer-events-none`}
        >
          <span
            className={`${textSize} font-bold text-white max-w-[${overflowSize}px] truncate drop-shadow-lg text-center`}
            style={{ maxWidth: `${maxWidth}px` }}
          >
            {name}
          </span>
        </div>
      )}
    </motion.div>
  );
}
