package fixed_test

import (
	"math"
	"testing"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func TestFromInt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in int64
	}{
		{0},
		{1},
		{-1},
		{2},
		{-2},
		{12345},
		{-12345},
		{999999999999999999},
		{-999999999999999999},
		{math.MaxInt64},
		{math.MinInt64},
	} {
		n := fixed.FromInt(tc.in)
		out := int64(n.Int())
		if out != tc.in {
			t.Errorf("expected %d, got %d", tc.in, out)
		}
	}
}

func TestFromIntTypes(t *testing.T) {
	t.Parallel()

	check := func(in int64, n fixed.Number) {
		t.Helper()
		if out := int64(n.Int()); out != in {
			t.Errorf("expected %d, got %d", in, out)
		}
	}
	for _, v := range []int8{0, 1, -1, math.MaxInt8, math.MinInt8} {
		check(int64(v), fixed.FromInt(v))
	}
	for _, v := range []int16{0, 1, -1, math.MaxInt16, math.MinInt16} {
		check(int64(v), fixed.FromInt(v))
	}
	for _, v := range []int32{0, 1, -1, math.MaxInt32, math.MinInt32} {
		check(int64(v), fixed.FromInt(v))
	}
	for _, v := range []int{0, 1, -1, math.MaxInt, math.MinInt} {
		check(int64(v), fixed.FromInt(v))
	}
}

func TestFromUint(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in uint64
	}{
		{0},
		{1},
		{12345},
		{math.MaxInt64},
		{9999999999999999999},
	} {
		n := fixed.FromUint(tc.in)
		out := uint64(n.Uint())
		if out != tc.in {
			t.Errorf("expected %d, got %d", tc.in, out)
		}
	}
	for _, in := range []uint64{
		10000000000000000000,
		math.MaxUint64,
	} {
		if n := fixed.FromUint(in); n != fixed.Overflow {
			t.Errorf("expected overflow for %d, got %q", in, n)
		}
	}
}

func TestFromUintTypes(t *testing.T) {
	t.Parallel()

	check := func(in uint64, n fixed.Number) {
		t.Helper()
		if out := uint64(n.Uint()); out != in {
			t.Errorf("expected %d, got %d", in, out)
		}
	}
	for _, v := range []uint8{0, 1, math.MaxUint8} {
		check(uint64(v), fixed.FromUint(v))
	}
	for _, v := range []uint16{0, 1, math.MaxUint16} {
		check(uint64(v), fixed.FromUint(v))
	}
	for _, v := range []uint32{0, 1, math.MaxUint32} {
		check(uint64(v), fixed.FromUint(v))
	}
	for _, v := range []uint{0, 1, 9999999999999999999} {
		check(uint64(v), fixed.FromUint(v))
	}
	if n := fixed.FromUint(uint(math.MaxUint)); n != fixed.Overflow {
		t.Errorf("expected overflow, got %q", n)
	}
}

func TestInt(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, tc := range []struct {
		n fixed.Number
		e int64
	}{
		{n("0"), 0},
		{n("0.9999999999999999999"), 0},
		{n("-0.9999999999999999999"), 0},
		{n("0.0000000000000000001"), 0},
		{n("-0.0000000000000000001"), 0},
		{n("2.5"), 2},
		{n("-2.5"), -2},
		{n("-3"), -3},
		{n("-3.0000000000000000001"), -3},
		{n("9223372036854775807.9999999999999999999"), math.MaxInt64},
		{n("-9223372036854775808.9999999999999999999"), math.MinInt64},

		// The integer part wraps around once it leaves the int64 range.
		{n("9223372036854775808"), math.MinInt64},
		{n("-9223372036854775809"), math.MaxInt64},
		{n("9999999999999999999.9999999999999999999"), -8446744073709551617},
		{n("-9999999999999999999.9999999999999999999"), 8446744073709551617},

		{fixed.NaN, 0},
		{fixed.Overflow, 0},
	} {
		t.Run(tc.n.String(), func(t *testing.T) {
			if r := int64(tc.n.Int()); r != tc.e {
				t.Errorf("expected %d, got %d", tc.e, r)
			}
		})
	}
}

