package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		contentType  string
		body         string
		status       int
		errorMessage string
		wantResult   *float64
	}{
		{
			name:       "add",
			body:       `{"operation":"add","a":2,"b":3}`,
			status:     http.StatusOK,
			wantResult: floatPtr(5),
		},
		{
			name:       "subtract negative result",
			body:       `{"operation":"subtract","a":3,"b":5}`,
			status:     http.StatusOK,
			wantResult: floatPtr(-2),
		},
		{
			name:       "multiply decimals",
			body:       `{"operation":"multiply","a":2.5,"b":-4}`,
			status:     http.StatusOK,
			wantResult: floatPtr(-10),
		},
		{
			name:       "divide",
			body:       `{"operation":"divide","a":-9,"b":2}`,
			status:     http.StatusOK,
			wantResult: floatPtr(-4.5),
		},
		{
			name:       "zero is a valid operand",
			body:       `{"operation":"add","a":0,"b":0}`,
			status:     http.StatusOK,
			wantResult: floatPtr(0),
		},
		{
			name:       "power",
			body:       `{"operation":"power","a":2,"b":8}`,
			status:     http.StatusOK,
			wantResult: floatPtr(256),
		},
		{
			name:       "sqrt",
			body:       `{"operation":"sqrt","a":9}`,
			status:     http.StatusOK,
			wantResult: floatPtr(3),
		},
		{
			name:       "percent",
			body:       `{"operation":"percent","a":25}`,
			status:     http.StatusOK,
			wantResult: floatPtr(0.25),
		},
		{
			name:         "division by zero",
			body:         `{"operation":"divide","a":5,"b":0}`,
			status:       http.StatusUnprocessableEntity,
			errorMessage: "division by zero",
		},
		{
			name:         "negative square root",
			body:         `{"operation":"sqrt","a":-4}`,
			status:       http.StatusUnprocessableEntity,
			errorMessage: "square root of a negative number is not defined",
		},
		{
			name:         "non finite result",
			body:         `{"operation":"multiply","a":1e308,"b":1e308}`,
			status:       http.StatusUnprocessableEntity,
			errorMessage: "result is not a finite number",
		},
		{
			name:         "non finite operand",
			body:         `{"operation":"add","a":1e309,"b":1}`,
			status:       http.StatusBadRequest,
			errorMessage: "operands must be finite numbers",
		},
		{
			name:         "unknown operation",
			body:         `{"operation":"average","a":1,"b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "unknown operation",
		},
		{
			name:         "operation omitted",
			body:         `{"a":1,"b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "operation is required",
		},
		{
			name:         "operation null",
			body:         `{"operation":null,"a":1,"b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "operation is required",
		},
		{
			name:         "operation empty",
			body:         `{"operation":"","a":1,"b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "operation is required",
		},
		{
			name:         "operation wrong type",
			body:         `{"operation":1,"a":1,"b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "operation must be a string",
		},
		{
			name:         "operand a omitted",
			body:         `{"operation":"add","b":1}`,
			status:       http.StatusBadRequest,
			errorMessage: "operand a is required",
		},
		{
			name:         "sqrt missing value",
			body:         `{"operation":"sqrt"}`,
			status:       http.StatusBadRequest,
			errorMessage: "operand a is required",
		},
		{
			name:         "operand a null",
			body:         `{"operation":"add","a":null,"b":1}`,
			status:       http.StatusBadRequest,
			errorMessage: "operand a is required",
		},
		{
			name:         "operand b omitted",
			body:         `{"operation":"divide","a":1}`,
			status:       http.StatusBadRequest,
			errorMessage: "operand b is required",
		},
		{
			name:         "operand wrong type",
			body:         `{"operation":"add","a":"1","b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "operands must be numbers",
		},
		{
			name:         "unary rejects b",
			body:         `{"operation":"sqrt","a":9,"b":0}`,
			status:       http.StatusBadRequest,
			errorMessage: "operand b is not allowed for this operation",
		},
		{
			name:         "unknown field",
			body:         `{"operation":"add","a":1,"b":2,"c":3}`,
			status:       http.StatusBadRequest,
			errorMessage: "unknown field in request body",
		},
		{
			name:         "malformed json",
			body:         `{"operation":`,
			status:       http.StatusBadRequest,
			errorMessage: "invalid JSON",
		},
		{
			name:         "empty body",
			body:         ``,
			status:       http.StatusBadRequest,
			errorMessage: "invalid JSON",
		},
		{
			name:         "trailing value",
			body:         `{"operation":"add","a":1,"b":2}{"operation":"add","a":1,"b":2}`,
			status:       http.StatusBadRequest,
			errorMessage: "invalid JSON",
		},
		{
			name:         "json array",
			body:         `[1,2]`,
			status:       http.StatusBadRequest,
			errorMessage: "invalid JSON",
		},
		{
			name:         "missing content type",
			contentType:  "omit",
			body:         `{"operation":"add","a":1,"b":2}`,
			status:       http.StatusUnsupportedMediaType,
			errorMessage: "Content-Type must be application/json",
		},
		{
			name:         "wrong content type",
			contentType:  "text/plain",
			body:         `{"operation":"add","a":1,"b":2}`,
			status:       http.StatusUnsupportedMediaType,
			errorMessage: "Content-Type must be application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType := tt.contentType
			if contentType == "" {
				contentType = "application/json"
			}
			rec := perform(t, http.MethodPost, contentType, tt.body)
			assertJSON(t, rec, tt.status, tt.errorMessage, tt.wantResult)
		})
	}
}

func TestContentTypeCharset(t *testing.T) {
	rec := perform(t, http.MethodPost, "application/json; charset=utf-8", `{"operation":"subtract","a":1,"b":0}`)
	assertJSON(t, rec, http.StatusOK, "", floatPtr(1))
}

func TestMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/calculate", nil)
	rec := httptest.NewRecorder()
	NewMux().ServeHTTP(rec, req)

	assertJSON(t, rec, http.StatusMethodNotAllowed, "method not allowed", nil)
	if rec.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("Allow = %q", rec.Header().Get("Allow"))
	}
}

func TestBodyTooLarge(t *testing.T) {
	body := `{"operation":"add","a":1,"b":` + strings.Repeat("1", maxBodyBytes) + `}`
	rec := perform(t, http.MethodPost, "application/json", body)
	assertJSON(t, rec, http.StatusBadRequest, "request body is too large", nil)
}

func TestUnknownPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	NewMux().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func perform(t *testing.T, method, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/api/calculate", strings.NewReader(body))
	if contentType != "omit" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	NewMux().ServeHTTP(rec, req)
	return rec
}

func assertJSON(t *testing.T, rec *httptest.ResponseRecorder, status int, errorMessage string, wantResult *float64) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, status, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q", got)
	}

	payload, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if errorMessage != "" {
		var parsed errorResponse
		if err := json.Unmarshal(payload, &parsed); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if parsed.Error != errorMessage {
			t.Fatalf("error = %q, want %q", parsed.Error, errorMessage)
		}
		return
	}

	var parsed resultResponse
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if wantResult == nil {
		t.Fatal("expected a result")
	}
	if parsed.Result != *wantResult {
		t.Fatalf("result = %v, want %v", parsed.Result, *wantResult)
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
