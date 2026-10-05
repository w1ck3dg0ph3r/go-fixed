package fixed

import (
	"math"
	"math/bits"
)

// Add adds v to n.
func (n Number) Add(v Number) Number {
	if !n.valid() {
		return n
	}
	if !v.valid() {
		return v
	}
	lo, carry := bits.Add64(n.Lo, v.Lo, 0)
	hi := n.Hi + v.Hi + carry
	ovf := ((n.Hi ^ hi) &^ (n.Hi ^ v.Hi)) >> 63
	if ovf > 0 || !inBounds(lo, hi) {
		return Overflow
	}
	return Number{lo, hi}
}

// AddInt adds the integer v to n.
func (n Number) AddInt(v int) Number {
	return n.Add(FromInt(v))
}

// Sub subtracts v from n.
func (n Number) Sub(v Number) Number {
	if !n.valid() {
		return n
	}
	if !v.valid() {
		return v
	}
	lo, borrow := bits.Sub64(n.Lo, v.Lo, 0)
	hi := n.Hi - v.Hi - borrow
	ovf := ((n.Hi ^ v.Hi) & (n.Hi ^ hi)) >> 63
	if ovf > 0 || !inBounds(lo, hi) {
		return Overflow
	}
	return Number{lo, hi}
}

// SubInt subtracts the integer v from n.
func (n Number) SubInt(v int) Number {
	return n.Sub(FromInt(v))
}

func mulCheck(n, v Number) (Number, bool) {
	if !n.valid() {
		return n, false
	}
	if !v.valid() {
		return v, false
	}
	if n.IsZero() || v.IsZero() {
		return Number{}, false
	}
	return Number{}, true
}

// Mul multiplies n by v, truncating the product to scale.
//
// Suits intermediate steps of a calculation, with the rounding that matters
// done once by [Number.Round]. Use [Number.MulRound] to round the product
// itself.
func (n Number) Mul(v Number) Number {
	return n.MulRound(v, RoundZero)
}

// MulRound multiplies n by v, rounding the product to scale according to mode.
func (n Number) MulRound(v Number, mode RoundingMode) Number {
	if r, ok := mulCheck(n, v); !ok {
		return r
	}

	neg := n.IsNegative() != v.IsNegative()
	a := n.Abs()
	b := v.Abs()
	lo, hi := mul128(a, b)

	q, r := quoRem256ByScale(hi, lo)
	if q == Overflow {
		return Overflow
	}

	if r != 0 && mode != RoundZero {
		const half = uint64(pow10u64Scale) / 2
		var ovf bool
		q, ovf = add64Checked(q, mode.increment(boolu64(neg), boolu64(r > half), boolu64(r == half), q.Lo&1))
		if ovf {
			return Overflow
		}
	}

	if q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow
	}

	if neg {
		return q.Neg()
	}
	return q
}

// MulExact multiplies n by v.
// Returns NaN and ErrInexact if the product has more than Scale decimal places.
// Other failures return a sentinel and a nil error, as in [Number.Mul].
func (n Number) MulExact(v Number) (Number, error) {
	if r, ok := mulCheck(n, v); !ok {
		return r, nil
	}

	neg := n.IsNegative() != v.IsNegative()
	a := n.Abs()
	b := v.Abs()
	lo, hi := mul128(a, b)

	q, r := quoRem256ByScale(hi, lo)
	if q == Overflow {
		return Overflow, nil
	}

	if q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow, nil
	}

	if r != 0 {
		return NaN, ErrInexact
	}

	if neg {
		return q.Neg(), nil
	}
	return q, nil
}

// MulInt multiplies n by the integer v.
//
// The product is exact, so unlike Mul it needs no rounding.
func (n Number) MulInt(v int) Number {
	if !n.valid() {
		return n
	}

	neg := n.IsNegative() != (v < 0)
	q, ovf := mul64Checked(n.Abs(), abs64(int64(v)))
	if ovf || q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow
	}

	if neg {
		return q.Neg()
	}
	return q
}

func divCheck(n, v Number) (Number, bool) {
	if !n.valid() {
		return n, false
	}
	if !v.valid() {
		return v, false
	}
	if v.IsZero() {
		return DivisionByZero, false
	}
	if n.IsZero() {
		return Zero, false
	}
	return Number{}, true
}

