const STORAGE_KEY_DISPLAY_NAME = "turnTracker_defaultDisplayName";
const STORAGE_KEY_COLOR = "turnTracker_defaultColor";

export interface UserProfile {
  displayName?: string;
  color?: string;
}

export function getDefaultProfile(): UserProfile {
  return {
    displayName: localStorage.getItem(STORAGE_KEY_DISPLAY_NAME) || undefined,
    color: localStorage.getItem(STORAGE_KEY_COLOR) || undefined,
  };
}

export function saveDefaultProfile(displayName?: string, color?: string): void {
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
}

