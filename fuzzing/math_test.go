package fuzzing

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func FuzzAdd(f *testing.F) {
	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok {
			return
		}

		c := a.Add(b)
		if !c.Valid() {
			return
		}

		aa, err := decimal.NewFromString(a.String())
		if err != nil {
			t.Error(err)
		}
		bb, err := decimal.NewFromString(b.String())
		if err != nil {
			t.Error(err)
		}
		cc := aa.Add(bb)

		if c.String() != cc.String() {
			t.Errorf("%s + %s : %q != %q (%d, %d)\n", a, b, c, cc, c.Lo, c.Hi)
		}
	})
}

func FuzzSub(f *testing.F) {
	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok {
			return
		}

		c := a.Sub(b)
		if !c.Valid() {
			return
		}

		aa, err := decimal.NewFromString(a.String())
		if err != nil {
			t.Error(err)
		}
		bb, err := decimal.NewFromString(b.String())
		if err != nil {
			t.Error(err)
		}
		cc := aa.Sub(bb)

		if c.String() != cc.String() {
			t.Errorf("%s - %s : %q != %q (%d, %d)\n", a, b, c, cc, c.Lo, c.Hi)
		}
	})
}

func FuzzMulZero(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundZero,
		func(a, b decimal.Decimal) decimal.Decimal {
			return a.Mul(b).RoundDown(fixed.Scale)
		},
	)
}

func FuzzMulInf(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundInf,
		func(a, b decimal.Decimal) decimal.Decimal {
			return a.Mul(b).RoundUp(fixed.Scale)
		},
	)
}

func FuzzMulFloor(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundFloor,
		func(a, b decimal.Decimal) decimal.Decimal {
			return a.Mul(b).RoundFloor(fixed.Scale)
		},
	)
}

func FuzzMulCeil(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundCeil,
		func(a, b decimal.Decimal) decimal.Decimal {
			return a.Mul(b).RoundCeil(fixed.Scale)
		},
	)
}

func FuzzMulHalfZero(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundHalfZero,
		func(a, b decimal.Decimal) decimal.Decimal {
			// shopspring has no half towards zero, so ties are truncated explicitly.
			p := a.Mul(b)
			t := p.RoundDown(fixed.Scale)
			if p.Sub(t).Abs().Shift(fixed.Scale + 1).Equal(decimal.NewFromInt(5)) {
				return t
			}
			return p.Round(fixed.Scale)
		},
	)
}

func FuzzMulHalfInf(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundHalfInf,
		func(a, b decimal.Decimal) decimal.Decimal {
			return a.Mul(b).Round(fixed.Scale)
		},
	)
}

func FuzzMulBank(f *testing.F) {
	fuzzMul(
		f,
		fixed.RoundBank,
		func(a, b decimal.Decimal) decimal.Decimal {
			return a.Mul(b).RoundBank(fixed.Scale)
		},
	)
}

func fuzzMul(
	f *testing.F,
	mode fixed.RoundingMode,
	df func(decimal.Decimal, decimal.Decimal) decimal.Decimal,
) {
	f.Helper()

	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok {
			return
		}

		c := a.MulRound(b, mode)
		if !c.Valid() {
			return
		}

		aa := toDecimal(a)
		bb := toDecimal(b)
		cc := df(aa, bb)

		if c.String() != cc.String() {
			t.Errorf("%s * %s : %q != %q (%d, %d)\n", a, b, c, cc, c.Lo, c.Hi)
		}
	})
}

func FuzzQuoRem(f *testing.F) {
	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok || b.IsZero() {
			return
		}

		c, r := a.QuoRem(b)
		if !c.Valid() || !r.Valid() {
			return
		}

		qq := toDecimal(c).Mul(toDecimal(b)).RoundDown(fixed.Scale).Add(toDecimal(r))
		if !qq.Equal(toDecimal(a)) {
			t.Errorf("%s / %s : axiom failed %q != %q\n", a, b, qq, a)
			t.Errorf("divisor:   %q", c)
			t.Errorf("remainder: %q", r)
			t.Errorf("product:   %q", c.Mul(b))
		}

		aa := toDecimal(a)
		bb := toDecimal(b)
		cc, rr := aa.QuoRem(bb, fixed.Scale)
		// QuoRem rounds the remainder up, so that RoundDown(q·b) + r == a.
		rr = rr.Abs().RoundUp(fixed.Scale)

		if c.String() != cc.String() {
			t.Errorf("%s / %s : quotient %q != %q\n", a, b, c, cc)
		}

		if !toDecimal(r.Abs()).Equal(rr) {
			t.Errorf("%s / %s : remainder %q != %q\n", a, b, r, rr)
		}
	})
}

func FuzzDivZero(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundZero,
		func(neg bool, half int, odd bool) bool {
			return false
		},
	)
}

func FuzzDivInf(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundInf,
		func(neg bool, half int, odd bool) bool {
			return true
		},
	)
}

func FuzzDivFloor(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundFloor,
		func(neg bool, half int, odd bool) bool {
			return neg
		},
	)
}

func FuzzDivCeil(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundCeil,
		func(neg bool, half int, odd bool) bool {
			return !neg
		},
	)
}

func FuzzDivHalfZero(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundHalfZero,
		func(neg bool, half int, odd bool) bool {
			return half > 0
		},
	)
}

func FuzzDivHalfInf(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundHalfInf,
		func(neg bool, half int, odd bool) bool {
			return half >= 0
		},
	)
}

func FuzzDivBank(f *testing.F) {
	fuzzDiv(
		f,
		fixed.RoundBank,
		func(neg bool, half int, odd bool) bool {
			return half > 0 || (half == 0 && odd)
		},
	)
}

