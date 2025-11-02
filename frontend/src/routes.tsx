import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("components/Home.tsx"),

  // Offline page
  route("offline", "components/BackendUnavailable.tsx"),

  // Game routes
  route("game/:gameID?", "components/GameContainer.tsx", [
    { index: true, file: "components/GameHome.tsx" },
    { path: "turn/:clientID", file: "components/PlayerTurn.tsx" },
  ]),

  // Options routes
  route("options", "components/OptionsPage.tsx"),

  // Dev-only routes
  route("dev/toasts", "components/ToastExamples.tsx"),
] satisfies RouteConfig;
