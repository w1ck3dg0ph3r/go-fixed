package fixed

import (
	"math"
	"math/bits"
)

// Sqrt returns the square root of n, truncating any remainder towards zero.
//
// The root of a positive number is positive, so this is also the floor of the
// exact root.
func (n Number) Sqrt() Number {
	if !n.valid() {
		return n
	}
	if n.IsZero() {
		return Zero
	}
	if n.IsNegative() {
		return NaN
	}

	hi, lo := mul128By64(n, pow10u64Scale)
	return sqrt192(hi, lo)
}

// PowInt raises n to the integer power m, truncating the result to Scale
// places. Use [Number.PowIntRound] to round it.
//
// Returns DivisionByZero for zero to a negative power, and One for any number
// to the power of zero.
func (n Number) PowInt(m int) Number {
	return n.PowIntRound(m, RoundZero)
}

// PowIntRound raises n to the integer power m, rounding the result to Scale
// places according to mode.
//
// The power is computed with about 57 significant digits and rounded once, so
// the result is the exact power correctly rounded.
func (n Number) PowIntRound(m int, mode RoundingMode) Number {
	if !n.valid() {
		return n
	}
	if m == 0 {
		return One
	}
	if n.IsZero() {
		if m < 0 {
			return DivisionByZero
		}
		return Zero
	}

	neg := n.IsNegative() && m&1 != 0
	k := uint(m)
	base := wideFromNumber(n)
	if m < 0 {
		k = -k
		base = wideReciprocal(n.Abs())
	}

	// Square and multiply. Once the result must be out of range, or so small
	// that it rounds like any value below a hundredth of the last place, the
	// rest of the exponent cannot change it, which also keeps the decimal
	// exponents from growing without bound.
	acc := wideOne
	for {
		if k&1 != 0 {
			acc = acc.Mul(base)
			if lo, hi := acc.Magnitude(); lo >= 20 {
				return Overflow
			} else if hi <= -Scale-2 {
				return acc.Round(neg, mode)
			}
		}
		k >>= 1
		if k == 0 {
			return acc.Round(neg, mode)
		}
		base = base.Mul(base)

		// What is left of the exponent multiplies acc by base at least once.
		// Above one, acc is at least one too, and below one it is at most one.
		if lo, hi := base.Magnitude(); lo >= 20 {
			return Overflow
		} else if hi <= -Scale-2 {
			return acc.Mul(base).Round(neg, mode)
		}
	}
}

// Pow raises n to the power y, truncating the result to Scale places. Use
// [Number.PowRound] to round it.
func (n Number) Pow(y Number) Number {
	return n.PowRound(y, RoundZero)
}

// PowRound raises n to the power y, rounding the result to Scale places
// according to mode.
//
// An integer y gives the same result as [Number.PowIntRound]. Otherwise n must
// not be negative, and the power is e^(y ln n), computed to about 50 significant
// digits. Results within that of a number with Scale places, or of halfway
// between two, are taken to be exact, so that powers like 4^0.5 and 1.21^1.5
// come out exactly. Returns NaN for a negative n and a fractional y, and
// DivisionByZero for zero to a negative power.
func (n Number) PowRound(y Number, mode RoundingMode) Number {
	if !n.valid() {
		return n
	}
	if !y.valid() {
		return y
	}
	if y.IsZero() {
		return One
	}

	whole := y.Round(0, RoundZero) == y
	if whole && y.Abs().Lte(powMaxInt) {
		return n.PowIntRound(y.Int(), mode)
	}
	if n.IsZero() {
		if y.IsNegative() {
			return DivisionByZero
		}
		return Zero
	}

	// Only an integer power of a negative number is real, and it is negative
	// for an odd one. An integer this large overflows or vanishes unless n is
	// one or minus one.
	neg := false
	x := n
	if n.IsNegative() {
		if !whole {
			return NaN
		}
		neg = y.Int()&1 != 0
		x = n.Neg()
	}
	if x == One {
		if neg {
			return One.Neg()
		}
		return One
	}

	l, lneg := lnWide(x)
	t := l.Mul(wideFromNumber(y))
	tneg := lneg != y.IsNegative()

	// e^44 is past the largest number, and e^-45 is below half the last place.
	if tn := t.Round(false, RoundZero); tn == Overflow || tn.Gte(expUnderflow.Neg()) {
		if tneg {
			return wide{m: [3]uint64{1}, e: -100, inexact: true}.Round(neg, mode)
		}
		return Overflow
	} else if !tneg && tn.Gte(expOverflow) {
		return Overflow
	}

	// powTrusted is how many significant digits of e^(y ln n) PowRound relies on.
	// They measure at 51.7 or more, and 50 still leaves at least 11 below the last
	// place of the largest results.
	const powTrusted = 50

	r := expWideOf(t, tneg)
	r.inexact = true
	return r.RoundSnap(neg, mode, powTrusted)
}

