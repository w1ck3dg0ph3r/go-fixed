package fixed

import (
	"math/bits"
)

// Number is a 128-bit fixed point decimal number.
type Number struct {
	Lo, Hi uint64
}

var (
	Zero = Number{}
	One  = FromInt(1)
	Two  = FromInt(2)
	Half = Number{5e18, 0}

	E  = MustFromString("2.7182818284590452354")
	Pi = MustFromString("3.1415926535897932385")

	Lowest  = Number{maxNegLo64, maxNegHi64}
	Highest = Number{maxPosLo64, maxPosHi64}
)

const (
	Precision = 38
	Scale     = 19

	maxPosHi64 = 0x4b3b4ca85a86c47a
	maxPosLo64 = 0x98a223fffffffff
	maxNegHi64 = 0xb4c4b357a5793b85
	maxNegLo64 = 0xf675ddc000000001

	maxStrLen = Precision + 2
)

// New constructs a [Number] coef*10^exp, e.g. New(12345, -2) is 123.45.
//
// Returns [Overflow] if it does not fit, and [NaN] if it has nonzero digits
// beyond Scale decimal places.
func New(coef int64, exp int) Number {
	if coef == 0 {
		return Zero
	}
	a := uint64(coef)
	if coef < 0 {
		a = -a
	}

	// The coefficient of the result is a*10^k. Checking exp first keeps k from
	// wrapping around for extreme exponents, and 10^Scale is already past the
	// largest number.
	if exp >= Scale {
		return Overflow
	}
	if exp < -2*Scale {
		return NaN
	}
	k := exp + Scale

	var n Number
	if k <= 0 {
		// a is below 10^19, so dividing by that or more always leaves digits.
		if k <= -Scale || a%pow10u64[-k] != 0 {
			return NaN
		}
		n = Number{a / pow10u64[-k], 0}
	} else {
		var ovf bool
		n = Number{a, 0}
		if k >= Scale {
			n, _ = mul64Checked(n, pow10u64Scale)
			k -= Scale
		}
		n, ovf = mul64Checked(n, pow10u64[k])
		if ovf || n.Hi > maxPosHi64 || (n.Hi == maxPosHi64 && n.Lo > maxPosLo64) {
			return Overflow
		}
	}

	if coef < 0 {
		return n.Neg()
	}
	return n
}

// NewWords constructs a [Number] from its low and high words.
//
// Returns the error if the words are not a valid number.
func NewWords(lo, hi uint64) (Number, error) {
	n := Number{lo, hi}
	if err := n.Err(); err != nil {
		return Number{}, err
	}
	return n, nil
}

// IsZero returns true if n is zero.
func (n Number) IsZero() bool {
	return n.Hi == 0 && n.Lo == 0
}

// Sign returns -1, 0 or 1 for negative, zero and positive n.
func (n Number) Sign() int {
	x := n.Hi | n.Lo
	return int(int64(n.Hi)>>63) | int((x|-x)>>63)
}

// IsPositive returns true if n is greater than zero.
func (n Number) IsPositive() bool {
	return !n.IsZero() && n.Hi&(1<<63) == 0
}

// IsNegative returns true if n is less than zero.
func (n Number) IsNegative() bool {
	return n.Hi&(1<<63) > 0
}

// Abs returns the absolute value of n.
func (n Number) Abs() Number {
	// Branchless version only wins on random signs, which is not a typical use
	// case.
	if !n.IsNegative() {
		return n
	}
	hi := ^n.Hi
	lo := ^n.Lo
	lo, carry := bits.Add64(lo, 1, 0)
	hi = hi + carry
	return Number{lo, hi}
}

// Neg returns negated value of n.
func (n Number) Neg() Number {
	if !n.valid() {
		return n
	}
	hi := ^n.Hi
	lo := ^n.Lo
	lo, carry := bits.Add64(lo, 1, 0)
	hi = hi + carry
	return Number{lo, hi}
}

// Cmp returns -1, 0 or 1 if n is less than, equal to or greater than b.
func (n Number) Cmp(b Number) int {
	if n.Hi != b.Hi {
		if int64(n.Hi) < int64(b.Hi) {
			return -1
		}
		return 1
	}
	if n.Lo != b.Lo {
		if n.Lo < b.Lo {
			return -1
		}
		return 1
	}
	return 0
}