func TestUint(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, tc := range []struct {
		n fixed.Number
		e uint64
	}{
		{n("0"), 0},
		{n("0.9999999999999999999"), 0},
		{n("-0.9999999999999999999"), 0},
		{n("2.5"), 2},
		{n("9999999999999999999.9999999999999999999"), 9999999999999999999},

		// A negative integer part wraps around, as uint(int64(v)) does.
		{n("-1"), math.MaxUint64},
		{n("-2.5"), math.MaxUint64 - 1},
		{n("-5"), math.MaxUint64 - 4},
		{n("-9999999999999999999.9999999999999999999"), 8446744073709551617},

		{fixed.NaN, 0},
		{fixed.Overflow, 0},
	} {
		t.Run(tc.n.String(), func(t *testing.T) {
			if r := uint64(tc.n.Uint()); r != tc.e {
				t.Errorf("expected %d, got %d", tc.e, r)
			}
		})
	}
}

func TestFromFloat64(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, tc := range []struct {
		in  float64
		exp fixed.Number
	}{
		{math.NaN(), fixed.NaN},
		{math.Inf(+1), fixed.NaN},
		{math.Inf(-1), fixed.NaN},
		{0.0, n("0")},
		{1.0, n("1")},
		{-1.0, n("-1")},
		{0.5, n("0.5")},
		{0.1, n("0.1")},
		{3.14159264, n("3.14159264")},
		{-3.14159264, n("-3.14159264")},
		{12345678.12345678, n("12345678.12345678")},
		{-12345678.12345678, n("-12345678.12345678")},
		{0.9999999999999999999, n("1")},
		{-0.9999999999999999999, n("-1")},
		{0.0000000000000000001, n("0.0000000000000000001")},
		{-0.0000000000000000001, n("-0.0000000000000000001")},
		{1000000000000000000, n("1000000000000000000")},
		{-1000000000000000000, n("-1000000000000000000")},
		{math.MaxInt64, n("9223372036854776000")},
		{1.7976931348623157, n("1.7976931348623157")},
		{math.MaxFloat64, fixed.Overflow},
		{math.Copysign(0, -1), n("0")},
		{9e18, n("9000000000000000000")},
		// Their products with a power of ten fall just short of the integer, so
		// the digits have to be rounded, not truncated.
		{4.35, n("4.35")},
		{0.29, n("0.29")},
		{1.005, n("1.005")},
		{-9.95, n("-9.95")},
		// Too long for the fast path.
		{math.Nextafter(0.3, 1), n("0.30000000000000004")},
		{1e19, fixed.Overflow},
		{3.15e19, fixed.Overflow},
		{-3.15e19, fixed.Overflow},

		// Digits beyond Scale are rounded half to even.
		{1e-20, n("0")},
		{-1e-20, n("0")},
		{5e-20, n("0")},
		{6e-20, n("0.0000000000000000001")},
		{1.5e-19, n("0.0000000000000000002")},
		{-1.5e-19, n("-0.0000000000000000002")},
		{2.5e-19, n("0.0000000000000000002")},
		{-2.5e-19, n("-0.0000000000000000002")},
		{1.2345678901234568e-05, n("0.0000123456789012346")},
		{1.0 / 30000, n("0.0000333333333333333")},
		{0.30000000000000004, n("0.30000000000000004")},
		{5.000000000000001e-20, n("0.0000000000000000001")},
		{9.999999999999999e-20, n("0.0000000000000000001")},
		{1.2345678901234567e-36, n("0")},
		{math.SmallestNonzeroFloat64, n("0")},
	} {
		r := fixed.FromFloat64(tc.in)
		if r != tc.exp {
			t.Errorf("expected %s, got %s", tc.exp.String(), r.String())
		}
	}
}