// QuoRem divides n by v, returning the quotient truncated towards zero and the
// remainder left over.
//
// This is not integer division: the quotient keeps all Scale fractional
// digits, so the remainder is not a modulo but the part of n that quotient was
// too narrow to represent. It is always below one unit in the last place of
// the quotient scaled back up, and it takes the sign of n.
//
// The exact remainder can have up to 2*Scale fractional digits, so it is
// rounded away from zero to Scale. That keeps q*v, truncated, plus r equal to n.
func (n Number) QuoRem(v Number) (Number, Number) {
	if r, ok := divCheck(n, v); !ok {
		return r, r
	}

	neg := n.IsNegative() != v.IsNegative()
	a := n.Abs()
	b := v.Abs()

	hi, lo := mul128By64(a, pow10u64Scale)
	q, r := quoRem192By128(hi, lo, b)
	if q == Overflow || q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow, Overflow
	}
	r, discarded := quoRem128ByScale(r)
	if discarded != 0 {
		r = add64(r, 1)
	}

	if neg {
		q = q.Neg()
	}
	if n.IsNegative() {
		r = r.Neg()
	}
	return q, r
}

// QuoRemInt divides n by the integer v, returning the quotient truncated
// towards zero and the remainder left over, as [Number.QuoRem] does. The
// remainder is exact and below v units in the last place.
func (n Number) QuoRemInt(v int) (Number, Number) {
	if !n.valid() {
		return n, n
	}
	if v == 0 {
		return DivisionByZero, DivisionByZero
	}

	q, r := quoRem64(n.Abs(), abs64(int64(v)))
	rem := Number{Lo: r}

	if n.IsNegative() != (v < 0) {
		q = q.Neg()
	}
	if n.IsNegative() {
		rem = rem.Neg()
	}
	return q, rem
}

// Div divides n by v, truncating the quotient to scale.
//
// Suits intermediate steps of a calculation, with the rounding that matters
// done once by [Number.Round]. Use [Number.DivRound] to round the quotient
// itself.
func (n Number) Div(v Number) Number {
	return n.DivRound(v, RoundZero)
}

// DivRound divides n by v, rounding the quotient to scale according to mode.
func (n Number) DivRound(v Number, mode RoundingMode) Number {
	if r, ok := divCheck(n, v); !ok {
		return r
	}

	neg := n.IsNegative() != v.IsNegative()
	a := n.Abs()
	b := v.Abs()

	hi, lo := mul128By64(a, pow10u64Scale)
	q, r := quoRem192By128(hi, lo, b)
	if q == Overflow || q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow
	}

	// Rounding up cannot overflow: a quotient of the largest number leaves no
	// remainder.
	if !r.IsZero() && mode != RoundZero {
		// The remainder is compared with what is left of the divisor, b-r,
		// through the borrow, so that no branch depends on it.
		h := sub128(b, r)
		_, borrow := bits.Sub64(h.Lo, r.Lo, 0)
		_, above := bits.Sub64(h.Hi, r.Hi, borrow)
		q = add64(q, mode.increment(boolu64(neg), above, boolu64(h == r), q.Lo&1))
	}

	if neg {
		return q.Neg()
	}
	return q
}

// DivInt divides n by the integer v, truncating the quotient to scale.
// Use [Number.DivRoundInt] to round the quotient itself.
func (n Number) DivInt(v int) Number {
	return n.DivRoundInt(v, RoundZero)
}

// DivRoundInt divides n by the integer v, rounding the quotient to scale
// according to mode.
func (n Number) DivRoundInt(v int, mode RoundingMode) Number {
	if !n.valid() {
		return n
	}
	if v == 0 {
		return DivisionByZero
	}

	neg := n.IsNegative() != (v < 0)
	d := abs64(int64(v))
	q, r := quoRem64(n.Abs(), d)

	// The quotient is no larger than n, and rounding it up cannot overflow:
	// only a divisor of one gives the largest quotient, and that leaves no
	// remainder.
	if r != 0 && mode != RoundZero {
		h := d - r
		q = add64(q, mode.increment(boolu64(neg), boolu64(r > h), boolu64(r == h), q.Lo&1))
	}

	if neg {
		return q.Neg()
	}
	return q
}

