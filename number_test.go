package fixed_test

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"testing"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func Example() {
	price := fixed.MustFromString("19.99")
	rate := fixed.MustFromString("0.0825")
	quantity := 3

	subtotal := price.MulInt(quantity)                  // 59.97
	tax := subtotal.Mul(rate).Round(2, fixed.RoundBank) // 4.947525 -> 4.95
	total := subtotal.Add(tax)                          // 64.92

	share := total.DivInt(7) // 9.2742857142857142857
	if err := share.Err(); err != nil {
		// NaN, overflow or division by zero anywhere above.
	}
	fmt.Println(share.Round(2, fixed.RoundHalfInf)) // 9.27

	//Output: 9.27
}

func TestNew(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		coef int64
		exp  int
		want fixed.Number
	}{
		{12345, -2, n("123.45")},
		{-12345, -2, n("-123.45")},
		{1, 0, fixed.One},
		{1, -1, n("0.1")},
		{0, 100, fixed.Zero},
		{0, -100, fixed.Zero},
		{0, math.MinInt, fixed.Zero},
		{5, -19, n("0.0000000000000000005")},
		{500, -21, n("0.0000000000000000005")},
		{5, -20, fixed.NaN},
		{501, -21, fixed.NaN},
		{math.MinInt64, -19, n("-0.9223372036854775808")},
		{math.MinInt64, -20, fixed.NaN},
		{1, math.MinInt, fixed.NaN},
		{123, 16, n("1230000000000000000")},
		{123, 17, fixed.Overflow},
		{315, 17, fixed.Overflow},
		{-315, 17, fixed.Overflow},
		{1, 18, n("1000000000000000000")},
		{1, 19, fixed.Overflow},
		{-1, 19, fixed.Overflow},
		{1, 20, fixed.Overflow},
		{1, 37, fixed.Overflow},
		{-1, 38, fixed.Overflow},
		{math.MaxInt64, 0, n("9223372036854775807")},
		{math.MaxInt64, 1, fixed.Overflow},
		{math.MinInt64, 0, n("-9223372036854775808")},
		{1, math.MaxInt, fixed.Overflow},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%de%d", tc.coef, tc.exp), func(t *testing.T) {
			if got := fixed.New(tc.coef, tc.exp); got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestFromWords(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		lo, hi uint64
		want   fixed.Number
		err    error
	}{
		{"zero", 0, 0, fixed.Zero, nil},
		{"one", fixed.One.Lo, fixed.One.Hi, fixed.One, nil},
		{"negative", fixed.FromInt(-1).Lo, fixed.FromInt(-1).Hi, fixed.FromInt(-1), nil},
		{"highest", fixed.Highest.Lo, fixed.Highest.Hi, fixed.Highest, nil},
		{"lowest", fixed.Lowest.Lo, fixed.Lowest.Hi, fixed.Lowest, nil},
		{"above highest", fixed.Highest.Lo + 2, fixed.Highest.Hi + 1, fixed.Number{}, fixed.ErrNaN},
		{"below lowest", fixed.Lowest.Lo - 1, fixed.Lowest.Hi, fixed.Number{}, fixed.ErrNaN},
		{"most negative words", 0, 1 << 63, fixed.Number{}, fixed.ErrNaN},
		{"nan", fixed.NaN.Lo, fixed.NaN.Hi, fixed.Number{}, fixed.ErrNaN},
		{"overflow", fixed.Overflow.Lo, fixed.Overflow.Hi, fixed.Number{}, fixed.ErrOverflow},
		{"division by zero", fixed.DivisionByZero.Lo, fixed.DivisionByZero.Hi, fixed.Number{}, fixed.ErrDivisionByZero},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, err := fixed.NewWords(tc.lo, tc.hi)
			if n != tc.want || err != tc.err {
				t.Errorf("expected (%v, %v), got (%v, %v)", tc.want, tc.err, n, err)
			}
		})
	}
}

