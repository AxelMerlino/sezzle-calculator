import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, calculate } from "./api";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("calculate", () => {
  it("posts the operation and returns the API result", async () => {
    const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ result: -4.5 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(calculate({ operation: "divide", a: -9, b: 2 }, "http://api.test/")).resolves.toBe(
      -4.5,
    );

    expect(fetchMock).toHaveBeenCalledWith(
      "http://api.test/api/calculate",
      expect.objectContaining({
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ operation: "divide", a: -9, b: 2 }),
      }),
    );
  });

  it("returns a valid zero", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ result: 0 })));
    await expect(calculate({ operation: "add", a: 0, b: 0 })).resolves.toBe(0);
  });

  it("uses the server error message and status", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(jsonResponse({ error: "division by zero" }, 422)),
    );

    await expect(calculate({ operation: "divide", a: 1, b: 0 })).rejects.toMatchObject({
      name: "ApiError",
      message: "division by zero",
      status: 422,
    });
  });

  it("falls back when the error payload has no message", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(jsonResponse({ error: " " }, 400)));

    await expect(calculate({ operation: "add", a: 1, b: 2 })).rejects.toThrow(
      "The calculation failed.",
    );
  });

  it("rejects an unreadable response", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(() => Promise.resolve(new Response("nope", { status: 502 }))),
    );

    const attempt = calculate({ operation: "add", a: 1, b: 2 });
    await expect(attempt).rejects.toBeInstanceOf(ApiError);
    await expect(attempt).rejects.toThrow("The server returned an unreadable response.");
  });

  it("rejects a success payload that is not a finite number", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(() =>
        Promise.resolve(new Response('{"result":1e309}', { status: 200 })),
      ),
    );

    await expect(calculate({ operation: "add", a: 1, b: 2 })).rejects.toThrow(
      "The server returned an unexpected response.",
    );
  });

  it("reports a connection failure", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));

    await expect(calculate({ operation: "add", a: 1, b: 2 })).rejects.toMatchObject({
      message: "Could not reach the server.",
      status: 0,
    });
  });
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