// Mod returns the remainder of n divided by v, taking the sign of n.
func (n Number) Mod(v Number) Number {
	if r, ok := divCheck(n, v); !ok {
		return r
	}

	_, r := quoRem192By128(0, n.Abs(), v.Abs())

	if n.IsNegative() {
		return r.Neg()
	}
	return r
}

// ModInt returns the remainder of n divided by the integer v, taking the sign
// of n.
func (n Number) ModInt(v int) Number {
	if !n.valid() {
		return n
	}
	if v == 0 {
		return DivisionByZero
	}

	// Only the integer part of n can hold multiples of v, so it is reduced
	// modulo v and the fraction added back.
	i, f := quoRem128ByScale(n.Abs())
	r := add64(mul64(Number{Lo: i.Lo % abs64(int64(v))}, pow10u64Scale), f)

	if n.IsNegative() {
		return r.Neg()
	}
	return r
}

// Shift returns n scaled by a power of ten, n*10^k.
//
// A right shift (k > 0) is exact and returns Overflow when the result no longer fits.
// A left shift (k < 0) discards digits below the last place, truncating towards zero.
func (n Number) Shift(k int) Number {
	if !n.valid() {
		return n
	}
	if n.IsZero() || k == 0 {
		return n
	}

	a := n.Abs()

	if k > 0 {
		if k >= Precision {
			return Overflow
		}
		var ovf bool
		if k > Scale {
			if a, ovf = mul64Checked(a, pow10u64Scale); ovf {
				return Overflow
			}
			k -= Scale
		}
		if a, ovf = mul64Checked(a, pow10u64[k]); ovf {
			return Overflow
		}
		if a.Hi > maxPosHi64 || (a.Hi == maxPosHi64 && a.Lo > maxPosLo64) {
			return Overflow
		}
	} else {
		if k <= -Precision {
			return Zero
		}
		k = -k
		if k > Scale {
			a, _ = quoRem128ByScale(a)
			k -= Scale
		}
		a, _ = quoRem128ByPow10(a, k)
	}

	if n.IsNegative() {
		return a.Neg()
	}
	return a
}

// leadingZeroes128 returns the number of leading zero bits in v.
func leadingZeroes128(v Number) uint {
	if v.Hi != 0 {
		return uint(bits.LeadingZeros64(v.Hi))
	} else {
		return uint(64 + bits.LeadingZeros64(v.Lo))
	}
}

// shiftLeft128 shifts v left by n bits, where n must be below 128.
func shiftLeft128(v Number, n uint) Number {
	if n >= 64 {
		return Number{Lo: 0, Hi: v.Lo << (n - 64)}
	}
	return Number{Lo: v.Lo << n, Hi: v.Hi<<n | v.Lo>>(64-n)}
}

// shiftRight128 shifts v right by n bits, where n must be below 128.
func shiftRight128(v Number, n uint) Number {
	if n >= 64 {
		return Number{Lo: v.Hi >> (n - 64), Hi: 0}
	}
	return Number{Lo: v.Hi<<(64-n) | v.Lo>>n, Hi: v.Hi >> n}
}

// shiftRight192 returns the low 64 bits of the 192-bit value hi:lo shifted
// right by n, where n must not exceed 128.
func shiftRight192(hi uint64, lo Number, n uint) uint64 {
	switch {
	case n == 0:
		return lo.Lo
	case n < 64:
		return lo.Lo>>n | lo.Hi<<(64-n)
	case n == 64:
		return lo.Hi
	case n < 128:
		return lo.Hi>>(n-64) | hi<<(128-n)
	default:
		return hi
	}
}

// add64 adds a to v, wrapping on overflow.
func add64(v Number, a uint64) Number {
	lo, carry := bits.Add64(v.Lo, a, 0)
	hi := v.Hi + carry
	return Number{lo, hi}
}

// add64Checked adds a to v, reporting whether the sum overflows 128 bits.
func add64Checked(v Number, a uint64) (Number, bool) {
	lo, carry := bits.Add64(v.Lo, a, 0)
	hi, carry := bits.Add64(v.Hi, 0, carry)
	if carry != 0 {
		return Number{}, true
	}
	return Number{lo, hi}, false
}

// add128 adds a to v, wrapping on overflow.
func add128(v Number, a Number) Number {
	lo, carry := bits.Add64(v.Lo, a.Lo, 0)
	hi, _ := bits.Add64(v.Hi, a.Hi, carry)
	return Number{lo, hi}
}