func TestSign(t *testing.T) {
	t.Parallel()

	cases := []struct {
		n fixed.Number
		e int
	}{
		{fixed.Number{}, 0},
		{fixed.FromInt(0), 0},
		{fixed.FromInt(1), 1},
		{fixed.Number{Lo: 0, Hi: 1}, 1},
		{fixed.MustFromString("0.0000000000000000001"), 1},
		{fixed.MustFromString("-0.0000000000000000001"), -1},
		{fixed.Highest, 1},
		{fixed.Lowest, -1},
		{fixed.FromInt(-1), -1},
		{fixed.FromInt(1234567), 1},
		{fixed.FromInt(-1234567), -1},
		{fixed.FromInt(math.MaxInt), 1},
		{fixed.FromInt(math.MinInt), -1},
	}
	for _, tc := range cases {
		r := tc.n.Sign()
		if r != tc.e {
			t.Errorf("expected %d, got %d", tc.e, r)
		}

		var neg, pos, zero bool
		switch tc.e {
		case -1:
			neg = true
		case 0:
			zero = true
		case +1:
			pos = true
		}
		if neg != tc.n.IsNegative() {
			t.Errorf("expected negative = %v", neg)
		}
		if zero != tc.n.IsZero() {
			t.Errorf("expected zero = %v", zero)
		}
		if pos != tc.n.IsPositive() {
			t.Errorf("expected positive = %v", pos)
		}
	}
}

func TestAbs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		n fixed.Number
		e int
	}{
		{fixed.FromInt(0), 0},
		{fixed.FromInt(1), 1},
		{fixed.FromInt(-1), 1},
		{fixed.FromInt(1234567), 1234567},
		{fixed.FromInt(-1234567), 1234567},
		{fixed.FromInt(math.MaxInt), 9223372036854775807},
		{fixed.FromInt(-9223372036854775807), 9223372036854775807},
		{fixed.FromInt(math.MinInt), -9223372036854775808},
	}
	for _, tc := range cases {
		r := tc.n.Abs().Int()
		if r != tc.e {
			t.Errorf("expected %d, got %d", tc.e, r)
		}
	}
}

func TestNeg(t *testing.T) {
	t.Parallel()

	cases := []struct {
		n fixed.Number
		e fixed.Number
	}{
		{fixed.FromInt(0), fixed.FromInt(0)},
		{fixed.FromInt(1), fixed.FromInt(-1)},
		{fixed.FromInt(-1), fixed.FromInt(1)},
		{fixed.FromInt(1234567), fixed.FromInt(-1234567)},
		{fixed.FromInt(-1234567), fixed.FromInt(1234567)},
		{fixed.FromInt(9223372036854775807), fixed.FromInt(-9223372036854775807)},
		{fixed.FromInt(-9223372036854775807), fixed.FromInt(9223372036854775807)},
		{fixed.Overflow, fixed.Overflow},
	}
	for _, tc := range cases {
		r := tc.n.Neg()
		if r != tc.e {
			t.Errorf("expected %s, got %s", tc.e, r)
		}
	}
}

func TestCmp(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		a, b fixed.Number
		e    int
	}{
		{n("0"), n("0"), 0},
		{n("1"), n("0"), 1},
		{n("0"), n("1"), -1},
		{n("0.0000000000000000001"), n("0"), 1},
		{n("0"), n("0.0000000000000000001"), -1},
		{n("-1"), n("0"), -1},
		{n("0"), n("-1"), 1},
		{n("-0.0000000000000000001"), n("0"), -1},
		{n("0"), n("-0.0000000000000000001"), 1},
		{n("2.7182"), n("3.1415"), -1},
		{n("-2.7182"), n("-3.1415"), 1},
		{n("3.1415"), n("2.7182"), 1},
		{n("-3.1415"), n("-2.7182"), -1},
		{n("9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999"), 1},
		{n("-9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999"), -1},
		{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999"), 0},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("Cmp(%s,%s)", tc.a, tc.b), func(t *testing.T) {
			if r := tc.a.Cmp(tc.b); r != tc.e {
				t.Errorf("expected %d, got %d", tc.e, r)
			}
			for _, c := range []struct {
				name string
				got  bool
				exp  bool
			}{
				{"Eq", tc.a.Eq(tc.b), tc.e == 0},
				{"Lt", tc.a.Lt(tc.b), tc.e < 0},
				{"Lte", tc.a.Lte(tc.b), tc.e <= 0},
				{"Gt", tc.a.Gt(tc.b), tc.e > 0},
				{"Gte", tc.a.Gte(tc.b), tc.e >= 0},
			} {
				if c.got != c.exp {
					t.Errorf("%s: expected %v, got %v", c.name, c.exp, c.got)
				}
			}
		})
	}
}

