// Package httpapi exposes the calculator over HTTP.
// Mathematical rules live in calc; this package only decodes the request,
// chooses a status code, and writes JSON.
package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"mime"
	"net/http"
	"strings"

	"calculator/internal/calc"
)

const maxBodyBytes = 1 << 20

var (
	errInvalidJSON       = errors.New("invalid JSON")
	errUnknownField      = errors.New("unknown field in request body")
	errTooLarge          = errors.New("request body is too large")
	errContentType       = errors.New("Content-Type must be application/json")
	errOperationRequired = errors.New("operation is required")
	errOperationType     = errors.New("operation must be a string")
	errOperandType       = errors.New("operands must be numbers")
)

type requestBody struct {
	Operation json.RawMessage `json:"operation"`
	A         json.RawMessage `json:"a"`
	B         json.RawMessage `json:"b"`
}

type resultResponse struct {
	Result float64 `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewMux returns the API routes. POST /api/calculate evaluates an operation.
// Any other method on that path is rejected with 405.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/calculate", handleCalculate)
	mux.HandleFunc("/api/calculate", handleMethodNotAllowed)
	return mux
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	operation, a, b, err := decodeRequest(r)
	if err != nil {
		writeJSON(w, statusForDecode(err), errorResponse{Error: err.Error()})
		return
	}

	result, err := calc.Evaluate(operation, a, b)
	if err != nil {
		status := statusForCalc(err)
		message := err.Error()
		if status == http.StatusInternalServerError {
			log.Printf("calculate: %v", err)
			message = "internal error"
		}
		writeJSON(w, status, errorResponse{Error: message})
		return
	}

	writeJSON(w, http.StatusOK, resultResponse{Result: result})
}

func handleMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", http.MethodPost)
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
}

func decodeRequest(r *http.Request) (string, *float64, *float64, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return "", nil, nil, errContentType
	}

	// Buffer the body with a hard limit so a truncated JSON token and an
	// oversized payload can be told apart. One megabyte is far above any
	// legitimate calculation request.
	limited := http.MaxBytesReader(nil, r.Body, maxBodyBytes)
	payload, err := io.ReadAll(limited)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return "", nil, nil, errTooLarge
		}
		return "", nil, nil, errInvalidJSON
	}

	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()

	var body requestBody
	if err := decoder.Decode(&body); err != nil {
		return "", nil, nil, classifyJSONError(err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return "", nil, nil, errInvalidJSON
	}

	operation, err := parseOperation(body.Operation)
	if err != nil {
		return "", nil, nil, err
	}
	a, err := parseOperand(body.A)
	if err != nil {
		return "", nil, nil, err
	}
	b, err := parseOperand(body.B)
	if err != nil {
		return "", nil, nil, err
	}
	return operation, a, b, nil
}

func parseOperation(raw json.RawMessage) (string, error) {
	if isAbsent(raw) {
		return "", errOperationRequired
	}
	var operation string
	if err := json.Unmarshal(raw, &operation); err != nil {
		return "", errOperationType
	}
	if operation == "" {
		return "", errOperationRequired
	}
	return operation, nil
}

// parseOperand returns nil for an omitted or null field. A JSON number,
// including zero, becomes a pointer to that number. Strings and other JSON
// types are rejected, so a missing operand is not confused with zero.
func parseOperand(raw json.RawMessage) (*float64, error) {
	if isAbsent(raw) {
		return nil, nil
	}
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		// A numeric token that does not fit in binary64 is a number, but not
		// a usable one. Other JSON types are rejected as the wrong type.
		if isJSONNumber(raw) {
			return nil, calc.ErrNonFiniteOperand
		}
		return nil, errOperandType
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, calc.ErrNonFiniteOperand
	}
	return &value, nil
}

func isJSONNumber(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return false
	}
	switch trimmed[0] {
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}

func isAbsent(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func classifyJSONError(err error) error {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return errInvalidJSON
	}
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		return errInvalidJSON
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return errInvalidJSON
	}
	if strings.HasPrefix(err.Error(), "json: unknown field") {
		return errUnknownField
	}
	return errInvalidJSON
}

func statusForDecode(err error) int {
	if errors.Is(err, errContentType) {
		return http.StatusUnsupportedMediaType
	}
	return http.StatusBadRequest
}

func statusForCalc(err error) int {
	switch {
	case errors.Is(err, calc.ErrUnknownOperation),
		errors.Is(err, calc.ErrMissingOperand),
		errors.Is(err, calc.ErrUnexpectedOperand),
		errors.Is(err, calc.ErrNonFiniteOperand):
		return http.StatusBadRequest
	case errors.Is(err, calc.ErrDivisionByZero),
		errors.Is(err, calc.ErrNegativeSqrt),
		errors.Is(err, calc.ErrNonFiniteResult):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write response: %v", err)
	}
}