// add128Checked adds a to v, reporting whether the sum overflows 128 bits.
func add128Checked(v Number, a Number) (Number, bool) {
	lo, carry := bits.Add64(v.Lo, a.Lo, 0)
	hi, carry := bits.Add64(v.Hi, a.Hi, carry)
	if carry != 0 {
		return Number{}, true
	}
	return Number{lo, hi}, false
}

// sub64 subtracts a from v, wrapping on underflow.
func sub64(v Number, a uint64) Number {
	lo, carry := bits.Sub64(v.Lo, a, 0)
	hi := v.Hi - carry
	return Number{lo, hi}
}

// sub128 subtracts b from a, wrapping on underflow.
func sub128(a, b Number) Number {
	lo, carry := bits.Sub64(a.Lo, b.Lo, 0)
	hi, _ := bits.Sub64(a.Hi, b.Hi, carry)
	return Number{lo, hi}
}

// sub128Checked subtracts b from a, reporting whether the result underflows.
func sub128Checked(a, b Number) (Number, bool) {
	lo, carry := bits.Sub64(a.Lo, b.Lo, 0)
	hi, carry := bits.Sub64(a.Hi, b.Hi, carry)
	if carry != 0 {
		return Number{}, true
	}
	return Number{lo, hi}, false
}

// abs64 returns the magnitude of k, which fits even for the smallest int64.
func abs64(v int64) uint64 {
	s := uint64(v >> 63)
	return (uint64(v) ^ s) - s
}

// neg256 negates the 256-bit value hi:lo.
func neg256(hi, lo Number) (Number, Number) {
	lo.Lo, lo.Hi, hi.Lo, hi.Hi = ^lo.Lo, ^lo.Hi, ^hi.Lo, ^hi.Hi
	var carry uint64
	lo.Lo, carry = bits.Add64(lo.Lo, 1, 0)
	lo.Hi, carry = bits.Add64(lo.Hi, 0, carry)
	hi.Lo, carry = bits.Add64(hi.Lo, 0, carry)
	hi.Hi, _ = bits.Add64(hi.Hi, 0, carry)
	return hi, lo
}

// mul64 multiplies v by m, keeping the low 128 bits of the product.
func mul64(v Number, m uint64) Number {
	hi, lo := bits.Mul64(v.Lo, m)
	_, lo2 := bits.Mul64(v.Hi, m)
	hi, _ = bits.Add64(hi, lo2, 0)
	return Number{lo, hi}
}

// mul64Checked multiplies v by m, reporting whether the product overflows
// 128 bits.
func mul64Checked(v Number, m uint64) (Number, bool) {
	hi, lo := bits.Mul64(v.Lo, m)
	hi2, lo2 := bits.Mul64(v.Hi, m)
	hi, carry := bits.Add64(hi, lo2, 0)
	if hi2 != 0 || carry != 0 {
		return Number{}, true
	}
	return Number{lo, hi}, false
}

// mul128 multiplies a by b, returning the low and high halves of the 256-bit
// product.
func mul128(a, b Number) (lo, hi Number) {
	// This is the schoolbook algorithm. A Karatsuba variant saves one Mul64 but
	// needs the 64-bit limb sums a.Lo+a.Hi and b.Lo+b.Hi, and carrying those
	// costs more than the multiplication it saves.

	llhi, lllo := bits.Mul64(a.Lo, b.Lo)
	lhhi, lhlo := bits.Mul64(a.Lo, b.Hi)
	hlhi, hllo := bits.Mul64(a.Hi, b.Lo)
	hhhi, hhlo := bits.Mul64(a.Hi, b.Hi)

	r1, c := bits.Add64(llhi, lhlo, 0)
	r2, c := bits.Add64(lhhi, hhlo, c)
	r3, _ := bits.Add64(hhhi, 0, c)

	r1, c = bits.Add64(r1, hllo, 0)
	r2, c = bits.Add64(r2, hlhi, c)
	r3, _ = bits.Add64(r3, 0, c)

	return Number{lllo, r1}, Number{r2, r3}
}

// mul128By64 multiplies the 128-bit a by the 64-bit b and returns the 192-bit
// product.
func mul128By64(a Number, b uint64) (hi uint64, lo Number) {
	h0, l0 := bits.Mul64(a.Lo, b)
	h1, l1 := bits.Mul64(a.Hi, b)
	mid, carry := bits.Add64(l1, h0, 0)
	return h1 + carry, Number{Lo: l0, Hi: mid}
}

