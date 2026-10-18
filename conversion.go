package fixed

import (
	"math"
	"math/bits"
	"strconv"
	"unsafe"
)

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// FromInt constructs [Number] from integer v.
func FromInt[T Int](v T) Number {
	i := int64(v)
	hi, lo := bits.Mul64(uint64(i), pow10u64Scale)
	return Number{lo, hi - (uint64(i>>63) & pow10u64Scale)}
}

// Int returns the integer part of n, truncated towards zero.
//
// Returns zero, if n is invalid.
// Wraps, if the result does not fit into an int.
func (n Number) Int() int {
	if !n.Valid() {
		return 0
	}
	return int(intPart(n))
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// FromUint constructs [Number] from unsigned integer v.
//
// Returns [Overflow] if v does not fit into Number.
func FromUint[T Uint](v T) Number {
	if uint64(v) >= pow10u64Scale {
		return Overflow
	}
	hi, lo := bits.Mul64(uint64(v), pow10u64Scale)
	return Number{lo, hi}
}

// Uint returns the integer part of n, truncated towards zero.
//
// Returns zero, if n is invalid.
// Wraps, if the result is negative or does not fit into an uint.
func (n Number) Uint() uint {
	if !n.Valid() {
		return 0
	}
	return uint(intPart(n))
}

// FromFloat64 constructs [Number] from f using its shortest representation.
//
// Digits of that representation beyond Scale are rounded half to even. Returns
// NaN if f is not finite, and Overflow if it does not fit.
func FromFloat64(f float64) Number {
	if n, ok := fromShortFloat64(f); ok {
		return n
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return NaN
	}

	// The shortest digits have to come from strconv, but the 'e' format keeps
	// the buffer small and makes the scan a single pass.
	var buf [32]byte
	b := strconv.AppendFloat(buf[:0], f, 'e', -1, 64)

	var neg bool
	if b[0] == '-' {
		neg = true
		b = b[1:]
	}

	// Scan the mantissa of d.ddddde±dd into a single integer, it never exceeds
	// 17 digits.
	var d uint64
	digits := 0
	i := 0
	for ; b[i] != 'e'; i++ {
		if b[i] == '.' {
			continue
		}
		d = d*10 + uint64(b[i]-'0')
		digits++
	}
	expNeg := b[i+1] == '-'
	exp := 0
	for i += 2; i < len(b); i++ {
		exp = exp*10 + int(b[i]-'0')
	}
	if expNeg {
		exp = -exp
	}

	// The value is d*10^(exp-digits+1), which the coefficient scales by
	// 10^Scale. A negative power would need more than Scale fractional digits,
	// so those digits are rounded off.
	k := exp - digits + 1 + Scale
	if k >= 2*Scale {
		return Overflow
	}
	if k < 0 {
		// d has at most 17 digits, so it is below half of 10^18 and rounds
		// to zero once that many digits go.
		if -k >= 18 {
			return Zero
		}
		p := pow10u64[-k]
		q, r := d/p, d%p
		if r > p/2 || (r == p/2 && q&1 != 0) {
			q++
		}
		d, k = q, 0
	}
	n := Number{d, 0}
	if k > Scale {
		// d has at most 17 digits, so this cannot overflow.
		n = mul64(n, pow10u64Scale)
		k -= Scale
	}
	n, ovf := mul64Checked(n, pow10u64[k])
	if ovf || n.Hi > maxPosHi64 || (n.Hi == maxPosHi64 && n.Lo > maxPosLo64) {
		return Overflow
	}

	if neg {
		return n.Neg()
	}
	return n
}

// fromShortFloat64 converts f without strconv when its shortest decimal has at
// most Scale fractional digits and fewer than 16 significant digits, which
// covers most amounts.
//
// Below 2^50, f*10^k is within a quarter of the integer that a k-place decimal
// of f would need, so rounding it finds that decimal, and dividing back, which
// is correctly rounded, checks that it parses to f. The first k that passes is
// the shortest decimal, the same strconv finds.
func fromShortFloat64(f float64) (Number, bool) {
	a := math.Abs(f)

	// The largest k that keeps a*10^k below 2^50, from the binary exponent;
	// 1233/4096 is just under log10(2), so the estimate errs low. NaN and
	// infinities come out negative.
	e := int(math.Float64bits(a)>>52) - 1023
	kmax := min((49-e)*1233>>12, Scale)
	if kmax < 0 {
		return Number{}, false
	}

	// A decimal that parses back at some k also does at every larger k, so
	// checking kmax first rejects longer floats in one step.
	if _, ok := shortDecimal(a, kmax); !ok {
		return Number{}, false
	}
	for k := 0; ; k++ {
		if d, ok := shortDecimal(a, k); ok {
			n := mul64(Number{Lo: d}, pow10u64[Scale-k])
			if f < 0 {
				return n.Neg(), true
			}
			return n, true
		}
	}
}

// shortDecimal returns the k-place decimal nearest to a, as an integer, and
// whether it parses back to a.
func shortDecimal(a float64, k int) (uint64, bool) {
	d := uint64(a*pow10f64[k] + 0.5)
	return d, float64(d)/pow10f64[k] == a
}

// FromFloat64Exact constructs [Number] using the exact binary value of f rounded half to even.
//
// Returns NaN if f is not finite, and Overflow if it does not fit.
func FromFloat64Exact(f float64) Number {
	b := math.Float64bits(f)
	sign := b & (1 << 63)
	e := int(b>>52) & (1<<11 - 1)
	m := b & (1<<52 - 1)

	if e == 1<<11-1 {
		return NaN
	}
	if e == 0 {
		// Zero, or a subnormal far below the last place.
		return Zero
	}
	m |= 1 << 52
	e -= 1075 // f is m*2^e

	// m*10^Scale needs at most 117 bits, so the coefficient of an exact value
	// is just the product shifted by the exponent.
	p := mul64(Number{m, 0}, pow10u64Scale)

	if e >= 0 {
		if uint(e) >= leadingZeroes128(p) {
			return Overflow
		}
		p = shiftLeft128(p, uint(e))
		if !inBounds(p.Lo, p.Hi) {
			return Overflow
		}
		if sign != 0 {
			return p.Neg()
		}
		return p
	}

	sh := uint(-e)
	if sh >= 128 {
		return Zero
	}
	q := shiftRight128(p, sh)

	// Round half to even on the discarded bits.
	var round, sticky uint64
	if sh <= 64 {
		d := p.Lo & (1<<sh - 1)
		round = d >> (sh - 1)
		sticky = d & (1<<(sh-1) - 1)
	} else {
		d := p.Hi & (1<<(sh-64) - 1)
		round = d >> (sh - 65)
		sticky = d&(1<<(sh-65)-1) | p.Lo
	}
	if round != 0 && (sticky != 0 || q.Lo&1 != 0) {
		q = add64(q, 1)
	}

	if sign != 0 {
		return q.Neg()
	}
	return q
}

// Float64 returns the float64 nearest to n, or NaN if n is not valid.
func (n Number) Float64() float64 {
	if !n.Valid() {
		return math.NaN()
	}
	if n.IsZero() {
		return 0
	}
	signBit := n.Hi & (1 << 63)
	n = n.Abs()

	// Normalizing the top bit to 126 keeps the high word below 10^19, which is
	// what lets a single division produce the whole quotient.
	s := leadingZeroes128(n) - 1
	n = shiftLeft128(n, s)
	p := pow10Reciprocals[Scale]
	q, r := divMollerGranlund(n.Hi, n.Lo, p.d, p.v)

	// Round the 63 or 64 bit quotient to a 53 bit significand, the remainder
	// acting as the sticky bit.
	sh := uint(64-bits.LeadingZeros64(q)) - 53
	rem := q & (1<<sh - 1)
	half := uint64(1) << (sh - 1)
	q >>= sh

	// Assemble the result directly: the significand's implicit bit carries into
	// the exponent field, so the bias is one short, and a rounding increment
	// carries the same way. The exponent stays well inside float64 range, n is
	// at most 10^38.
	b := uint64(int(sh)-int(s)+1074)<<52 + q
	if rem > half || (rem == half && (r != 0 || q&1 != 0)) {
		b++
	}
	return math.Float64frombits(signBit | b)
}

// FromString constructs [Number] from decimal string s.
//
// An optional sign, digits with at most one point, and at least one digit,
// e.g. "-12", "+0.5", "3.", ".25". Exponent, whitespace or digit separators
// are not accepted.
//
// Returns ErrNaN for invalid syntax, ErrOverflow for more than 19 integer
// digits, and ErrInexact for nonzero digits beyond Scale fractional places.
func FromString(s string) (Number, error) {
	if s == "" {
		return Number{}, ErrNaN
	}

	neg := false
	var i uint
	switch s[0] {
	case '-':
		neg = true
		i = 1
	case '+':
		i = 1
	}

	overflow, inexact := false, false
	start := i
	for i < uint(len(s)) && s[i] == '0' {
		i++
	}

	var intVal uint64
	intDigits := 0
	for ; i < uint(len(s)) && s[i] >= '0' && s[i] <= '9'; i++ {
		intVal = intVal*10 + uint64(s[i]-'0')
		intDigits++
	}
	if intDigits > Scale {
		overflow = true
	}
	hasDigits := i > start

	var fracVal uint64
	fracDigits := 0
	if i < uint(len(s)) && s[i] == '.' {
		i++
		fracStart := i
		for ; i < uint(len(s)) && s[i] >= '0' && s[i] <= '9'; i++ {
			d := s[i] - '0'
			if fracDigits < Scale {
				fracVal = fracVal*10 + uint64(d)
				fracDigits++
			} else if d != 0 {
				inexact = true
			}
		}
		hasDigits = hasDigits || i > fracStart
	}

	// Anything left over is not part of a number, and a syntax error takes
	// precedence over a value that does not fit.
	if i != uint(len(s)) {
		return Number{}, ErrNaN
	}
	if !hasDigits {
		return Number{}, ErrNaN
	}
	if overflow {
		return Number{}, ErrOverflow
	}
	if inexact {
		return Number{}, ErrInexact
	}

	n := mul64(Number{intVal, 0}, pow10u64Scale)
	n = add64(n, fracVal*pow10u64[Scale-fracDigits])

	if neg {
		n = n.Neg()
	}
	return n, nil
}

// MustFromString is like [FromString], but panics on error.
func MustFromString(s string) Number {
	n, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return n
}

// FromBytes is like [FromString], but parses b without copying it to a string.
func FromBytes(b []byte) (Number, error) {
	return FromString(unsafe.String(unsafe.SliceData(b), len(b)))
}

// String returns a string representation of the number.
func (n Number) String() string {
	var buf [maxStrLen]byte
	return string(n.Append(buf[:0]))
}

// Append appends the string representation of the number to b.
func (n Number) Append(b []byte) []byte {
	if err := n.Err(); err != nil {
		return append(b, err.Error()...)
	}
	return n.format(b)
}

// intPart returns the low 64 bits of the integer part of n, truncated towards
// zero.
func intPart(n Number) uint64 {
	neg := n.IsNegative()
	if neg {
		n.Hi += pow10u64Scale
	}
	p := pow10Reciprocals[Scale]
	q, r := divMollerGranlund(n.Hi, n.Lo, p.d, p.v)
	if neg && r != 0 {
		q++
	}
	return q
}

// format appends the digits of n to b. The number must be valid.
func (n Number) format(b []byte) []byte {
	if n.IsZero() {
		return append(b, '0')
	}

	// Digits are produced from the last, so they go into the end of buf first.
	var buf [maxStrLen]byte
	start := len(buf)
	neg := n.IsNegative()
	if neg {
		n = n.Abs()
	}

	q, frac := quoRem128ByScale(n)

	if frac != 0 {
		// Drop the trailing zeroes first, so only significant digits are emitted.
		d := Scale
		if frac%1e16 == 0 {
			frac /= 1e16
			d -= 16
		}
		if frac%1e8 == 0 {
			frac /= 1e8
			d -= 8
		}
		if frac%1e4 == 0 {
			frac /= 1e4
			d -= 4
		}
		if frac%100 == 0 {
			frac /= 100
			d -= 2
		}
		if frac%10 == 0 {
			frac /= 10
			d--
		}
		for ; d >= 2; d -= 2 {
			j := frac % 100 * 2
			frac /= 100
			start -= 2
			buf[start] = formatDigitPairs[j]
			buf[start+1] = formatDigitPairs[j+1]
		}
		if d == 1 {
			start--
			buf[start] = '0' + byte(frac)
		}
		start--
		buf[start] = '.'
	}

	u := q.Lo
	for u >= 100 {
		j := u % 100 * 2
		u /= 100
		start -= 2
		buf[start] = formatDigitPairs[j]
		buf[start+1] = formatDigitPairs[j+1]
	}
	if u >= 10 {
		start -= 2
		buf[start] = formatDigitPairs[u*2]
		buf[start+1] = formatDigitPairs[u*2+1]
	} else {
		start--
		buf[start] = '0' + byte(u)
	}

	if neg {
		start--
		buf[start] = '-'
	}
	return append(b, buf[start:]...)
}

var formatDigitPairs = [...]byte{
	'0', '0', '0', '1', '0', '2', '0', '3', '0', '4', '0', '5', '0', '6', '0', '7', '0', '8', '0', '9',
	'1', '0', '1', '1', '1', '2', '1', '3', '1', '4', '1', '5', '1', '6', '1', '7', '1', '8', '1', '9',
	'2', '0', '2', '1', '2', '2', '2', '3', '2', '4', '2', '5', '2', '6', '2', '7', '2', '8', '2', '9',
	'3', '0', '3', '1', '3', '2', '3', '3', '3', '4', '3', '5', '3', '6', '3', '7', '3', '8', '3', '9',
	'4', '0', '4', '1', '4', '2', '4', '3', '4', '4', '4', '5', '4', '6', '4', '7', '4', '8', '4', '9',
	'5', '0', '5', '1', '5', '2', '5', '3', '5', '4', '5', '5', '5', '6', '5', '7', '5', '8', '5', '9',
	'6', '0', '6', '1', '6', '2', '6', '3', '6', '4', '6', '5', '6', '6', '6', '7', '6', '8', '6', '9',
	'7', '0', '7', '1', '7', '2', '7', '3', '7', '4', '7', '5', '7', '6', '7', '7', '7', '8', '7', '9',
	'8', '0', '8', '1', '8', '2', '8', '3', '8', '4', '8', '5', '8', '6', '8', '7', '8', '8', '8', '9',
	'9', '0', '9', '1', '9', '2', '9', '3', '9', '4', '9', '5', '9', '6', '9', '7', '9', '8', '9', '9',
}
