package fixed_test

import (
	"fmt"
	"testing"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func TestSqrt(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		a fixed.Number
		e fixed.Number
	}{
		{n("0"), n("0")},
		{n("1"), n("1")},
		{n("4"), n("2")},
		{n("9"), n("3")},
		{n("2.25"), n("1.5")},
		{n("6.25"), n("2.5")},
		{n("1.44"), n("1.2")},
		{n("0.25"), n("0.5")},
		{n("0.000000000000000001"), n("0.000000001")},
		{n("1000000000000000000"), n("1000000000")},
		{n("2"), n("1.4142135623730950488")},
		{n("3"), n("1.7320508075688772935")},
		{n("10"), n("3.1622776601683793319")},
		{n("0.5"), n("0.7071067811865475244")},
		{n("0.0000000000000000001"), n("0.000000000316227766")},
		{n("0.0000000000000000002"), n("0.0000000004472135954")},
		{n("9999999999999999999.9999999999999999999"), n("3162277660.1683793319988935444")},

		{n("-1"), fixed.NaN},
		{n("-0.0000000000000000001"), fixed.NaN},
		{n("-9999999999999999999.9999999999999999999"), fixed.NaN},

		{fixed.NaN, fixed.NaN},
		{fixed.Overflow, fixed.Overflow},
		{fixed.DivisionByZero, fixed.DivisionByZero},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("Sqrt(%s)", tc.a), func(t *testing.T) {
			r := tc.a.Sqrt()
			if r != tc.e {
				t.Errorf("expected %q, got %q", tc.e.String(), r.String())
			}
		})
	}
}

func TestExp(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	all := func(e fixed.Number) [7]fixed.Number { return [7]fixed.Number{e, e, e, e, e, e, e} }
	// cols lays out a positive result that is not a tie: the modes towards zero
	// and down take down, the modes away from zero and up take up, and the
	// nearest modes take up only past half the last place.
	cols := func(down, up fixed.Number, nearestUp bool) [7]fixed.Number {
		near := down
		if nearestUp {
			near = up
		}
		return [7]fixed.Number{down, up, down, up, near, near, near}
	}
	cases := []struct {
		a fixed.Number
		// zero, inf, floor, ceil, halfZero, halfInf, bank
		e [7]fixed.Number
	}{
		{n("0"), all(fixed.One)},
		{n("1"), cols(n("2.7182818284590452353"), n("2.7182818284590452354"), true)},
		{n("-1"), cols(n("0.3678794411714423215"), n("0.3678794411714423216"), true)},
		{n("0.5"), cols(n("1.6487212707001281468"), n("1.6487212707001281469"), false)},
		{n("2"), cols(n("7.3890560989306502272"), n("7.3890560989306502273"), false)},
		{n("10"), cols(n("22026.4657948067165169579"), n("22026.465794806716516958"), false)},
		{n("-10"), cols(n("0.0000453999297624848"), n("0.0000453999297624849"), true)},
		{n("43"), cols(n("4727839468229346561.4744575627442803708"), n("4727839468229346561.4744575627442803709"), false)},
		{n("0.6931471805599453094"), cols(n("1.9999999999999999999"), n("2"), true)},
		{n("2.302585092994045684"), cols(n("9.9999999999999999998"), n("9.9999999999999999999"), false)},
		{n("0.0000000000000000001"), cols(n("1.0000000000000000001"), n("1.0000000000000000002"), false)},
		{n("-0.0000000000000000001"), cols(n("0.9999999999999999999"), n("1"), false)},

		// Either side of the largest power that fits.
		{n("43.7491167668868679963"), cols(n("9999999999999999999.5816236099708005557"), n("9999999999999999999.5816236099708005558"), true)},
		{n("43.7491167668868679964"), all(fixed.Overflow)},
		{n("44"), all(fixed.Overflow)},
		{n("1000000000000000000"), all(fixed.Overflow)},

		// Either side of half the last place, and well below it.
		{n("-44"), cols(n("0"), n("0.0000000000000000001"), true)},
		{n("-44.4422639474468133057"), cols(n("0"), n("0.0000000000000000001"), true)},
		{n("-44.4422639474468133058"), cols(n("0"), n("0.0000000000000000001"), false)},
		{n("-45"), cols(n("0"), n("0.0000000000000000001"), false)},
		{n("-1000000000000000000"), cols(n("0"), n("0.0000000000000000001"), false)},

		{fixed.NaN, all(fixed.NaN)},
		{fixed.Overflow, all(fixed.Overflow)},
	}
	modes := [...]fixed.RoundingMode{fixed.RoundZero, fixed.RoundInf, fixed.RoundFloor, fixed.RoundCeil, fixed.RoundHalfZero, fixed.RoundHalfInf, fixed.RoundBank}
	for _, tc := range cases {
		t.Run(tc.a.String(), func(t *testing.T) {
			for i, mode := range modes {
				if r := tc.a.ExpRound(mode); r != tc.e[i] {
					t.Errorf("mode %d: expected %q, got %q", mode, tc.e[i].String(), r.String())
				}
			}
			if r := tc.a.Exp(); r != tc.e[0] {
				t.Errorf("Exp: expected %q, got %q", tc.e[0].String(), r.String())
			}
		})
	}
}

