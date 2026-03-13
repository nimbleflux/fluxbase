/**
 * Standard response type for Fluxbase SDK methods
 * @category Common
 */
export interface FluxbaseResponse<T> {
  /** The response data, or null if there was an error */
  data: T | null;
  /** The error, or null if successful */
  error: Error | null;
}
