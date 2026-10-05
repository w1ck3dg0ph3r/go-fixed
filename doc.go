// Package fixed provides an allocation-free 128-bit fixed-point decimal type
// for money and other exact base-10 values.
//
// A [Number] has 19 decimal places and a range of
// ±9999999999999999999.9999999999999999999. Mul and Div truncate at the last
// place; [Number.Round], [Number.MulRound] and [Number.DivRound] take an
// explicit rounding mode. Failed operations return a sentinel (NaN, Overflow,
// DivisionByZero) that propagates through later operations; check it once with
// [Number.Err].
package fixed
