package fixed_test

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"go.w1ck3dg0ph3r.dev/fixed"
)

var serializationValues = []string{
	"0",
	"1",
	"-1",
	"123.45",
	"-123.45",
	"0.0000000000000000001",
	"-0.0000000000000000001",
	"9999999999999999999.9999999999999999999",
	"-9999999999999999999.9999999999999999999",
}

var invalidNumbers = []fixed.Number{
	fixed.NaN,
	fixed.Overflow,
	fixed.DivisionByZero,
}

func TestText(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, s := range serializationValues {
		t.Run(s, func(t *testing.T) {
			b, err := n(s).MarshalText()
			if err != nil || string(b) != s {
				t.Fatalf("expected %q, got %q (%v)", s, b, err)
			}
			b, err = n(s).AppendText([]byte("x"))
			if err != nil || string(b) != "x"+s {
				t.Fatalf("expected %q, got %q (%v)", "x"+s, b, err)
			}
			var r fixed.Number
			if err := r.UnmarshalText(b[1:]); err != nil || r != n(s) {
				t.Fatalf("expected %q, got %q (%v)", s, r, err)
			}
		})
	}
	for _, v := range invalidNumbers {
		t.Run(v.String(), func(t *testing.T) {
			if _, err := v.MarshalText(); err != v.Err() {
				t.Errorf("expected err %v, got %v", v.Err(), err)
			}
		})
	}

	cases := []struct {
		in  string
		err error
	}{
		{"", fixed.ErrNaN},
		{"abc", fixed.ErrNaN},
		{"1e5", fixed.ErrNaN},
		{"10000000000000000000", fixed.ErrOverflow},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("Unmarshal(%q)", tc.in), func(t *testing.T) {
			r := fixed.MustFromString("42")
			if err := r.UnmarshalText([]byte(tc.in)); err != tc.err {
				t.Errorf("expected err %v, got %v", tc.err, err)
			}
			if r != fixed.MustFromString("42") {
				t.Errorf("value changed on error: %q", r)
			}
		})
	}
}

func TestBinary(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, s := range serializationValues {
		t.Run(s, func(t *testing.T) {
			in := n(s)
			b, err := in.MarshalBinary()
			if err != nil || len(b) != 16 {
				t.Fatalf("expected 16 bytes, got %d (%v)", len(b), err)
			}
			a, err := in.AppendBinary([]byte("x"))
			if err != nil || !bytes.Equal(a[1:], b) {
				t.Fatalf("expected %x, got %x (%v)", b, a[1:], err)
			}
			var out fixed.Number
			if err := out.UnmarshalBinary(b); err != nil || out != n(s) {
				t.Fatalf("expected %q, got %q (%v)", s, out, err)
			}
			if out != in {
				t.Error("out != in")
			}
		})
	}
	for _, v := range invalidNumbers {
		t.Run(v.String(), func(t *testing.T) {
			if _, err := v.MarshalBinary(); err != v.Err() {
				t.Errorf("expected err %v, got %v", v.Err(), err)
			}
		})
	}

	raw := func(v fixed.Number) []byte {
		b := make([]byte, 16)
		for i := range 8 {
			b[i] = byte((v.Hi ^ 1<<63) >> (56 - 8*i))
			b[8+i] = byte(v.Lo >> (56 - 8*i))
		}
		return b
	}
	cases := []struct {
		name string
		in   []byte
		err  error
	}{
		{"empty", nil, fixed.ErrNaN},
		{"short", make([]byte, 15), fixed.ErrNaN},
		{"long", make([]byte, 17), fixed.ErrNaN},
		{"above max", raw(fixed.Number{Lo: fixed.Highest.Lo + 1, Hi: fixed.Highest.Hi}), fixed.ErrNaN},
		{"below min", raw(fixed.Number{Lo: fixed.Lowest.Lo - 1, Hi: fixed.Lowest.Hi}), fixed.ErrNaN},
		{"NaN", raw(fixed.NaN), fixed.ErrNaN},
		{"Overflow", raw(fixed.Overflow), fixed.ErrOverflow},
		{"DivisionByZero", raw(fixed.DivisionByZero), fixed.ErrDivisionByZero},
	}
	for _, tc := range cases {
		t.Run("Unmarshal "+tc.name, func(t *testing.T) {
			var r fixed.Number
			if err := r.UnmarshalBinary(tc.in); err != tc.err {
				t.Errorf("expected err %v, got %v", tc.err, err)
			}
		})
	}

	// Byte order must match numeric order, including across the sign boundary.
	ordered := []string{
		"-9999999999999999999.9999999999999999999",
		"-123.45",
		"-1",
		"-0.0000000000000000001",
		"0",
		"0.0000000000000000001",
		"1",
		"123.45",
		"9999999999999999999.9999999999999999999",
	}
	for i := 1; i < len(ordered); i++ {
		a, _ := n(ordered[i-1]).MarshalBinary()
		b, _ := n(ordered[i]).MarshalBinary()
		if bytes.Compare(a, b) >= 0 {
			t.Errorf("expected %s < %s, got %x >= %x", ordered[i-1], ordered[i], a, b)
		}
	}
}

