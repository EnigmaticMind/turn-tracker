export function isValidGameID(gameID: string): boolean {
  return !!gameID && gameID.length === 6 && /^[A-Z0-9]+$/.test(gameID.toUpperCase());
}

