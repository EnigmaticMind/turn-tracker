import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import GameContainer from "../GameContainer";
import {
  createMockWebSocketManager,
  resetMocks,
  mockUseLoaderData,
  mockUseNavigate,
} from "./testUtils";
import type { PeerInfo } from "../../lib/websocket/handlers/types";

// Mock toast module
vi.mock("../ToastProvider", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
  dispatchToast: vi.fn(),
  useToast: vi.fn(),
  ToastProvider: ({ children }: { children: React.ReactNode }) => (
    <>{children}</>
  ),
}));

// Mock dependencies
vi.mock("../Options", () => ({
  default: ({ onClose, ws }: { onClose: () => void; ws?: any }) => (
    <div data-testid="options-modal-content">
      <button data-testid="close-options" onClick={onClose}>
        Close
      </button>
      {ws && <div data-testid="ws-provided">WS provided</div>}
    </div>
  ),
}));

vi.mock("lucide-react", () => ({
  Home: () => <svg data-testid="home-icon" />,
  Settings: () => <svg data-testid="settings-icon" />,
}));

// Add mock for TopLoadingBar
vi.mock("../TopLoadingBar", () => ({
  default: vi.fn(() => null),
}));

// Mock persistentConnection module
vi.mock("../../lib/websocket/persistentConnection", () => ({
  getPersistentConnection: vi.fn(),
  destroyPersistentConnection: vi.fn(),
}));

// Mock joinGame module
vi.mock("../../lib/websocket/handlers/joinGame", () => ({
  joinGame: vi.fn(),
}));

// Mock createGame module
vi.mock("../../lib/websocket/handlers/createGame", () => ({
  createGame: vi.fn(),
}));

// Define type for navigation mock to match React Router's Navigation
type MockNavigation = {
  state: "idle" | "loading" | "submitting";
  location: { pathname: string } | null;
};

// Add mock for useNavigation
const mockUseNavigation = vi.fn(
  (): MockNavigation => ({
    state: "idle",
    location: null,
  })
);

// Update the React Router mock to use a simpler redirect
vi.mock("react-router", async () => {
  // Simple redirect that throws with path information
  const redirect = (path: string) => {
    const error = new Error(`Redirect to ${path}`);
    (error as any).status = 302;
    (error as any).headers = { location: path };
    throw error;
  };

  const { mockUseLoaderData, mockUseNavigate } = await import("./testUtils");

  return {
    Outlet: ({ context }: { context?: any }) => (
      <div data-testid="outlet">
        {context && (
          <div data-testid="outlet-context">
            {JSON.stringify(context, null, 2)}
          </div>
        )}
      </div>
    ),
    useLoaderData: () => mockUseLoaderData(),
    useNavigate: () => mockUseNavigate(),
    useParams: () => ({}),
    useNavigation: () => mockUseNavigation(),
    useLocation: () => ({ pathname: "/game/TEST123" }),
    useBeforeUnload: vi.fn(),
    redirect,
  };
});

// Mock wake lock API
const mockWakeLock = {
  request: vi.fn().mockResolvedValue({
    release: vi.fn().mockResolvedValue(undefined),
  }),
};

Object.defineProperty(navigator, "wakeLock", {
  value: mockWakeLock,
  writable: true,
});

