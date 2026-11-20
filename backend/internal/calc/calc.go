// Package calc evaluates calculator operations.
//
// Operands are pointers so a missing value stays distinct from a valid zero.
// Arithmetic uses IEEE-754 binary64, which is also how encoding/json
// represents JSON numbers. This is appropriate for a general calculator and
// unsuitable for currency.
package calc

import (
	"errors"
	"fmt"
	"math"
)

const (
	OpAdd      = "add"
	OpSubtract = "subtract"
	OpMultiply = "multiply"
	OpDivide   = "divide"
	OpPower    = "power"
	OpSqrt     = "sqrt"
	// OpPercent converts a number to its equivalent divided by 100.
	// percent(25) is 0.25. It does not mean "x percent of y".
	OpPercent = "percent"
)

var (
	ErrUnknownOperation  = errors.New("unknown operation")
	ErrMissingOperand    = errors.New("missing operand")
	ErrUnexpectedOperand = errors.New("unexpected operand")
	ErrNonFiniteOperand  = errors.New("operands must be finite numbers")
	ErrDivisionByZero    = errors.New("division by zero")
	ErrNegativeSqrt      = errors.New("square root of a negative number is not defined")
	ErrNonFiniteResult   = errors.New("result is not a finite number")
)

// OperandError identifies an operand that was required and omitted, or
// supplied to an operation that does not accept it.
type OperandError struct {
	Name string
	Err  error
}

func (e *OperandError) Error() string {
	if errors.Is(e.Err, ErrUnexpectedOperand) {
		return fmt.Sprintf("operand %s is not allowed for this operation", e.Name)
	}
	return fmt.Sprintf("operand %s is required", e.Name)
}

func (e *OperandError) Unwrap() error {
	return e.Err
}

// Evaluate applies operation to a and b. Unary operations accept only a.
func Evaluate(operation string, a, b *float64) (float64, error) {
	switch operation {
	case OpAdd, OpSubtract, OpMultiply, OpDivide, OpPower:
		left, right, err := binaryOperands(a, b)
		if err != nil {
			return 0, err
		}
		value, err := applyBinary(operation, left, right)
		if err != nil {
			return 0, err
		}
		return finish(value)
	case OpSqrt:
		value, err := unaryOperand(a, b)
		if err != nil {
			return 0, err
		}
		if value < 0 {
			return 0, ErrNegativeSqrt
		}
		return finish(math.Sqrt(value))
	case OpPercent:
		value, err := unaryOperand(a, b)
		if err != nil {
			return 0, err
		}
		return finish(value / 100)
	default:
		return 0, ErrUnknownOperation
	}
}

func binaryOperands(a, b *float64) (float64, float64, error) {
	if a == nil {
		return 0, 0, &OperandError{Name: "a", Err: ErrMissingOperand}
	}
	if b == nil {
		return 0, 0, &OperandError{Name: "b", Err: ErrMissingOperand}
	}
	if err := checkFinite(*a, *b); err != nil {
		return 0, 0, err
	}
	return *a, *b, nil
}

func unaryOperand(a, b *float64) (float64, error) {
	if a == nil {
		return 0, &OperandError{Name: "a", Err: ErrMissingOperand}
	}
	if b != nil {
		return 0, &OperandError{Name: "b", Err: ErrUnexpectedOperand}
	}
	if err := checkFinite(*a); err != nil {
		return 0, err
	}
	return *a, nil
}

func checkFinite(values ...float64) error {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return ErrNonFiniteOperand
		}
	}
	return nil
}

func applyBinary(operation string, a, b float64) (float64, error) {
	switch operation {
	case OpAdd:
		return a + b, nil
	case OpSubtract:
		return a - b, nil
	case OpMultiply:
		return a * b, nil
	case OpDivide:
		if b == 0 {
			return 0, ErrDivisionByZero
		}
		return a / b, nil
	case OpPower:
		return math.Pow(a, b), nil
	default:
		return 0, ErrUnknownOperation
	}
}

func finish(value float64) (float64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, ErrNonFiniteResult
	}
	return value, nil
}