// quoRem64 divides the 128-bit n by d.
func quoRem64(n Number, d uint64) (q Number, r uint64) {
	s := uint(bits.LeadingZeros64(d))
	d <<= s
	v := reciprocal64(d)
	u2 := n.Hi >> (64 - s)
	u1 := n.Hi<<s | n.Lo>>(64-s)
	u0 := n.Lo << s
	if u2 == 0 && u1 < d {
		q.Lo, r = divMollerGranlund(u1, u0, d, v)
		return q, r >> s
	}
	q.Hi, r = divMollerGranlund(u2, u1, d, v)
	q.Lo, r = divMollerGranlund(r, u0, d, v)
	return q, r >> s
}

// quoRem192By64 divides the 192-bit value hi:mid:lo by d, returning the
// 128-bit quotient q1:q0 and remainder r.
//
// The quotient must fit in 128 bits, i.e. hi < d.
func quoRem192By64(hi, mid, lo, d uint64) (q0, q1, r uint64) {
	// hi < d, so normalizing shifts nothing out of hi.
	s := uint(bits.LeadingZeros64(d))
	d <<= s
	a2 := hi<<s | mid>>(64-s)
	a1 := mid<<s | lo>>(64-s)
	a0 := lo << s

	v := reciprocal64(d)
	q1, r = divMollerGranlund(a2, a1, d, v)
	q0, r = divMollerGranlund(r, a0, d, v)
	return q0, q1, r >> s
}

// quoRem256By64 divides the 256-bit value hi:lo by d.
//
// Returns Overflow if the quotient does not fit in 128 bits.
func quoRem256By64(hi, lo Number, d uint64) (q Number, r uint64) {
	if hi.Hi == 0 {
		if hi.Lo == 0 {
			return quoRem64(lo, d)
		}
		if hi.Lo >= d {
			return Overflow, 0
		}
		q0, q1, rr := quoRem192By64(hi.Lo, lo.Hi, lo.Lo, d)
		return Number{q0, q1}, rr
	}
	// The dividend is at least 2^192 and d below 2^64.
	return Overflow, 0
}

// quoRem192By128 divides the 192-bit value hi:lo by d.
//
// The quotient must fit in 128 bits, i.e. hi < d.
func quoRem192By128(hi uint64, lo Number, d Number) (Number, Number) {
	if d.Hi == 0 {
		// Dividend fits in 192 bits: one fewer Div64 call than the general 4-limb
		// quoRem256By64.
		if hi >= d.Lo {
			return Overflow, Number{}
		}
		q0, q1, r := quoRem192By64(hi, lo.Hi, lo.Lo, d.Lo)
		return Number{q0, q1}, Number{Lo: r}
	}

	a0, a1, a2, a3 := lo.Lo, lo.Hi, hi, uint64(0)
	v0, v1 := d.Lo, d.Hi

	// Left-shift divisor and dividend to normalize. a3 starts at 0 since the
	// dividend is only 3 limbs wide; normalizing can still shift bits out of the
	// top of a2 into it.
	s := uint(bits.LeadingZeros64(v1))
	if s > 0 {
		v1 = v1<<s | v0>>(64-s)
		v0 <<= s
		a3 = a2 >> (64 - s)
		a2 = a2<<s | a1>>(64-s)
		a1 = a1<<s | a0>>(64-s)
		a0 <<= s
	}

	divisor := Number{v0, v1}
	rcp := reciprocal64(v1)
	var q0, q1 uint64

	// A single quotient digit is only enough while [a2,a1,a0] < v*2^64, which
	// a2 == v1 does not guarantee, so that case needs two digits.
	//
	// a3 < 2^s <= v1 always, as a3 is the shifted-out overflow from a single
	// 64-bit limb, so the saturation and k=2 cases here are unreachable for this
	// digit - see the trailing digit for the general case.
	if a3 > 0 || a2 >= v1 {
		tq, r := divMollerGranlund(a3, a2, v1, rcp)

		c1h, c1l := bits.Mul64(tq, v0)
		c1 := Number{c1l, c1h}
		c2 := Number{a1, r}

		var rem Number
		if cmpUnsigned128(c1, c2) > 0 {
			delta := sub128(c1, c2)
			tq--
			rem = sub128(divisor, delta)
		} else {
			rem = sub128(c2, c1)
		}

		q1 = tq
		a2, a1 = rem.Hi, rem.Lo
	}

	{
		var tq, r, rCarry uint64
		if a2 == v1 {
			tq = math.MaxUint64
			r, rCarry = bits.Add64(a1, v1, 0)
		} else {
			tq, r = divMollerGranlund(a2, a1, v1, rcp)
		}

		c1h, c1l := bits.Mul64(tq, v0)
		c1 := Number{c1l, c1h}
		c2 := Number{a0, r}

		var rem Number
		if rCarry == 0 && cmpUnsigned128(c1, c2) > 0 {
			delta := sub128(c1, c2)
			tq--
			if cmpUnsigned128(delta, divisor) > 0 {
				tq--
				delta = sub128(delta, divisor)
			}
			rem = sub128(divisor, delta)
		} else {
			rem = sub128(c2, c1)
		}

		q0 = tq
		a1, a0 = rem.Hi, rem.Lo
	}

	// Unnormalize remainder back to original scale.
	if s > 0 {
		a0 = a0>>s | a1<<(64-s)
		a1 >>= s
	}

	return Number{q0, q1}, Number{a0, a1}
}

