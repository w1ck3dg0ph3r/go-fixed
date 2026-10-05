package jxfixed_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/go-faster/jx"
	"go.w1ck3dg0ph3r.dev/fixed"
	"go.w1ck3dg0ph3r.dev/fixed/codecs/jxfixed"
)

var values = []string{
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

func TestEncode(t *testing.T) {
	t.Parallel()

	for _, s := range values {
		t.Run(s, func(t *testing.T) {
			n := fixed.MustFromString(s)
			var e jx.Encoder
			if err := jxfixed.Encode(&e, n); err != nil {
				t.Fatal(err)
			}
			exp, _ := json.Marshal(n)
			if string(e.Bytes()) != string(exp) {
				t.Errorf("expected %s, got %s", exp, e.Bytes())
			}
		})
	}
	for _, v := range []fixed.Number{fixed.NaN, fixed.Overflow, fixed.DivisionByZero} {
		t.Run(v.String(), func(t *testing.T) {
			var e jx.Encoder
			if err := jxfixed.Encode(&e, v); err != v.Err() {
				t.Errorf("expected err %v, got %v", v.Err(), err)
			}
			if len(e.Bytes()) != 0 {
				t.Errorf("expected nothing written, got %s", e.Bytes())
			}
		})
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()

	for _, s := range values {
		t.Run(s, func(t *testing.T) {
			r, err := jxfixed.Decode(jx.DecodeStr(`"` + s + `"`))
			if err != nil || r != fixed.MustFromString(s) {
				t.Errorf("expected %q, got %q (%v)", s, r, err)
			}
		})
	}

	cases := []struct {
		in  string
		exp fixed.Number
		err error
	}{
		{"\"\x5cu0031.5\"", fixed.MustFromString("1.5"), nil},
		{`1.5`, fixed.Number{}, fixed.ErrNaN},
		{`null`, fixed.Number{}, fixed.ErrNaN},
		{`true`, fixed.Number{}, fixed.ErrNaN},
		{`"1e5"`, fixed.Number{}, fixed.ErrNaN},
		{`"abc"`, fixed.Number{}, fixed.ErrNaN},
		{`"10000000000000000000"`, fixed.Number{}, fixed.ErrOverflow},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			r, err := jxfixed.Decode(jx.DecodeStr(tc.in))
			if !errors.Is(err, tc.err) {
				t.Errorf("expected err %v, got %v", tc.err, err)
			}
			if r != tc.exp {
				t.Errorf("expected %q, got %q", tc.exp, r)
			}
		})
	}

	t.Run("not consumed", func(t *testing.T) {
		d := jx.DecodeStr(`1.5`)
		if _, err := jxfixed.Decode(d); err != fixed.ErrNaN {
			t.Fatalf("expected err %v, got %v", fixed.ErrNaN, err)
		}
		if f, err := d.Float64(); err != nil || f != 1.5 {
			t.Errorf("expected 1.5 left to read, got %v (%v)", f, err)
		}
	})

	t.Run("unterminated", func(t *testing.T) {
		if _, err := jxfixed.Decode(jx.DecodeStr(`"1.5`)); err == nil {
			t.Error("expected error")
		}
	})
}

// TestObject round-trips Numbers as object fields, and checks that a value
// Decode rejects is left for the caller to handle.
func TestObject(t *testing.T) {
	t.Parallel()

	a, b := fixed.MustFromString("1.5"), fixed.MustFromString("-0.25")
	var e jx.Encoder
	e.Obj(func(e *jx.Encoder) {
		e.Field("a", func(e *jx.Encoder) { _ = jxfixed.Encode(e, a) })
		e.Field("b", func(e *jx.Encoder) { _ = jxfixed.Encode(e, b) })
	})
	if exp := `{"a":"1.5","b":"-0.25"}`; string(e.Bytes()) != exp {
		t.Fatalf("expected %s, got %s", exp, e.Bytes())
	}

	got := map[string]fixed.Number{}
	err := jx.DecodeBytes(e.Bytes()).Obj(func(d *jx.Decoder, key string) error {
		n, err := jxfixed.Decode(d)
		got[key] = n
		return err
	})
	if err != nil || got["a"] != a || got["b"] != b {
		t.Errorf("expected %q and %q, got %v (%v)", a, b, got, err)
	}

	d := jx.DecodeStr(`[null, "2"]`)
	var r []fixed.Number
	err = d.Arr(func(d *jx.Decoder) error {
		if d.Next() == jx.Null {
			return d.Null()
		}
		n, err := jxfixed.Decode(d)
		r = append(r, n)
		return err
	})
	if err != nil || len(r) != 1 || r[0] != fixed.MustFromString("2") {
		t.Errorf("expected [2], got %v (%v)", r, err)
	}
}

func TestAllocs(t *testing.T) {
	n := fixed.MustFromString("-1234567890.1234567890123456789")
	var e jx.Encoder
	if a := testing.AllocsPerRun(100, func() {
		e.Reset()
		_ = jxfixed.Encode(&e, n)
	}); a != 0 {
		t.Errorf("encode: %v allocs", a)
	}

	src := []byte(`"-1234567890.1234567890123456789"`)
	d := jx.DecodeBytes(src)
	if a := testing.AllocsPerRun(100, func() {
		d.ResetBytes(src)
		_, _ = jxfixed.Decode(d)
	}); a != 0 {
		t.Errorf("decode: %v allocs", a)
	}
}

// ogen's json.Marshaler and json.Unmarshaler, which x-ogen-type looks for.
var (
	_ interface{ Encode(*jx.Encoder) }       = jxfixed.Number{}
	_ interface{ Decode(*jx.Decoder) error } = (*jxfixed.Number)(nil)
)

func TestNumber(t *testing.T) {
	t.Parallel()

	for _, s := range values {
		t.Run(s, func(t *testing.T) {
			n := jxfixed.Number{Number: fixed.MustFromString(s)}
			var e jx.Encoder
			n.Encode(&e)
			if exp := `"` + s + `"`; string(e.Bytes()) != exp {
				t.Fatalf("expected %s, got %s", exp, e.Bytes())
			}
			var r jxfixed.Number
			if err := r.Decode(jx.DecodeBytes(e.Bytes())); err != nil || r != n {
				t.Errorf("expected %q, got %q (%v)", n, r, err)
			}
		})
	}
	for _, v := range []fixed.Number{fixed.NaN, fixed.Overflow, fixed.DivisionByZero} {
		t.Run(v.String(), func(t *testing.T) {
			var e jx.Encoder
			jxfixed.Number{Number: v}.Encode(&e)
			if string(e.Bytes()) != "null" {
				t.Errorf("expected null, got %s", e.Bytes())
			}
		})
	}

	t.Run("error", func(t *testing.T) {
		r := jxfixed.Number{Number: fixed.MustFromString("42")}
		if err := r.Decode(jx.DecodeStr(`1.5`)); err != fixed.ErrNaN {
			t.Errorf("expected err %v, got %v", fixed.ErrNaN, err)
		}
		if r.Number != fixed.MustFromString("42") {
			t.Errorf("value changed on error: %q", r)
		}
	})

	t.Run("promoted", func(t *testing.T) {
		n := jxfixed.Number{Number: fixed.MustFromString("1.5")}
		if b, err := n.MarshalText(); err != nil || string(b) != "1.5" {
			t.Errorf("expected 1.5, got %s (%v)", b, err)
		}
		if b, err := json.Marshal(n); err != nil || string(b) != `"1.5"` {
			t.Errorf(`expected "1.5", got %s (%v)`, b, err)
		}
	})
}
