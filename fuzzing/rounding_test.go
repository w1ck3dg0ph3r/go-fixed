package fuzzing

import (
	"testing"

	"github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func FuzzRoundZero(f *testing.F) {
	fuzzRound(f, fixed.RoundZero, decimal.Decimal.RoundDown)
}

func FuzzRoundInf(f *testing.F) {
	fuzzRound(f, fixed.RoundInf, decimal.Decimal.RoundUp)
}

func FuzzRoundFloor(f *testing.F) {
	fuzzRound(f, fixed.RoundFloor, decimal.Decimal.RoundFloor)
}

func FuzzRoundCeil(f *testing.F) {
	fuzzRound(f, fixed.RoundCeil, decimal.Decimal.RoundCeil)
}

func FuzzRoundHalfZero(f *testing.F) {
	// shopspring has no half towards zero, so ties are truncated explicitly.
	fuzzRound(f, fixed.RoundHalfZero, func(d decimal.Decimal, places int32) decimal.Decimal {
		t := d.RoundDown(places)
		if d.Sub(t).Abs().Shift(places + 1).Equal(decimal.NewFromInt(5)) {
			return t
		}
		return d.Round(places)
	})
}

func FuzzRoundHalfInf(f *testing.F) {
	fuzzRound(f, fixed.RoundHalfInf, decimal.Decimal.Round)
}

func FuzzRoundBank(f *testing.F) {
	fuzzRound(f, fixed.RoundBank, decimal.Decimal.RoundBank)
}

func fuzzRound(
	f *testing.F,
	mode fixed.RoundingMode,
	df func(decimal.Decimal, int32) decimal.Decimal,
) {
	f.Helper()

	scales := []int{0, 1, 2, 18, 19, 20, -1, -2, -18, -19, -20, -40}
	for _, c := range corpus {
		for _, k := range scales {
			f.Add(c.neg, c.hi, c.lo, k)
		}
	}
	for _, s := range []string{
		"125",
		"135",
		"125.0000000000000000001",
		"124.9999999999999999999",
		"5000000000000000000",
		"4999999999999999999.9999999999999999999",
		"500000000000000000",
		"9500000000000000000",
		"1200.5",
	} {
		v := fixed.MustFromString(s)
		for _, k := range scales {
			f.Add(false, v.Hi, v.Lo, k)
			f.Add(true, v.Hi, v.Lo, k)
		}
	}

	lowest := toDecimal(fixed.Lowest)
	highest := toDecimal(fixed.Highest)

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, k int) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		if k < -64 || k > 64 {
			return
		}

		r := a.Round(k, mode)

		want := df(toDecimal(a), int32(k))
		if want.GreaterThan(highest) || want.LessThan(lowest) {
			if r != fixed.Overflow {
				t.Errorf("%s round %d : expected overflow, got %q\n", a, k, r)
			}
			return
		}

		if !r.Valid() {
			t.Fatalf("%s round %d : unexpected %q, %s is representable\n", a, k, r, want)
		}
		if r.String() != want.String() {
			t.Errorf("%s round %d : %q != %q\n", a, k, r, want)
		}
	})
}