// divMollerGranlund divides u1:u0 by the normalized d with reciprocal v.
//
// The quotient must fit in one word, i.e. u1 < d.
//
// Dividing by the invariant replaces the hardware divide with a multiplication
// by a precomputed reciprocal, the method of Granlund & Montgomery, "Division
// by Invariant Integers using Multiplication" (1994), in the two words by one
// form given by Möller & Granlund, "Improved Division by Invariant Integers"
// (2011).
func divMollerGranlund(u1, u0, d, v uint64) (q, r uint64) {
	q1, q0 := bits.Mul64(v, u1)
	q0, carry := bits.Add64(q0, u0, 0)
	q1, _ = bits.Add64(q1, u1, carry)
	q1++
	r = u0 - q1*d
	if r > q0 {
		q1--
		r += d
	}
	if r >= d {
		q1++
		r -= d
	}
	return q1, r
}

// pow10Reciprocals holds power of ten divisors normalized by a left shift,
// with its reciprocals for [divMollerGranlund].
//
// Reciprocal is computed as floor((2^128-1)/d)-2^64 (where d is normalized).
//
//	s := bits.LeadingZeros64(i)
//	d := d << s
//	v := bits.Div64(^d, ^uint64(0), d)
var pow10Reciprocals = [...]pow10Reciprocal{
	{0x8000000000000000, 0xffffffffffffffff, 63},
	{0xa000000000000000, 0x9999999999999999, 60},
	{0xc800000000000000, 0x47ae147ae147ae14, 57},
	{0xfa00000000000000, 0x0624dd2f1a9fbe76, 54},
	{0x9c40000000000000, 0xa36e2eb1c432ca57, 50},
	{0xc350000000000000, 0x4f8b588e368f0846, 47},
	{0xf424000000000000, 0x0c6f7a0b5ed8d36b, 44},
	{0x9896800000000000, 0xad7f29abcaf48578, 40},
	{0xbebc200000000000, 0x5798ee2308c39df9, 37},
	{0xee6b280000000000, 0x12e0be826d694b2e, 34},
	{0x9502f90000000000, 0xb7cdfd9d7bdbab7d, 30},
	{0xba43b74000000000, 0x5fd7fe17964955fd, 27},
	{0xe8d4a51000000000, 0x19799812dea11197, 24},
	{0x9184e72a00000000, 0xc25c268497681c26, 20},
	{0xb5e620f480000000, 0x6849b86a12b9b01e, 17},
	{0xe35fa931a0000000, 0x203af9ee756159b2, 14},
	{0x8e1bc9bf04000000, 0xcd2b297d889bc2b6, 10},
	{0xb1a2bc2ec5000000, 0x70ef54646d496892, 7},
	{0xde0b6b3a76400000, 0x2725dd1d243aba0e, 4},
	{0x8ac7230489e80000, 0xd83c94fb6d2ac34a, 0},
}

type pow10Reciprocal struct {
	d uint64 // Normalized divisor.
	v uint64 // Reciprocal.
	s uint   // Shift used to normalize the divisor.
}

