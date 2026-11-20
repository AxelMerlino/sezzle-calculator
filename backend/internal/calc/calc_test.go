package calc

import (
	"errors"
	"math"
	"testing"
)

func num(value float64) *float64 {
	return &value
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		a         *float64
		b         *float64
		want      float64
		wantErr   error
	}{
		{name: "add", operation: OpAdd, a: num(2), b: num(3), want: 5},
		{name: "add zero", operation: OpAdd, a: num(0), b: num(0), want: 0},
		{name: "add negative and decimal", operation: OpAdd, a: num(-1.5), b: num(2), want: 0.5},

		{name: "subtract", operation: OpSubtract, a: num(5), b: num(3), want: 2},
		{name: "subtract to negative", operation: OpSubtract, a: num(3), b: num(5), want: -2},
		{name: "subtract decimals", operation: OpSubtract, a: num(0.5), b: num(1.5), want: -1},

		{name: "multiply", operation: OpMultiply, a: num(2.5), b: num(4), want: 10},
		{name: "multiply negatives", operation: OpMultiply, a: num(-2), b: num(-4), want: 8},
		{name: "multiply by zero", operation: OpMultiply, a: num(5), b: num(0), want: 0},

		{name: "divide", operation: OpDivide, a: num(5), b: num(2), want: 2.5},
		{name: "divide negative", operation: OpDivide, a: num(-9), b: num(2), want: -4.5},
		{name: "divide zero by number", operation: OpDivide, a: num(0), b: num(5), want: 0},
		{name: "divide by zero", operation: OpDivide, a: num(5), b: num(0), wantErr: ErrDivisionByZero},

		{name: "power", operation: OpPower, a: num(2), b: num(8), want: 256},
		{name: "power negative exponent", operation: OpPower, a: num(2), b: num(-1), want: 0.5},
		{name: "power zero to zero follows math.Pow", operation: OpPower, a: num(0), b: num(0), want: 1},
		{name: "power non finite", operation: OpPower, a: num(-1), b: num(0.5), wantErr: ErrNonFiniteResult},
		{name: "power overflow", operation: OpPower, a: num(0), b: num(-1), wantErr: ErrNonFiniteResult},

		{name: "sqrt", operation: OpSqrt, a: num(9), want: 3},
		{name: "sqrt zero", operation: OpSqrt, a: num(0), want: 0},
		{name: "sqrt two", operation: OpSqrt, a: num(2), want: math.Sqrt(2)},
		{name: "sqrt negative", operation: OpSqrt, a: num(-1), wantErr: ErrNegativeSqrt},
		{name: "sqrt missing value", operation: OpSqrt, wantErr: ErrMissingOperand},
		{name: "percent non finite", operation: OpPercent, a: num(math.NaN()), wantErr: ErrNonFiniteOperand},

		{name: "percent", operation: OpPercent, a: num(25), want: 0.25},
		{name: "percent zero", operation: OpPercent, a: num(0), want: 0},
		{name: "percent negative", operation: OpPercent, a: num(-50), want: -0.5},

		{name: "unknown operation", operation: "average", a: num(1), b: num(2), wantErr: ErrUnknownOperation},
		{name: "missing a", operation: OpAdd, b: num(1), wantErr: ErrMissingOperand},
		{name: "missing b", operation: OpSubtract, a: num(1), wantErr: ErrMissingOperand},
		{name: "sqrt rejects b", operation: OpSqrt, a: num(4), b: num(1), wantErr: ErrUnexpectedOperand},
		{name: "percent rejects b", operation: OpPercent, a: num(4), b: num(1), wantErr: ErrUnexpectedOperand},
		{name: "non finite operand", operation: OpAdd, a: num(math.Inf(1)), b: num(1), wantErr: ErrNonFiniteOperand},
		{name: "overflow multiply", operation: OpMultiply, a: num(1e308), b: num(1e308), wantErr: ErrNonFiniteResult},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.operation, tt.a, tt.b)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Evaluate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Evaluate() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecimalAdditionUsesBinary64(t *testing.T) {
	// 0.1 and 0.2 are already binary64 values here, so this is the runtime sum.
	// A constant expression such as 0.1 + 0.2 would be rounded differently.
	left := 0.1
	right := 0.2
	got, err := Evaluate(OpAdd, &left, &right)
	if err != nil {
		t.Fatalf("Evaluate() unexpected error = %v", err)
	}
	want := left + right
	if got != want {
		t.Fatalf("Evaluate() = %v, want %v", got, want)
	}
	if got == 0.3 {
		t.Fatal("binary64 sum of 0.1 and 0.2 is not exactly 0.3")
	}
}

func TestOperandErrorMessages(t *testing.T) {
	_, err := Evaluate(OpAdd, nil, num(1))
	var missing *OperandError
	if !errors.As(err, &missing) {
		t.Fatalf("error = %v, want OperandError", err)
	}
	if missing.Error() != "operand a is required" {
		t.Fatalf("message = %q", missing.Error())
	}

	_, err = Evaluate(OpSqrt, num(4), num(1))
	var unexpected *OperandError
	if !errors.As(err, &unexpected) {
		t.Fatalf("error = %v, want OperandError", err)
	}
	if unexpected.Error() != "operand b is not allowed for this operation" {
		t.Fatalf("message = %q", unexpected.Error())
	}
}
