import { describe, expect, it } from "vitest";
import { formatExpression } from "./display";
import type { Operation } from "./types";

describe("formatExpression", () => {
  it.each<[Operation, string]>([
    ["add", "2 + 3"],
    ["subtract", "2 - 3"],
    ["multiply", "2 * 3"],
    ["divide", "2 / 3"],
    ["power", "2 ^ 3"],
    ["sqrt", "sqrt(2)"],
    ["percent", "2 / 100"],
  ])("formats %s", (operation, expected) => {
    expect(formatExpression(operation, " 2 ", " 3 ")).toBe(expected);
  });
});
