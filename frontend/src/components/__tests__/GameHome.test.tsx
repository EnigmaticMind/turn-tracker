import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GameHome from "../GameHome";
import {
  createMockWebSocketManager,
  renderWithRouter,
  resetMocks,
  mockUseOutletContext,
} from "./testUtils";
import type { PeerInfo } from "../../lib/websocket/handlers/types";

// Mock React Router
vi.mock("react-router", async () => {
  const actual = await vi.importActual("react-router");
  const { mockUseOutletContext } = await import("./testUtils");
  return {
    ...actual,
    useOutletContext: () => mockUseOutletContext(),
  };
});

// Mock dependencies
vi.mock("../PlayerGrid", () => ({
  PlayerGrid: ({ players, onSelect, currentTurn }: any) => (
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
      {currentTurn && (
        <div data-testid="current-turn">{currentTurn.client_id}</div>
      )}
    </div>
  ),
}));

vi.mock("../../lib/websocket/handlers/startTurn", () => ({
  startTurn: vi.fn().mockResolvedValue(undefined),
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

describe("GameHome", () => {
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
  ];

  beforeEach(() => {
    resetMocks();
    vi.clearAllMocks();
    // Set default outlet context mock
    mockUseOutletContext.mockReturnValue({
      ws: null,
      peers: [],
      currentTurn: null,
      turnStartTime: null,
    });
  });

  it('should render "Connecting..." when ws is null', () => {
    renderWithRouter(<GameHome />, {
      outletContext: {
        ws: null,
        peers: [],
        currentTurn: null,
        turnStartTime: null,
      },
    });

    expect(screen.getByText("Connecting...")).toBeInTheDocument();
  });

  it("should render player grid with peers", () => {
    const mockWS = createMockWebSocketManager();
    renderWithRouter(<GameHome />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn: null,
        turnStartTime: null,
      },
    });

    expect(screen.getByTestId("player-grid")).toBeInTheDocument();
    expect(screen.getByTestId("player-1")).toHaveTextContent("Alice");
    expect(screen.getByTestId("player-2")).toHaveTextContent("Bob");
  });

  it("should call startTurn when a player is selected", async () => {
    const { startTurn } = await import(
      "../../lib/websocket/handlers/startTurn"
    );
    const mockWS = createMockWebSocketManager();

    renderWithRouter(<GameHome />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn: null,
        turnStartTime: null,
      },
    });

    const aliceButton = screen.getByTestId("player-1");
    fireEvent.click(aliceButton);

    await waitFor(() => {
      expect(startTurn).toHaveBeenCalledWith(mockWS, "1");
      expect(startTurn).toHaveBeenCalledTimes(1);
    });
  });

  it("should not call startTurn when ws is null", async () => {
    const { startTurn } = await import(
      "../../lib/websocket/handlers/startTurn"
    );

    renderWithRouter(<GameHome />, {
      outletContext: {
        ws: null,
        peers: mockPeers,
        currentTurn: null,
        turnStartTime: null,
      },
    });

    // Component should render "Connecting..." and not the grid
    expect(screen.getByText("Connecting...")).toBeInTheDocument();
    expect(screen.queryByTestId("player-grid")).not.toBeInTheDocument();
  });

  it("should display instruction text", () => {
    const mockWS = createMockWebSocketManager();
    renderWithRouter(<GameHome />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn: null,
        turnStartTime: null,
      },
    });

    expect(
      screen.getByText("Tap a player to start their turn")
    ).toBeInTheDocument();
  });

  it("should handle startTurn errors gracefully", async () => {
    const { startTurn } = await import(
      "../../lib/websocket/handlers/startTurn"
    );
    const { toast } = await import("../ToastProvider");
    vi.mocked(startTurn).mockRejectedValueOnce(new Error("Network error"));
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const mockWS = createMockWebSocketManager();
    renderWithRouter(<GameHome />, {
      outletContext: {
        ws: mockWS,
        peers: mockPeers,
        currentTurn: null,
        turnStartTime: null,
      },
    });

    const aliceButton = screen.getByTestId("player-1");
    fireEvent.click(aliceButton);

    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "GameHome: ",
        expect.any(Error)
      );
      expect(toast.error).toHaveBeenCalledWith("Network error");
      expect(toast.error).toHaveBeenCalledTimes(1);
    });

    consoleErrorSpy.mockRestore();
  });
});
