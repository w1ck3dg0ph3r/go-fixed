// Package jxfixed encodes and decodes fixed.Number with go-faster/jx.
package jxfixed

import (
	"github.com/go-faster/jx"
	"go.w1ck3dg0ph3r.dev/fixed"
)

// Encode writes n as a JSON string.
// Returns the number's error if it is not valid, without writing anything.
func Encode(e *jx.Encoder, n fixed.Number) error {
	var buf [fixed.Precision + 3]byte
	b, err := n.AppendText(buf[:0])
	if err != nil {
		return err
	}
	e.ByteStr(b)
	return nil
}

// Decode reads a Number from a JSON string.
// Returns ErrNaN without consuming the value if it is not a string.
func Decode(d *jx.Decoder) (fixed.Number, error) {
	if d.Next() != jx.String {
		return fixed.Number{}, fixed.ErrNaN
	}
	b, err := d.StrBytes()
	if err != nil {
		return fixed.Number{}, err
	}
	return fixed.FromBytes(b)
}

// Number is fixed.Number with the Encode and Decode methods ogen uses for a
// schema's x-ogen-type, which are faster than the MarshalJSON and
// UnmarshalJSON it would use for fixed.Number:
//
//	x-ogen-type: go.w1ck3dg0ph3r.dev/fixed/codecs/jxfixed.Number
type Number struct {
	fixed.Number
}

// Encode implements ogen's json.Marshaler. Since it cannot return an error, an
// invalid number is written as null.
func (n Number) Encode(e *jx.Encoder) {
	if err := Encode(e, n.Number); err != nil {
		e.Null()
	}
}

// Decode implements ogen's json.Unmarshaler, the same way as the Decode
// function.
func (n *Number) Decode(d *jx.Decoder) error {
	v, err := Decode(d)
	if err != nil {
		return err
	}
	n.Number = v
	return nil
}
