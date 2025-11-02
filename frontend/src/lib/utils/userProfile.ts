const STORAGE_KEY_DISPLAY_NAME = "turnTracker_defaultDisplayName";
const STORAGE_KEY_COLOR = "turnTracker_defaultColor";

export interface UserProfile {
  displayName?: string;
  color?: string;
}

export function getDefaultProfile(): UserProfile {
  try {
    return {
      displayName: localStorage.getItem(STORAGE_KEY_DISPLAY_NAME) || undefined,
      color: localStorage.getItem(STORAGE_KEY_COLOR) || undefined,
    };
  } catch (error) {
    // localStorage might be disabled in incognito mode or blocked
    console.warn("Failed to read profile from localStorage:", error);
    return {};
  }
}

export function saveDefaultProfile(displayName?: string, color?: string): void {
  try {
    if (displayName) {
      localStorage.setItem(STORAGE_KEY_DISPLAY_NAME, displayName);
    } else {
      localStorage.removeItem(STORAGE_KEY_DISPLAY_NAME);
    }

    if (color) {
      localStorage.setItem(STORAGE_KEY_COLOR, color);
    } else {
      localStorage.removeItem(STORAGE_KEY_COLOR);
    }
  } catch (error) {
    // localStorage might be disabled in incognito mode or blocked
    console.warn("Failed to save profile to localStorage:", error);
    // Don't throw - allow app to continue without persistence
  }
}

