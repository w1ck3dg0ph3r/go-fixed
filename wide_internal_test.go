package fixed

import "testing"

func TestWideRound(t *testing.T) {
	t.Parallel()

	n := MustFromString
	cases := []struct {
		w    wide
		neg  bool
		mode RoundingMode
		e    Number
	}{
		// Integer coefficients, scaled up rather than divided.
		{wide{m: [3]uint64{5}}, false, RoundZero, n("5")},
		{wide{m: [3]uint64{5}, e: 2}, true, RoundZero, n("-500")},
		{wide{m: [3]uint64{1}, e: 18}, false, RoundZero, n("1000000000000000000")},
		{wide{m: [3]uint64{10}, e: 18}, false, RoundZero, Overflow},
		{wide{m: [3]uint64{1}, e: 40}, false, RoundZero, Overflow},
		{wide{m: [3]uint64{0, 0, 1}, e: -Scale}, false, RoundZero, Overflow},

		// A guard digit and sticky digits below the last place.
		{wide{m: [3]uint64{15}, e: -Scale - 1}, false, RoundBank, n("0.0000000000000000002")},
		{wide{m: [3]uint64{25}, e: -Scale - 1}, false, RoundBank, n("0.0000000000000000002")},
		{wide{m: [3]uint64{25}, e: -Scale - 1, inexact: true}, false, RoundBank, n("0.0000000000000000003")},
		{wide{m: [3]uint64{2501}, e: -Scale - 3}, false, RoundHalfZero, n("0.0000000000000000003")},
		{wide{m: [3]uint64{2500}, e: -Scale - 3}, false, RoundHalfZero, n("0.0000000000000000002")},
		{wide{m: [3]uint64{20}, e: -Scale - 1, inexact: true}, false, RoundCeil, n("0.0000000000000000003")},
		{wide{m: [3]uint64{20}, e: -Scale - 1, inexact: true}, true, RoundCeil, n("-0.0000000000000000002")},

		// Far below the last place, where every digit goes.
		{wide{m: [3]uint64{1}, e: -100}, false, RoundInf, n("0.0000000000000000001")},
		{wide{m: [3]uint64{1}, e: -100}, false, RoundHalfInf, n("0")},
		{wide{e: -100}, false, RoundInf, n("0")},

		// Too many digits left after dividing.
		{wide{m: [3]uint64{0, 0, 10}, e: -Scale - 1}, false, RoundZero, Overflow},
	}
	for _, tc := range cases {
		if r := tc.w.Round(tc.neg, tc.mode); r != tc.e {
			t.Errorf("%+v neg=%v mode=%d: expected %q, got %q", tc.w, tc.neg, tc.mode, tc.e.String(), r.String())
		}
	}
}

func TestAddWide(t *testing.T) {
	t.Parallel()

	n := MustFromString
	max192 := [3]uint64{^uint64(0), ^uint64(0), ^uint64(0)}
	cases := []struct {
		a, b wide
		e    Number
		// inexact is whether the sum had to drop nonzero digits.
		inexact bool
	}{
		{wide{m: [3]uint64{15}, e: -1}, wide{m: [3]uint64{25}, e: -2}, n("1.75"), false},
		{wide{}, wide{m: [3]uint64{3}}, n("3"), false},
		{wide{m: [3]uint64{3}}, wide{}, n("3"), false},
		{wide{inexact: true}, wide{m: [3]uint64{3}}, n("3"), true},

		// Too far apart for any digit of b to survive.
		{wide{m: [3]uint64{1}}, wide{m: [3]uint64{1}, e: -100}, n("1"), true},
		{wide{m: [3]uint64{1}, e: -100}, wide{m: [3]uint64{1}}, n("1"), true},

		// The sum carries out of the top word, so a digit is dropped: a zero,
		// then a nonzero one.
		{wide{m: max192, e: -57}, wide{m: max192, e: -57}, n("12.5542034707733615276"), false},
		{wide{m: max192, e: -57}, wide{m: [3]uint64{^uint64(0) - 1, ^uint64(0), ^uint64(0)}, e: -57}, n("12.5542034707733615276"), true},
	}
	for _, tc := range cases {
		s := tc.a.Add(tc.b)
		if r := s.Round(false, RoundZero); r != tc.e || s.inexact != tc.inexact {
			t.Errorf("%+v + %+v: expected %q inexact=%v, got %q inexact=%v", tc.a, tc.b, tc.e.String(), tc.inexact, r.String(), s.inexact)
		}
	}
}