// Exp returns e raised to the power n, truncated to Scale places. Use
// [Number.ExpRound] to round it.
func (n Number) Exp() Number {
	return n.ExpRound(RoundZero)
}

// ExpRound returns e raised to the power n, rounded to Scale places according
// to mode.
//
// The power is computed to about 53 significant digits before it is rounded,
// so the result is correctly rounded unless the exact one lies within about
// 10^-15 of a unit in the last place from a rounding boundary.
func (n Number) ExpRound(mode RoundingMode) Number {
	if !n.valid() {
		return n
	}
	if n.IsZero() {
		return One
	}

	// e^44 is past the largest number, and e^-45 is below half the last place,
	// where only the directed modes still round up to it.
	if n.Gte(expOverflow) {
		return Overflow
	}
	if n.Lte(expUnderflow) {
		return wide{m: [3]uint64{1}, e: -100, inexact: true}.Round(false, mode)
	}

	// e^n is irrational for any nonzero n, so it is never exactly a tie.
	r := expWide(n)
	r.inexact = true
	return r.Round(false, mode)
}

// Ln returns the natural logarithm of n, truncated to Scale places, or NaN if
// n is not positive. Use [Number.LnRound] to round it.
func (n Number) Ln() Number {
	return n.LnRound(RoundZero)
}

// LnRound returns the natural logarithm of n, rounded to Scale places according
// to mode, or NaN if n is not positive.
//
// The logarithm is computed to within about 10^-42 before it is rounded, so the
// result is correctly rounded unless the exact one lies within about 10^-23 of
// a unit in the last place from a rounding boundary.
func (n Number) LnRound(mode RoundingMode) Number {
	if !n.valid() {
		return n
	}
	if n.Sign() <= 0 {
		return NaN
	}
	if n == One {
		return Zero
	}

	// ln n is irrational for any positive n other than one, so it is never
	// exactly a tie.
	y, neg := lnWide(n)
	y.inexact = true
	return y.Round(neg, mode)
}

// sqrt192 returns the integer square root of the 192-bit value hi:lo.
//
// The value must be non-zero.
func sqrt192(hi uint64, lo Number) Number {
	// First, seed root r by over-estimate of the integer square root of the
	// non-zero 192-bit value hi:lo, good to about 48 bits.
	var e uint
	if hi != 0 {
		e = 128 + uint(bits.Len64(hi))
	} else {
		e = 128 - leadingZeroes128(lo)
	}

	// Split the value as m*2^(2k) so that the root factors as sqrt(m)*2^k. The
	// shift has to be even for that to work, which leaves m holding the top 63
	// or 64 bits.
	var k uint
	if e > 64 {
		k = (e - 64 + 1) / 2
	}
	m := shiftRight192(hi, lo, 2*k)

	// math.Sqrt is correctly rounded, so it hands back all 53 bits of mantissa
	// it can hold. Scaling by 2^21 before truncating to an integer keeps them
	// instead of discarding everything below 2^k, which is what makes one Heron
	// step enough.
	s := uint64(math.Sqrt(float64(m)) * (1 << 21))

	var r Number
	if k >= 21 {
		r = shiftLeft128(Number{Lo: s}, k-21)
	} else {
		r = Number{Lo: s >> (21 - k)}
	}

	// Push the estimate past its own error bound: Heron's iteration converges to
	// the root only from above. The relative error is under 2^-51 and the shift
	// back down above can drop one more unit, so 2^-48 plus two units covers it.
	r = add128(r, add64(shiftRight128(r, 48), 2))

	// A single step of Heron's iteration. The step squares the relative error,
	// so the seed's 48 good bits cover the 95 the root can need, and the value
	// lands within one unit of the root. Iterating to a fixed point instead
	// would cost a second division purely to observe that it had converged,
	// which is more expensive than the multiplication that settles the last unit
	// below.
	q, _ := quoRem192By128(hi, lo, r)
	r = shiftRight128(add128(r, q), 1)

	// Settle the last unit against the exact square. An integer Heron step never
	// lands below the root, for any positive r: floor((r + floor(N/r))/2) equals
	// floor((r + N/r)/2), which is at least floor(sqrt(N)). So stepping down is
	// all that can be needed, and it runs for well under one input in a hundred.
	x := Number{Lo: hi}
	for {
		sqlo, sqhi := mul128(r, r)
		if cmpUnsigned256(sqhi, sqlo, x, lo) <= 0 {
			return r
		}
		r = sub64(r, 1)
	}
}