func TestLn(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	all := func(e fixed.Number) [7]fixed.Number { return [7]fixed.Number{e, e, e, e, e, e, e} }
	row := func(zero, inf, floor, ceil, halfZero, halfInf, bank string) [7]fixed.Number {
		return [7]fixed.Number{n(zero), n(inf), n(floor), n(ceil), n(halfZero), n(halfInf), n(bank)}
	}
	cases := []struct {
		a fixed.Number
		// zero, inf, floor, ceil, halfZero, halfInf, bank
		e [7]fixed.Number
	}{
		{n("1"), all(fixed.Zero)},
		{n("2"), row("0.6931471805599453094", "0.6931471805599453095", "0.6931471805599453094", "0.6931471805599453095", "0.6931471805599453094", "0.6931471805599453094", "0.6931471805599453094")},
		{n("10"), row("2.302585092994045684", "2.3025850929940456841", "2.302585092994045684", "2.3025850929940456841", "2.302585092994045684", "2.302585092994045684", "2.302585092994045684")},
		{n("0.5"), row("-0.6931471805599453094", "-0.6931471805599453095", "-0.6931471805599453095", "-0.6931471805599453094", "-0.6931471805599453094", "-0.6931471805599453094", "-0.6931471805599453094")},
		{n("0.0001"), row("-9.210340371976182736", "-9.2103403719761827361", "-9.2103403719761827361", "-9.210340371976182736", "-9.2103403719761827361", "-9.2103403719761827361", "-9.2103403719761827361")},
		{n("1.05"), row("0.048790164169432003", "0.0487901641694320031", "0.048790164169432003", "0.0487901641694320031", "0.0487901641694320031", "0.0487901641694320031", "0.0487901641694320031")},
		{n("123.456"), row("4.8158848172832638831", "4.8158848172832638832", "4.8158848172832638831", "4.8158848172832638832", "4.8158848172832638831", "4.8158848172832638831", "4.8158848172832638831")},

		// Either side of e.
		{n("2.7182818284590452353"), row("0.9999999999999999999", "1", "0.9999999999999999999", "1", "1", "1", "1")},
		{n("2.7182818284590452354"), row("1", "1.0000000000000000001", "1", "1.0000000000000000001", "1", "1", "1")},

		// Next to one, where the logarithm is below the last place.
		{n("1.0000000000000000001"), row("0", "0.0000000000000000001", "0", "0.0000000000000000001", "0.0000000000000000001", "0.0000000000000000001", "0.0000000000000000001")},
		{n("0.9999999999999999999"), row("-0.0000000000000000001", "-0.0000000000000000002", "-0.0000000000000000002", "-0.0000000000000000001", "-0.0000000000000000001", "-0.0000000000000000001", "-0.0000000000000000001")},

		// The range limits.
		{fixed.Highest, row("43.7491167668868679963", "43.7491167668868679964", "43.7491167668868679963", "43.7491167668868679964", "43.7491167668868679963", "43.7491167668868679963", "43.7491167668868679963")},
		{n("0.0000000000000000001"), row("-43.7491167668868679963", "-43.7491167668868679964", "-43.7491167668868679964", "-43.7491167668868679963", "-43.7491167668868679963", "-43.7491167668868679963", "-43.7491167668868679963")},

		{n("0"), all(fixed.NaN)},
		{n("-1"), all(fixed.NaN)},
		{fixed.Lowest, all(fixed.NaN)},
		{fixed.Overflow, all(fixed.Overflow)},
	}
	modes := [...]fixed.RoundingMode{fixed.RoundZero, fixed.RoundInf, fixed.RoundFloor, fixed.RoundCeil, fixed.RoundHalfZero, fixed.RoundHalfInf, fixed.RoundBank}
	for _, tc := range cases {
		t.Run(tc.a.String(), func(t *testing.T) {
			for i, mode := range modes {
				if r := tc.a.LnRound(mode); r != tc.e[i] {
					t.Errorf("mode %d: expected %q, got %q", mode, tc.e[i].String(), r.String())
				}
			}
			if r := tc.a.Ln(); r != tc.e[0] {
				t.Errorf("Ln: expected %q, got %q", tc.e[0].String(), r.String())
			}
		})
	}
}

