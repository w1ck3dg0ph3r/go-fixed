package fixed

import "errors"

// Sentinel numbers returned by operations that may fail.
//
// Sentinel numbers:
//   - propagate through further operations (One.Div(Zero).Add(One) == DivisionByZero)
//   - are positive (NaN.IsPositive() && NaN.Sign() == 1)
//   - compare greater than valid numbers
//   - are equal only to themselves
//
// Use [Number.Err] to get the matching error.
var (
	NaN            = Number{firstCodeLo + codeNaN, firstCodeHi}
	Overflow       = Number{firstCodeLo + codeOverflow, firstCodeHi}
	DivisionByZero = Number{firstCodeLo + codeDivisionByZero, firstCodeHi}
)

var (
	// ErrNaN reports invalid input or a NaN result.
	ErrNaN = errors.New("NaN")
	// ErrOverflow reports a value out of the representable range.
	ErrOverflow = errors.New("overflow")
	// ErrDivisionByZero reports a DivisionByZero result.
	ErrDivisionByZero = errors.New("division by zero")
	// ErrInexact reports a value with more than Scale decimal places.
	ErrInexact = errors.New("inexact")
)

// Valid reports whether n is a valid representable number.
func (n Number) Valid() bool {
	return inBounds(n.Lo, n.Hi)
}

// Err returns the error for an invalid n, or nil for a valid one.
func (n Number) Err() error {
	if inBounds(n.Lo, n.Hi) {
		return nil
	}
	if n.Hi == firstCodeHi {
		switch n.Lo - firstCodeLo {
		case codeNaN:
			return ErrNaN
		case codeOverflow:
			return ErrOverflow
		case codeDivisionByZero:
			return ErrDivisionByZero
		}
	}
	return ErrNaN
}

// valid reports whether n is a number rather than a sentinel.
func (n Number) valid() bool {
	return !(n.Hi == firstCodeHi && n.Lo-firstCodeLo < _codeCount)
}

const (
	codeNaN = iota
	codeOverflow
	codeDivisionByZero

	_codeCount
)

const (
	firstCodeHi = 0b0100101100111011010011001010100001011010100001101100010001111010
	firstCodeLo = 0b0000100110001010001000100100000000000000000000000000000000000000
)
