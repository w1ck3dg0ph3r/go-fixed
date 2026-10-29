// Package pgxfixed lets pgx encode and scan fixed.Number as PostgreSQL numeric.
package pgxfixed

import (
	"encoding/binary"
	"errors"
	"math/bits"

	"github.com/jackc/pgx/v5/pgtype"
	"go.w1ck3dg0ph3r.dev/fixed"
)

// Register makes m handle fixed.Number as numeric.
func Register(m *pgtype.Map) {
	t := &pgtype.Type{Name: "numeric", OID: pgtype.NumericOID, Codec: NumericCodec{}}
	m.RegisterType(t)
	m.RegisterType(&pgtype.Type{Name: "_numeric", OID: pgtype.NumericArrayOID, Codec: &pgtype.ArrayCodec{ElementType: t}})

	m.RegisterDefaultPgType(fixed.Number{}, "numeric")
	m.RegisterDefaultPgType(&fixed.Number{}, "numeric")
	m.RegisterDefaultPgType([]fixed.Number{}, "_numeric")
	m.RegisterDefaultPgType([]*fixed.Number{}, "_numeric")
}

// NumericCodec is pgtype.NumericCodec that also encodes fixed.Number and scans
// into *fixed.Number.
//
// NULL scans as zero. Scan into **fixed.Number or sql.Null[fixed.Number] to
// tell NULL apart.
type NumericCodec struct {
	pgtype.NumericCodec
}

func (c NumericCodec) PlanEncode(m *pgtype.Map, oid uint32, format int16, value any) pgtype.EncodePlan {
	if _, ok := value.(fixed.Number); ok {
		switch format {
		case pgtype.BinaryFormatCode:
			return encodePlanBinary{}
		case pgtype.TextFormatCode:
			return encodePlanText{}
		}
	}
	return c.NumericCodec.PlanEncode(m, oid, format, value)
}

func (c NumericCodec) PlanScan(m *pgtype.Map, oid uint32, format int16, target any) pgtype.ScanPlan {
	if _, ok := target.(*fixed.Number); ok {
		switch format {
		case pgtype.BinaryFormatCode:
			return scanPlanBinary{}
		case pgtype.TextFormatCode:
			return scanPlanText{}
		}
	}
	return c.NumericCodec.PlanScan(m, oid, format, target)
}

var errInvalidBinary = errors.New("pgxfixed: invalid numeric binary data")

const (
	signPos    = 0x0000
	signNeg    = 0x4000
	signNaN    = 0xc000
	signPosInf = 0xd000
	signNegInf = 0xf000
)

type encodePlanText struct{}

func (encodePlanText) Encode(value any, buf []byte) ([]byte, error) {
	return value.(fixed.Number).AppendText(buf)
}

type encodePlanBinary struct{}

// Encode writes n as base 10000 digits around the decimal point: up to five
// for the whole part and five for the fraction, without leading and trailing
// zero digits.
func (encodePlanBinary) Encode(value any, buf []byte) ([]byte, error) {
	n := value.(fixed.Number)
	if !n.Valid() {
		return nil, n.Err()
	}
	sign := uint16(signPos)
	if n.IsNegative() {
		sign = signNeg
		n = n.Abs()
	}

	// n is below 10^38, so the high word is below 10^19 and one division splits
	// it into the whole part and 19 fractional digits.
	q, f := bits.Div64(n.Hi, n.Lo, 1e19)

	dscale := 0
	if f != 0 {
		dscale = fixed.Scale
		for t := f; t%10 == 0; t /= 10 {
			dscale--
		}
	}

	var digits [10]uint16
	for i := 4; i >= 0; i-- {
		digits[i] = uint16(q % 10000)
		q /= 10000
	}
	// The last fractional digit holds the 19th decimal digit padded with zero.
	digits[9] = uint16(f % 1000 * 10)
	f /= 1000
	for i := 8; i >= 5; i-- {
		digits[i] = uint16(f % 10000)
		f /= 10000
	}

	first := 0
	for first < len(digits) && digits[first] == 0 {
		first++
	}
	if first == len(digits) {
		return append(buf, 0, 0, 0, 0, 0, 0, 0, 0), nil
	}
	last := len(digits) - 1
	for digits[last] == 0 {
		last--
	}

	buf = binary.BigEndian.AppendUint16(buf, uint16(last-first+1))
	buf = binary.BigEndian.AppendUint16(buf, uint16(int16(4-first)))
	buf = binary.BigEndian.AppendUint16(buf, sign)
	buf = binary.BigEndian.AppendUint16(buf, uint16(dscale))
	for _, d := range digits[first : last+1] {
		buf = binary.BigEndian.AppendUint16(buf, d)
	}
	return buf, nil
}

