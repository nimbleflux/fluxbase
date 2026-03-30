import type { AxiosError } from "axios";

export interface ApiErrorResponse {
  error?: string;
  message?: string;
}

export function getApiErrorMessage(
  error: unknown,
  fallback = "An unexpected error occurred",
): string {
  if (!error) return fallback;

  if (isAxiosError(error)) {
    const data = error.response?.data as ApiErrorResponse | undefined;
    return data?.error || data?.message || error.message || fallback;
  }

  if (error instanceof Error) {
    return error.message || fallback;
  }

  if (typeof error === "string") {
    return error || fallback;
  }

  return fallback;
}

function isAxiosError(error: unknown): error is AxiosError {
  return (
    typeof error === "object" &&
    error !== null &&
    "isAxiosError" in error &&
    (error as AxiosError).isAxiosError === true
  );
}
