package fixed_test

import (
	"testing"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func TestValid(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, tc := range []struct {
		n   fixed.Number
		exp bool
	}{
		{fixed.Number{}, true},
		{n("1"), true},
		{n("8248727485765637238.3886788824556356949"), true},
		{n("9999999999999999999.9999999999999999999"), true},
		{n("9999999999999999999").Add(n("1")), false},
		{n("-9999999999999999999.9999999999999999999"), true},
		{n("-9999999999999999999").Sub(n("1")), false},
		{fixed.NaN, false},
		{fixed.Overflow, false},
		{fixed.DivisionByZero, false},
		{fixed.Number{Hi: 1 << 63}, false},
		{fixed.Number{Lo: 0x98a224000000003, Hi: 0x4b3b4ca85a86c47a}, false},
	} {
		t.Run(tc.n.String(), func(t *testing.T) {
			res := tc.n.Valid()
			if res != tc.exp {
				t.Errorf("expected %v, got %v", tc.exp, res)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		n     fixed.Number
		valid bool
		err   error
	}{
		{"Zero", fixed.Zero, true, nil},
		{"One", fixed.One, true, nil},
		{"NaN", fixed.NaN, false, fixed.ErrNaN},
		{"Overflow", fixed.Overflow, false, fixed.ErrOverflow},
		{"DivisionByZero", fixed.DivisionByZero, false, fixed.ErrDivisionByZero},
		{"ValidHi", fixed.MustFromString("-1.7759344522308878336"), true, nil},
		{"OutOfBounds", fixed.Number{Hi: 1 << 63}, false, fixed.ErrNaN},
		{"Sentinel", fixed.Number{Lo: 0x98a224000000003, Hi: 0x4b3b4ca85a86c47a}, false, fixed.ErrNaN},
	} {
		t.Run(tc.name, func(t *testing.T) {
			valid := tc.n.Valid()
			if valid != tc.valid {
				t.Errorf("expected valid to be %v, got %v", tc.valid, valid)
			}
			err := tc.n.Err()
			if err != tc.err {
				t.Errorf("expected error to be %v, got %v", tc.err, err)
			}
		})
	}
}
