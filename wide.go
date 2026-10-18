package fixed

import (
	"math/bits"
	"slices"
)

// wide is a decimal floating point value m*10^e with a 192-bit mantissa, about
// 57 digits, for intermediate results that need more precision than a Number.
//
// Operations truncate the mantissa, so a wide value never exceeds the exact
// one, and inexact records that some nonzero digits were dropped.
type wide struct {
	m       [3]uint64 // Little-endian words.
	e       int
	inexact bool
}

var wideOne = wide{m: [3]uint64{1}}

// wideFromNumber returns the magnitude of n.
func wideFromNumber(n Number) wide {
	a := n.Abs()
	return wide{m: [3]uint64{a.Lo, a.Hi}, e: -Scale}
}

// IsZero reports whether w is zero.
func (a wide) IsZero() bool {
	return a.m[0]|a.m[1]|a.m[2] == 0
}

// Magnitude bounds the decimal exponent of w: 10^lo <= w < 10^hi for a nonzero w.
func (a wide) Magnitude() (lo, hi int) {
	b := bitLenWords(a.m[:])
	return (b-1)*1233>>12 + a.e, b*1233>>12 + 1 + a.e
}

// Normalize scales the mantissa of w up to at least 56 digits, so that later
// truncation keeps as many of them as possible.
func (a wide) Normalize() wide {
	if a.IsZero() {
		return a
	}
	// m*10^d stays below 2^192 for d = floor((192-b)*log10(2)); 1233/4096 is
	// just under log10(2).
	d := (192 - bitLenWords(a.m[:])) * 1233 >> 12
	for d > 0 {
		c := min(d, Scale)
		mulWordsUint64(a.m[:], pow10u64[c])
		a.e -= c
		d -= c
	}
	return a
}

// Add adds a and b, truncating the sum to 192 bits.
func (a wide) Add(b wide) wide {
	if a.IsZero() {
		b.inexact = b.inexact || a.inexact
		return b
	}
	if b.IsZero() {
		a.inexact = a.inexact || b.inexact
		return a
	}

	// With full mantissas, the operand with the smaller exponent is aligned to
	// the other one by dropping its lowest digits.
	a, b = a.Normalize(), b.Normalize()
	if a.e < b.e {
		a, b = b, a
	}
	inexact := a.inexact || b.inexact
	if d := a.e - b.e; d >= 60 {
		inexact = true
		b.m = [3]uint64{}
	} else {
		for d > 0 {
			c := min(d, Scale)
			inexact = divWordsPow10(b.m[:], c) != 0 || inexact
			d -= c
		}
	}

	var s [4]uint64
	var carry uint64
	for i := range a.m {
		s[i], carry = bits.Add64(a.m[i], b.m[i], carry)
	}
	s[3] = carry

	e := a.e
	if carry != 0 {
		inexact = divWordsPow10(s[:], 1) != 0 || inexact
		e++
	}
	return wide{m: [3]uint64{s[0], s[1], s[2]}, e: e, inexact: inexact}
}

// Sub returns |a - b| and whether a is smaller, truncating the difference
// to 192 bits.
func (a wide) Sub(b wide) (wide, bool) {
	if b.IsZero() {
		a.inexact = a.inexact || b.inexact
		return a, false
	}
	if a.IsZero() {
		b.inexact = b.inexact || a.inexact
		return b, true
	}

	// Align as addWide does, then subtract the smaller mantissa.
	a, b = a.Normalize(), b.Normalize()
	neg := false
	if a.e < b.e {
		a, b, neg = b, a, true
	}
	inexact := a.inexact || b.inexact
	if d := a.e - b.e; d >= 60 {
		inexact = true
		b.m = [3]uint64{}
	} else {
		for d > 0 {
			c := min(d, Scale)
			inexact = divWordsPow10(b.m[:], c) != 0 || inexact
			d -= c
		}
	}
	if cmpWords(a.m[:], b.m[:]) < 0 {
		a.m, b.m, neg = b.m, a.m, !neg
	}

	var borrow uint64
	for i := range a.m {
		a.m[i], borrow = bits.Sub64(a.m[i], b.m[i], borrow)
	}
	a.inexact = inexact
	return a, neg
}

// Mul multiplies a by b, truncating the product to 192 bits.
func (a wide) Mul(b wide) wide {
	var p [6]uint64
	for i := range a.m {
		var carry uint64
		for j := range b.m {
			hi, lo := bits.Mul64(a.m[i], b.m[j])
			var c uint64
			lo, c = bits.Add64(lo, p[i+j], 0)
			hi += c
			lo, c = bits.Add64(lo, carry, 0)
			hi += c
			p[i+j] = lo
			carry = hi
		}
		p[i+3] = carry
	}

	return shrinkWords(p, a.e+b.e, a.inexact || b.inexact)
}