describe("GameContainer", () => {
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

  let mockWS: any;
  let mockNavigate: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    resetMocks();
    vi.clearAllMocks();
    mockWS = createMockWebSocketManager({ gameID: "TEST123" });
    mockNavigate = vi.fn();

    // Setup mocks
    mockUseLoaderData.mockReturnValue(mockWS);
    mockUseNavigate.mockReturnValue(mockNavigate);

    // Reset navigation to default state
    mockUseNavigation.mockReturnValue({
      state: "idle",
      location: null,
    });

    // Reset wake lock
    mockWakeLock.request.mockClear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("should render game code in bottom bar", () => {
    render(<GameContainer />);
    expect(screen.getByText("TEST123")).toBeInTheDocument();
  });

  it("should render outlet with context", () => {
    render(<GameContainer />);
    expect(screen.getByTestId("outlet")).toBeInTheDocument();
    expect(screen.getByTestId("outlet-context")).toBeInTheDocument();
  });

  it("should navigate to home when home button is clicked", () => {
    render(<GameContainer />);
    const homeButton = screen.getByTestId("home-icon").closest("button");
    fireEvent.click(homeButton!);

    expect(mockNavigate).toHaveBeenCalledWith("/");
  });

  it("should copy game code to clipboard when clicked", async () => {
    // Mock clipboard API
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockResolvedValue(undefined),
      },
    });

    render(<GameContainer />);

    const gameCodeElement = screen.getByText("TEST123");
    fireEvent.click(gameCodeElement);

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith("TEST123");
    });
  });

  it("should open options modal when settings button is clicked", () => {
    render(<GameContainer />);
    const settingsButton = screen
      .getByTestId("settings-icon")
      .closest("button");
    fireEvent.click(settingsButton!);

    expect(screen.getByTestId("options-modal-content")).toBeInTheDocument();
    expect(screen.getByTestId("ws-provided")).toBeInTheDocument();
  });

  it("should close options modal when close button is clicked", () => {
    render(<GameContainer />);
    const settingsButton = screen
      .getByTestId("settings-icon")
      .closest("button");
    fireEvent.click(settingsButton!);

    expect(screen.getByTestId("options-modal-content")).toBeInTheDocument();

    const closeButton = screen.getByTestId("close-options");
    fireEvent.click(closeButton);

    expect(
      screen.queryByTestId("options-modal-content")
    ).not.toBeInTheDocument();
  });

  it("should close options modal when backdrop is clicked", () => {
    render(<GameContainer />);
    const settingsButton = screen
      .getByTestId("settings-icon")
      .closest("button");
    fireEvent.click(settingsButton!);

    expect(screen.getByTestId("options-modal-content")).toBeInTheDocument();

    // Click the backdrop directly using test ID
    const backdrop = screen.getByTestId("options-modal-backdrop");
    fireEvent.click(backdrop);

    // Modal should close
    expect(
      screen.queryByTestId("options-modal-content")
    ).not.toBeInTheDocument();
  });

  it("should not close options modal when modal content is clicked", () => {
    render(<GameContainer />);
    const settingsButton = screen
      .getByTestId("settings-icon")
      .closest("button");
    fireEvent.click(settingsButton!);

    expect(screen.getByTestId("options-modal-content")).toBeInTheDocument();

    const modalContent = screen.getByTestId("options-modal-content");
    fireEvent.click(modalContent);

    // Modal should still be open
    expect(screen.getByTestId("options-modal-content")).toBeInTheDocument();
  });

  it("should unsubscribe from WebSocket events on unmount", () => {
    const unsubscribePeers = vi.fn();
    const unsubscribeTurn = vi.fn();

    mockWS.onPeersChange.mockReturnValue(unsubscribePeers);
    mockWS.onTurnChange.mockReturnValue(unsubscribeTurn);

    const { unmount } = render(<GameContainer />);
    unmount();

    expect(unsubscribeTurn).toHaveBeenCalled();
    expect(unsubscribePeers).toHaveBeenCalled();
  });

  it("should request wake lock on mount", async () => {
    render(<GameContainer />);

    await waitFor(() => {
      expect(mockWakeLock.request).toHaveBeenCalledWith("screen");
    });
  });

  it("should handle wake lock request failure gracefully", async () => {
    mockWakeLock.request.mockRejectedValueOnce(
      new Error("Wake lock not supported")
    );
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    render(<GameContainer />);

    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalled();
    });

    consoleErrorSpy.mockRestore();
  });

  describe("clientLoader", () => {
    it("should not redirect when params.gameID matches actual gameID", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");

      const mockWSWithGameID = createMockWebSocketManager({
        gameID: "MATCH123",
      });

      // Now getPersistentConnection is a mock, so we can use mockReturnValue
      vi.mocked(getPersistentConnection).mockReturnValue(
        mockWSWithGameID as any
      );

      // Call loader with matching gameID
      const params = { gameID: "MATCH123" };

      // Should resolve without redirect (no error thrown)
      const result = await clientLoader({ params });

      // Verify it returns the WebSocket manager
      expect(result).toBe(mockWSWithGameID);
    });
    it("should redirect when params.gameID is undefined and game has gameID", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");
      const mockWSWithGameID = createMockWebSocketManager({ gameID: "NEW456" });
      vi.mocked(getPersistentConnection).mockReturnValue(
        mockWSWithGameID as any
      );
      // Call loader without gameID in params
      const params = {};
      // Catch the redirect error and verify the path
      try {
        await clientLoader({ params });
        expect.fail("Should have thrown a redirect");
      } catch (error: any) {
        expect(error.message).toContain("Redirect");
        expect(error.message).toContain("/game/NEW456");
      }
    });

    it("when createGame returns an error, should redirect to home and show toast", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");
      // Import toast from the mocked module
      const { toast } = await import("../ToastProvider");
      // Import createGame from the mocked module
      const { createGame: mockCreateGame } = await import(
        "../../lib/websocket/handlers/createGame"
      );

      const mockWSWithGameID = createMockWebSocketManager({ gameID: "NEW456" });
      vi.mocked(getPersistentConnection).mockReturnValue(
        mockWSWithGameID as any
      );

      // Mock createGame to reject - it's already a mock function from vi.mock()
      (mockCreateGame as any).mockRejectedValue(
        new Error("Failed to create game")
      );

      const params = {};

      // Call clientLoader first - it should throw redirect
      await expect(clientLoader({ params })).rejects.toMatchObject({
        status: 302,
        headers: { location: "/" },
      });

      // Then verify toast.error was called AFTER clientLoader runs
      expect(toast.error).toHaveBeenCalledWith("Failed to create game");
    });
    it("when joinGame returns an error, should redirect to home and show toast", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");
      // Import toast from the mocked module
      const { toast } = await import("../ToastProvider");
      // Import joinGame from the mocked module
      const { joinGame: mockJoinGame } = await import(
        "../../lib/websocket/handlers/joinGame"
      );

      const mockWSWithGameID = createMockWebSocketManager({ gameID: null });
      vi.mocked(getPersistentConnection).mockReturnValue(
        mockWSWithGameID as any
      );

      // Mock createGame to reject - it's already a mock function from vi.mock()
      (mockJoinGame as any).mockRejectedValue(new Error("Failed to join game"));

      const params = { gameID: "NEW456" };

      // Call clientLoader and catch the redirect
      await expect(clientLoader({ params })).rejects.toMatchObject({
        status: 302,
        headers: { location: "/?code=NEW456" },
      });

      // Then verify toast.error was called AFTER clientLoader runs
      expect(toast.error).toHaveBeenCalledWith("Failed to join game");
    });

    it("when joinGame returns a success, should not redirect if already joined the game", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");
      // Import toast from the mocked module
      const { toast } = await import("../ToastProvider");
      // Import joinGame from the mocked module
      const { joinGame: mockJoinGame } = await import(
        "../../lib/websocket/handlers/joinGame"
      );

      const mockWS = createMockWebSocketManager({ gameID: "NEW456" });
      vi.mocked(getPersistentConnection).mockReturnValue(mockWS as any);

      // Mock createGame to reject - it's already a mock function from vi.mock()
      (mockJoinGame as any).mockRejectedValue(new Error("Failed to join game"));

      const params = { gameID: "NEW456" };

      // Call clientLoader - should return without calling joinGame
      const result = await clientLoader({ params });

      // Verify joinGame was NOT called (because ws.gameID === params.gameID)
      expect(mockJoinGame).not.toHaveBeenCalled();

      // Verify it returns the WebSocket manager
      expect(result).toBe(mockWS);
    });

    it("should use the same persistent connection when called twice", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");

      // Import joinGame from the mocked module
      const { joinGame: mockJoinGame } = await import(
        "../../lib/websocket/handlers/joinGame"
      );

      // Create a single mock WebSocket manager that will be reused
      const mockWS = createMockWebSocketManager({ gameID: "TEST123" });

      // Mock getPersistentConnection to return the same instance both times
      vi.mocked(getPersistentConnection).mockReturnValue(mockWS as any);

      // Mock joinGame to succeed (since we're testing with a gameID)
      (mockJoinGame as any).mockResolvedValue({
        type: "room_joined",
        data: { room_id: "TEST123" },
      });

      const params = { gameID: "TEST123" };

      // First call to clientLoader
      const result1 = await clientLoader({ params });

      // Verify getPersistentConnection was called
      expect(getPersistentConnection).toHaveBeenCalled();
      const callCount1 = vi.mocked(getPersistentConnection).mock.calls.length;

      // Second call to clientLoader with same params
      const result2 = await clientLoader({ params });

      // Verify getPersistentConnection was called again
      expect(getPersistentConnection).toHaveBeenCalledTimes(callCount1 + 1);

      // Verify both calls returned the SAME WebSocket instance
      expect(result1).toBe(result2);
      expect(result1).toBe(mockWS);
      expect(result2).toBe(mockWS);

      // Verify the WebSocket wasn't destroyed or recreated
      expect(mockWS.destroy).not.toHaveBeenCalled();

      // Verify no duplicate game operations occurred (e.g., createGame/joinGame called once each at most)
      // Since we're using the same gameID, joinGame should only be called once per clientLoader call
      // but we can't directly verify this since joinGame is mocked in the module

      // Verify the connection is still the same object reference
      expect(getPersistentConnection).toHaveReturnedWith(mockWS);
    });

    it("should handle multiple sequential clientLoader calls without errors", async () => {
      const { getPersistentConnection } = await import(
        "../../lib/websocket/persistentConnection"
      );
      const { clientLoader } = await import("../GameContainer");

      // Import joinGame from the mocked module
      const { joinGame: mockJoinGame } = await import(
        "../../lib/websocket/handlers/joinGame"
      );

      const mockWS = createMockWebSocketManager({ gameID: "MULTI123" });
      vi.mocked(getPersistentConnection).mockReturnValue(mockWS as any);

      // Mock joinGame to succeed (since we're testing with a gameID)
      (mockJoinGame as any).mockResolvedValue({
        type: "room_joined",
        data: { room_id: "MULTI123" },
      });

      const params = { gameID: "MULTI123" };

      // Call clientLoader multiple times in sequence
      const results = await Promise.all([
        clientLoader({ params }),
        clientLoader({ params }),
        clientLoader({ params }),
      ]);

      // All should return the same instance
      expect(results[0]).toBe(mockWS);
      expect(results[1]).toBe(mockWS);
      expect(results[2]).toBe(mockWS);
      expect(results[0]).toBe(results[1]);
      expect(results[1]).toBe(results[2]);

      // Verify getPersistentConnection was called (at least once, may be optimized)
      expect(getPersistentConnection).toHaveBeenCalled();

      // Verify the connection is still valid and not destroyed
      expect(mockWS.destroy).not.toHaveBeenCalled();
    });
  });

  describe("Loading states", () => {
    beforeEach(() => {
      // Reset only the navigation mock for clean state
      mockUseNavigation.mockClear();
    });

    it("should show loading indicator when ws is null (loader still running)", () => {
      mockUseLoaderData.mockReturnValue(null);
      mockUseNavigation.mockReturnValue({
        state: "loading",
        location: null,
      });

      // Render should not throw when ws is null
      expect(() => render(<GameContainer />)).not.toThrow();

      // When loader is still running, gameID should not be displayed
      expect(screen.queryByText("TEST123")).not.toBeInTheDocument();
      // Component should still render something
      expect(screen.getByTestId("outlet")).toBeInTheDocument();
    });

    it("should handle loading state when creating a new game", () => {
      mockUseLoaderData.mockReturnValue(null);
      mockUseNavigation.mockReturnValue({
        state: "loading" as const,
        location: { pathname: "/game" },
      });

      expect(() => render(<GameContainer />)).not.toThrow();

      // During loading, component should render but without gameID
      expect(screen.queryByText("TEST123")).not.toBeInTheDocument();
      expect(screen.getByTestId("outlet")).toBeInTheDocument();
    });

    it("should handle loading state when joining an existing game", () => {
      mockUseLoaderData.mockReturnValue(null);
      mockUseNavigation.mockReturnValue({
        state: "loading" as const,
        location: { pathname: "/game/ABC123" },
      });

      expect(() => render(<GameContainer />)).not.toThrow();

      // During loading, gameID from URL param shouldn't be in bottom bar yet
      // (because ws.gameID is not set until loader completes)
      expect(screen.queryByText("ABC123")).not.toBeInTheDocument();
      expect(screen.getByTestId("outlet")).toBeInTheDocument();
    });

    it("should render game code once loader completes", () => {
      mockUseLoaderData.mockReturnValue(mockWS);
      mockUseNavigation.mockReturnValue({
        state: "idle" as const,
        location: null,
      });

      render(<GameContainer />);

      expect(screen.getByText("TEST123")).toBeInTheDocument();
    });

    it("should not render game code when ws is null", () => {
      mockUseLoaderData.mockReturnValue(null);
      mockUseNavigation.mockReturnValue({
        state: "idle" as const,
        location: null,
      });

      expect(() => render(<GameContainer />)).not.toThrow();

      // When ws is null, gameID should be undefined, so no game code shown
      const gameCodeElement = screen.queryByText("TEST123");
      expect(gameCodeElement).not.toBeInTheDocument();
      // Component should still render
      expect(screen.getByTestId("outlet")).toBeInTheDocument();
    });
  });
});
