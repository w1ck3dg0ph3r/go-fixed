package fuzzing

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func FuzzFloat64(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.neg, c.hi, c.lo)
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}

		// String is exact, so parsing it back is the correctly rounded result.
		want, err := strconv.ParseFloat(a.String(), 64)
		if err != nil {
			t.Fatal(err)
		}

		if got := a.Float64(); got != want {
			t.Errorf("Float64(%s): %v != %v\n", a, got, want)
		}
	})
}

func FuzzFromFloat64(f *testing.F) {
	for _, v := range corpusFloat64 {
		f.Add(v)
	}

	f.Fuzz(func(t *testing.T, v float64) {
		checkFromFloat64(t, v)
	})
}

// FuzzFromFloat64Decimal builds floats from short decimals, which take the
// conversion's fast path, unlike most of the floats FuzzFromFloat64 tries.
func FuzzFromFloat64Decimal(f *testing.F) {
	for _, c := range []struct {
		coef   int64
		places uint8
	}{
		{435, 2}, {29, 2}, {1005, 3}, {-995, 2}, {1, 19}, {123456789012345, 2},
		{1 << 50, 0}, {1<<53 + 1, 3}, {9007199254740993, 16}, {math.MaxInt64, 5},
	} {
		f.Add(c.coef, c.places)
	}

	f.Fuzz(func(t *testing.T, coef int64, places uint8) {
		s := decimal.New(coef, -int32(places%25)).String()
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			t.Fatalf("ParseFloat(%q): %v", s, err)
		}
		checkFromFloat64(t, v)
	})
}

// checkFromFloat64 checks FromFloat64 against the shortest representation of v
// rounded half to even to Scale, which is what the conversion is defined to
// produce.
func checkFromFloat64(t *testing.T, v float64) {
	t.Helper()

	if math.IsNaN(v) || math.IsInf(v, 0) {
		return
	}

	shortest := decimal.RequireFromString(strconv.FormatFloat(v, 'f', -1, 64))
	rounded := shortest.RoundBank(fixed.Scale)
	want, err := fixed.FromString(rounded.String())
	if err != nil {
		want = fixed.Overflow
	}

	got := fixed.FromFloat64(v)
	if got != want {
		t.Fatalf("FromFloat64(%v): %q != %q (%q)\n", v, got, want, rounded)
	}

	// Without rounding the conversion is exact, so the value converts back
	// unchanged.
	if got.Valid() && rounded.Equal(shortest) {
		if back := got.Float64(); back != v {
			t.Errorf("FromFloat64(%v): converts back to %v\n", v, back)
		}
	}
}

func FuzzFromFloat64Exact(f *testing.F) {
	for _, v := range corpusFloat64 {
		f.Add(math.Float64bits(v))
	}

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(fixed.Scale), nil)
	max := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(10), big.NewInt(fixed.Precision), nil), big.NewInt(1))

	f.Fuzz(func(t *testing.T, bits uint64) {
		v := math.Float64frombits(bits)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return
		}

		// The exact value of the float, scaled and rounded half to even.
		r := new(big.Rat).SetFloat64(v)
		r.Mul(r, new(big.Rat).SetInt(scale))
		want, rem := new(big.Int).QuoRem(r.Num(), r.Denom(), new(big.Int))
		switch rem.Abs(rem).Lsh(rem, 1).Cmp(r.Denom()) {
		case 1:
			want.Add(want, big.NewInt(int64(r.Sign())))
		case 0:
			if want.Bit(0) == 1 {
				want.Add(want, big.NewInt(int64(r.Sign())))
			}
		}

		got := fixed.FromFloat64Exact(v)
		if want.CmpAbs(max) > 0 {
			if got != fixed.Overflow {
				t.Fatalf("FromFloat64Exact(%v): expected overflow, got %q\n", v, got)
			}
			return
		}
		if !got.Valid() {
			t.Fatalf("FromFloat64Exact(%v): unexpected %q, %s is representable\n", v, got, want)
		}
		if toBigInt(got).Cmp(want) != 0 {
			t.Fatalf("FromFloat64Exact(%v): %s != %s (%q)\n", v, toBigInt(got), want, got)
		}
	})
}

func FuzzString(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.neg, c.hi, c.lo)
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}

		want := toDecimal(a).String()

		if got := a.String(); got != want {
			t.Fatalf("string: %q != %q\n", got, want)
		}

		if got := string(a.Append([]byte("x"))); got != "x"+want {
			t.Fatalf("append: %q != %q\n", got, "x"+want)
		}
	})
}

func FuzzFromString(f *testing.F) {
	for _, s := range corpusString {
		f.Add(s)
	}

	highest := toDecimal(fixed.Highest)
	lowest := toDecimal(fixed.Lowest)

	// check parses s and compares the outcome with shopspring decimal.
	check := func(t *testing.T, s string) {
		t.Helper()

		d, derr := decimal.NewFromString(s)
		n, nerr := fixed.FromString(s)

		if derr != nil {
			if nerr != nil {
				return
			}
			t.Fatalf("FromString(%q): expected error %v", s, derr)
		}

		if nerr != nil {
			switch nerr {
			case fixed.ErrNaN:
				if strings.ContainsAny(s, "eEoOxX") {
					// We don't support scientific, octal or hexadecimal notations.
					return
				}
				if strings.Contains(s, ".+") || strings.Contains(s, ".-") {
					// We don't support weird notation as well.
					return
				}
			case fixed.ErrOverflow:
				if d.LessThan(lowest) || d.GreaterThan(highest) {
					// We have narrower range.
					return
				}
			case fixed.ErrInexact:
				if d.Exponent() < -fixed.Scale {
					// We reject numbers that can't be represented exactly.
					return
				}
			}
			// Anything else is an unexpected error.
			t.Fatalf("FromString(%q): unexpected error %v", s, nerr)
		}

		have := toDecimal(n)
		if !have.Equal(d) {
			t.Fatalf("FromString(%q): expected %s, got %s", s, d, have)
		}

		// Formatting the result and parsing it back is exact.
		back, err := fixed.FromString(n.String())
		if err != nil || back != n {
			t.Fatalf("FromString(%q): %q parses back as (%q, %v)\n", s, n, back, err)
		}
	}

	f.Fuzz(func(t *testing.T, s string) {
		check(t, s)

		// Map the input onto the decimal alphabet as well, so that random bytes
		// also have a chance to explore well formed numbers.
		b := []byte(s)
		for i := range b {
			b[i] = "0123456789.+-0123456789"[int(b[i])%23]
		}
		check(t, string(b))
	})
}

func toBigInt(n fixed.Number) *big.Int {
	neg := n.IsNegative()
	a := n.Abs()
	v := new(big.Int).SetUint64(a.Hi)
	v.Lsh(v, 64).Or(v, new(big.Int).SetUint64(a.Lo))
	if neg {
		v.Neg(v)
	}
	return v
}