// DivUint64 divides w by k, truncating the quotient to 192 bits.
func (a wide) DivUint64(k uint64) wide {
	a = a.Normalize()
	var r uint64
	for i := len(a.m) - 1; i >= 0; i-- {
		a.m[i], r = bits.Div64(r, a.m[i], k)
	}
	a.inexact = a.inexact || r != 0
	return a
}

// DivNumber divides w by the nonzero magnitude c, truncating the quotient
// to 192 bits.
func (a wide) DivNumber(c Number) wide {
	// w/c = (m*10^38/C) * 10^(e-38+Scale) for the coefficient C. Extending the
	// mantissa by 10^38 keeps at least 56 digits in the quotient.
	a = a.Normalize()
	u := [6]uint64{a.m[0], a.m[1], a.m[2]}
	mulWordsUint64(u[:], pow10u64Scale)
	mulWordsUint64(u[:], pow10u64Scale)

	var q [6]uint64
	var r Number
	for i := len(u) - 1; i >= 0; i-- {
		var qi Number
		qi, r = quoRem192By128(r.Hi, Number{u[i], r.Lo}, c)
		q[i] = qi.Lo
	}
	return shrinkWords(q, a.e-Scale, a.inexact || !r.IsZero())
}

// Round returns w rounded to Scale places according to mode, negated if neg is
// set, or Overflow if it does not fit.
func (a wide) Round(neg bool, mode RoundingMode) Number {
	k := a.e + Scale
	if k >= 0 {
		// An integer coefficient, which also covers every inexact value of this
		// size: those keep at least 56 digits and so overflow.
		if a.m[2] != 0 {
			return Overflow
		}
		n := Number{a.m[0], a.m[1]}
		for k > 0 {
			c := min(k, Scale)
			var ovf bool
			if n, ovf = mul64Checked(n, pow10u64[c]); ovf {
				return Overflow
			}
			k -= c
		}
		if n.Hi > maxPosHi64 || (n.Hi == maxPosHi64 && n.Lo > maxPosLo64) {
			return Overflow
		}
		if neg {
			return n.Neg()
		}
		return n
	}

	// Keep one digit below the last place to tell where the remainder lies,
	// and fold everything below it into sticky.
	m := a.m
	sticky := a.inexact
	var guard uint64
	if d := -k - 1; d >= 60 {
		// 192 bits hold under 60 digits, so all of them go.
		sticky = sticky || !a.IsZero()
		m = [3]uint64{}
	} else {
		for d > 0 {
			c := min(d, Scale)
			sticky = divWordsPow10(m[:], c) != 0 || sticky
			d -= c
		}
		guard = divWordsPow10(m[:], 1)
	}
	if m[2] != 0 {
		return Overflow
	}

	q := Number{m[0], m[1]}
	if (guard != 0 || sticky) && mode != RoundZero {
		above := guard > 5 || (guard == 5 && sticky)
		tie := guard == 5 && !sticky
		q = add64(q, mode.increment(boolu64(neg), boolu64(above), boolu64(tie), q.Lo&1))
	}
	if q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow
	}
	if neg {
		return q.Neg()
	}
	return q
}

// RoundSnap rounds an approximation w whose first trusted significant digits
// are right, treating a value within that error of a number with Scale places,
// or of halfway between two, as exactly that.
//
// A true value so close to either without being exactly there is about as
// likely as a run of that many zeros or nines in its digits, while results
// that are exact, such as 4^0.5, land there by construction.
func (a wide) RoundSnap(neg bool, mode RoundingMode, trusted int) Number {
	k := a.e + Scale
	if k >= 0 {
		return a.Round(neg, mode)
	}

	// Keep the trusted digits below the last place, up to Scale of them; with
	// fewer than 10 the snapping would catch too much, so round as usual.
	_, hi := a.Magnitude()
	keep := min(Scale, trusted-hi-Scale, -k)
	if keep < 10 {
		return a.Round(neg, mode)
	}

	m := a.m
	if drop := -k - keep; drop >= 60 {
		m = [3]uint64{}
	} else {
		for drop > 0 {
			c := min(drop, Scale)
			divWordsPow10(m[:], c)
			drop -= c
		}
	}
	g := divWordsPow10(m[:], keep)
	if m[2] != 0 {
		return Overflow
	}

	q := Number{m[0], m[1]}
	unit := pow10u64[keep]
	half := unit / 2
	switch {
	case g == unit-1:
		// Just below the next number, so exactly it.
		q = add64(q, 1)
	case g == 0 && !q.IsZero():
		// Exactly q. A positive result is never exactly zero, so a zero q is
		// left to round below.
	case g == half || g == half-1:
		q = add64(q, mode.increment(boolu64(neg), 0, 1, q.Lo&1))
	case mode != RoundZero:
		q = add64(q, mode.increment(boolu64(neg), boolu64(g > half), 0, q.Lo&1))
	}
	if q.Hi > maxPosHi64 || (q.Hi == maxPosHi64 && q.Lo > maxPosLo64) {
		return Overflow
	}
	if neg {
		return q.Neg()
	}
	return q
}

