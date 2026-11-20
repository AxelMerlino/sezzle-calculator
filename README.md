# Calculator

A small calculator with a React interface and a Go API. The browser collects the operation and the operands, and every calculation is performed by the server.

The required operations are addition, subtraction, multiplication, and division. Power, square root, and percent are included as small extras. Percent means “divide this number by 100”: `percent(25)` is `0.25`. It does not mean “x percent of y”.

There is no database, authentication, or external service.

## Requirements

Versions used while building and testing this repository:

| Tool | Version |
| --- | --- |
| Go | 1.27.4 |
| Node.js | 24.19.0 |
| npm | 11.17.0 |
| React | 19.3.1 |
| Vite | 8.3.8 |
| TypeScript | 7.0.2 |
| Vitest | 5.1.1 |

Go uses only the standard library. Frontend dependencies are declared in `frontend/package.json` and locked in `frontend/package-lock.json`.

## Run it

From `backend`:

```powershell
go run ./cmd/server

```

From `frontend`, in another terminal:

```powershell
npm install
npm run dev

```

Open http://127.0.0.1:5173/. The API listens on http://127.0.0.1:8080/.

## Ports and configuration

| Setting | Default | Purpose |
| --- | --- | --- |
| `PORT` | `8080` | Port for the Go server. `9090` and `:9090` are both accepted. |
| Vite | `5173` | Development UI. Requests to `/api` are proxied to `http://127.0.0.1:8080`. |
| `STATIC_DIR` | unset | When set, the Go server also serves that directory. Docker uses this so one port hosts the UI and the API. |

No CORS middleware is enabled. In development the Vite proxy keeps the browser on one origin. In Docker the same process serves both.

## API

`POST /api/calculate`

`Content-Type` must be `application/json`. The body is one JSON object:

| Field | Required | Meaning |
| --- | --- | --- |
| `operation` | always | One of the names below. |
| `a` | always | First operand, or the only operand for a unary operation. |
| `b` | binary operations only | Second operand. Sending it with a unary operation is rejected. |

Operation names:

| Name | Calculation |
| --- | --- |
| `add` | `a + b` |
| `subtract` | `a - b` |
| `multiply` | `a * b` |
| `divide` | `a / b` |
| `power` | `a` raised to `b` |
| `sqrt` | square root of `a` |
| `percent` | `a / 100` |

An omitted field and JSON `null` are missing. A numeric `0` is a real operand. Unknown JSON fields are rejected.

Success:

```json
{ "result": 5 }

```

Failure:

```json
{ "error": "division by zero" }

```

### Examples

Addition:

```powershell
curl.exe -s -X POST http://127.0.0.1:8080/api/calculate -H "Content-Type: application/json" -d "{\"operation\":\"add\",\"a\":2,\"b\":3}"
```

```json
{ "result": 5 }

```

Zero is valid:

```json
{ "operation": "add", "a": 0, "b": 0 }

```

```json
{ "result": 0 }

```

Square root and percent:

```json
{ "operation": "sqrt", "a": 9 }

```

```json
{ "result": 3 }

```

```json
{ "operation": "percent", "a": 25 }

```

```json
{ "result": 0.25 }

```

Errors:

| Request | Status | Body |
| --- | --- | --- |
| `{"operation":"divide","a":1,"b":0}` | 422 | `{"error":"division by zero"}` |
| `{"operation":"sqrt","a":-4}` | 422 | `{"error":"square root of a negative number is not defined"}` |
| `{"operation":"multiply","a":1e308,"b":1e308}` | 422 | `{"error":"result is not a finite number"}` |
| `{"operation":"average","a":1,"b":2}` | 400 | `{"error":"unknown operation"}` |
| `{"operation":"add","b":1}` or `"a": null` | 400 | `{"error":"operand a is required"}` |
| `{"operation":"add","a":"1","b":2}` | 400 | `{"error":"operands must be numbers"}` |
| `{"operation":"add","a":1e309,"b":1}` | 400 | `{"error":"operands must be finite numbers"}` |
| `{"operation":"sqrt","a":9,"b":0}` | 400 | `{"error":"operand b is not allowed for this operation"}` |
| malformed JSON, empty body, or a second JSON value | 400 | `{"error":"invalid JSON"}` |
| missing or non-JSON `Content-Type` | 415 | `{"error":"Content-Type must be application/json"}` |
| `GET /api/calculate` | 405 | `{"error":"method not allowed"}` and `Allow: POST` |

400 means the request does not match the contract. 422 means the JSON is valid but the calculation is not defined or is not finite. Responses are JSON either way. A path other than `/api/calculate` gets the standard library’s plain-text 404, unless `STATIC_DIR` is set.

## Tests, coverage, and build

Backend, from `backend`:

```powershell
go test ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go build -o bin/server.exe ./cmd/server

```

Frontend, from `frontend`:

```powershell
npm test
npm run test:coverage
npm run test:integration
npm run build

```

`npm test` is the unit suite and mocks the network. `npm run test:integration` starts the Go API and drives the React form against it. `npm run build` typechecks with `tsc --noEmit` and then builds the production assets.

### Verification results

Commands that passed:

* `go test ./internal/calc`, `./internal/httpapi`, and `./cmd/server`
* `npm run test:coverage` — 4 files, 47 tests
* `npm run test:integration` — 3 tests against the live API
* `npx tsc --noEmit` and `npm run build`
* `go run ./cmd/server`, then the Vite app on port 5173

