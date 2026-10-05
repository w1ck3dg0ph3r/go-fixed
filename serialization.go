package fixed

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/json/jsontext"
)

// AppendText implements encoding.TextAppender.
func (n Number) AppendText(b []byte) ([]byte, error) {
	if err := n.Err(); err != nil {
		return b, err
	}
	return n.Append(b), nil
}

// MarshalText implements encoding.TextMarshaler.
func (n Number) MarshalText() ([]byte, error) {
	return n.AppendText(make([]byte, 0, Precision+3))
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *Number) UnmarshalText(b []byte) error {
	v, err := FromBytes(b)
	if err != nil {
		return err
	}
	*n = v
	return nil
}

// AppendBinary implements encoding.BinaryAppender.
//
// The encoding is 16 bytes big-endian, with the sign bit flipped, so byte
// order matches numeric order.
func (n Number) AppendBinary(b []byte) ([]byte, error) {
	if err := n.Err(); err != nil {
		return b, err
	}
	b = binary.BigEndian.AppendUint64(b, n.Hi^1<<63)
	return binary.BigEndian.AppendUint64(b, n.Lo), nil
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (n Number) MarshalBinary() ([]byte, error) {
	return n.AppendBinary(make([]byte, 0, 16))
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (n *Number) UnmarshalBinary(b []byte) error {
	if len(b) != 16 {
		return ErrNaN
	}
	hi := binary.BigEndian.Uint64(b) ^ 1<<63
	lo := binary.BigEndian.Uint64(b[8:])

	if res, err := NewWords(lo, hi); err != nil {
		return err
	} else {
		*n = res
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (n Number) MarshalJSON() ([]byte, error) {
	return n.appendJSON(make([]byte, 0, Precision+5))
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Number) UnmarshalJSON(b []byte) error {
	return n.unmarshalJSON(b)
}

// MarshalJSONTo implements json.MarshalerTo.
func (n Number) MarshalJSONTo(enc *jsontext.Encoder) error {
	var buf [Precision + 5]byte
	b, err := n.appendJSON(buf[:0])
	if err != nil {
		return err
	}
	return enc.WriteValue(b)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (n *Number) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	v, err := dec.ReadValue()
	if err != nil {
		return err
	}
	return n.unmarshalJSON(v)
}

func (n Number) appendJSON(b []byte) ([]byte, error) {
	if err := n.Err(); err != nil {
		return b, err
	}
	b = append(b, '"')
	b = n.Append(b)
	return append(b, '"'), nil
}

func (n *Number) unmarshalJSON(v jsontext.Value) error {
	b := []byte(v)
	switch v.Kind() {
	case 'n':
		return nil
	case '"':
		if len(b) < 2 {
			return ErrNaN
		}
		if bytes.IndexByte(b, '\\') < 0 {
			b = b[1 : len(b)-1]
		} else {
			var buf [64]byte
			var err error
			if b, err = jsontext.AppendUnquote(buf[:0], b); err != nil {
				return err
			}
		}
	default:
		return ErrNaN
	}
	return n.UnmarshalText(b)
}

// Value implements driver.Valuer.
func (n Number) Value() (driver.Value, error) {
	if err := n.Err(); err != nil {
		return nil, err
	}
	return n.String(), nil
}

// Scan implements sql.Scanner.
//
// NULL scans as zero. Use sql.Null[Number] to tell NULL apart.
func (n *Number) Scan(src any) error {
	var v Number
	var err error
	switch src := src.(type) {
	case nil:
		v = Zero
	case []byte:
		v, err = FromBytes(src)
	case string:
		v, err = FromString(src)
	case int64:
		v = FromInt(src)
	case float64:
		v = FromFloat64(src)
		err = v.Err()
	default:
		return ErrNaN
	}
	if err != nil {
		return err
	}
	*n = v
	return nil
}
