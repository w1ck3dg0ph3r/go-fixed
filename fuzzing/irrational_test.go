package fuzzing

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func FuzzSqrt(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.neg, c.hi, c.lo)
	}

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(fixed.Scale), nil)

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}

		if a.IsNegative() {
			if r := a.Sqrt(); r != fixed.NaN {
				t.Fatalf("Sqrt(%s): expected NaN, got %q", a, r)
			}
			return
		}

		// sqrt(a/10^Scale) == sqrt(a*10^Scale)/10^Scale, so the expected result
		// at full scale is the integer square root of the scaled coefficient.
		x := new(big.Int).SetUint64(ahi)
		x.Lsh(x, 64).Or(x, new(big.Int).SetUint64(alo))
		x.Mul(x, scale)
		root := new(big.Int).Sqrt(x)

		r := a.Sqrt()
		got := new(big.Int).SetUint64(r.Hi)
		got.Lsh(got, 64).Or(got, new(big.Int).SetUint64(r.Lo))
		if got.Cmp(root) != 0 {
			t.Fatalf("Sqrt(%s): expected %s, got %s (%q)", a, root, got, r)
		}
	})
}

// FuzzExp checks ExpRound against shopspring's ExpTaylor to 45 places. Inputs
// are an integer part and 19 fractional digits, which keeps them around the
// range where e^x fits.
func FuzzExp(f *testing.F) {
	for _, s := range []struct {
		whole int8
		frac  uint64
	}{
		{0, 0}, {0, 1}, {1, 0}, {-1, 0}, {0, 5_000_000_000_000_000_000}, {2, 0}, {10, 0}, {-10, 0},
		{43, 7_491_167_668_868_679_963}, {43, 7_491_167_668_868_679_964}, {-45, 5_577_360_525_531_866_943},
		{-44, 0}, {-45, 0}, {0, 6_931_471_805_599_453_094}, {-1, 5_000_000_000_000_000_000}, {-3, 1_234_567_890_123_456_789},
	} {
		f.Add(s.whole, s.frac, uint8(fixed.RoundBank))
	}

	limit := decimal.RequireFromString(fixed.Highest.String())
	unit := decimal.New(1, -fixed.Scale)

	f.Fuzz(func(t *testing.T, whole int8, frac uint64, mode uint8) {
		x := fixed.FromInt(int(whole) % 47).Add(fixed.Number{Lo: frac % 10_000_000_000_000_000_000})
		rm := fixed.RoundingMode(mode % uint8(fixed.RoundBank+1))

		v, err := toDecimal(x).ExpTaylor(45)
		if err != nil {
			t.Fatal(err)
		}

		// The reference is itself rounded at 45 places, so values too close to
		// a rounding boundary of the 19th cannot be decided from it.
		below := v.Sub(v.Truncate(fixed.Scale)).Div(unit)
		for _, edge := range []decimal.Decimal{decimal.Zero, decimal.NewFromFloat(0.5), decimal.NewFromInt(1)} {
			if below.Sub(edge).Abs().LessThan(decimal.New(1, -25)) {
				return
			}
		}

		var want decimal.Decimal
		switch rm {
		case fixed.RoundZero, fixed.RoundFloor:
			want = v.Truncate(fixed.Scale)
		case fixed.RoundInf, fixed.RoundCeil:
			want = v.Truncate(fixed.Scale).Add(unit)
		default:
			want = v.Round(fixed.Scale)
		}

		got := x.ExpRound(rm)
		if want.GreaterThan(limit) {
			if got != fixed.Overflow {
				t.Fatalf("exp(%s), mode %d: %q, want overflow", x, rm, got)
			}
			return
		}
		if got != fixed.MustFromString(want.String()) {
			t.Fatalf("exp(%s), mode %d: %q != %s", x, rm, got, want)
		}
	})
}