// wideReciprocal returns 1/a for a nonzero magnitude a.
//
// It is 10^Scale/C for the coefficient C, divided out to at least 54 digits.
// The quotient is exact whenever 1/a is a short enough decimal.
func wideReciprocal(a Number) wide {
	// 10^k/C stays below 2^192 for k = 57+floor((b-1)*log10(2)), where b is the
	// bit length of C; 1233/4096 is just under log10(2).
	b := 128 - int(leadingZeroes128(a))
	k := 57 + (b-1)*1233>>12

	// The dividend 10^k, in words.
	var u [6]uint64
	u[0] = 1
	for left := k; left > 0; {
		c := min(left, Scale)
		mulWordsUint64(u[:], pow10u64[c])
		left -= c
	}

	// Long division by the two-word divisor, one word at a time.
	var q [6]uint64
	var r Number
	for i := len(u) - 1; i >= 0; i-- {
		var qi Number
		qi, r = quoRem192By128(r.Hi, Number{u[i], r.Lo}, a)
		q[i] = qi.Lo
	}

	return wide{m: [3]uint64{q[0], q[1], q[2]}, e: Scale - k, inexact: !r.IsZero()}
}

// bitLenWords returns the number of bits needed for the little-endian words w.
func bitLenWords(w []uint64) int {
	for i, v := range slices.Backward(w) {
		if v != 0 {
			return i*64 + bits.Len64(v)
		}
	}
	return 0
}

// cmpWords compares the little-endian words a and b of the same length.
func cmpWords(a, b []uint64) int {
	b = b[:len(a)]
	for i, v := range slices.Backward(a) {
		if v != b[i] {
			if v < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// shrinkWords returns p*10^e as a wide value, dropping decimal digits until p
// fits in three words.
func shrinkWords(p [6]uint64, e int, inexact bool) wide {
	// Only the words still in use are divided. The estimate errs low, so at
	// most a digit or two more are dropped one at a time.
	if l := bitLenWords(p[:]); l > 192 {
		k := max((l-192)*1233>>12, 1)
		for k > 0 {
			c := min(k, Scale)
			inexact = divWordsPow10(p[:(l+63)/64], c) != 0 || inexact
			e += c
			k -= c
			l = bitLenWords(p[:])
		}
		for l > 192 {
			inexact = divWordsPow10(p[:(l+63)/64], 1) != 0 || inexact
			e++
			l = bitLenWords(p[:])
		}
	}
	return wide{m: [3]uint64{p[0], p[1], p[2]}, e: e, inexact: inexact}
}

// mulWordsUint64 multiplies the little-endian words w by m in place, dropping
// any carry out of the top word.
func mulWordsUint64(w []uint64, m uint64) {
	var carry uint64
	for i := range w {
		hi, lo := bits.Mul64(w[i], m)
		var c uint64
		w[i], c = bits.Add64(lo, carry, 0)
		carry = hi + c
	}
}

// divWordsPow10 divides the little-endian words w by 10^k in place, where
// 0 <= k <= Scale, and returns the remainder.
func divWordsPow10(w []uint64, k int) uint64 {
	p := pow10Reciprocals[k]

	// The divisor is normalized by a left shift, so the dividend is shifted the
	// same way as it is read. The bits shifted out of the top word start the
	// remainder, which stays below the divisor.
	r := w[len(w)-1] >> (64 - p.s)
	for i := len(w) - 1; i >= 0; i-- {
		u := w[i] << p.s
		if i > 0 {
			u |= w[i-1] >> (64 - p.s)
		}
		w[i], r = divMollerGranlund(r, u, p.d, p.v)
	}
	return r >> p.s
}
