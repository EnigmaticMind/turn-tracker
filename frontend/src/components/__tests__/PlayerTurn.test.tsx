import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import PlayerTurn from "../PlayerTurn";
import {
  createMockWebSocketManager,
  renderWithRouter,
  resetMocks,
  mockUseOutletContext,
  mockUseParams,
} from "./testUtils";
import type { PeerInfo } from "../../lib/websocket/handlers/types";

// Mock React Router
vi.mock("react-router", async () => {
  const actual = await vi.importActual("react-router");
  const { mockUseOutletContext, mockUseParams } = await import("./testUtils");
  return {
    ...actual,
    useOutletContext: () => mockUseOutletContext(),
    useParams: () => mockUseParams(),
  };
});

// Mock dependencies
vi.mock("../LiquidAvatar", () => ({
  LiquidAvatar: ({
    name,
    color,
    active,
  }: {
    name: string;
    color: string;
    active?: boolean;
  }) => (
    <div
      data-testid={`avatar-${name}`}
      data-active={active}
      style={{ backgroundColor: color }}
    >
      {name}
    </div>
  ),
}));

vi.mock("../PlayerGrid", () => ({
  PlayerGrid: ({ players, onSelect }: any) => (
    <div data-testid="player-grid">
      {players.map((p: any) => (
        <button
          key={p.client_id}
          data-testid={`player-${p.client_id}`}
          onClick={() => onSelect?.(p.client_id)}
        >
          {p.display_name}
        </button>
      ))}
    </div>
  ),
}));

vi.mock("../../lib/websocket/handlers/startTurn", () => ({
  startTurn: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("../../lib/websocket/handlers/endTurn", () => ({
  endTurn: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("../ToastProvider", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
  dispatchToast: vi.fn(),
  ToastProvider: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}));

describe("PlayerTurn", () => {
  const mockPeers: PeerInfo[] = [
    {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    },
    {
      client_id: "2",
      display_name: "Bob",
      color: "#33FF57",
      total_turn_time: 0,
    },
    {
      client_id: "3",
      display_name: "Charlie",
      color: "#3357FF",
      total_turn_time: 0,
    },
  ];

  beforeEach(async () => {
    resetMocks();
    vi.clearAllMocks();

    // Clear mocks
    const { startTurn } = await import(
      "../../lib/websocket/handlers/startTurn"
    );
    const { endTurn } = await import("../../lib/websocket/handlers/endTurn");
    vi.mocked(startTurn).mockClear();
    vi.mocked(endTurn).mockClear();

    // Set default mocks - removed turnStartTime, clientID not used anymore
    mockUseParams.mockReturnValue({ clientID: "1" }); // Still mocked but not used by component
    mockUseOutletContext.mockReturnValue({
      ws: null,
      peers: [],
      currentTurn: null,
    });
  });

  afterEach(() => {
    // Remove vi.useRealTimers() since we're not using fake timers anymore
  });

  it('should render "Connecting..." when ws is null', () => {
    renderWithRouter(<PlayerTurn />, {
      params: { clientID: "1" },
      outletContext: {
        ws: null,
        peers: mockPeers,
        currentTurn: null,
        turnStartTime: null,
      },
    });

    expect(screen.getByText("Connecting...")).toBeInTheDocument();
  });

  it('should render "Player not found" when currentTurn is null', () => {
    const mockWS = createMockWebSocketManager();
    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn: null, // This is what triggers "Player not found"
      },
    });

    expect(screen.getByText("Player not found")).toBeInTheDocument();
  });

  it("should render current player information", () => {
    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn,
      },
    });

    expect(screen.getByText("Current Turn")).toBeInTheDocument();
    expect(screen.getByText("Tap below to end turn")).toBeInTheDocument();
    expect(screen.getByTestId("avatar-Alice")).toBeInTheDocument();
    expect(screen.getByTestId("avatar-Alice")).toHaveAttribute(
      "data-active",
      "true"
    );
  });

  it("should display other players in the grid", () => {
    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn,
      },
    });

    // Should not show current player (Alice) in the grid
    expect(screen.queryByTestId("player-1")).not.toBeInTheDocument();
    // Should show other players
    expect(screen.getByTestId("player-2")).toHaveTextContent("Bob");
    expect(screen.getByTestId("player-3")).toHaveTextContent("Charlie");
  });

  it("should call endTurn when current turn area is clicked", async () => {
    const { endTurn } = await import("../../lib/websocket/handlers/endTurn");
    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn,
      },
    });

    // Find the clickable area - it's a div with cursor-pointer class containing "Current Turn"
    const currentTurnHeading = screen.getByText("Current Turn");
    const clickableArea = currentTurnHeading.closest(".cursor-pointer");

    expect(clickableArea).toBeInTheDocument();
    fireEvent.click(clickableArea!);

    await waitFor(
      () => {
        expect(endTurn).toHaveBeenCalledWith(mockWS);
      },
      { timeout: 2000 }
    );
  });

  it("should call startTurn when another player is selected", async () => {
    const { startTurn } = await import(
      "../../lib/websocket/handlers/startTurn"
    );
    vi.mocked(startTurn).mockClear();

    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn,
      },
    });

    // Wait for the player grid to render with otherPlayers
    await waitFor(() => {
      expect(screen.getByTestId("player-grid")).toBeInTheDocument();
    });

    const bobButton = screen.getByTestId("player-2");
    expect(bobButton).toBeInTheDocument();

    fireEvent.click(bobButton);

    await waitFor(
      () => {
        expect(startTurn).toHaveBeenCalledWith(mockWS, "2");
        expect(startTurn).toHaveBeenCalledTimes(1);
      },
      { timeout: 3000 }
    );
  });

  it("should hide instruction text when no other players", () => {
    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };
    const singlePlayer: PeerInfo[] = [mockPeers[0]];

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: singlePlayer,
        currentTurn,
      },
    });

    const instructionText = screen.getByText(
      "Tap below to start the next player's turn"
    );
    expect(instructionText).toHaveStyle({ visibility: "hidden" });
  });

  it("should show toast error when endTurn fails", async () => {
    const { endTurn } = await import("../../lib/websocket/handlers/endTurn");
    const { toast } = await import("../ToastProvider");
    vi.mocked(endTurn).mockRejectedValueOnce(new Error("Network error"));
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn,
      },
    });

    const currentTurnHeading = screen.getByText("Current Turn");
    const clickableArea = currentTurnHeading.closest(".cursor-pointer");
    fireEvent.click(clickableArea!);

    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "PlayerTurn: ",
        expect.any(Error)
      );
      expect(toast.error).toHaveBeenCalledWith("Network error");
    });

    consoleErrorSpy.mockRestore();
  });

  it("should show toast error when startTurn fails", async () => {
    const { startTurn } = await import(
      "../../lib/websocket/handlers/startTurn"
    );
    const { toast } = await import("../ToastProvider");
    vi.mocked(startTurn).mockRejectedValueOnce(new Error("Network error"));
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const mockWS = createMockWebSocketManager();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    renderWithRouter(<PlayerTurn />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn,
      },
    });

    const bobButton = screen.getByTestId("player-2");
    fireEvent.click(bobButton);

    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "PlayerTurn: ",
        expect.any(Error)
      );
      expect(toast.error).toHaveBeenCalledWith("Network error");
    });

    consoleErrorSpy.mockRestore();
  });
});
