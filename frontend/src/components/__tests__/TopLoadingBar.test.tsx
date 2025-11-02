import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import TopLoadingBar, { startLoading, stopLoading } from "../TopLoadingBar";

// Mock React Router hooks
const mockUseNavigation = vi.fn();
const mockUseLocation = vi.fn();

vi.mock("react-router", async () => {
  const actual = await vi.importActual("react-router");
  return {
    ...actual,
    useNavigation: () => mockUseNavigation(),
    useLocation: () => mockUseLocation(),
  };
});

// Mock framer-motion to avoid animation issues in tests
vi.mock("framer-motion", () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
  },
}));

describe("TopLoadingBar", () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Set default mocks for navigation and location
    mockUseNavigation.mockReturnValue({
      state: "idle",
      location: null,
    });

    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: "default",
    });
  });

  afterEach(() => {
    // Clean up any active loading states
    stopLoading();
  });

  it("should not render when no loading is active", () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    expect(screen.queryByText(/Loading/i)).not.toBeInTheDocument();
  });

  it("should show loading bar when startLoading is called", async () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    startLoading();

    await waitFor(() => {
      expect(screen.getByText("Loading...")).toBeInTheDocument();
    });
  });

  it("should show custom message when startLoading is called with message", async () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    startLoading("Creating game...");

    await waitFor(() => {
      expect(screen.getByText("Creating game...")).toBeInTheDocument();
    });
  });

  it("should hide loading bar when stopLoading is called", async () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    startLoading("Test message");

    await waitFor(() => {
      expect(screen.getByText("Test message")).toBeInTheDocument();
    });

    stopLoading();

    await waitFor(() => {
      expect(screen.queryByText("Test message")).not.toBeInTheDocument();
    });
  });

  it("should show loading bar for navigation loading state", () => {
    mockUseNavigation.mockReturnValue({
      state: "loading",
      location: null,
    });

    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    expect(screen.getByText("Loading...")).toBeInTheDocument();
  });

  it("should show 'Creating game...' message when navigating to /game", () => {
    mockUseNavigation.mockReturnValue({
      state: "loading",
      location: {
        pathname: "/game",
        search: "",
        hash: "",
        state: null,
        key: "test",
      },
    });

    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    expect(screen.getByText("Creating game...")).toBeInTheDocument();
  });

  it("should show 'Joining game {gameID}...' message when navigating to /game/{gameID}", () => {
    mockUseNavigation.mockReturnValue({
      state: "loading",
      location: {
        pathname: "/game/ABCD",
        search: "",
        hash: "",
        state: null,
        key: "test",
      },
    });

    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    expect(screen.getByText("Joining game ABCD...")).toBeInTheDocument();
  });

  it("should prioritize custom loading message over navigation message", async () => {
    mockUseNavigation.mockReturnValue({
      state: "loading",
      location: {
        pathname: "/game",
        search: "",
        hash: "",
        state: null,
        key: "test",
      },
    });

    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    // Navigation message should show first
    expect(screen.getByText("Creating game...")).toBeInTheDocument();

    // Custom message should override
    startLoading("Custom message");

    await waitFor(() => {
      expect(screen.getByText("Custom message")).toBeInTheDocument();
      expect(screen.queryByText("Creating game...")).not.toBeInTheDocument();
    });
  });

  it("should handle multiple start/stop calls correctly", async () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    startLoading("First");

    await waitFor(() => {
      expect(screen.getByText("First")).toBeInTheDocument();
    });

    stopLoading();

    await waitFor(() => {
      expect(screen.queryByText("First")).not.toBeInTheDocument();
    });

    startLoading("Second");

    await waitFor(() => {
      expect(screen.getByText("Second")).toBeInTheDocument();
    });

    stopLoading();

    await waitFor(() => {
      expect(screen.queryByText("Second")).not.toBeInTheDocument();
    });
  });

  it("should render loading bar element when visible", async () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    startLoading();

    await waitFor(() => {
      // Check for the loading bar container (has specific classes)
      const loadingBar = document.querySelector(".fixed.top-0.left-0.right-0");
      expect(loadingBar).toBeInTheDocument();
    });
  });

  it("should render spinner when visible", async () => {
    render(
      <MemoryRouter>
        <TopLoadingBar />
      </MemoryRouter>
    );

    startLoading();

    await waitFor(() => {
      // Check for spinner element
      const spinner = document.querySelector(".border-4.border-sky-400");
      expect(spinner).toBeInTheDocument();
    });
  });
});