func TestFromFloat64Exact(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, tc := range []struct {
		in  float64
		exp fixed.Number
	}{
		{math.NaN(), fixed.NaN},
		{math.Inf(+1), fixed.NaN},
		{math.Inf(-1), fixed.NaN},
		{0.0, n("0")},
		{math.Copysign(0, -1), n("0")},
		{1.0, n("1")},
		{-1.0, n("-1")},
		{0.5, n("0.5")},
		{0.1, n("0.1000000000000000056")},
		{-0.1, n("-0.1000000000000000056")},
		{3.14159264, n("3.1415926399999998253")},
		{12345678.12345678, n("12345678.1234567798674106598")},
		{1.7976931348623157, n("1.7976931348623157447")},
		{math.MaxInt64, n("9223372036854775808")},
		{1 << 62, n("4611686018427387904")},
		{-(1 << 62), n("-4611686018427387904")},
		{9e18, n("9000000000000000000")},
		{1e-19, n("0.0000000000000000001")},
		{1e-20, n("0")},
		{1e-30, n("0")},
		{0x1p-1022, n("0")},
		{math.SmallestNonzeroFloat64, n("0")},
		{1e19, fixed.Overflow},
		// Between 2^127 and 2^128, so the top bit would read as a sign.
		{3e19, fixed.Overflow},
		{-3e19, fixed.Overflow},
		// Exact ties at the 19th place, rounded to even, with the discarded bits
		// ending at the 64-bit boundary and past it.
		{0x101p-20, n("0.0002450942993164062")},
		{0x103p-20, n("0.0002470016479492188")},
		{0x1p-20, n("0.0000009536743164062")},
		{0x1.0000000000001p-20, n("0.0000009536743164063")}, // just above that tie
		{0x3p-20, n("0.0000028610229492188")},
		{1.1e-6, n("0.0000011")},
		{-1e19, fixed.Overflow},
		{math.MaxFloat64, fixed.Overflow},
	} {
		r := fixed.FromFloat64Exact(tc.in)
		if r != tc.exp {
			t.Errorf("%v: expected %s, got %s", tc.in, tc.exp.String(), r.String())
		}
	}
}

func TestFloat64(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	for _, tc := range []struct {
		in  fixed.Number
		exp float64
	}{
		{fixed.NaN, math.NaN()},
		{fixed.Overflow, math.NaN()},
		{fixed.DivisionByZero, math.NaN()},
		{n("0"), 0.0},
		{n("1"), 1.0},
		{n("-1"), -1.0},
		{n("3.14159264"), 3.14159264},
		{n("-3.14159264"), -3.14159264},
		{n("12345678.12345678"), 12345678.12345678},
		{n("-12345678.12345678"), -12345678.12345678},
		{n("0.0000000000000000001"), 0.0000000000000000001},
		{n("-0.0000000000000000001"), -0.0000000000000000001},
		{n("1000000000000000000"), 1000000000000000000},
		{n("-1000000000000000000"), -1000000000000000000},
		{n("9223372036854776000"), math.MaxInt64},
		{n("1.7976931348623157"), 1.7976931348623157},
		// Ties at the 53rd bit round to even, unless digits follow.
		{n("9007199254740993"), 9007199254740992},
		{n("9007199254740995"), 9007199254740996},
		{n("9007199254740993.0000000000000000001"), 9007199254740994},
		{n("0.5"), 0.5},
		{n("0.1"), 0.1},
		{fixed.Highest, 1e19},
		{fixed.Lowest, -1e19},
	} {
		r := tc.in.Float64()
		if math.IsNaN(r) != math.IsNaN(tc.exp) {
			t.Errorf("expected %f, got %f", tc.exp, r)
		} else if !math.IsNaN(r) && !math.IsNaN(tc.exp) {
			if r != tc.exp {
				t.Errorf("expected %f, got %f", tc.exp, r)
			}
		}
	}
}