// fuzzDiv checks ff against the quotient truncated by shopspring's QuoRem and
// rounded away from zero when away says so. away is only called for inexact
// quotients: neg is the sign of the quotient, half compares the remainder to
// half a unit in the last place, and odd is the parity of the last digit.
func fuzzDiv(
	f *testing.F,
	mode fixed.RoundingMode,
	away func(neg bool, half int, odd bool) bool,
) {
	f.Helper()

	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	unit := decimal.New(1, -fixed.Scale)

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok || b.IsZero() {
			return
		}

		c := a.DivRound(b, mode)
		if !c.Valid() {
			return
		}

		aa := toDecimal(a)
		bb := toDecimal(b)
		cc, rr := aa.QuoRem(bb, fixed.Scale)
		if !rr.IsZero() {
			neg := a.IsNegative() != b.IsNegative()
			half := rr.Abs().Mul(decimal.NewFromInt(2)).Cmp(bb.Abs().Mul(unit))
			odd := cc.Shift(fixed.Scale).BigInt().Bit(0) == 1
			if away(neg, half, odd) {
				if neg {
					cc = cc.Sub(unit)
				} else {
					cc = cc.Add(unit)
				}
			}
		}

		if c.String() != cc.String() {
			t.Errorf("%s / %s : %q != %q (%d, %d)\n", a, b, c, cc, c.Lo, c.Hi)
		}
	})
}

func FuzzMod(f *testing.F) {
	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok || b.IsZero() {
			return
		}

		r := a.Mod(b)
		if !r.Valid() {
			t.Fatalf("%s %% %s : Mod can never fail, got %q", a, b, r)
		}

		rr := toDecimal(a).Mod(toDecimal(b))
		if r.String() != rr.String() {
			t.Errorf("%s %% %s : %q != %q\n", a, b, r, rr)
		}

		q, _ := toDecimal(a).QuoRem(toDecimal(b), 0)
		if sum := q.Mul(toDecimal(b)).Add(toDecimal(r)); !sum.Equal(toDecimal(a)) {
			t.Errorf("%s %% %s : axiom failed %q != %q\n", a, b, sum, a)
		}

		if !r.IsZero() {
			if r.Abs().Gte(b.Abs()) {
				t.Errorf("%s %% %s : remainder %q is not below the divisor\n", a, b, r)
			}
			if r.IsNegative() != a.IsNegative() {
				t.Errorf("%s %% %s : remainder %q does not take the sign of the dividend\n", a, b, r)
			}
		}
	})
}

func FuzzShift(f *testing.F) {
	for _, c := range corpus {
		for _, k := range []int{0, 1, -1, 18, -18, 19, -19, 20, -20, 37, -37, 38, -38} {
			f.Add(c.neg, c.hi, c.lo, k)
		}
	}

	highest := toDecimal(fixed.Highest)
	lowest := toDecimal(fixed.Lowest)

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, k int) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		if k < -64 || k > 64 {
			return
		}

		r := a.Shift(k)

		// Truncate towards zero first: shopspring keeps the digits a left shift
		// pushes below the last place.
		want := toDecimal(a).Shift(int32(k)).Truncate(fixed.Scale)
		if want.GreaterThan(highest) || want.LessThan(lowest) {
			if r != fixed.Overflow {
				t.Errorf("%s shift %d : expected overflow, got %q\n", a, k, r)
			}
			return
		}

		if !r.Valid() {
			t.Fatalf("%s shift %d : unexpected %q, %s is representable\n", a, k, r, want)
		}
		if r.String() != want.String() {
			t.Errorf("%s shift %d : %q != %q\n", a, k, r, want)
		}

		// A right shift is exact, so shifting back recovers the value.
		if k > 0 {
			if back := r.Shift(-k); back != a {
				t.Errorf("%s shift %d shift %d : %q != %q\n", a, k, -k, back, a)
			}
		}
	})
}

func FuzzIntVariants(f *testing.F) {
	for _, c := range corpus {
		for _, v := range corpusInt {
			f.Add(c.neg, c.hi, c.lo, v)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, v int) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b := fixed.FromInt(v)

		check := func(op string, got, want fixed.Number) {
			if got != want {
				t.Errorf("%s(%s, %d): %q != %q", op, a, v, got, want)
			}
		}
		check("AddInt", a.AddInt(v), a.Add(b))
		check("SubInt", a.SubInt(v), a.Sub(b))
		check("MulInt", a.MulInt(v), a.Mul(b))
		check("DivInt", a.DivInt(v), a.Div(b))
		for mode := fixed.RoundZero; mode <= fixed.RoundBank; mode++ {
			check(fmt.Sprintf("DivRoundInt/%d", mode), a.DivRoundInt(v, mode), a.DivRound(b, mode))
		}
		check("ModInt", a.ModInt(v), a.Mod(b))

		q, r := a.QuoRemInt(v)
		wq, wr := a.QuoRem(b)
		check("QuoRemInt quotient", q, wq)
		check("QuoRemInt remainder", r, wr)
	})
}

func number(neg bool, hi, lo uint64) (fixed.Number, bool) {
	n := fixed.Number{Lo: lo, Hi: hi}
	if !n.Valid() || n.IsNegative() {
		return n, false
	}
	if neg {
		n = n.Neg()
	}
	return n, true
}

func toDecimal(n fixed.Number) decimal.Decimal {
	neg := n.IsNegative()
	if neg {
		n = n.Abs()
	}
	var bytes [16]byte
	binary.BigEndian.PutUint64(bytes[0:], n.Hi)
	binary.BigEndian.PutUint64(bytes[8:], n.Lo)
	bint := new(big.Int).SetBytes(bytes[:])
	if neg {
		bint = bint.Neg(bint)
	}
	d := decimal.NewFromBigInt(bint, -fixed.Scale)
	return d
}