type scanPlanText struct{}

func (scanPlanText) Scan(src []byte, target any) error {
	dst := target.(*fixed.Number)
	if src == nil {
		*dst = fixed.Zero
		return nil
	}
	return dst.UnmarshalText(src)
}

type scanPlanBinary struct{}

func (scanPlanBinary) Scan(src []byte, target any) error {
	dst := target.(*fixed.Number)
	if src == nil {
		*dst = fixed.Zero
		return nil
	}
	n, err := decodeBinary(src)
	if err != nil {
		return err
	}
	*dst = n
	return nil
}

// Scaling of a base 10000 digit by its position: 10000^p for the whole part,
// and 10^(Scale+4p) for the fraction, where the fifth fractional digit loses
// its last decimal digit.
var (
	wholeScale = [...]uint64{1, 1e4, 1e8, 1e12, 1e16}
	fracScale  = [...]uint64{0, 1e15, 1e11, 1e7, 1e3}
)

// decodeBinary returns ErrOverflow for values that are out of range,
// ErrInexact for more than Scale significant fractional digits, and ErrNaN for
// NaN and infinities.
func decodeBinary(src []byte) (fixed.Number, error) {
	if len(src) < 8 {
		return fixed.Number{}, errInvalidBinary
	}
	ndigits := int(binary.BigEndian.Uint16(src))
	weight := int(int16(binary.BigEndian.Uint16(src[2:])))
	sign := binary.BigEndian.Uint16(src[4:])
	if len(src) != 8+2*ndigits {
		return fixed.Number{}, errInvalidBinary
	}
	switch sign {
	case signPos, signNeg:
	case signNaN, signPosInf, signNegInf:
		return fixed.Number{}, fixed.ErrNaN
	default:
		return fixed.Number{}, errInvalidBinary
	}

	var q, f uint64
	inexact := false
	for i := range ndigits {
		d := uint64(binary.BigEndian.Uint16(src[8+2*i:]))
		if d >= 10000 {
			return fixed.Number{}, errInvalidBinary
		}
		if d == 0 {
			continue
		}
		switch p := weight - i; {
		case p >= len(wholeScale):
			return fixed.Number{}, fixed.ErrOverflow
		case p >= 0:
			hi, lo := bits.Mul64(d, wholeScale[p])
			var carry uint64
			q, carry = bits.Add64(q, lo, 0)
			if hi|carry != 0 {
				return fixed.Number{}, fixed.ErrOverflow
			}
		case p >= -4:
			f += d * fracScale[-p]
		case p == -5 && d%10 == 0:
			f += d / 10
		default:
			// Keep scanning, so that overflow takes precedence.
			inexact = true
		}
	}
	if q >= 1e19 {
		return fixed.Number{}, fixed.ErrOverflow
	}
	if inexact {
		return fixed.Number{}, fixed.ErrInexact
	}

	hi, lo := bits.Mul64(q, 1e19)
	lo, carry := bits.Add64(lo, f, 0)
	n := fixed.Number{Lo: lo, Hi: hi + carry}
	if sign == signNeg {
		n = n.Neg()
	}
	return n, nil
}
