package fuzzing

import (
	"bytes"
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"testing"
	"unicode/utf8"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func FuzzText(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.neg, c.hi, c.lo)
	}

	f.Fuzz(func(t *testing.T, neg bool, hi uint64, lo uint64) {
		a, ok := number(neg, hi, lo)
		if !ok {
			return
		}

		b, err := a.MarshalText()
		if err != nil {
			t.Fatalf("marshal %q: %v", a, err)
		}
		if string(b) != a.String() {
			t.Fatalf("marshal %q: %q != String", a, b)
		}
		if got, _ := a.AppendText([]byte("x")); string(got) != "x"+string(b) {
			t.Fatalf("append %q: %q", a, got)
		}

		var r fixed.Number
		if err := r.UnmarshalText(b); err != nil || r != a {
			t.Fatalf("unmarshal %q: (%q, %v)", b, r, err)
		}
	})
}

func FuzzUnmarshalText(f *testing.F) {
	for _, s := range []string{"", "0", "-1.5", ".5", "1.", "abc", "1e5", fixed.Highest.String(), "10000000000000000000", "0.00000000000000000001"} {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		want, wantErr := fixed.FromString(string(b))

		r := fixed.FromInt(42)
		err := r.UnmarshalText(b)
		if err != wantErr {
			t.Fatalf("unmarshal %q: error %v != %v", b, err, wantErr)
		}
		if err != nil {
			if r != fixed.FromInt(42) {
				t.Fatalf("unmarshal %q: changed to %q on error", b, r)
			}
			return
		}
		if r != want {
			t.Fatalf("unmarshal %q: %q != %q", b, r, want)
		}
	})
}

func FuzzBinary(f *testing.F) {
	for _, c := range corpus {
		for _, d := range corpus {
			f.Add(c.neg, c.hi, c.lo, d.neg, d.hi, d.lo)
		}
	}

	f.Fuzz(func(t *testing.T, aneg bool, ahi uint64, alo uint64, bneg bool, bhi uint64, blo uint64) {
		a, ok := number(aneg, ahi, alo)
		if !ok {
			return
		}
		b, ok := number(bneg, bhi, blo)
		if !ok {
			return
		}

		ea, err := a.MarshalBinary()
		if err != nil || len(ea) != 16 {
			t.Fatalf("marshal %q: (%x, %v)", a, ea, err)
		}
		eb, err := b.MarshalBinary()
		if err != nil {
			t.Fatalf("marshal %q: %v", b, err)
		}

		var r fixed.Number
		if err := r.UnmarshalBinary(ea); err != nil || r != a {
			t.Fatalf("unmarshal %x: (%q, %v), want %q", ea, r, err, a)
		}

		// Byte order matches numeric order.
		if got, want := bytes.Compare(ea, eb), a.Cmp(b); got != want {
			t.Fatalf("%q vs %q: bytes compare %d, numbers %d", a, b, got, want)
		}
	})
}

func FuzzUnmarshalBinary(f *testing.F) {
	for _, n := range []fixed.Number{fixed.Zero, fixed.One, fixed.Highest, fixed.Lowest} {
		b, _ := n.MarshalBinary()
		f.Add(b)
	}
	f.Add([]byte{})
	f.Add(make([]byte, 16))
	f.Add(bytes.Repeat([]byte{0xff}, 16))

	f.Fuzz(func(t *testing.T, b []byte) {
		r := fixed.FromInt(42)
		if err := r.UnmarshalBinary(b); err != nil {
			if r != fixed.FromInt(42) {
				t.Fatalf("unmarshal %x: changed to %q on error", b, r)
			}
			return
		}
		if !r.Valid() {
			t.Fatalf("unmarshal %x: accepted invalid %q", b, r)
		}

		// The encoding is a bijection, so whatever decodes encodes back the same.
		if e, err := r.MarshalBinary(); err != nil || !bytes.Equal(e, b) {
			t.Fatalf("unmarshal %x: %q encodes back as (%x, %v)", b, r, e, err)
		}
	})
}

type jsonField struct {
	N fixed.Number `json:"n"`
}

func FuzzJSON(f *testing.F) {
	for _, c := range corpus {
		f.Add(c.neg, c.hi, c.lo)
	}

	f.Fuzz(func(t *testing.T, neg bool, hi uint64, lo uint64) {
		a, ok := number(neg, hi, lo)
		if !ok {
			return
		}
		want := `{"n":"` + a.String() + `"}`

		for name, marshal := range map[string]func(any) ([]byte, error){
			"v1": jsonv1.Marshal,
			"v2": func(v any) ([]byte, error) { return json.Marshal(v) },
		} {
			b, err := marshal(jsonField{a})
			if err != nil || string(b) != want {
				t.Fatalf("%s marshal %q: (%s, %v)", name, a, b, err)
			}
		}

		for name, unmarshal := range map[string]func([]byte, any) error{
			"v1": jsonv1.Unmarshal,
			"v2": func(b []byte, v any) error { return json.Unmarshal(b, v) },
		} {
			var r jsonField
			if err := unmarshal([]byte(want), &r); err != nil || r.N != a {
				t.Fatalf("%s unmarshal %s: (%q, %v)", name, want, r.N, err)
			}
		}
	})
}

func FuzzUnmarshalJSON(f *testing.F) {
	for _, s := range []string{
		`"1.5"`,
		`"-0.5"`,
		`null`,
		`1.5`,
		`""`,
		`"abc"`,
		`"\u0031.5"`,
		`"1\n"`,
		`"10000000000000000000"`,
		`"0.00000000000000000001"`,
		`{}`,
		`[]`,
		`"`,
	} {
		f.Add([]byte(s))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		// v1 replaces invalid UTF-8 anywhere in the document while v2 rejects
		// it, so the two only have to agree on valid input.
		if !utf8.Valid(b) {
			return
		}
		doc := append(append([]byte(`{"n":`), b...), '}')

		var r1, r2 jsonField
		err1 := jsonv1.Unmarshal(doc, &r1)
		err2 := json.Unmarshal(doc, &r2)

		// Both front ends run the same method, so they agree on the outcome.
		if (err1 == nil) != (err2 == nil) {
			t.Fatalf("unmarshal %s: v1 error %v, v2 error %v", doc, err1, err2)
		}
		if err1 != nil {
			return
		}
		if r1.N != r2.N {
			t.Fatalf("unmarshal %s: v1 %q != v2 %q", doc, r1.N, r2.N)
		}
		if !r1.N.Valid() {
			t.Fatalf("unmarshal %s: accepted invalid %q", doc, r1.N)
		}

		// Whatever is accepted is the value of a string that FromString parses,
		// so formatting it and parsing it back is exact.
		var back jsonField
		e, _ := json.Marshal(r1)
		if err := json.Unmarshal(e, &back); err != nil || back.N != r1.N {
			t.Fatalf("unmarshal %s: %q does not round trip via %s", doc, r1.N, e)
		}
	})
}