func TestFromString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		s    string
		exp  string
		err  string
	}{
		{"simple 01", "0", "0", ""},
		{"simple 02", "0.5", "0.5", ""},
		{"simple 03", "-0.5", "-0.5", ""},
		{"simple 04", "1.5", "1.5", ""},
		{"simple 05", "0.25", "0.25", ""},
		{"simple 06", "0.0625", "0.0625", ""},
		{"simple 07", "1", "1", ""},
		{"simple 08", "-1", "-1", ""},
		{"simple 09", "42", "42", ""},
		{"simple 10", "-42", "-42", ""},
		{"simple 11", "1234567", "1234567", ""},
		{"simple 12", "-1234567", "-1234567", ""},

		{"max int", "9223372036854775807", "9223372036854775807", ""},
		{"min int", "-9223372036854775808", "-9223372036854775808", ""},

		{"max fixed", "9999999999999999999.9999999999999999999", "9999999999999999999.9999999999999999999", ""},
		{"min fixed", "-9999999999999999999.9999999999999999999", "-9999999999999999999.9999999999999999999", ""},
		{"max whole", "9999999999999999999", "9999999999999999999", ""},
		{"min whole", "-9999999999999999999", "-9999999999999999999", ""},
		{"max fract", "0.9999999999999999999", "0.9999999999999999999", ""},
		{"min fract", "-0.9999999999999999999", "-0.9999999999999999999", ""},
		{"max precision", "0.0000000000000000001", "0.0000000000000000001", ""},
		{"max precision negative", "-0.0000000000000000001", "-0.0000000000000000001", ""},
		{"trim fract zeroes", "0.001000000000000000000000000000000000", "0.001", ""},
		{"trim fract zeroes 2", "0.5000000000000000000000", "0.5", ""},
		{"trim fract zeroes 3", "1000.000", "1000", ""},

		{"overflow 01", "10000000000000000000", "", "overflow"},
		{"overflow 02", "-10000000000000000000", "", "overflow"},
		{"overflow 03", "10000000000000000000.9999999999999999999", "", "overflow"},
		{"inexact 01", "0.11111111111111111111", "", "inexact"},
		{"inexact 02", "-1.00000000000000000001", "", "inexact"},
		{"overflow over inexact", "10000000000000000000.00000000000000000001", "", "overflow"},
		{"overflow 05", "100000000000000000000000000000000000000", "", "overflow"},
		{"overflow 06", "170141183460469231731687303715884105727", "", "overflow"},
		{"overflow 07", "-170141183460469231731687303715884105728", "", "overflow"},
		{"overflow 08", "24028236692093846346", "", "overflow"},
		{"overflow 09", "34028236692093846346", "", "overflow"},
		{"overflow 10", "-25000000000000000000.25", "", "overflow"},
		{"overflow 11", "-30288027783578028822.34358421438", "", "overflow"},

		{"leading zeroes 01", "0000000000000000000000000000000000000001", "1", ""},
		{"leading zeroes 02", "-0000000000000000000000000000000000000001.5", "-1.5", ""},
		{"leading zeroes 03", "0000000000000000000009999999999999999999.9999999999999999999", "9999999999999999999.9999999999999999999", ""},
		{"leading zeroes 04", "00000000000000000000000000000000000000000", "0", ""},
		{"leading zeroes 05", "0000000000000000000010000000000000000000", "", "overflow"},
		{"fraction only 01", ".0", "0", ""},
		{"fraction only 02", "-.000", "0", ""},
		{"fraction only 03", "+.5", "0.5", ""},
		{"fraction only 04", "00.00", "0", ""},

		{"invalid 01", "", "", "NaN"},
		{"invalid 02", "-", "", "NaN"},
		{"invalid 03", "+", "", "NaN"},
		{"invalid 04", ".", "", "NaN"},
		{"invalid 05", "1.2.3", "", "NaN"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := func(t *testing.T, n fixed.Number, err error) {
				if tc.err == "" && err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if tc.err != "" {
					if err == nil {
						t.Errorf("expected %s, got no error (%s)", tc.err, n.String())
						return
					}
					if err.Error() != tc.err {
						t.Errorf("expected %s error, got %s", tc.err, err.Error())
					}
					return
				}
				res := n.String()
				if res != tc.exp {
					t.Errorf("expected %q, got %q", tc.s, res)
				}
			}
			t.Run("string", func(t *testing.T) {
				n, err := fixed.FromString(tc.s)
				check(t, n, err)
			})
			t.Run("bytes", func(t *testing.T) {
				n, err := fixed.FromBytes([]byte(tc.s))
				check(t, n, err)
			})
		})
	}
}