var jsonUnmarshalCases = []struct {
	in  string
	exp fixed.Number
	err error
}{
	{`"1.5"`, fixed.MustFromString("1.5"), nil},
	{`1.5`, fixed.MustFromString("42"), fixed.ErrNaN},
	{`-1.5`, fixed.MustFromString("42"), fixed.ErrNaN},
	{"\"\x5cu0031.5\"", fixed.MustFromString("1.5"), nil},
	{"\"\x5cu002d1.5\"", fixed.MustFromString("-1.5"), nil},
	{"\"1\x5cn\"", fixed.MustFromString("42"), fixed.ErrNaN},
	{`null`, fixed.MustFromString("42"), nil},
	{`"1e5"`, fixed.MustFromString("42"), fixed.ErrNaN},
	{`"abc"`, fixed.MustFromString("42"), fixed.ErrNaN},
	{`true`, fixed.MustFromString("42"), fixed.ErrNaN},
	{`"10000000000000000000"`, fixed.MustFromString("42"), fixed.ErrOverflow},
}

func TestJSONMethods(t *testing.T) {
	t.Parallel()

	methods := []struct {
		name      string
		marshal   func(fixed.Number) ([]byte, error)
		unmarshal func(*fixed.Number, []byte) error
	}{
		{
			"MarshalJSON",
			fixed.Number.MarshalJSON,
			(*fixed.Number).UnmarshalJSON,
		},
		{
			"MarshalJSONTo",
			func(n fixed.Number) ([]byte, error) {
				var buf bytes.Buffer
				err := n.MarshalJSONTo(jsontext.NewEncoder(&buf))
				return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), err
			},
			func(n *fixed.Number, b []byte) error {
				return n.UnmarshalJSONFrom(jsontext.NewDecoder(bytes.NewReader(b)))
			},
		},
	}
	n := fixed.MustFromString

	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			for _, s := range serializationValues {
				t.Run(s, func(t *testing.T) {
					b, err := m.marshal(n(s))
					exp := `"` + s + `"`
					if err != nil || string(b) != exp {
						t.Fatalf("expected %s, got %s (%v)", exp, b, err)
					}
					var r fixed.Number
					if err := m.unmarshal(&r, b); err != nil || r != n(s) {
						t.Fatalf("expected %q, got %q (%v)", s, r, err)
					}
				})
			}
			for _, v := range invalidNumbers {
				t.Run(v.String(), func(t *testing.T) {
					if _, err := m.marshal(v); err != v.Err() {
						t.Errorf("expected err %v, got %v", v.Err(), err)
					}
				})
			}
			for _, tc := range jsonUnmarshalCases {
				t.Run(fmt.Sprintf("Unmarshal(%s)", tc.in), func(t *testing.T) {
					r := n("42")
					if err := m.unmarshal(&r, []byte(tc.in)); err != tc.err {
						t.Errorf("expected err %v, got %v", tc.err, err)
					}
					if r != tc.exp {
						t.Errorf("expected %q, got %q", tc.exp, r)
					}
				})
			}
		})
	}
}

// encoding/json validates input before calling UnmarshalJSON, but a direct
// call can pass anything.
func TestUnmarshalJSONMalformed(t *testing.T) {
	t.Parallel()

	var r fixed.Number
	if err := r.UnmarshalJSON([]byte(`"`)); err != fixed.ErrNaN {
		t.Errorf("expected err %v, got %v", fixed.ErrNaN, err)
	}

	if err := r.UnmarshalJSON([]byte(`"0\FF"`)); err == nil || !strings.Contains(err.Error(), "invalid escape sequence") {
		t.Errorf("expected err, got %v", err)
	}
}