func TestSort(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString

	in := []fixed.Number{
		fixed.NaN,
		fixed.Overflow,
		fixed.DivisionByZero,
		n("0.0000000000000000001"),
		n("-1"),
		n("-2.7182"),
		n("0"),
		n("1"),
		n("-0.0000000000000000001"),
		n("9999999999999999999.9999999999999999999"),
		n("3.1415"),
		n("2.7182"),
		n("-9999999999999999999.9999999999999999999"),
		n("-3.1415"),
	}

	exp := []fixed.Number{
		n("-9999999999999999999.9999999999999999999"),
		n("-3.1415"),
		n("-2.7182"),
		n("-1"),
		n("-0.0000000000000000001"),
		n("0"),
		n("0.0000000000000000001"),
		n("1"),
		n("2.7182"),
		n("3.1415"),
		n("9999999999999999999.9999999999999999999"),
		fixed.NaN,
		fixed.Overflow,
		fixed.DivisionByZero,
	}

	slices.SortFunc(in, fixed.Number.Cmp)

	if !reflect.DeepEqual(in, exp) {
		t.Errorf("expected %v, got %v", exp, in)
	}
}

func TestMin(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		ns []fixed.Number
		e  fixed.Number
	}{
		{nil, fixed.NaN},
		{[]fixed.Number{n("0")}, n("0")},
		{[]fixed.Number{n("42")}, n("42")},
		{[]fixed.Number{n("-42")}, n("-42")},
		{[]fixed.Number{n("1"), n("2")}, n("1")},
		{[]fixed.Number{n("2"), n("1")}, n("1")},
		{[]fixed.Number{n("-1"), n("1")}, n("-1")},
		{[]fixed.Number{n("1"), n("-1")}, n("-1")},
		{[]fixed.Number{n("0"), n("0")}, n("0")},
		{[]fixed.Number{n("0"), n("-0.0000000000000000001")}, n("-0.0000000000000000001")},
		{[]fixed.Number{n("0.0000000000000000001"), n("0.0000000000000000002")}, n("0.0000000000000000001")},
		{[]fixed.Number{n("3"), n("1"), n("2")}, n("1")},
		{[]fixed.Number{n("-3"), n("-1"), n("-2")}, n("-3")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("-9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999")}, n("9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("-9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("-9999999999999999999.9999999999999999999")},

		{[]fixed.Number{fixed.NaN}, fixed.NaN},
		{[]fixed.Number{fixed.NaN, n("1")}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.NaN}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.Overflow, n("2")}, fixed.Overflow},
		{[]fixed.Number{n("1"), fixed.DivisionByZero}, fixed.DivisionByZero},
		{[]fixed.Number{fixed.NaN, fixed.Overflow}, fixed.NaN},
		{[]fixed.Number{fixed.Overflow, fixed.NaN}, fixed.Overflow},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.ns), func(t *testing.T) {
			r := fixed.Min(tc.ns...)
			if r != tc.e {
				t.Errorf("expected %q, got %q", tc.e.String(), r.String())
			}
		})
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		ns []fixed.Number
		e  fixed.Number
	}{
		{nil, fixed.NaN},
		{[]fixed.Number{n("0")}, n("0")},
		{[]fixed.Number{n("42")}, n("42")},
		{[]fixed.Number{n("-42")}, n("-42")},
		{[]fixed.Number{n("1"), n("2")}, n("2")},
		{[]fixed.Number{n("2"), n("1")}, n("2")},
		{[]fixed.Number{n("-1"), n("1")}, n("1")},
		{[]fixed.Number{n("1"), n("-1")}, n("1")},
		{[]fixed.Number{n("0"), n("0")}, n("0")},
		{[]fixed.Number{n("0"), n("0.0000000000000000001")}, n("0.0000000000000000001")},
		{[]fixed.Number{n("-0.0000000000000000002"), n("-0.0000000000000000001")}, n("-0.0000000000000000001")},
		{[]fixed.Number{n("1"), n("3"), n("2")}, n("3")},
		{[]fixed.Number{n("-1"), n("-3"), n("-2")}, n("-1")},
		{[]fixed.Number{n("-9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999")}, n("9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999")}, n("9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("-9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("-9999999999999999999.9999999999999999999")},

		{[]fixed.Number{fixed.NaN}, fixed.NaN},
		{[]fixed.Number{fixed.NaN, n("1")}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.NaN}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.Overflow, n("2")}, fixed.Overflow},
		{[]fixed.Number{n("1"), fixed.DivisionByZero}, fixed.DivisionByZero},
		{[]fixed.Number{fixed.NaN, fixed.Overflow}, fixed.NaN},
		{[]fixed.Number{fixed.Overflow, fixed.NaN}, fixed.Overflow},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.ns), func(t *testing.T) {
			r := fixed.Max(tc.ns...)
			if r != tc.e {
				t.Errorf("expected %q, got %q", tc.e.String(), r.String())
			}
		})
	}
}

