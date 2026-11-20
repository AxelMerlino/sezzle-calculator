import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("App", () => {
  it("sends both operands and shows the result returned by the API", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse({ result: 123 }));
    render(<App />);

    await user.selectOptions(screen.getByLabelText("Operation"), "multiply");
    await user.type(screen.getByLabelText("First operand"), "-2.5");
    await user.type(screen.getByLabelText("Second operand"), "4");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/calculate",
      expect.objectContaining({
        body: JSON.stringify({ operation: "multiply", a: -2.5, b: 4 }),
      }),
    );
    const status = await screen.findByRole("status");
    expect(status).toHaveTextContent("-2.5 * 4");
    expect(resultValue()).toHaveTextContent("123");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows a valid zero instead of the empty state", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ result: 0 }));
    render(<App />);

    await user.type(screen.getByLabelText("First operand"), "0");
    await user.type(screen.getByLabelText("Second operand"), "0");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(await screen.findByRole("status")).toHaveTextContent("0 + 0");
    expect(resultValue()).toHaveTextContent("0");
  });

  it("submits when Enter is pressed in a field", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ result: 5 }));
    render(<App />);

    await user.type(screen.getByLabelText("First operand"), "2");
    await user.type(screen.getByLabelText("Second operand"), "3{Enter}");

    expect(await screen.findByRole("status")).toHaveTextContent("2 + 3");
    expect(resultValue()).toHaveTextContent("5");
  });

  it("does not call the API when a field is empty or invalid", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse({ result: 1 }));
    render(<App />);

    await user.click(screen.getByRole("button", { name: "Calculate" }));
    expect(screen.getByLabelText("First operand")).toHaveAttribute("aria-invalid", "true");
    expect(screen.getAllByText("Enter a number.")).toHaveLength(2);

    await user.type(screen.getByLabelText("First operand"), "1,5");
    await user.type(screen.getByLabelText("Second operand"), "2");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(screen.getByText("Enter a valid number. Use a decimal point, not a comma.")).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("shows an API error and drops the previous result", async () => {
    const user = userEvent.setup();
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ result: 99 }))
      .mockResolvedValueOnce(jsonResponse({ error: "division by zero" }, 422));
    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    await user.selectOptions(screen.getByLabelText("Operation"), "divide");
    await user.type(screen.getByLabelText("First operand"), "8");
    await user.type(screen.getByLabelText("Second operand"), "2");
    await user.click(screen.getByRole("button", { name: "Calculate" }));
    expect(await screen.findByRole("status")).toHaveTextContent("8 / 2");
    expect(resultValue()).toHaveTextContent("99");

    await user.clear(screen.getByLabelText("Second operand"));
    await user.type(screen.getByLabelText("Second operand"), "0");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("division by zero");
    expect(resultValue()).toHaveTextContent("No result yet.");
    expect(screen.getByRole("status")).not.toHaveTextContent("99");
  });

  it("shows a connection failure", async () => {
    const user = userEvent.setup();
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));
    render(<App />);

    await user.type(screen.getByLabelText("First operand"), "1");
    await user.type(screen.getByLabelText("Second operand"), "1");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("Could not reach the server.");
  });

  it("disables the actions and ignores a second submit while a request is in flight", async () => {
    const user = userEvent.setup();
    let resolveFetch: (response: Response) => void = () => {};
    const fetchMock = vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          resolveFetch = resolve;
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    render(<App />);

    await user.type(screen.getByLabelText("First operand"), "2");
    await user.type(screen.getByLabelText("Second operand"), "2");
    fireEvent.submit(screen.getByRole("form", { name: "Calculation" }));
    fireEvent.submit(screen.getByRole("form", { name: "Calculation" }));

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", { name: "Calculate" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Clear" })).toBeDisabled();
    expect(screen.getByRole("status")).toHaveTextContent("Calculating…");

    resolveFetch(jsonResponse({ result: 4 }));
    await waitFor(() => {
      expect(resultValue()).toHaveTextContent("4");
    });
    expect(screen.getByRole("button", { name: "Calculate" })).toBeEnabled();
  });

  it("sends only the value for a square root", async () => {
    const user = userEvent.setup();
    const fetchMock = mockFetch(jsonResponse({ result: 3 }));
    render(<App />);

    await user.selectOptions(screen.getByLabelText("Operation"), "sqrt");
    expect(screen.queryByLabelText("Second operand")).not.toBeInTheDocument();
    expect(screen.getByLabelText("Operation")).not.toHaveAttribute("aria-describedby");

    await user.type(screen.getByLabelText("Number"), "9");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/calculate",
      expect.objectContaining({
        body: JSON.stringify({ operation: "sqrt", a: 9 }),
      }),
    );
    expect(await screen.findByText("sqrt(9)")).toBeInTheDocument();
    expect(resultValue()).toHaveTextContent("3");
  });

  it("ignores an unknown operation value", () => {
    render(<App />);
    fireEvent.change(screen.getByLabelText("Operation"), { target: { value: "average" } });
    expect(screen.getByLabelText("Operation")).toHaveValue("add");
  });

  it("clears the operation, fields, errors, and result", async () => {
    const user = userEvent.setup();
    mockFetch(jsonResponse({ result: 5 }));
    render(<App />);

    await user.type(screen.getByLabelText("First operand"), "2");
    await user.type(screen.getByLabelText("Second operand"), "3{Enter}");
    expect(await screen.findByRole("status")).toHaveTextContent("2 + 3");
    expect(resultValue()).toHaveTextContent("5");

    await user.click(screen.getByRole("button", { name: "Clear" }));

    expect(screen.getByLabelText("Operation")).toHaveValue("add");
    expect(screen.getByLabelText("First operand")).toHaveValue("");
    expect(screen.getByLabelText("Second operand")).toHaveValue("");
    expect(screen.getByRole("status")).toHaveTextContent("No result yet.");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});

function resultValue(): HTMLElement {
  const value = screen.getByRole("status").querySelector(".value");
  if (!(value instanceof HTMLElement)) {
    throw new Error("Result value is missing");
  }
  return value;
}

function mockFetch(response: Response) {
  const fetchMock = vi.fn().mockResolvedValue(response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