func TestPow(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	all := func(e fixed.Number) [7]fixed.Number { return [7]fixed.Number{e, e, e, e, e, e, e} }
	row := func(zero, inf, floor, ceil, halfZero, halfInf, bank string) [7]fixed.Number {
		return [7]fixed.Number{n(zero), n(inf), n(floor), n(ceil), n(halfZero), n(halfInf), n(bank)}
	}
	cases := []struct {
		a, y fixed.Number
		// zero, inf, floor, ceil, halfZero, halfInf, bank
		e [7]fixed.Number
	}{
		{n("2"), n("0.5"), row("1.4142135623730950488", "1.4142135623730950489", "1.4142135623730950488", "1.4142135623730950489", "1.4142135623730950488", "1.4142135623730950488", "1.4142135623730950488")},
		{n("10"), n("1.5"), row("31.6227766016837933199", "31.62277660168379332", "31.6227766016837933199", "31.62277660168379332", "31.62277660168379332", "31.62277660168379332", "31.62277660168379332")},
		{n("2"), n("-0.5"), row("0.7071067811865475244", "0.7071067811865475245", "0.7071067811865475244", "0.7071067811865475245", "0.7071067811865475244", "0.7071067811865475244", "0.7071067811865475244")},
		{n("0.5"), n("-1.5"), row("2.8284271247461900976", "2.8284271247461900977", "2.8284271247461900976", "2.8284271247461900977", "2.8284271247461900976", "2.8284271247461900976", "2.8284271247461900976")},
		{n("1.05"), n("0.5"), row("1.0246950765959598383", "1.0246950765959598384", "1.0246950765959598383", "1.0246950765959598384", "1.0246950765959598383", "1.0246950765959598383", "1.0246950765959598383")},
		{n("123.456"), n("2.7182818284590452354"), row("484528.8172018086663059364", "484528.8172018086663059365", "484528.8172018086663059364", "484528.8172018086663059365", "484528.8172018086663059364", "484528.8172018086663059364", "484528.8172018086663059364")},

		// Exact results, and one just short of being exact.
		{n("4"), n("0.5"), all(n("2"))},
		{n("2.25"), n("0.5"), all(n("1.5"))},
		{n("16"), n("0.25"), all(n("2"))},
		{n("1.21"), n("1.5"), all(n("1.331"))},
		{n("1.21"), n("-0.5"), [7]fixed.Number{n("0.9090909090909090909"), n("0.909090909090909091"), n("0.9090909090909090909"), n("0.909090909090909091"), n("0.9090909090909090909"), n("0.9090909090909090909"), n("0.9090909090909090909")}},
		{n("8"), n("0.3333333333333333333"), row("1.9999999999999999998", "1.9999999999999999999", "1.9999999999999999998", "1.9999999999999999999", "1.9999999999999999999", "1.9999999999999999999", "1.9999999999999999999")},

		// An exact tie at the last place.
		{n("0.00000025"), n("2.5"), row("0.0000000000000000312", "0.0000000000000000313", "0.0000000000000000312", "0.0000000000000000313", "0.0000000000000000312", "0.0000000000000000313", "0.0000000000000000312")},

		// Integer exponents, including ones too large for PowIntRound.
		{n("-2"), n("3"), all(n("-8"))},
		{n("1.0000000000000000001"), n("9999999999999999999"), row("2.7182818284590452349", "2.718281828459045235", "2.7182818284590452349", "2.718281828459045235", "2.718281828459045235", "2.718281828459045235", "2.718281828459045235")},
		{n("-0.5"), n("9999999999999999999"), row("0", "-0.0000000000000000001", "-0.0000000000000000001", "0", "0", "0", "0")},
		{n("-1"), n("9999999999999999999"), all(n("-1"))},
		{n("-2"), n("9999999999999999998"), all(fixed.Overflow)},

		// Out of range either way.
		{n("2"), n("63.5"), all(fixed.Overflow)},
		{n("0.5"), n("64.5"), row("0", "0.0000000000000000001", "0", "0.0000000000000000001", "0", "0", "0")},

		{n("1"), n("123.456"), all(fixed.One)},
		{n("123.456"), n("0"), all(fixed.One)},
		{n("0"), n("0.5"), all(fixed.Zero)},
		{n("0"), n("-0.5"), all(fixed.DivisionByZero)},
		{n("-8"), n("0.5"), all(fixed.NaN)},
		{fixed.NaN, n("0.5"), all(fixed.NaN)},
		{n("2"), fixed.Overflow, all(fixed.Overflow)},
	}
	modes := [...]fixed.RoundingMode{fixed.RoundZero, fixed.RoundInf, fixed.RoundFloor, fixed.RoundCeil, fixed.RoundHalfZero, fixed.RoundHalfInf, fixed.RoundBank}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s^%s", tc.a, tc.y), func(t *testing.T) {
			for i, mode := range modes {
				if r := tc.a.PowRound(tc.y, mode); r != tc.e[i] {
					t.Errorf("mode %d: expected %q, got %q", mode, tc.e[i].String(), r.String())
				}
			}
			if r := tc.a.Pow(tc.y); r != tc.e[0] {
				t.Errorf("Pow: expected %q, got %q", tc.e[0].String(), r.String())
			}
		})
	}
}