func TestMustFromString(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		_ = fixed.MustFromString("123.45")
	})
	t.Run("panic", func(t *testing.T) {
		defer func() {
			err := recover()
			if err == nil {
				t.Error("expected panic")
			}
		}()
		_ = fixed.MustFromString("haha")
	})
}

func TestString(t *testing.T) {
	t.Parallel()

	cases := []struct {
		n fixed.Number
		e string
	}{
		{fixed.FromInt(0), "0"},
		{fixed.Half, "0.5"},
		{fixed.Half.Neg(), "-0.5"},
		{fixed.Half.Mul(fixed.FromInt(3)), "1.5"},
		{fixed.Half.DivRound(fixed.Two, fixed.RoundHalfInf), "0.25"},
		{fixed.Half.DivRound(fixed.FromInt(8), fixed.RoundHalfInf), "0.0625"},
		{fixed.FromInt(1), "1"},
		{fixed.FromInt(-1), "-1"},
		{fixed.FromInt(42), "42"},
		{fixed.FromInt(-42), "-42"},
		{fixed.FromInt(100), "100"},
		{fixed.FromInt(1000), "1000"},
		{fixed.FromInt(1000).Add(fixed.Half), "1000.5"},
		{fixed.MustFromString("-10000.5"), "-10000.5"},
		{fixed.FromInt(1234567), "1234567"},
		{fixed.FromInt(-1234567), "-1234567"},
		{fixed.FromInt(math.MaxInt64), "9223372036854775807"},
		{fixed.FromInt(math.MinInt64), "-9223372036854775808"},
		{fixed.MustFromString("2398475923874510023.120381209458129412"), "2398475923874510023.120381209458129412"},
		{fixed.MustFromString("-2398475923874510023.120381209458129412"), "-2398475923874510023.120381209458129412"},
		{fixed.Highest, "9999999999999999999.9999999999999999999"},
		{fixed.Lowest, "-9999999999999999999.9999999999999999999"},

		// Coefficients beyond the representable range are invalid, even when built directly.
		{fixed.Number{Lo: 1<<64 - 1, Hi: 1<<63 - 1}, "NaN"},
		{fixed.Number{Lo: 0, Hi: 1 << 63}, "NaN"},
		{fixed.Number{Lo: 0, Hi: 1}, "1.8446744073709551616"},
	}
	for _, tc := range cases {
		r := tc.n.String()
		if r != tc.e {
			t.Errorf("expected %q, got %q", tc.e, r)
		}
	}
}

func TestAppend(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		n fixed.Number
		e string
	}{
		{fixed.Zero, "0"},
		{fixed.MustFromString("1234.56"), "1234.56"},
		{fixed.MustFromString("-1234.56"), "-1234.56"},
		{fixed.MustFromString("0.0000000000000000001"), "0.0000000000000000001"},
		{fixed.NaN, "NaN"},
		{fixed.Overflow, "overflow"},
	} {
		r := string(tc.n.Append([]byte("v=")))
		if r != "v="+tc.e {
			t.Errorf("expected %q, got %q", "v="+tc.e, r)
		}
		if s := tc.n.String(); s != tc.e {
			t.Errorf("String disagrees with Append: %q != %q", s, tc.e)
		}
	}
}
