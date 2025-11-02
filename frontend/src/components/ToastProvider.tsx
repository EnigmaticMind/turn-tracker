import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from "react";
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

export function dispatchToast(message: string, type: ToastType = "info"): void {
  if (typeof window === "undefined") return;

  window.dispatchEvent(
    new CustomEvent("toast", {
      detail: { type, message },
    })
  );
}

export const toast = {
  success: (message: string) => dispatchToast(message, "success"),
  error: (message: string) => dispatchToast(message, "error"),
  info: (message: string) => dispatchToast(message, "info"),
};

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

  // Listen for custom toast events (from loaders, WebSocket errors, etc.)
  useEffect(() => {
    const handleToastEvent = (event: Event) => {
      const customEvent = event as CustomEvent<{
        type: ToastType;
        message: string;
      }>;
      if (customEvent.detail?.message) {
        showToast(
          customEvent.detail.message,
          customEvent.detail.type || "info"
        );
      }
    };

    window.addEventListener("toast", handleToastEvent);
    return () => {
      window.removeEventListener("toast", handleToastEvent);
    };
  }, [showToast]);
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
                    ? "border-emerald-800 bg-emerald-800"
                    : toast.type === "error"
                      ? "border-red-800 bg-red-800"
                      : "border-sky-800 bg-sky-800"
                }`}
            >
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