func TestSum(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		ns []fixed.Number
		e  fixed.Number
	}{
		{nil, n("0")},
		{[]fixed.Number{n("42")}, n("42")},
		{[]fixed.Number{n("-42")}, n("-42")},
		{[]fixed.Number{n("21"), n("21")}, n("42")},
		{[]fixed.Number{n("1"), n("-1")}, n("0")},
		{[]fixed.Number{n("1"), n("2"), n("3"), n("4")}, n("10")},
		{[]fixed.Number{n("0.1"), n("0.2")}, n("0.3")},
		{[]fixed.Number{n("0.0000000000000000001"), n("0.0000000000000000001")}, n("0.0000000000000000002")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("0")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999998"), n("0.0000000000000000001")}, n("9999999999999999999.9999999999999999999")},

		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("0.0000000000000000001")}, fixed.Overflow},
		{[]fixed.Number{n("-9999999999999999999.9999999999999999999"), n("-0.0000000000000000001")}, fixed.Overflow},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, fixed.Overflow},

		{[]fixed.Number{fixed.NaN}, fixed.NaN},
		{[]fixed.Number{fixed.NaN, n("1")}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.NaN}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.DivisionByZero, n("2")}, fixed.DivisionByZero},
		{[]fixed.Number{fixed.NaN, fixed.Overflow}, fixed.NaN},
		{[]fixed.Number{fixed.Overflow, fixed.NaN}, fixed.Overflow},
		{[]fixed.Number{fixed.Highest, fixed.Highest, fixed.NaN}, fixed.Overflow},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.ns), func(t *testing.T) {
			r := fixed.Sum(tc.ns...)
			if r != tc.e {
				t.Errorf("expected %q, got %q", tc.e.String(), r.String())
			}
		})
	}
}

func TestAvg(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	cases := []struct {
		ns []fixed.Number
		e  fixed.Number
	}{
		{nil, fixed.NaN},
		{[]fixed.Number{n("0")}, n("0")},
		{[]fixed.Number{n("42")}, n("42")},
		{[]fixed.Number{n("-42")}, n("-42")},
		{[]fixed.Number{n("1"), n("2")}, n("1.5")},
		{[]fixed.Number{n("7"), n("11")}, n("9")},
		{[]fixed.Number{n("1"), n("2"), n("3")}, n("2")},
		{[]fixed.Number{n("1"), n("-1")}, n("0")},
		{[]fixed.Number{n("-1"), n("-2"), n("-3")}, n("-2")},
		{[]fixed.Number{n("0.1"), n("0.2")}, n("0.15")},
		{[]fixed.Number{n("0.0000000000000000001"), n("0.0000000000000000003")}, n("0.0000000000000000002")},

		// Truncated towards zero.
		{[]fixed.Number{n("10"), n("0"), n("0")}, n("3.3333333333333333333")},
		{[]fixed.Number{n("-10"), n("0"), n("0")}, n("-3.3333333333333333333")},
		{[]fixed.Number{n("1"), n("2"), n("2")}, n("1.6666666666666666666")},
		{[]fixed.Number{n("-1"), n("-2"), n("-2")}, n("-1.6666666666666666666")},
		{[]fixed.Number{n("0.0000000000000000001"), n("0")}, n("0")},
		{[]fixed.Number{n("-0.0000000000000000001"), n("0")}, n("0")},

		// The sum is accumulated in 256 bits and never overflows.
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999")}, n("9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("-9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("-9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999")}, n("9999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("0")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("0")}, n("4999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("-9999999999999999999.9999999999999999999"), n("0")}, n("-4999999999999999999.9999999999999999999")},
		{[]fixed.Number{n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999"), n("9999999999999999999.9999999999999999999"), n("-9999999999999999999.9999999999999999999")}, n("4999999999999999999.9999999999999999999")},

		{[]fixed.Number{fixed.NaN}, fixed.NaN},
		{[]fixed.Number{fixed.NaN, n("1")}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.NaN}, fixed.NaN},
		{[]fixed.Number{n("1"), fixed.DivisionByZero, n("2")}, fixed.DivisionByZero},
		{[]fixed.Number{fixed.NaN, fixed.Overflow}, fixed.NaN},
		{[]fixed.Number{fixed.Overflow, fixed.NaN}, fixed.Overflow},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.ns), func(t *testing.T) {
			r := fixed.Avg(tc.ns...)
			if r != tc.e {
				t.Errorf("expected %q, got %q", tc.e.String(), r.String())
			}
		})
	}
}