func TestJSON(t *testing.T) {
	t.Parallel()

	type Doc struct {
		N fixed.Number  `json:"n"`
		P *fixed.Number `json:"p,omitempty"`
	}
	codecs := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"v1", jsonv1.Marshal, jsonv1.Unmarshal},
		{"v2", func(v any) ([]byte, error) { return json.Marshal(v) }, func(b []byte, v any) error { return json.Unmarshal(b, v) }},
	}
	n := fixed.MustFromString

	for _, c := range codecs {
		t.Run(c.name, func(t *testing.T) {
			for _, s := range serializationValues {
				t.Run(s, func(t *testing.T) {
					b, err := c.marshal(Doc{N: n(s)})
					exp := `{"n":"` + s + `"}`
					if err != nil || string(b) != exp {
						t.Fatalf("expected %s, got %s (%v)", exp, b, err)
					}
					var doc Doc
					if err := c.unmarshal(b, &doc); err != nil || doc.N != n(s) {
						t.Fatalf("expected %q, got %q (%v)", s, doc.N, err)
					}
				})
			}
			for _, v := range invalidNumbers {
				t.Run(v.String(), func(t *testing.T) {
					if _, err := c.marshal(Doc{N: v}); !errors.Is(err, v.Err()) {
						t.Errorf("expected err %v, got %v", v.Err(), err)
					}
				})
			}

			for _, tc := range jsonUnmarshalCases {
				in := `{"n":` + tc.in + `}`
				t.Run(fmt.Sprintf("Unmarshal(%s)", in), func(t *testing.T) {
					doc := Doc{N: n("42")}
					err := c.unmarshal([]byte(in), &doc)
					if !errors.Is(err, tc.err) {
						t.Errorf("expected err %v, got %v", tc.err, err)
					}
					if doc.N != tc.exp {
						t.Errorf("expected %q, got %q", tc.exp, doc.N)
					}
				})
			}

			t.Run("pointer", func(t *testing.T) {
				p := n("-0.5")
				b, err := c.marshal(Doc{P: &p})
				exp := `{"n":"0","p":"-0.5"}`
				if err != nil || string(b) != exp {
					t.Fatalf("expected %s, got %s (%v)", exp, b, err)
				}
				var doc Doc
				if err := c.unmarshal(b, &doc); err != nil || doc.P == nil || *doc.P != p {
					t.Fatalf("expected %q, got %v (%v)", p, doc.P, err)
				}
			})
		})
	}
}

func TestSQL(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, s := range serializationValues {
		t.Run(s, func(t *testing.T) {
			v, err := n(s).Value()
			if err != nil || v != s || !driver.IsValue(v) {
				t.Fatalf("expected %q, got %#v (%v)", s, v, err)
			}
			var r fixed.Number
			if err := r.Scan(v); err != nil || r != n(s) {
				t.Fatalf("expected %q, got %q (%v)", s, r, err)
			}
		})
	}
	for _, v := range invalidNumbers {
		t.Run(v.String(), func(t *testing.T) {
			if _, err := v.Value(); err != v.Err() {
				t.Errorf("expected err %v, got %v", v.Err(), err)
			}
		})
	}

	cases := []struct {
		in  any
		exp fixed.Number
		err error
	}{
		{"1.5", n("1.5"), nil},
		{[]byte("-1.5"), n("-1.5"), nil},
		{int64(-42), n("-42"), nil},
		{int64(math.MaxInt64), n("9223372036854775807"), nil},
		{0.1, n("0.1"), nil},
		{-1.5, n("-1.5"), nil},
		{nil, fixed.Zero, nil},
		{true, n("42"), fixed.ErrNaN},
		{time.Time{}, n("42"), fixed.ErrNaN},
		{"abc", n("42"), fixed.ErrNaN},
		{[]byte("1.00000000000000000001"), n("42"), fixed.ErrInexact},
		{"10000000000000000000", n("42"), fixed.ErrOverflow},
		{1e19, n("42"), fixed.ErrOverflow},
		{math.NaN(), n("42"), fixed.ErrNaN},
		{math.Inf(1), n("42"), fixed.ErrNaN},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("Scan(%#v)", tc.in), func(t *testing.T) {
			r := n("42")
			if err := r.Scan(tc.in); err != tc.err {
				t.Errorf("expected err %v, got %v", tc.err, err)
			}
			if r != tc.exp {
				t.Errorf("expected %q, got %q", tc.exp, r)
			}
		})
	}

	t.Run("Null", func(t *testing.T) {
		var r sql.Null[fixed.Number]
		if err := r.Scan(nil); err != nil || r.Valid {
			t.Fatalf("expected invalid, got %v (%v)", r, err)
		}
		if err := r.Scan([]byte("1.5")); err != nil || !r.Valid || r.V != n("1.5") {
			t.Fatalf("expected 1.5, got %v (%v)", r, err)
		}
	})
}
