import { useEffect } from "react";
import { useNavigate } from "react-router";
import { useToast } from "./ToastProvider";

export default function ToastExamples() {
  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!import.meta.env.DEV) {
      navigate("/", { replace: true });
    }
  }, [navigate]);

  if (!import.meta.env.DEV) {
    return null;
  }

  return (
    <div className="h-full flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className=" rounded-lg backdrop-blur-md px-5 py-6 space-y-6 text-center">
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold text-slate-200">
              Toast Examples (Dev Only)
            </h1>
            <p className="text-sm text-slate-400">
              Toasts appear at the top of the screen. Click any toast to dismiss
              it.
            </p>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-3">
            <button
              onClick={() => showToast("Heads up: informational toast", "info")}
              className="px-4 py-2 rounded-lg text-white border border-slate-700/50 bg-linear-to-r from-sky-600/30 to-indigo-600/30 hover:from-sky-600/40 hover:to-indigo-600/40 transition-colors"
            >
              Show Info Toast
            </button>
            <button
              onClick={() => showToast("Saved successfully!", "success")}
              className="px-4 py-2 rounded-lg text-white border border-slate-700/50 bg-linear-to-r from-emerald-600/30 to-teal-600/30 hover:from-emerald-600/40 hover:to-teal-600/40 transition-colors"
            >
              Show Success Toast
            </button>
            <button
              onClick={() => showToast("Something went wrong", "error")}
              className="px-4 py-2 rounded-lg text-white border border-slate-700/50 bg-linear-to-r from-red-600/30 to-rose-600/30 hover:from-red-600/40 hover:to-rose-600/40 transition-colors"
            >
              Show Error Toast
            </button>
          </div>

          <div className="flex items-center justify-center">
            <button
              onClick={() => {
                showToast("Queued info", "info");
                setTimeout(() => showToast("Queued success", "success"), 300);
                setTimeout(() => showToast("Queued error", "error"), 600);
              }}
              className="px-4 py-2 rounded-lg text-white border border-slate-700/50 bg-slate-700/30 hover:bg-slate-700/40 transition-colors"
            >
              Burst 3 Toasts
            </button>
          </div>

          <div className="text-xs text-slate-500">
            Route: <code className="text-slate-300">/dev/toasts</code>
          </div>
        </div>
      </div>
    </div>
  );
}