// quoRem128ByPow10 divides the 128-bit n by 10^k, where 0 <= k <= Scale.
func quoRem128ByPow10(n Number, k int) (q Number, r uint64) {
	p := pow10Reciprocals[k]
	u2 := n.Hi >> (64 - p.s)
	u1 := n.Hi<<p.s | n.Lo>>(64-p.s)
	u0 := n.Lo << p.s
	if u2 == 0 && u1 < p.d {
		q.Lo, r = divMollerGranlund(u1, u0, p.d, p.v)
		return q, r >> p.s
	}
	q.Hi, r = divMollerGranlund(u2, u1, p.d, p.v)
	q.Lo, r = divMollerGranlund(r, u0, p.d, p.v)
	return q, r >> p.s
}

// quoRem128ByScale divides the 128-bit n by 10^Scale.
func quoRem128ByScale(n Number) (q Number, r uint64) {
	p := pow10Reciprocals[Scale]
	if n.Hi < p.d {
		q.Lo, r = divMollerGranlund(n.Hi, n.Lo, p.d, p.v)
		return q, r
	}
	q.Hi, r = divMollerGranlund(0, n.Hi, p.d, p.v)
	q.Lo, r = divMollerGranlund(r, n.Lo, p.d, p.v)
	return q, r
}

// quoRem256ByScale divides the 256-bit value hi:lo by 10^Scale.
//
// Returns Overflow if the quotient does not fit in 128 bits.
func quoRem256ByScale(hi, lo Number) (q Number, r uint64) {
	p := pow10Reciprocals[Scale]
	if hi.Hi == 0 {
		if hi.Lo == 0 {
			return quoRem128ByScale(lo)
		}
		if hi.Lo >= p.d {
			return Overflow, 0
		}
		q1, r := divMollerGranlund(hi.Lo, lo.Hi, p.d, p.v)
		q0, r := divMollerGranlund(r, lo.Lo, p.d, p.v)
		return Number{q0, q1}, r
	}
	// The dividend is at least 2^192 and 10^Scale below 2^64.
	return Overflow, 0
}

// reciprocal64 returns the reciprocal of the arbitrary normalized divisor d
// for [divMollerGranlund], floor((2^128-1)/d)-2^64, without a hardware divide.
//
// This is Algorithm 2 of Möller & Granlund: a 9-bit table seed refined by
// Newton steps to 11, 21, 34 and finally 64 bits.
func reciprocal64(d uint64) uint64 {
	d0 := d & 1
	d40 := d>>24 + 1
	d63 := d>>1 + d0
	v0 := uint64(reciprocalTable[d>>55-256])
	v1 := v0<<11 - (v0*v0*d40)>>40 - 1
	v2 := v1<<13 + (v1*(1<<60-v1*d40))>>47
	e := (v2>>1)&-d0 - v2*d63
	h, _ := bits.Mul64(v2, e)
	v3 := v2<<31 + h>>1
	ph, pl := bits.Mul64(v3, d)
	_, carry := bits.Add64(pl, d, 0)
	ph, _ = bits.Add64(ph, d, carry)
	return v3 - ph
}

