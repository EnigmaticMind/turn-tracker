import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useNavigation,
} from "react-router";
import { Suspense } from "react";

import type { Route } from "./+types/root";
import "./app.css";
import { ToastProvider } from "./components/ToastProvider";
import TopLoadingBar from "./components/TopLoadingBar";

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
          {/* Animated Background Elements */}
          <div className="absolute inset-0">
            {/* Floating Particles */}
            <div className="absolute top-1/4 left-1/4 w-2 h-2 bg-blue-400 rounded-full animate-bounce opacity-60"></div>
            <div className="absolute top-1/3 right-1/3 w-1 h-1 bg-cyan-400 rounded-full animate-pulse opacity-40"></div>
            <div className="absolute bottom-1/3 left-1/3 w-1.5 h-1.5 bg-teal-400 rounded-full animate-bounce delay-1000 opacity-50"></div>
            <div className="absolute bottom-1/4 right-1/4 w-2 h-2 bg-purple-400 rounded-full animate-pulse delay-2000 opacity-60"></div>

            {/* Geometric Shapes */}
            {/* <div className="absolute top-20 left-20 w-8 h-8 border border-blue-400/30 rotate-45 animate-spin"></div> */}
            <div className="absolute top-40 right-20 w-6 h-6 border border-cyan-400/30 rounded-full animate-pulse"></div>
            {/* <div className="absolute bottom-20 left-40 w-4 h-4 bg-gradient-to-r from-teal-400 to-purple-400 rounded-full animate-bounce delay-500"></div> */}
            <div className="absolute bottom-40 left-30 w-5 h-5 border border-purple-400/20 rounded-full animate-pulse delay-1000"></div>

            {/* Gradient Orbs */}
            <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-gradient-to-r from-blue-500/10 to-purple-500/10 rounded-full blur-3xl animate-pulse"></div>
            <div className="absolute top-1/3 right-1/3 w-64 h-64 bg-gradient-to-r from-cyan-500/10 to-teal-500/10 rounded-full blur-2xl animate-pulse delay-1000"></div>
            <div className="absolute bottom-1/3 left-1/3 w-80 h-80 bg-gradient-to-r from-purple-500/10 to-pink-500/10 rounded-full blur-3xl animate-pulse delay-2000"></div>
          </div>

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
    <div>
      <TopLoadingBar />
    </div>
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
