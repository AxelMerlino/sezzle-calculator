import { spawn, type ChildProcess } from "node:child_process";
import { createServer } from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";
import { App } from "../App";

const backendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../../backend");

let server: ChildProcess | undefined;
let apiBase = "";
let serverLogs = "";

beforeAll(async () => {
  const port = await freePort();
  server = spawn(
    process.env.GO_BIN ?? "go",
    ["test", "-count=1", "-timeout", "120s", "-run", "^TestHoldOpen$", "./cmd/server"],
    {
      cwd: backendDir,
      env: { ...process.env, HOLD_SERVER: "1", PORT: String(port) },
      stdio: ["ignore", "pipe", "pipe"],
    },
  );
  server.stdout?.on("data", (chunk: Buffer) => {
    serverLogs += chunk.toString();
  });
  server.stderr?.on("data", (chunk: Buffer) => {
    serverLogs += chunk.toString();
  });
  server.on("error", (error) => {
    serverLogs += `${error.message}\n`;
  });
  apiBase = `http://127.0.0.1:${port}`;
  try {
    await waitForServer(`${apiBase}/api/calculate`, server);
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : "server did not start";
    throw new Error(`${message}\n${serverLogs}`);
  }
}, 60000);

afterAll(() => {
  stopServer(server);
});

afterEach(() => {
  cleanup();
});

describe("frontend and backend", () => {
  it("shows a result produced by the Go API", async () => {
    const user = userEvent.setup();
    render(<App apiBase={apiBase} />);

    await user.selectOptions(screen.getByLabelText("Operation"), "multiply");
    await user.type(screen.getByLabelText("First operand"), "6");
    await user.type(screen.getByLabelText("Second operand"), "7");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(await screen.findByText("42")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("6 * 7");
  });

  it("shows a division by zero error from the Go API", async () => {
    const user = userEvent.setup();
    render(<App apiBase={apiBase} />);

    await user.selectOptions(screen.getByLabelText("Operation"), "divide");
    await user.type(screen.getByLabelText("First operand"), "1");
    await user.type(screen.getByLabelText("Second operand"), "0");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("division by zero");
    expect(screen.getByRole("status")).toHaveTextContent("No result yet.");
  });

  it("evaluates a unary percent through the Go API", async () => {
    const user = userEvent.setup();
    render(<App apiBase={apiBase} />);

    await user.selectOptions(screen.getByLabelText("Operation"), "percent");
    await user.type(screen.getByLabelText("Number"), "25");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(await screen.findByText("0.25")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("25 / 100");
  });
});

function stopServer(child: ChildProcess | undefined): void {
  if (!child || child.pid === undefined || child.exitCode !== null) {
    return;
  }
  if (process.platform === "win32") {
    spawn("taskkill", ["/PID", String(child.pid), "/T", "/F"], { stdio: "ignore" });
    return;
  }
  child.kill();
}

function freePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const probe = createServer();
    probe.once("error", reject);
    probe.listen(0, "127.0.0.1", () => {
      const address = probe.address();
      if (address === null || typeof address === "string") {
        probe.close();
        reject(new Error("Could not allocate a port"));
        return;
      }
      const { port } = address;
      probe.close((error) => (error ? reject(error) : resolve(port)));
    });
  });
}

async function waitForServer(url: string, child: ChildProcess): Promise<void> {
  const deadline = Date.now() + 45000;
  let lastError = "no response";
  while (Date.now() < deadline) {
    if (child.exitCode !== null) {
      throw new Error(`server exited with code ${child.exitCode}`);
    }
    try {
      const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: "{}",
      });
      if (response.status === 400) {
        return;
      }
      lastError = `status ${response.status}`;
    } catch (error: unknown) {
      lastError = error instanceof Error ? error.message : "connection failed";
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error(`server did not start: ${lastError}`);
}