// Eq returns true if n is equal to b.
func (n Number) Eq(b Number) bool {
	return n == b
}

// Lt returns true if n is less than b.
func (n Number) Lt(b Number) bool {
	return n.Cmp(b) < 0
}

// Lte returns true if n is less than or equal to b.
func (n Number) Lte(b Number) bool {
	return n.Cmp(b) <= 0
}

// Gt returns true if n is greater than b.
func (n Number) Gt(b Number) bool {
	return n.Cmp(b) > 0
}

// Gte returns true if n is greater than or equal to b.
func (n Number) Gte(b Number) bool {
	return n.Cmp(b) >= 0
}

// Min returns the smallest of the given numbers.
// Returns NaN if called without arguments.
func Min(ns ...Number) Number {
	if len(ns) == 0 {
		return NaN
	}
	res := ns[0]
	if !res.valid() {
		return res
	}
	for _, n := range ns[1:] {
		if !n.valid() {
			return n
		}
		if n.Lt(res) {
			res = n
		}
	}
	return res
}

// Max returns the largest of the given numbers.
// Returns NaN if called without arguments.
func Max(ns ...Number) Number {
	if len(ns) == 0 {
		return NaN
	}
	res := ns[0]
	if !res.valid() {
		return res
	}
	for _, n := range ns[1:] {
		if !n.valid() {
			return n
		}
		if n.Gt(res) {
			res = n
		}
	}
	return res
}

// Sum adds the given numbers in order.
// Returns Overflow if any intermediate sum is out of range.
func Sum(ns ...Number) Number {
	res := Zero
	for _, n := range ns {
		res = res.Add(n)
		if !res.valid() {
			return res
		}
	}
	return res
}

// Avg returns the mean of the given numbers, rounding towards zero.
//
// The sum is accumulated in 256 bits, so it never overflows.
// Returns NaN if called without arguments.
func Avg(ns ...Number) Number {
	if len(ns) == 0 {
		return NaN
	}

	var hi, lo Number
	for _, n := range ns {
		if !n.valid() {
			return n
		}
		ext := -(n.Hi >> 63)
		var carry uint64
		lo.Lo, carry = bits.Add64(lo.Lo, n.Lo, 0)
		lo.Hi, carry = bits.Add64(lo.Hi, n.Hi, carry)
		hi.Lo, carry = bits.Add64(hi.Lo, ext, carry)
		hi.Hi, _ = bits.Add64(hi.Hi, ext, carry)
	}

	neg := hi.Hi&(1<<63) > 0
	if neg {
		hi, lo = neg256(hi, lo)
	}

	q, _ := quoRem256By64(hi, lo, uint64(len(ns)))
	if neg {
		return q.Neg()
	}
	return q
}

func inBounds(lo, hi uint64) bool {
	// The range is symmetric, so a high word strictly between the extremes needs
	// one unsigned comparison. Only the two extreme words need the low.
	if hi+maxPosHi64 < 2*maxPosHi64 {
		return true
	}
	return (hi == maxPosHi64 && lo <= maxPosLo64) || (hi == maxNegHi64 && lo >= maxNegLo64)
}

const pow10u64Scale = 1e19

var pow10u64 = [...]uint64{
	1e0,
	1e1,
	1e2,
	1e3,
	1e4,
	1e5,
	1e6,
	1e7,
	1e8,
	1e9,
	1e10,
	1e11,
	1e12,
	1e13,
	1e14,
	1e15,
	1e16,
	1e17,
	1e18,
	1e19,
}

var pow10f64 = [...]float64{1e0,
	1e1,
	1e2,
	1e3,
	1e4,
	1e5,
	1e6,
	1e7,
	1e8,
	1e9,
	1e10,
	1e11,
	1e12,
	1e13,
	1e14,
	1e15,
	1e16,
	1e17,
	1e18,
	1e19,
}

var negpow10 = [...]Number{
	{1e19, 0},
	{1e18, 0},
	{1e17, 0},
	{1e16, 0},
	{1e15, 0},
	{1e14, 0},
	{1e13, 0},
	{1e12, 0},
	{1e11, 0},
	{1e10, 0},
	{1e9, 0},
	{1e8, 0},
	{1e7, 0},
	{1e6, 0},
	{1e5, 0},
	{1e4, 0},
	{1e3, 0},
	{1e2, 0},
	{1e1, 0},
	{1e0, 0},
}