// powMaxInt is the largest integer power PowRound hands to PowIntRound.
var powMaxInt = FromInt(math.MaxInt)

// expWide returns e^n for -45 < n < 44, never above the exact value.
func expWide(n Number) wide {
	// e^n = e^k * e^f for the integer k = floor(n) and 0 <= f < 1.
	fl := n.Round(0, RoundFloor)
	k := fl.Int()
	return expTimesPow(expFraction(n.Sub(fl)), k)
}

// expWideOf returns e^t, or e^-t if neg is set, for a magnitude t below 45.
func expWideOf(t wide, neg bool) wide {
	// e^t = e^k * e^f for the integer part k and 0 <= f < 1, and
	// e^-t = e^-(k+1) * e^(1-f) keeps the fraction positive.
	k, f := splitInteger(t)
	if neg && !f.IsZero() {
		k++
		f, _ = wideOne.Sub(f)
	}
	if neg {
		k = -k
	}

	// f/2^8 is f*5^8/10^8, exactly but for the digits that do not fit.
	f = f.Normalize()
	p := [6]uint64{f.m[0], f.m[1], f.m[2]}
	mulWordsUint64(p[:], 390625)
	return expTimesPow(expSeries(shrinkWords(p, f.e-expHalvings, f.inexact)), k)
}

// expTimesPow multiplies r by e^k for |k| below 64.
func expTimesPow(r wide, k int) wide {
	table := &expPow2
	if k < 0 {
		k = -k
		table = &expNegPow2
	}
	for i := 0; k != 0; i, k = i+1, k>>1 {
		if k&1 != 0 {
			r = r.Mul(table[i])
		}
	}
	return r
}

// splitInteger returns the integer part of the magnitude t, which must be
// small, and the fraction left over.
func splitInteger(t wide) (int, wide) {
	if t.e >= 0 {
		return int(t.m[0] * pow10u64[t.e]), wide{}
	}
	if -t.e >= 60 {
		return 0, t
	}
	q := t.m
	for d := -t.e; d > 0; {
		c := min(d, Scale)
		divWordsPow10(q[:], c)
		d -= c
	}
	f, _ := t.Sub(wide{m: [3]uint64{q[0]}})
	return int(q[0]), f
}

var (
	expOverflow  = FromInt(44)
	expUnderflow = FromInt(-45)
)

// expHalvings is how many times expFraction halves its argument before the
// series and squares the sum after it.
const expHalvings = 8

// expFraction returns e^f for 0 <= f < 1.
//
// The series for e^(f/2^8) needs about 17 terms below 192 bits, and squaring
// the sum 8 times gives e^f. Every term is positive and truncated, so the sum
// never exceeds the exact value.
func expFraction(f Number) wide {
	// f/2^8 is f*5^8/10^8, exactly.
	hi, lo := bits.Mul64(f.Lo, 390625)
	return expSeries(wide{m: [3]uint64{lo, hi}, e: -Scale - expHalvings})
}

// expSeries returns e^(2^8*x) for 0 <= x < 1/2^8: the series for e^x, squared 8 times.
func expSeries(x wide) wide {
	sum, term := wideOne, wideOne
	for k := uint64(1); ; k++ {
		term = term.Mul(x).DivUint64(k)
		if term.IsZero() {
			break
		}
		// Terms shrink at least 250 times each step, so once one is below the
		// last of the sum's 57 digits the rest of the series cannot reach it.
		slo, _ := sum.Magnitude()
		if _, thi := term.Magnitude(); thi < slo-58 {
			sum.inexact = true
			break
		}
		sum = sum.Add(term)
	}

	for range expHalvings {
		sum = sum.Mul(sum)
	}
	return sum
}

