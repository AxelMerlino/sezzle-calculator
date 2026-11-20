export const BINARY_OPERATIONS = ["add", "subtract", "multiply", "divide", "power"] as const;
export const UNARY_OPERATIONS = ["sqrt", "percent"] as const;

export type BinaryOperation = (typeof BINARY_OPERATIONS)[number];
export type UnaryOperation = (typeof UNARY_OPERATIONS)[number];
export type Operation = BinaryOperation | UnaryOperation;

export type CalculateRequest =
  | { operation: BinaryOperation; a: number; b: number }
  | { operation: UnaryOperation; a: number };

export type OperationChoice = {
  value: Operation;
  label: string;
  description?: string;
};

// Descriptions are only for rules the label does not already state:
// operand order, or the meaning of percent.
export const OPERATION_CHOICES: readonly OperationChoice[] = [
  { value: "add", label: "Add" },
  {
    value: "subtract",
    label: "Subtract",
    description: "Subtracts the second operand from the first.",
  },
  { value: "multiply", label: "Multiply" },
  {
    value: "divide",
    label: "Divide",
    description: "Divides the first operand by the second.",
  },
  {
    value: "power",
    label: "Power",
    description: "Raises the first operand to the power of the second.",
  },
  { value: "sqrt", label: "Square root" },
  {
    value: "percent",
    label: "Percent",
    description: "Divides the number by 100.",
  },
];

export function isOperation(value: string): value is Operation {
  return OPERATION_CHOICES.some((choice) => choice.value === value);
}

export function isBinary(operation: Operation): operation is BinaryOperation {
  return (BINARY_OPERATIONS as readonly string[]).includes(operation);
}
