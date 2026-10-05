package fuzzing

import (
	"math/big"
	"testing"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func FuzzPowInt(f *testing.F) {
	for _, c := range corpus {
		for _, m := range []int16{-365, -20, -3, -1, 0, 1, 2, 3, 12, 20, 365} {
			f.Add(c.neg, c.hi, c.lo, m, uint8(fixed.RoundBank))
		}
	}
	// Exact ties at the last place, directly and through a reciprocal.
	f.Add(false, uint64(0), uint64(5_000_000_000), int16(2), uint8(fixed.RoundBank))
	f.Add(false, uint64(1), uint64(1_553_255_926_290_448_384), int16(-20), uint8(fixed.RoundHalfZero)) // 2

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, m int16, mode uint8) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		rm := fixed.RoundingMode(mode % uint8(fixed.RoundBank+1))

		want, ok := powIntOracle(a, int(m), rm)
		if !ok {
			return
		}
		if got := a.PowIntRound(int(m), rm); got != want {
			t.Fatalf("%s ^ %d, mode %d: %q != %q", a, m, rm, got, want)
		}
	})
}

// powIntOracle rounds the exact power a^m with mode, reporting false for powers
// too large to work out exactly in reasonable time.
func powIntOracle(a fixed.Number, m int, mode fixed.RoundingMode) (fixed.Number, bool) {
	if m == 0 {
		return fixed.One, true
	}
	if a.IsZero() {
		if m < 0 {
			return fixed.DivisionByZero, true
		}
		return fixed.Zero, true
	}

	c := toBigInt(a)
	neg := c.Sign() < 0 && m%2 != 0
	c.Abs(c)
	am := max(m, -m)
	if c.BitLen()*am > 200_000 {
		return fixed.Number{}, false
	}

	// The coefficient of a is c, so a^m has coefficient c^m/10^(Scale*(m-1)),
	// or 10^(Scale*(1-m))/c^-m for a negative power.
	scale := func(k int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(fixed.Scale*k)), nil) }
	p := new(big.Int).Exp(c, big.NewInt(int64(am)), nil)
	num, den := p, scale(am-1)
	if m < 0 {
		num, den = scale(am+1), p
	}

	q, r := new(big.Int).QuoRem(num, den, new(big.Int))
	if r.Sign() != 0 {
		half := new(big.Int).Lsh(r, 1).Cmp(den)
		var up bool
		switch mode {
		case fixed.RoundInf:
			up = true
		case fixed.RoundFloor:
			up = neg
		case fixed.RoundCeil:
			up = !neg
		case fixed.RoundHalfZero:
			up = half > 0
		case fixed.RoundHalfInf:
			up = half >= 0
		case fixed.RoundBank:
			up = half > 0 || (half == 0 && q.Bit(0) == 1)
		}
		if up {
			q.Add(q, big.NewInt(1))
		}
	}

	if q.Cmp(toBigInt(fixed.Highest)) > 0 {
		return fixed.Overflow, true
	}
	if neg {
		q.Neg(q)
	}
	return fromBigInt(q), true
}

// fromBigInt converts a coefficient that is known to be in range.
func fromBigInt(c *big.Int) fixed.Number {
	a := new(big.Int).Abs(c)
	lo := new(big.Int).And(a, new(big.Int).SetUint64(^uint64(0))).Uint64()
	hi := new(big.Int).Rsh(a, 64).Uint64()
	n := fixed.Number{Lo: lo, Hi: hi}
	if c.Sign() < 0 {
		n = n.Neg()
	}
	return n
}
