package benchmarks

import "github.com/cockroachdb/apd/v3"

// apdContext returns a context with enough significant digits for the exact
// product of two operands and for every result to keep fixed.Scale
// fractional digits, rounding with r.
func apdContext(r apd.Rounder) *apd.Context {
	c := apd.BaseContext.WithPrecision(50)
	c.Rounding = r
	return c
}

func apdParse(s string) *apd.Decimal {
	d, _, err := apd.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}
