import React, { createContext, useContext, useState, useCallback } from "react";
import { AnimatePresence, motion } from "framer-motion";

const MAX_TOASTS = 3;
const TOAST_DURATION = 3500;

type ToastType = "success" | "error" | "info";
interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

interface ToastContextValue {
  showToast: (message: string, type?: ToastType) => void;
}

const ToastContext = createContext<ToastContextValue>({ showToast: () => {} });

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const showToast = useCallback((message: string, type: ToastType = "info") => {
    const id = Date.now();
    setToasts((prev) => {
      const next = [...prev, { id, message, type }];
      while (next.length > MAX_TOASTS) next.shift();
      return next;
    });
    setTimeout(
      () => setToasts((prev) => prev.filter((t) => t.id !== id)),
      TOAST_DURATION
    );
  }, []);
  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}

      {/* Toast Container (top of screen) */}
      <div className="fixed top-4 left-1/2 -translate-x-1/2 flex flex-col items-center gap-3 z-50 w-full max-w-sm px-4">
        <AnimatePresence>
          {toasts.map((toast) => (
            <motion.div
              key={toast.id}
              onClick={() =>
                setToasts((prev) => prev.filter((t) => t.id !== toast.id))
              }
              initial={{ opacity: 0, y: -24, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -12, scale: 0.9 }}
              transition={{ duration: 0.25 }}
              className={`relative overflow-hidden rounded-xl shadow-lg backdrop-blur-md border 
                text-sm text-white px-4 py-3 w-full text-center
                ${
                  toast.type === "success"
                    ? "border-emerald-500/40 bg-emerald-600/20"
                    : toast.type === "error"
                      ? "border-red-500/40 bg-red-600/20"
                      : "border-sky-500/40 bg-sky-600/20"
                }`}
            >
              {/* === 🔮 Optional Styling Enhancements === */}
              {/* Glowing background animation */}
              <div
                className={`absolute inset-0 blur-lg opacity-40 animate-pulse
                ${
                  toast.type === "success"
                    ? "bg-gradient-to-r from-emerald-400 to-emerald-700"
                    : toast.type === "error"
                      ? "bg-gradient-to-r from-red-400 to-red-700"
                      : "bg-gradient-to-r from-sky-400 to-indigo-700"
                }`}
              ></div>

              {/* Animated liquid shimmer */}
              <div className="absolute inset-0 opacity-25 bg-[radial-gradient(circle_at_30%_50%,rgba(255,255,255,0.2),transparent_70%)] animate-[pulse_3s_ease-in-out_infinite]"></div>

              {/* Toast text (click toast to dismiss) */}
              <span className="relative z-10 font-medium tracking-wide drop-shadow-md">
                {toast.message}
              </span>

              {/* Progress bar (top, solid) */}
              <div className="pointer-events-none absolute inset-x-0 top-0 h-0.5 bg-white/10 overflow-hidden">
                <motion.div
                  initial={{ width: "100%" }}
                  animate={{ width: 0 }}
                  transition={{
                    duration: TOAST_DURATION / 1000,
                    ease: "linear",
                  }}
                  className={`h-full ${
                    toast.type === "success"
                      ? "bg-emerald-400/80"
                      : toast.type === "error"
                        ? "bg-red-400/80"
                        : "bg-sky-400/80"
                  }`}
                />
              </div>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  );
};

export const useToast = () => useContext(ToastContext);
