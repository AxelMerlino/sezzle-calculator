import { describe, expect, it } from "vitest";
import { parseOperand, validateForm } from "./form";

describe("parseOperand", () => {
  it.each([
    ["0", 0],
    ["-0", -0],
    ["2", 2],
    ["-1.5", -1.5],
    [".5", 0.5],
    ["1.", 1],
    ["+2", 2],
    ["1e2", 100],
    ["  8  ", 8],
  ])("accepts %j", (raw, expected) => {
    expect(parseOperand(raw)).toEqual({ ok: true, value: expected });
  });

  it.each(["", "   "])("rejects empty input %j", (raw) => {
    expect(parseOperand(raw)).toEqual({ ok: false, reason: "empty" });
  });

  it.each(["abc", "1,5", "1.2.3", "--1", "1e", "Infinity", "0x10", "1e309"])(
    "rejects invalid input %j",
    (raw) => {
      expect(parseOperand(raw)).toEqual({ ok: false, reason: "invalid" });
    },
  );
});

describe("validateForm", () => {
  it("builds a binary request and preserves zero", () => {
    expect(validateForm("add", "0", "0")).toEqual({
      ok: true,
      request: { operation: "add", a: 0, b: 0 },
    });
  });

  it("builds a unary request without the second operand", () => {
    expect(validateForm("sqrt", "9", "not-a-number")).toEqual({
      ok: true,
      request: { operation: "sqrt", a: 9 },
    });
  });

  it("reports empty and invalid operands separately", () => {
    expect(validateForm("divide", " ", "1,2")).toEqual({
      ok: false,
      errors: {
        a: "Enter a number.",
        b: "Enter a valid number. Use a decimal point, not a comma.",
      },
    });
  });

  it("does not require the second operand for percent", () => {
    expect(validateForm("percent", "", "")).toEqual({
      ok: false,
      errors: { a: "Enter a number." },
    });
  });
});
