import { vi } from "vitest";
import { render } from "@testing-library/react";
import type { ReactNode } from "react";
import type { useNavigate } from "react-router";

// Mock React Router hooks - must be defined before any imports that use them
export const mockUseLoaderData = vi.fn();
export const mockUseParams = vi.fn();
export const mockUseNavigate = vi.fn();
export const mockUseOutletContext = vi.fn();

// Helper to create mock WebSocketManager
export function createMockWebSocketManager(overrides?: Partial<any>) {
  const mockManager: any = {
    gameID: "TEST123",
    connection: {
      send: vi.fn().mockResolvedValue(undefined),
      sendAndWait: vi.fn().mockResolvedValue(undefined),
    },
    onPeersChange: vi.fn(() => () => {}), // Returns unsubscribe function
    onTurnChange: vi.fn(() => () => {}), // Returns unsubscribe function
    destroy: vi.fn(),
    peers: [],
    currentTurn: null,
    turnStartTime: null,
    ...overrides,
  };
  return mockManager;
}

// Helper to render with router context
export function renderWithRouter(
  component: ReactNode,
  options: {
    loaderData?: any;
    params?: Record<string, string>;
    outletContext?: any;
    navigate?: ReturnType<typeof useNavigate>;
  } = {}
) {
  mockUseLoaderData.mockReturnValue(options.loaderData);
  mockUseParams.mockReturnValue(options.params || {});
  mockUseNavigate.mockReturnValue(options.navigate || vi.fn());
  mockUseOutletContext.mockReturnValue(options.outletContext);

  return render(<>{component}</>);
}

// Store mocks for GameContainer test
(renderWithRouter as any).mockLoaderData = null;
(renderWithRouter as any).mockNavigate = vi.fn();
(renderWithRouter as any).mockParams = {};

// Helper for cleanup
export function resetMocks() {
  mockUseLoaderData.mockReset();
  mockUseParams.mockReset();
  mockUseNavigate.mockReset();
  mockUseOutletContext.mockReset();
}
