import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { PlayerGrid, type Player } from "../PlayerGrid";
import type { PeerInfo } from "../../lib/websocket/handlers/types";

// Mock LiquidAvatar
vi.mock("../LiquidAvatar", () => ({
  LiquidAvatar: ({ name, color }: { name: string; color: string }) => (
    <div data-testid={`avatar-${name}`} style={{ backgroundColor: color }}>
      {name}
    </div>
  ),
}));

describe("PlayerGrid", () => {
  const mockPlayers: Player[] = [
    { client_id: "1", display_name: "Alice", color: "#FF5733" },
    { client_id: "2", display_name: "Bob", color: "#33FF57" },
    { client_id: "3", display_name: "Charlie", color: "#3357FF" },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should render "Waiting for players..." when no players', () => {
    render(<PlayerGrid players={[]} />);
    expect(screen.getByText("Waiting for players...")).toBeInTheDocument();
  });

  it("should render all players", () => {
    render(<PlayerGrid players={mockPlayers} />);
    expect(screen.getByTestId("avatar-Alice")).toBeInTheDocument();
    expect(screen.getByTestId("avatar-Bob")).toBeInTheDocument();
    expect(screen.getByTestId("avatar-Charlie")).toBeInTheDocument();
  });

  it("should call onSelect when a player is clicked", () => {
    const onSelect = vi.fn();
    render(<PlayerGrid players={mockPlayers} onSelect={onSelect} />);

    const aliceAvatar = screen.getByTestId("avatar-Alice").parentElement;
    fireEvent.click(aliceAvatar!);

    expect(onSelect).toHaveBeenCalledWith("1");
    expect(onSelect).toHaveBeenCalledTimes(1);
  });

  it("should not call onSelect if not provided", () => {
    render(<PlayerGrid players={mockPlayers} />);
    const aliceAvatar = screen.getByTestId("avatar-Alice").parentElement;
    expect(() => fireEvent.click(aliceAvatar!)).not.toThrow();
  });

  it("should apply correct grid columns based on cols prop", () => {
    const { container } = render(<PlayerGrid players={mockPlayers} cols={3} />);
    const grid = container.querySelector(".player-grid") as HTMLElement;
    expect(grid?.style.gridTemplateColumns).toBe("repeat(3, 1fr)");
  });

  it("should enable scrolling when shouldScroll is true", () => {
    const { container } = render(
      <PlayerGrid players={mockPlayers} shouldScroll={true} />
    );
    const scrollContainer = container.querySelector(".overflow-y-auto");
    expect(scrollContainer).toBeInTheDocument();
  });

  it("should enable scrolling when players exceed rows * cols", () => {
    const manyPlayers = Array.from({ length: 10 }, (_, i) => ({
      client_id: `${i}`,
      display_name: `Player ${i}`,
      color: "#FF0000",
    }));
    const { container } = render(
      <PlayerGrid players={manyPlayers} rows={2} cols={2} />
    );
    const scrollContainer = container.querySelector(".overflow-y-auto");
    expect(scrollContainer).toBeInTheDocument();
  });

  it("should display elapsed time for current turn player", async () => {
    vi.useRealTimers();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };
    const turnStartTime = Date.now() - 5000; // 5 seconds ago

    render(
      <PlayerGrid
        players={mockPlayers}
        currentTurn={currentTurn}
        turnStartTime={turnStartTime}
      />
    );

    // Wait for initial time calculation
    await waitFor(
      () => {
        const timeElement = screen.getByText(/0:0[5-6]/);
        expect(timeElement).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("should update elapsed time every second", async () => {
    vi.useRealTimers();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };
    const turnStartTime = Date.now() - 2000; // 2 seconds ago

    render(
      <PlayerGrid
        players={mockPlayers}
        currentTurn={currentTurn}
        turnStartTime={turnStartTime}
      />
    );

    // Initial time should be around 2 seconds
    await waitFor(
      () => {
        expect(screen.getByText(/0:0[2-3]/)).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    // Wait a bit for the timer to update
    await new Promise((resolve) => setTimeout(resolve, 1500));

    // Time should now be around 3-4 seconds
    await waitFor(
      () => {
        const timeText = screen.getByText(/0:0[3-4]/);
        expect(timeText).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("should not display elapsed time for non-current turn players", () => {
    const currentTurn: PeerInfo = {
      client_id: "2",
      display_name: "Bob",
      color: "#33FF57",
      total_turn_time: 0,
    };
    const turnStartTime = Date.now();

    render(
      <PlayerGrid
        players={mockPlayers}
        currentTurn={currentTurn}
        turnStartTime={turnStartTime}
      />
    );

    // Alice should not have time displayed
    const aliceContainer = screen.getByTestId("avatar-Alice").parentElement;
    const timeElements = aliceContainer?.querySelectorAll(".font-mono");
    expect(timeElements?.length).toBe(0);
  });

  it("should format time correctly for hours", async () => {
    vi.useRealTimers();
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };
    const turnStartTime = Date.now() - 3665000; // 1 hour, 1 minute, 5 seconds

    render(
      <PlayerGrid
        players={mockPlayers}
        currentTurn={currentTurn}
        turnStartTime={turnStartTime}
      />
    );

    await waitFor(
      () => {
        expect(screen.getByText(/1:01:0[4-6]/)).toBeInTheDocument();
      },
      { timeout: 2000 }
    );
  });

  it("should reset elapsed time when turnStartTime is null", () => {
    const currentTurn: PeerInfo = {
      client_id: "1",
      display_name: "Alice",
      color: "#FF5733",
      total_turn_time: 0,
    };

    const { rerender } = render(
      <PlayerGrid
        players={mockPlayers}
        currentTurn={currentTurn}
        turnStartTime={Date.now()}
      />
    );

    rerender(
      <PlayerGrid
        players={mockPlayers}
        currentTurn={currentTurn}
        turnStartTime={null}
      />
    );

    const timeElements = screen.queryAllByText(/0:0[0-9]/);
    expect(timeElements.length).toBe(0);
  });
});
