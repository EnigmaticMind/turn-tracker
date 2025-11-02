import { startLoading, stopLoading } from "../../../components/TopLoadingBar";

/**
 * Wraps an async handler function with loading state management
 */
export function withLoading<T extends (...args: any[]) => Promise<any>>(
  handler: T,
  loadingMessage?: string
): T {
  return (async (...args: Parameters<T>) => {
    startLoading(loadingMessage);
    try {
      const result = await handler(...args);
      return result;
    } finally {
      stopLoading();
    }
  }) as T;
}

/**
 * Helper to create loading-aware handlers
 */
export function createLoadingHandler<T extends (...args: any[]) => Promise<any>>(
  handler: T,
  loadingMessage: string
): T {
  return withLoading(handler, loadingMessage);
}
