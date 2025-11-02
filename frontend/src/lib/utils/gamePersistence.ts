const STORAGE_KEY_CLIENT_ID = "turnTracker_clientID";

// Use sessionStorage in dev mode (per-tab), localStorage in production (shared)
function getStorage(): Storage {
  if (import.meta.env.DEV) {
    return sessionStorage; // Per-tab in development
  }
  return localStorage; // Shared in production
}

export function getPersistedClientID(): string | null {
  if (typeof window === "undefined") return null;
  try {
    return getStorage().getItem(STORAGE_KEY_CLIENT_ID);
  } catch (error) {
    console.warn("Persist: Failed to read clientID from storage:", error);
    return null;
  }
}

export function saveClientID(clientID: string): void {
  if (typeof window === "undefined") return;
  try {
    getStorage().setItem(STORAGE_KEY_CLIENT_ID, clientID);
  } catch (error) {
    console.warn("Persist: Failed to save clientID to storage:", error);
  }
}

export function clearClientID(): void {
  if (typeof window === "undefined") return;
  try {
    getStorage().removeItem(STORAGE_KEY_CLIENT_ID);
  } catch (error) {
    console.warn("Persist: Failed to clear clientID from storage:", error);
  }
}

