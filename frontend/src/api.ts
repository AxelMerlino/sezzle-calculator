import type { CalculateRequest } from "./types";

export class ApiError extends Error {
  readonly status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}

// apiBase is empty when the UI is served behind the Vite proxy or from the
// same origin as the API. Tests and alternate hosts pass an absolute origin.
export async function calculate(request: CalculateRequest, apiBase = ""): Promise<number> {
  const url = `${apiBase.replace(/\/$/, "")}/api/calculate`;

  let response: Response;
  try {
    response = await fetch(url, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(request),
    });
  } catch {
    throw new ApiError("Could not reach the server.", 0);
  }

  let payload: unknown;
  try {
    payload = await response.json();
  } catch {
    throw new ApiError("The server returned an unreadable response.", response.status);
  }

  if (!response.ok) {
    throw new ApiError(errorMessage(payload), response.status);
  }
  if (!isResult(payload)) {
    throw new ApiError("The server returned an unexpected response.", response.status);
  }
  return payload.result;
}

function isResult(payload: unknown): payload is { result: number } {
  return (
    typeof payload === "object" &&
    payload !== null &&
    "result" in payload &&
    typeof payload.result === "number" &&
    Number.isFinite(payload.result)
  );
}

function errorMessage(payload: unknown): string {
  if (
    typeof payload === "object" &&
    payload !== null &&
    "error" in payload &&
    typeof payload.error === "string" &&
    payload.error.trim() !== ""
  ) {
    return payload.error;
  }
  return "The calculation failed.";
}
