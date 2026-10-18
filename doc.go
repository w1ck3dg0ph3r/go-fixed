// Package fixed provides a fast, allocation-free 128-bit fixed-point decimal type for Go, designed for
// monetary and other values that must be exact in base 10.
//
// A number is a plain, immutable 16-byte value that is safe to copy and compare bytewise.
// Operations that can round take an explicit rounding mode (with short aliases that truncate).
// Failed operations return sentinel values that propagate through the operation chain.
package fixed
