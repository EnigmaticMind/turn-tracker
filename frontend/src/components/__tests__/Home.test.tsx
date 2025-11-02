import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import Home from "../Home";

// Mock useLocation to return different query parameters
const mockUseLocation = vi.fn();

// Mock react-router - ensure this is before Home import
vi.mock("react-router", async () => {
  const actual = await vi.importActual("react-router");
  return {
    ...actual,
    useLocation: () => mockUseLocation(),
    useNavigate: vi.fn(() => vi.fn()),
  };
});

describe("Home", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Set default location
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: "default",
    });
  });

  it("should prefill game ID input when code query parameter is present", () => {
    // Mock location with code query parameter
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "?code=ABC123",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    // Component normalizes: uppercase, alphanumeric only, max 4 chars
    expect(input.value).toBe("ABC1");
  });

  it("should not prefill game ID input when code query parameter is absent", () => {
    // Mock location without code query parameter
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    expect(input.value).toBe("");
  });

  it("should prefill game ID input with code from query parameter", () => {
    // Mock location with code query parameter
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "?code=xyz789",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    // Component normalizes: uppercase, alphanumeric only, max 4 chars
    expect(input.value).toBe("XYZ7");
  });

  it("should update game ID input when code query parameter changes", () => {
    const { rerender } = render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    // Initially no code
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "",
      hash: "",
      state: null,
      key: "default",
    });

    let input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    expect(input.value).toBe("");

    // Now with code
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "?code=NEW456",
      hash: "",
      state: null,
      key: "default",
    });

    rerender(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    // Component normalizes: uppercase, alphanumeric only, max 4 chars
    expect(input.value).toBe("NEW4");
  });

  it("should filter out invalid characters from query parameter", () => {
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "?code=abc-123!",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    // Should remove invalid chars (- and !), uppercase, and limit to 4 chars
    expect(input.value).toBe("ABC1");
  });

  it("should handle lowercase input and convert to uppercase", () => {
    mockUseLocation.mockReturnValue({
      pathname: "/",
      search: "?code=test",
      hash: "",
      state: null,
      key: "default",
    });

    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    );

    const input = screen.getByPlaceholderText("Enter code") as HTMLInputElement;
    // Should uppercase and limit to 4 chars
    expect(input.value).toBe("TEST");
  });
});