// reciprocalTable seeds [reciprocal64] with floor((2^19-3*2^8)/d9) for the
// top 9 bits d9 of the divisor.
//
//	reciprocalTable[i] = uint16((1<<19 - 3<<8) / (i + 256))
var reciprocalTable = [...]uint16{
	0x07fd, 0x07f5, 0x07ed, 0x07e5, 0x07dd, 0x07d5, 0x07ce, 0x07c6,
	0x07bf, 0x07b7, 0x07b0, 0x07a8, 0x07a1, 0x079a, 0x0792, 0x078b,
	0x0784, 0x077d, 0x0776, 0x076f, 0x0768, 0x0761, 0x075b, 0x0754,
	0x074d, 0x0747, 0x0740, 0x0739, 0x0733, 0x072c, 0x0726, 0x0720,
	0x0719, 0x0713, 0x070d, 0x0707, 0x0700, 0x06fa, 0x06f4, 0x06ee,
	0x06e8, 0x06e2, 0x06dc, 0x06d6, 0x06d1, 0x06cb, 0x06c5, 0x06bf,
	0x06ba, 0x06b4, 0x06ae, 0x06a9, 0x06a3, 0x069e, 0x0698, 0x0693,
	0x068d, 0x0688, 0x0683, 0x067d, 0x0678, 0x0673, 0x066e, 0x0669,
	0x0664, 0x065e, 0x0659, 0x0654, 0x064f, 0x064a, 0x0645, 0x0640,
	0x063c, 0x0637, 0x0632, 0x062d, 0x0628, 0x0624, 0x061f, 0x061a,
	0x0616, 0x0611, 0x060c, 0x0608, 0x0603, 0x05ff, 0x05fa, 0x05f6,
	0x05f1, 0x05ed, 0x05e9, 0x05e4, 0x05e0, 0x05dc, 0x05d7, 0x05d3,
	0x05cf, 0x05cb, 0x05c6, 0x05c2, 0x05be, 0x05ba, 0x05b6, 0x05b2,
	0x05ae, 0x05aa, 0x05a6, 0x05a2, 0x059e, 0x059a, 0x0596, 0x0592,
	0x058e, 0x058a, 0x0586, 0x0583, 0x057f, 0x057b, 0x0577, 0x0574,
	0x0570, 0x056c, 0x0568, 0x0565, 0x0561, 0x055e, 0x055a, 0x0556,
	0x0553, 0x054f, 0x054c, 0x0548, 0x0545, 0x0541, 0x053e, 0x053a,
	0x0537, 0x0534, 0x0530, 0x052d, 0x052a, 0x0526, 0x0523, 0x0520,
	0x051c, 0x0519, 0x0516, 0x0513, 0x050f, 0x050c, 0x0509, 0x0506,
	0x0503, 0x0500, 0x04fc, 0x04f9, 0x04f6, 0x04f3, 0x04f0, 0x04ed,
	0x04ea, 0x04e7, 0x04e4, 0x04e1, 0x04de, 0x04db, 0x04d8, 0x04d5,
	0x04d2, 0x04cf, 0x04cc, 0x04ca, 0x04c7, 0x04c4, 0x04c1, 0x04be,
	0x04bb, 0x04b9, 0x04b6, 0x04b3, 0x04b0, 0x04ad, 0x04ab, 0x04a8,
	0x04a5, 0x04a3, 0x04a0, 0x049d, 0x049b, 0x0498, 0x0495, 0x0493,
	0x0490, 0x048d, 0x048b, 0x0488, 0x0486, 0x0483, 0x0481, 0x047e,
	0x047c, 0x0479, 0x0477, 0x0474, 0x0472, 0x046f, 0x046d, 0x046a,
	0x0468, 0x0465, 0x0463, 0x0461, 0x045e, 0x045c, 0x0459, 0x0457,
	0x0455, 0x0452, 0x0450, 0x044e, 0x044b, 0x0449, 0x0447, 0x0444,
	0x0442, 0x0440, 0x043e, 0x043b, 0x0439, 0x0437, 0x0435, 0x0432,
	0x0430, 0x042e, 0x042c, 0x042a, 0x0428, 0x0425, 0x0423, 0x0421,
	0x041f, 0x041d, 0x041b, 0x0419, 0x0417, 0x0414, 0x0412, 0x0410,
	0x040e, 0x040c, 0x040a, 0x0408, 0x0406, 0x0404, 0x0402, 0x0400,
}

// cmpUnsigned128 compares a and b as unsigned 128-bit values.
func cmpUnsigned128(a, b Number) int {
	if a.Hi < b.Hi {
		return -1
	}
	if a.Hi > b.Hi {
		return 1
	}
	if a.Lo < b.Lo {
		return -1
	}
	if a.Lo > b.Lo {
		return 1
	}
	return 0
}

// cmpUnsigned256 compares ahi:alo and bhi:blo as unsigned 256-bit values.
func cmpUnsigned256(ahi, alo, bhi, blo Number) int {
	if ahi.Hi < bhi.Hi {
		return -1
	}
	if ahi.Hi > bhi.Hi {
		return 1
	}
	if ahi.Lo < bhi.Lo {
		return -1
	}
	if ahi.Lo > bhi.Lo {
		return 1
	}
	if alo.Hi < blo.Hi {
		return -1
	}
	if alo.Hi > blo.Hi {
		return 1
	}
	if alo.Lo < blo.Lo {
		return -1
	}
	if alo.Lo > blo.Lo {
		return 1
	}
	return 0
}

func boolu64(b bool) (v uint64) {
	if b {
		v = 1
	}
	return
}