Backend coverage, measured with `go test -coverprofile` and `go tool cover -func`:

| Package | Statements |
| --- | --- |
| `calculator/internal/calc` | 98.1% (52/53) |
| `calculator/internal/httpapi` | 87.8% (108/123) |
| `calculator/cmd/server` | 58.8% (10/17) |
| Combined | 88.1% (170/193) |

The uncovered Go statement in `calc` is the default branch of `applyBinary`. `Evaluate` never calls it with an unknown operation. The HTTP gaps are defensive branches: an unexpected internal error, a failed JSON write, and a few decode paths that normal requests do not hit. `main` is not executed by unit tests; `newServer` and `listenAddr` are.

Frontend coverage from `npm run test:coverage`:

|  | Statements | Branches | Functions | Lines |
| --- | --- | --- | --- | --- |
| All included source | 100% (117/117) | 98.87% (88/89) | 100% (23/23) | 100% (113/113) |

`src/main.tsx` is excluded as the bootstrap. The uncovered branch is the fallback message used when `calculate` throws something other than `ApiError`; the client only throws `ApiError`. The integration test renders the form, multiplies 6 by 7 through the Go server, shows a real division-by-zero error, and evaluates `percent(25)`.

Checked against the running dev servers:

* `GET /` on port 5173 returned the app.
* `POST /api/calculate` through the Vite proxy returned `{"result":0.5}` for `-1.5 + 2`, `{"result":3}` for `sqrt(9)`, and `{"result":256}` for `2 ^ 8`.
* `POST` directly to port 8080 returned `{"result":0.25}` for percent and HTTP 422 `{"error":"division by zero"}`.
* Malformed JSON returned HTTP 400 `{"error":"invalid JSON"}`.

Keyboard submission and focusable controls are covered by the interface tests. A manual check of the layout on a narrow screen was not recorded.

## Design decisions

Arithmetic lives in `internal/calc`. `internal/httpapi` only decodes JSON, chooses a status code, and writes the response. The process entrypoint is `cmd/server`.

Operands are pointers. A missing JSON field and `null` stay `nil`, while `0` is a pointer to zero, so the two cases do not collapse.

Numbers are IEEE-754 binary64, the same representation `encoding/json` uses. That matches a general calculator. It is the wrong type for money: `0.1 + 0.2` is not exactly `0.3`.

Unary operations reject `b` instead of ignoring it, so a stale second operand cannot change the request quietly. The form hides that field and does not send it.

The form uses text inputs with `inputMode="decimal"`. `type="number"` changes with the locale and blocks intermediate values such as `-` and `1.`. The decimal separator is a point. A comma is invalid. A short note appears only when the operand order matters, or to state that percent divides the number by 100.

The displayed number is the value returned by the API, next to the expression that was submitted. A new submit clears the previous result before the response arrives. A result of `0` is shown as a result, not as the empty state.

The dev UI calls a relative `/api/calculate`. An optional `apiBase` prop exists so a test, or another host, can point at a concrete origin.

`power` uses `math.Pow`. `0^0` is `1`, which is that function’s result. A negative base with a fractional exponent, and overflow such as `0^-1`, are rejected because the result is not finite.

## Layout

```text
backend/cmd/server          process entrypoint and address parsing
backend/internal/calc       operations
backend/internal/httpapi    JSON API and optional static files
frontend/src/App.tsx        form, loading, errors, result
frontend/src/api.ts         fetch client
frontend/src/form.ts        empty and invalid operands
frontend/src/integration    React form against the real API

```

## Docker

The Dockerfile compiles the UI, compiles a static Go binary, and serves both on port 8080. Verified locally on Windows with Docker Desktop:

```powershell
docker build -t sezzle-calculator .
docker run --rm -p 8080:8080 sezzle-calculator

```

The build ended with `FINISHED`. The container logged:

```text
serving files from /app/static
listening on :8080

```

Open http://localhost:8080. Stop the container with Ctrl+C.

When `STATIC_DIR` is set, `/` and other extensionless paths return `index.html`. A missing asset, such as `/assets/missing.js`, stays a 404. `/api/calculate` is unchanged.

### Manual checks of the running calculator

Manual verification of core operations and empty-field validation in the browser:

| Check | Result |
| --- | --- |
| 2 + 3 | 5 |
| 5 − 8 | −3 |
| 6 × 7 | 42 |
| 9 ÷ 2 | 4.5 |
| 2 raised to the power of 8 | 256 |
| Square root of 9 | 3 |
| Percent of 25 | 0.25 |
| Division by zero | An error message |
| Square root of a negative number | An error message |
| Empty fields | Validation asking for numbers to be entered |

## Limitations and next steps

* Binary64 cannot represent every decimal. Do not use this API for currency without changing the numeric type.
* `0^0` follows `math.Pow` and returns `1`.
* Negative bases with fractional exponents are rejected when the result is not finite.
* There is no history, keypad, or thousands separator.
* Unknown URLs are plain-text 404s unless static files are enabled.
* A short history of submitted expressions would be the useful UI follow-up. Decimal or integer minor units would be required before using this for payments.

## Use of AI

AI assistance (Cursor) was used during development to validate edge-case handling, expand unit test coverage, review UI accessibility details, and refine documentation. The key prompts used are documented in `PROMPTS.md`.