// FuzzLn checks LnRound against shopspring's Ln to 45 places.
func FuzzLn(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.hi, c.lo, uint8(fixed.RoundBank))
	}
	for _, s := range []string{"2", "10", "0.5", "1.0000000000000000001", "0.9999999999999999999", "123.456", "0.0001"} {
		n := fixed.MustFromString(s)
		f.Add(n.Hi, n.Lo, uint8(fixed.RoundHalfZero))
	}

	unit := decimal.New(1, -fixed.Scale)

	f.Fuzz(func(t *testing.T, hi uint64, lo uint64, mode uint8) {
		x, ok := number(false, hi, lo)
		if !ok || x.IsZero() {
			return
		}
		rm := fixed.RoundingMode(mode % uint8(fixed.RoundBank+1))

		v, err := toDecimal(x).Ln(45)
		if err != nil {
			t.Fatal(err)
		}

		// The reference is itself rounded at 45 places, so values too close to
		// a rounding boundary of the 19th cannot be decided from it.
		below := v.Sub(v.Truncate(fixed.Scale)).Abs().Div(unit)
		for _, edge := range []decimal.Decimal{decimal.Zero, decimal.NewFromFloat(0.5), decimal.NewFromInt(1)} {
			if below.Sub(edge).Abs().LessThan(decimal.New(1, -25)) {
				return
			}
		}

		// The logarithm is never exact here, so the directed modes move the
		// truncated value one unit away from zero or not at all.
		toward := v.Truncate(fixed.Scale)
		away := toward.Add(unit)
		if v.IsNegative() {
			away = toward.Sub(unit)
		}
		var want decimal.Decimal
		switch rm {
		case fixed.RoundZero:
			want = toward
		case fixed.RoundInf:
			want = away
		case fixed.RoundFloor:
			want = toward
			if v.IsNegative() {
				want = away
			}
		case fixed.RoundCeil:
			want = away
			if v.IsNegative() {
				want = toward
			}
		default:
			want = v.Round(fixed.Scale)
		}

		if got := x.LnRound(rm); got != fixed.MustFromString(want.String()) {
			t.Fatalf("ln(%s), mode %d: %q != %s", x, rm, got, want)
		}
	})
}

// FuzzPow checks PowRound against shopspring's ExpTaylor(y Ln(x)), with ln to 60
// places and exp to 45. Exponents stay below 50 in magnitude, which keeps that
// reference within 10^-20 of a unit in the last place.
func FuzzPow(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.hi, c.lo, int8(0), uint64(5_000_000_000_000_000_000), uint8(fixed.RoundBank))
	}
	for _, s := range []struct {
		x     string
		whole int8
		frac  uint64
	}{
		{"2", 0, 5_000_000_000_000_000_000}, {"10", 1, 5_000_000_000_000_000_000}, {"0.5", -1, 5_000_000_000_000_000_000},
		{"1.05", 30, 1_234_567_890_123_456_789}, {"123.456", 2, 7_182_818_284_590_452_354}, {"0.9", -40, 0},
	} {
		x := fixed.MustFromString(s.x)
		f.Add(x.Hi, x.Lo, s.whole, s.frac, uint8(fixed.RoundHalfZero))
	}

	limit := decimal.RequireFromString(fixed.Highest.String())
	unit := decimal.New(1, -fixed.Scale)

	f.Fuzz(func(t *testing.T, hi uint64, lo uint64, whole int8, frac uint64, mode uint8) {
		x, ok := number(false, hi, lo)
		if !ok || x.IsZero() {
			return
		}
		y := fixed.FromInt(int(whole) % 50).Add(fixed.Number{Lo: frac % 10_000_000_000_000_000_000})
		if y.Round(0, fixed.RoundZero) == y {
			return // integer powers are FuzzPowInt's
		}
		rm := fixed.RoundingMode(mode % uint8(fixed.RoundBank+1))

		l, err := toDecimal(x).Ln(60)
		if err != nil {
			t.Fatal(err)
		}
		t60 := l.Mul(toDecimal(y))
		if t60.GreaterThan(decimal.NewFromInt(45)) {
			if got := x.PowRound(y, rm); got != fixed.Overflow {
				t.Fatalf("%s^%s, mode %d: %q, want overflow", x, y, rm, got)
			}
			return
		}
		if t60.LessThan(decimal.NewFromInt(-46)) {
			return // below the last place, which TestPow covers
		}
		v, err := t60.ExpTaylor(45)
		if err != nil {
			t.Fatal(err)
		}

		// Values this close to a rounding boundary, exact results among them,
		// cannot be decided from the reference.
		below := v.Sub(v.Truncate(fixed.Scale)).Div(unit)
		for _, edge := range []decimal.Decimal{decimal.Zero, decimal.NewFromFloat(0.5), decimal.NewFromInt(1)} {
			if below.Sub(edge).Abs().LessThan(decimal.New(1, -15)) {
				return
			}
		}

		var want decimal.Decimal
		switch rm {
		case fixed.RoundZero, fixed.RoundFloor:
			want = v.Truncate(fixed.Scale)
		case fixed.RoundInf, fixed.RoundCeil:
			want = v.Truncate(fixed.Scale).Add(unit)
		default:
			want = v.Round(fixed.Scale)
		}

		got := x.PowRound(y, rm)
		if want.GreaterThan(limit) {
			if got != fixed.Overflow {
				t.Fatalf("%s^%s, mode %d: %q, want overflow", x, y, rm, got)
			}
			return
		}
		if got != fixed.MustFromString(want.String()) {
			t.Fatalf("%s^%s, mode %d: %q != %s", x, y, rm, got, want)
		}
	})
}
