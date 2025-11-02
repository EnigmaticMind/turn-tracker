import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import Options from "../Options";
import { createMockWebSocketManager } from "./testUtils";

// Mock dependencies
vi.mock("../../lib/utils/userProfile", () => ({
  saveDefaultProfile: vi.fn(),
  getDefaultProfile: vi.fn(() => ({
    displayName: "",
    color: "#3498db",
  })),
}));

vi.mock("../../lib/websocket/handlers/updateProfile", () => ({
  updateProfile: vi.fn().mockResolvedValue(undefined),
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

describe("Options", () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should render settings form", () => {
    render(<Options onClose={mockOnClose} />);

    expect(screen.getByText("Settings")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Display name")).toBeInTheDocument();
    expect(screen.getByText("Save")).toBeInTheDocument();
    expect(screen.getByText("Cancel")).toBeInTheDocument();
  });

  it("should call onClose when cancel button is clicked", () => {
    render(<Options onClose={mockOnClose} />);

    const cancelButton = screen.getByText("Cancel");
    fireEvent.click(cancelButton);

    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it("should call onClose when save button is clicked", async () => {
    const { saveDefaultProfile } = await import("../../lib/utils/userProfile");
    render(<Options onClose={mockOnClose} />);

    const saveButton = screen.getByText("Save");
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(saveDefaultProfile).toHaveBeenCalled();
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });

  it("should update display name input", () => {
    render(<Options onClose={mockOnClose} />);

    const nameInput = screen.getByPlaceholderText(
      "Display name"
    ) as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: "Test Player" } });

    expect(nameInput.value).toBe("Test Player");
  });

  it("should update color picker", () => {
    const { container } = render(<Options onClose={mockOnClose} />);

    const colorInput = container.querySelector(
      'input[type="color"]'
    ) as HTMLInputElement;
    expect(colorInput).toBeInTheDocument();

    fireEvent.change(colorInput, { target: { value: "#FF5733" } });

    expect(colorInput.value.toUpperCase()).toBe("#FF5733");
  });

  it("should save profile to localStorage and update server when ws is provided", async () => {
    const { saveDefaultProfile } = await import("../../lib/utils/userProfile");
    const { updateProfile } = await import(
      "../../lib/websocket/handlers/updateProfile"
    );
    const mockWS = createMockWebSocketManager({ gameID: "TEST123" });

    render(<Options onClose={mockOnClose} ws={mockWS as any} />);

    const nameInput = screen.getByPlaceholderText("Display name");
    fireEvent.change(nameInput, { target: { value: "Test Player" } });

    const saveButton = screen.getByText("Save");
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(saveDefaultProfile).toHaveBeenCalledWith(
        "Test Player",
        expect.any(String)
      );
      expect(updateProfile).toHaveBeenCalledWith(
        mockWS,
        "Test Player",
        expect.any(String)
      );
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it("should not update server when ws is not provided", async () => {
    const { updateProfile } = await import(
      "../../lib/websocket/handlers/updateProfile"
    );
    render(<Options onClose={mockOnClose} />);

    const saveButton = screen.getByText("Save");
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(updateProfile).not.toHaveBeenCalled();
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it("should not update server when ws has no gameID", async () => {
    const { updateProfile } = await import(
      "../../lib/websocket/handlers/updateProfile"
    );
    const mockWS = createMockWebSocketManager({ gameID: undefined });

    render(<Options onClose={mockOnClose} ws={mockWS as any} />);

    const saveButton = screen.getByText("Save");
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(updateProfile).not.toHaveBeenCalled();
      expect(mockOnClose).toHaveBeenCalled();
    });
  });

  it("should load default profile from localStorage on mount", async () => {
    const { getDefaultProfile } = await import("../../lib/utils/userProfile");
    vi.mocked(getDefaultProfile).mockReturnValueOnce({
      displayName: "Saved Name",
      color: "#FF5733",
    });

    render(<Options onClose={mockOnClose} />);

    await waitFor(() => {
      const nameInput = screen.getByPlaceholderText(
        "Display name"
      ) as HTMLInputElement;
      expect(nameInput.value).toBe("Saved Name");
    });
  });

  it("should handle updateProfile errors gracefully", async () => {
    const { updateProfile } = await import(
      "../../lib/websocket/handlers/updateProfile"
    );
    const { toast } = await import("../ToastProvider");
    vi.mocked(updateProfile).mockRejectedValueOnce(new Error("Network error"));
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});

    const mockWS = createMockWebSocketManager({ gameID: "TEST123" });
    render(<Options onClose={mockOnClose} ws={mockWS as any} />);

    const saveButton = screen.getByText("Save");
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Options: ",
        "Network error"
      );
      expect(toast.error).toHaveBeenCalledWith(
        "Problem updating profile: Network error"
      );
      expect(mockOnClose).toHaveBeenCalledTimes(1); // Closes so user isn't stuck
    });

    consoleErrorSpy.mockRestore();
  });
});
