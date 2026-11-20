import { isBinary, type CalculateRequest, type Operation } from "./types";

const NUMBER_PATTERN = /^[+-]?(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?$/;

export type FieldErrors = {
  a?: string;
  b?: string;
};

export type ValidationResult =
  | { ok: true; request: CalculateRequest }
  | { ok: false; errors: FieldErrors };

type ParsedNumber = { ok: true; value: number } | { ok: false; reason: "empty" | "invalid" };

// Accepts a plain decimal form, including a leading sign and an exponent.
// Commas are rejected so an empty field, a valid zero, and a locale-formatted
// number stay distinct. The check is intentionally stricter than Number().
export function parseOperand(raw: string): ParsedNumber {
  const trimmed = raw.trim();
  if (trimmed === "") {
    return { ok: false, reason: "empty" };
  }
  if (!NUMBER_PATTERN.test(trimmed)) {
    return { ok: false, reason: "invalid" };
  }
  const value = Number(trimmed);
  if (!Number.isFinite(value)) {
    return { ok: false, reason: "invalid" };
  }
  return { ok: true, value };
}

export function validateForm(operation: Operation, rawA: string, rawB: string): ValidationResult {
  const errors: FieldErrors = {};
  const a = parseOperand(rawA);
  if (!a.ok) {
    errors.a = operandMessage(a.reason);
  }

  if (!isBinary(operation)) {
    if (!a.ok) {
      return { ok: false, errors };
    }
    return { ok: true, request: { operation, a: a.value } };
  }

  const b = parseOperand(rawB);
  if (!b.ok) {
    errors.b = operandMessage(b.reason);
  }
  if (!a.ok || !b.ok) {
    return { ok: false, errors };
  }
  return { ok: true, request: { operation, a: a.value, b: b.value } };
}

function operandMessage(reason: "empty" | "invalid"): string {
  if (reason === "empty") {
    return "Enter a number.";
  }
  return "Enter a valid number. Use a decimal point, not a comma.";
}
