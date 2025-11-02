import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useNavigation,
  useLocation,
} from "react-router";
import { Suspense } from "react";
import { motion } from "framer-motion";

import type { Route } from "./+types/root";
import "./app.css";
import { ToastProvider } from "./components/ToastProvider";
import TopLoadingBar from "./components/TopLoadingBar";
import AnimatedBackground from "./components/AnimatedBackground";

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        <div className="min-h-screen h-screen bg-linear-to-br from-slate-900 via-slate-800 to-slate-900 relative overflow-hidden">
          <AnimatedBackground />
          <div className="relative z-10 h-full">{children}</div>
        </div>
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}
export function HydrateFallback() {
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

      <div className="min-h-screen h-screen bg-linear-to-br from-slate-900 via-slate-800 to-slate-900 relative overflow-hidden">
        <AnimatedBackground />
        <div className="relative flex inset-0 z-10 h-full items-center justify-center">
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
              Connecting to remote server...
            </p>
          </motion.div>
        </div>
      </div>
    </>
  );
}

export default function App() {
  return (
    <ToastProvider>
      <Suspense fallback={null}>
        <Outlet />
      </Suspense>
    </ToastProvider>
  );
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  let message = "Oops!";
  let details = "An unexpected error occurred.";
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? "404" : "Error";
    details =
      error.status === 404
        ? "The requested page could not be found."
        : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <main className="pt-16 p-4 container mx-auto">
      <h1>{message}</h1>
      <p>{details}</p>
      {stack && (
        <pre className="w-full p-4 overflow-x-auto">
          <code>{stack}</code>
        </pre>
      )}
    </main>
  );
}
