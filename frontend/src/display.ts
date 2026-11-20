import { isBinary, type Operation } from "./types";

export function formatExpression(operation: Operation, rawA: string, rawB: string): string {
  const left = rawA.trim();
  const right = rawB.trim();
  switch (operation) {
    case "add":
      return `${left} + ${right}`;
    case "subtract":
      return `${left} - ${right}`;
    case "multiply":
      return `${left} * ${right}`;
    case "divide":
      return `${left} / ${right}`;
    case "power":
      return `${left} ^ ${right}`;
    case "sqrt":
      return `sqrt(${left})`;
    case "percent":
      return `${left} / 100`;
  }
}

export function operandLabel(operation: Operation): string {
  return isBinary(operation) ? "First operand" : "Number";
}
