import { useRef, useState, type FormEvent } from "react";
import { calculate, ApiError } from "./api";
import { formatExpression, operandLabel } from "./display";
import { validateForm, type FieldErrors } from "./form";
import { OPERATION_CHOICES, isBinary, isOperation, type Operation } from "./types";

type AppProps = {
  apiBase?: string;
};

type CompletedResult = {
  expression: string;
  value: number;
};

export function App({ apiBase = "" }: AppProps) {
  const [operation, setOperation] = useState<Operation>("add");
  const [a, setA] = useState("");
  const [b, setB] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [result, setResult] = useState<CompletedResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);
  const inFlight = useRef(false);

  const description = OPERATION_CHOICES.find((item) => item.value === operation)?.description;
  const binary = isBinary(operation);
  const valueLabel = operandLabel(operation);

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (inFlight.current) {
      return;
    }

    const parsed = validateForm(operation, a, b);
    if (!parsed.ok) {
      setFieldErrors(parsed.errors);
      setResult(null);
      setError(null);
      return;
    }

    inFlight.current = true;
    setPending(true);
    setFieldErrors({});
    setError(null);
    setResult(null);

    try {
      const value = await calculate(parsed.request, apiBase);
      setResult({
        expression: formatExpression(operation, a, b),
        value,
      });
    } catch (err: unknown) {
      setError(err instanceof ApiError ? err.message : "The calculation failed.");
    } finally {
      inFlight.current = false;
      setPending(false);
    }
  }

  function onClear() {
    setOperation("add");
    setA("");
    setB("");
    setFieldErrors({});
    setResult(null);
    setError(null);
  }

  function onOperationChange(value: string) {
    if (!isOperation(value)) {
      return;
    }
    setOperation(value);
    setFieldErrors({});
    setError(null);
    setResult(null);
  }

  const readoutValue = pending
    ? "Calculating…"
    : result
      ? String(result.value)
      : "No result yet.";

  return (
    <main className="app">
      <h1>Calculator</h1>
      <form aria-label="Calculation" aria-busy={pending} noValidate onSubmit={onSubmit}>
        <div className="field">
          <label htmlFor="operation">Operation</label>
          <select
            id="operation"
            value={operation}
            aria-describedby={description ? "operation-description" : undefined}
            onChange={(event) => onOperationChange(event.target.value)}
          >
            {OPERATION_CHOICES.map((item) => (
              <option key={item.value} value={item.value}>
                {item.label}
              </option>
            ))}
          </select>
          {description ? (
            <p id="operation-description" className="hint">
              {description}
            </p>
          ) : null}
        </div>

        <div className="field">
          <label htmlFor="operand-a">{valueLabel}</label>
          {/* type="text" keeps "-", ".", and partial decimals while the user types. */}
          <input
            id="operand-a"
            value={a}
            inputMode="decimal"
            autoComplete="off"
            autoCorrect="off"
            spellCheck={false}
            enterKeyHint={binary ? "next" : "go"}
            aria-invalid={fieldErrors.a ? true : undefined}
            aria-describedby={fieldErrors.a ? "operand-a-error" : undefined}
            onChange={(event) => {
              setA(event.target.value);
              setFieldErrors((current) => ({ ...current, a: undefined }));
            }}
          />
          {fieldErrors.a ? (
            <p id="operand-a-error" className="field-error">
              {fieldErrors.a}
            </p>
          ) : null}
        </div>

        {binary ? (
          <div className="field">
            <label htmlFor="operand-b">Second operand</label>
            <input
              id="operand-b"
              value={b}
              inputMode="decimal"
              autoComplete="off"
              autoCorrect="off"
              spellCheck={false}
              enterKeyHint="go"
              aria-invalid={fieldErrors.b ? true : undefined}
              aria-describedby={fieldErrors.b ? "operand-b-error" : undefined}
              onChange={(event) => {
                setB(event.target.value);
                setFieldErrors((current) => ({ ...current, b: undefined }));
              }}
            />
            {fieldErrors.b ? (
              <p id="operand-b-error" className="field-error">
                {fieldErrors.b}
              </p>
            ) : null}
          </div>
        ) : null}

        <div className="actions">
          <button type="submit" disabled={pending}>
            Calculate
          </button>
          <button type="button" onClick={onClear} disabled={pending}>
            Clear
          </button>
        </div>
      </form>

      {error ? (
        <p className="alert" role="alert">
          {error}
        </p>
      ) : null}

      <section className="readout">
        <h2>Result</h2>
        <div role="status">
          {result && !pending ? <p className="expression">{result.expression}</p> : null}
          <p className={result && !pending ? "value" : "value quiet"}>{readoutValue}</p>
        </div>
      </section>
    </main>
  );
}
