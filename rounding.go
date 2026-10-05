package fixed

// RoundingMode selects how discarded digits are rounded.
type RoundingMode uint8

const (
	RoundZero     RoundingMode = iota // Towards zero (truncate).
	RoundInf                          // Away from zero.
	RoundFloor                        // Towards negative infinity.
	RoundCeil                         // Towards positive infinity.
	RoundHalfZero                     // To nearest, ties towards zero.
	RoundHalfInf                      // To nearest, ties away from zero.
	RoundBank                         // To nearest, ties to even.
)

// Round rounds n to scale decimal places according to mode. A negative scale
// rounds to tens, hundreds and so on.
func (n Number) Round(scale int, mode RoundingMode) Number {
	if !n.valid() || scale >= Scale {
		return n
	}
	if scale < 0 {
		return n.roundWide(scale, mode)
	}

	factor := negpow10[scale].Lo
	q, r := quoRem128ByPow10(n.Abs(), Scale-scale)
	if r == 0 {
		return n
	}

	half := factor / 2
	q = add64(q, mode.increment(boolu64(n.IsNegative()), boolu64(r > half), boolu64(r == half), q.Lo&1))
	result := mul64(q, factor)
	if n.IsNegative() {
		result = result.Neg()
	}
	if !inBounds(result.Lo, result.Hi) {
		return Overflow
	}
	return result
}

// roundWide rounds to a negative scale, where the rounding unit 10^(Scale-scale)
// does not fit in 64 bits.
func (n Number) roundWide(scale int, mode RoundingMode) Number {
	if n.IsZero() {
		return n
	}
	if scale < -Scale {
		if mode.increment(boolu64(n.IsNegative()), 0, 0, 0) != 0 {
			return Overflow
		}
		return Zero
	}

	// The unit is 10^k * 10^Scale, and 10^k is even, so half of it is
	// half*10^Scale exactly.
	k := -scale
	q1, r1 := quoRem128ByScale(n.Abs())
	q, r2 := quoRem128ByPow10(q1, k)
	if r1 == 0 && r2 == 0 {
		return n
	}

	half := pow10u64[k] / 2
	above := boolu64(r2 > half) | boolu64(r2 == half)&boolu64(r1 != 0)
	tie := boolu64(r2 == half) & boolu64(r1 == 0)
	q = add64(q, mode.increment(boolu64(n.IsNegative()), above, tie, q.Lo&1))
	result := mul64(mul64(q, pow10u64[k]), pow10u64Scale)
	if n.IsNegative() {
		result = result.Neg()
	}
	if !inBounds(result.Lo, result.Hi) {
		return Overflow
	}
	return result
}

// increment returns 1 if the truncated magnitude must be increased and 0
// otherwise, given the sign and how the discarded remainder compares to half
// the unit, each passed as 0 or 1.
//
// Whether the remainder is above half is a coin flip for varied operands, so
// callers add the result unconditionally rather than branch on it.
func (m RoundingMode) increment(neg, above, tie, odd uint64) uint64 {
	switch m {
	case RoundInf:
		return 1
	case RoundFloor:
		return neg
	case RoundCeil:
		return neg ^ 1
	case RoundHalfZero:
		return above
	case RoundHalfInf:
		return above | tie
	case RoundBank:
		return above | tie&odd
	}
	return 0
}