// expPow2 holds e^(2^i), truncated to 192 bits.
var expPow2 = [...]wide{
	{m: [3]uint64{0x40f58f3b79a7d1b6, 0x1a10b7105509e711, 0x6edc2fbe41b6e8ca}, e: -57, inexact: true}, // 2.71828182845904523536…
	{m: [3]uint64{0xa650ebec30c980f8, 0x710db694dbcf8e4b, 0x1e2288d695049ebb}, e: -56, inexact: true}, // 7.38905609893065022723…
	{m: [3]uint64{0xaf8e31224fb73406, 0x74cfde2a6ee00060, 0xdeab20dd4f4324d9}, e: -56, inexact: true}, // 54.5981500331442390781…
	{m: [3]uint64{0xae2f817f176b25c0, 0x4d0bbe604d1cb747, 0x7992a74f5e4be599}, e: -54, inexact: true}, // 2980.95798704172827474…
	{m: [3]uint64{0xd17b4f7979daeaca, 0x9e9316c3d8715945, 0x243d8842cb9a8148}, e: -50, inexact: true}, // 8886110.52050787263676…
	{m: [3]uint64{0xd99073c082dcb036, 0x28050d90986d7256, 0x20341e22412a487f}, e: -43, inexact: true}, // 78962960182680.6951609…
}

// expNegPow2 holds e^-(2^i), truncated to 192 bits.
var expNegPow2 = [...]wide{
	{m: [3]uint64{0x7c33dfd89335bb76, 0x8165019622d42710, 0x960867e11ac47b92}, e: -58, inexact: true}, // 0.367879441171442321595…
	{m: [3]uint64{0x7093628fbecebdd9, 0x5069ceff0de0059b, 0x3731a9b0ce89729b}, e: -58, inexact: true}, // 0.135335283236612691893…
	{m: [3]uint64{0x10cce429166deedf, 0x7470c3b05bf666ea, 0x4ab26b5392e389f9}, e: -59, inexact: true}, // 0.0183156388887341802937…
	{m: [3]uint64{0x22ec79971f8ced45, 0x6e6c43f72809eea1, 0x88cfeebbefb25bb9}, e: -61, inexact: true}, // 0.000335462627902511838…
	{m: [3]uint64{0x3c50879cd16aca71, 0xe44c71b1f8269353, 0x2de5386fb3c749a9}, e: -64, inexact: true}, // 1.12535174719259114513…e-7
	{m: [3]uint64{0x7e6f0339c6691e45, 0x93742100c44d369f, 0x33a60179eac2f375}, e: -71, inexact: true}, // 1.26641655490941757231…e-14
}

// lnNearOne bounds |n-1| below which lnWide sums a series instead of
// correcting a float64 logarithm, whose error would swamp the tiny result.
var lnNearOne = MustFromString("0.01")

// lnWide returns the magnitude and sign of ln n for a positive n, to about 53
// significant digits.
func lnWide(n Number) (wide, bool) {
	if u := n.Sub(One); u.Abs().Lt(lnNearOne) {
		return lnSeries(u, n), u.IsNegative()
	}

	// Start from the float64 logarithm, good to about 15 digits, and correct it
	// with the series for ln(x/e^y0) = -ln(1 - d):
	//
	//	ln x = y0 + d + d^2/2 + d^3/3 + O(d^4)
	//
	// for d = (x - e^y0)/x. With d around 10^-15, the dropped terms are around
	// 10^-58.
	y0 := FromFloat64(math.Log(n.Float64()))
	diff, dneg := wideFromNumber(n).Sub(expWide(y0))
	d := diff.DivNumber(n)
	d2 := d.Mul(d)
	half := d2.DivUint64(2)
	third := d2.Mul(d).DivUint64(3)

	// The correction's magnitude is |d| + d^2/2 + |d|^3/3 for a positive d and
	// |d| - d^2/2 + |d|^3/3 for a negative one.
	var corr wide
	if dneg {
		corr, _ = d.Sub(half)
		corr = corr.Add(third)
	} else {
		corr = d.Add(half).Add(third)
	}

	y := wideFromNumber(y0)
	yneg := y0.IsNegative()
	if dneg == yneg {
		y = y.Add(corr)
	} else {
		var flip bool
		y, flip = y.Sub(corr)
		yneg = yneg != flip
	}
	return y, yneg
}

// lnSeries returns the magnitude of ln n for n = 1+u near one, from
//
//	ln n = 2 atanh(z) = 2(z + z^3/3 + z^5/5 + ...)
//
// for z = u/(n+1), which is below 0.005, so that z^2 shrinks each term at least
// 40000 times. Every term has the sign of u, so the sum keeps its precision
// however small it is.
func lnSeries(u, n Number) wide {
	z := wideFromNumber(u).DivNumber(n.Add(One))
	z2 := z.Mul(z)
	sum, term := z, z
	for k := uint64(3); ; k += 2 {
		term = term.Mul(z2)
		t := term.DivUint64(k)
		slo, _ := sum.Magnitude()
		if _, thi := t.Magnitude(); t.IsZero() || thi < slo-58 {
			sum.inexact = true
			break
		}
		sum = sum.Add(t)
	}
	return sum.Add(sum)
}
