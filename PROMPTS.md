# Prompts

This file documents the prompts used as assistance during the development of the calculator, focusing on edge-case validation, test coverage, and documentation refinement.

## 1. Backend API contract and edge-case validation

I am building a calculator API in Go using `net/http` with a `POST /api/calculate` endpoint that receives `operation` and numeric operands (`a`, `b`). 

Help me structure the request validation and error handling in Go so that:
- Missing operands are clearly distinguished from a valid `0` value (e.g., using pointers in the request struct).
- Division by zero, square root of negative numbers, and non-finite results (`NaN`, `Inf`) return clean JSON error responses with appropriate HTTP status codes (`400` / `422`).
- The mathematical logic remains decoupled from the HTTP handler for unit testing.

## 2. Unit and integration test cases

Review my Go math/handler implementation and my React + TypeScript calculator component, and suggest test cases for `testing`/`httptest` (Go) and Vitest + React Testing Library (frontend). 

Focus on covering:
- Standard operations (`add`, `subtract`, `multiply`, `divide`, `power`, `sqrt`, `percent`) with negative numbers, zero, and decimals.
- Malformed JSON, unknown operations, and missing operands on the backend.
- Form submission, loading states (preventing duplicate submits), client-side empty field validation, and API/network error rendering on the frontend.

## 3. UI accessibility and responsive layout review

Review the CSS and form structure of my React calculator component to ensure it works cleanly on both desktop and mobile viewports without extra UI libraries. 

Check for:
- Proper `<label>` associations and keyboard navigation (submitting with `Enter`, clear visible `:focus-visible` states).
- Clear visual distinction between the active operation, current inputs, and the returned result so previous results don't cause confusion.

## 4. Docker setup and README consistency

I verified the multi-stage Docker build locally on Windows using Docker Desktop:
- `docker build -t sezzle-calculator .`
- `docker run --rm -p 8080:8080 sezzle-calculator`

Both completed cleanly, with the Go binary serving the compiled Vite static assets on `:8080`. Help me organize the `README.md` in English so the setup instructions, API curl examples, test/coverage commands, and manual verification results are clear and concise for a reviewer.

## 5. Final code clarity and simplicity pass

Do a final review of the repository before submission to flag any unnecessary abstractions, redundant validations, vague variable names, or repetitive comments. Keep the codebase small, idiomatic, and easy to walk through during a technical interview.