func TestSubWide(t *testing.T) {
	t.Parallel()

	n := MustFromString
	cases := []struct {
		a, b    wide
		e       Number
		neg     bool
		inexact bool
	}{
		{wide{m: [3]uint64{175}, e: -2}, wide{m: [3]uint64{25}, e: -2}, n("1.5"), false, false},
		{wide{m: [3]uint64{25}, e: -2}, wide{m: [3]uint64{175}, e: -2}, n("1.5"), true, false},
		{wide{m: [3]uint64{5}}, wide{m: [3]uint64{5}}, n("0"), false, false},
		{wide{m: [3]uint64{3}}, wide{}, n("3"), false, false},
		{wide{}, wide{m: [3]uint64{3}}, n("3"), true, false},

		// The smaller exponent comes first, and too far apart for any digit of
		// it to survive.
		{wide{m: [3]uint64{1}, e: -100}, wide{m: [3]uint64{1}}, n("1"), true, true},
		{wide{m: [3]uint64{1}}, wide{m: [3]uint64{1}, e: -100}, n("1"), false, true},
	}
	for _, tc := range cases {
		d, neg := tc.a.Sub(tc.b)
		if r := d.Round(false, RoundZero); r != tc.e || neg != tc.neg || d.inexact != tc.inexact {
			t.Errorf("%+v - %+v: expected %q neg=%v inexact=%v, got %q neg=%v inexact=%v", tc.a, tc.b, tc.e.String(), tc.neg, tc.inexact, r.String(), neg, d.inexact)
		}
	}
}

func TestRoundSnap(t *testing.T) {
	t.Parallel()

	n := MustFromString
	nines := [3]uint64{0x923fffffffffffff, 0xc51999090b65f67d, 0xa70c3c40a64e6} // 10^54-1
	cases := []struct {
		w       wide
		neg     bool
		mode    RoundingMode
		trusted int
		e       Number
	}{
		// 9.99…9 and 0.99…9, with 54 nines, are within the trusted digits of 10
		// and 1.
		{wide{m: nines, e: -53, inexact: true}, false, RoundZero, 50, n("10")},
		{wide{m: nines, e: -54, inexact: true}, false, RoundZero, 50, n("1")},
		// With too few trusted digits below the last place, it rounds as usual.
		{wide{m: nines, e: -54, inexact: true}, false, RoundZero, 25, n("0.9999999999999999999")},
		// An integer coefficient needs no snapping.
		{wide{m: [3]uint64{5}}, true, RoundZero, 50, n("-5")},
		// A positive value far below the last place is never exactly zero.
		{wide{m: [3]uint64{1}, e: -100, inexact: true}, false, RoundInf, 50, n("0.0000000000000000001")},
		{wide{m: [3]uint64{1}, e: -100, inexact: true}, false, RoundHalfInf, 50, n("0")},
		// Too many digits left after dividing.
		{wide{m: [3]uint64{0, 0, 1 << 40}, e: -30}, false, RoundZero, 80, Overflow},
	}
	for _, tc := range cases {
		if r := tc.w.RoundSnap(tc.neg, tc.mode, tc.trusted); r != tc.e {
			t.Errorf("%+v trusted=%d mode=%d: expected %q, got %q", tc.w, tc.trusted, tc.mode, tc.e.String(), r.String())
		}
	}
}

func TestSplitInteger(t *testing.T) {
	t.Parallel()

	n := MustFromString
	cases := []struct {
		w    wide
		k    int
		frac Number
	}{
		{wide{m: [3]uint64{4}, e: 1}, 40, n("0")},
		{wide{m: [3]uint64{4225}, e: -2}, 42, n("0.25")},
		{wide{m: [3]uint64{1}, e: -100}, 0, n("0")},
	}
	for _, tc := range cases {
		k, f := splitInteger(tc.w)
		if r := f.Round(false, RoundZero); k != tc.k || r != tc.frac {
			t.Errorf("%+v: expected %d %q, got %d %q", tc.w, tc.k, tc.frac.String(), k, r.String())
		}
	}
}
