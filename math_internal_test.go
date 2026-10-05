package fixed

import (
	"fmt"
	"math/bits"
	"testing"
)

func TestLeadingZeroes128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v Number
		e uint
	}{
		{u128(0x0, 0x0), 128},
		{u128(0x0, 0x1), 127},
		{u128(0x0, 0xffffffffffffffff), 64},
		{u128(0x1, 0x0), 63},
		{u128(0x4000000000000000, 0x0), 1},
		{u128(0x8000000000000000, 0x0), 0},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0},
	}
	for _, tc := range cases {
		if r := leadingZeroes128(tc.v); r != tc.e {
			t.Errorf("leadingZeroes128(%s): expected %d, got %d", hex(tc.v), tc.e, r)
		}
	}
}

func TestShift128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v           Number
		n           uint
		left, right Number
	}{
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0, u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 1, u128(0x2468acf13579bdf, 0xfdb97530eca86420), u128(0x91a2b3c4d5e6f7, 0xff6e5d4c3b2a1908)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 4, u128(0x123456789abcdeff, 0xedcba98765432100), u128(0x123456789abcde, 0xffedcba987654321)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 63, u128(0xff6e5d4c3b2a1908, 0x0), u128(0x0, 0x2468acf13579bdf)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 64, u128(0xfedcba9876543210, 0x0), u128(0x0, 0x123456789abcdef)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 65, u128(0xfdb97530eca86420, 0x0), u128(0x0, 0x91a2b3c4d5e6f7)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 100, u128(0x6543210000000000, 0x0), u128(0x0, 0x123456)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 127, u128(0x0, 0x0), u128(0x0, 0x0)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 1, u128(0xffffffffffffffff, 0xfffffffffffffffe), u128(0x7fffffffffffffff, 0xffffffffffffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 4, u128(0xffffffffffffffff, 0xfffffffffffffff0), u128(0xfffffffffffffff, 0xffffffffffffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 63, u128(0xffffffffffffffff, 0x8000000000000000), u128(0x1, 0xffffffffffffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 64, u128(0xffffffffffffffff, 0x0), u128(0x0, 0xffffffffffffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 65, u128(0xfffffffffffffffe, 0x0), u128(0x0, 0x7fffffffffffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 100, u128(0xfffffff000000000, 0x0), u128(0x0, 0xfffffff)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 127, u128(0x8000000000000000, 0x0), u128(0x0, 0x1)},
		{u128(0x0, 0x1), 0, u128(0x0, 0x1), u128(0x0, 0x1)},
		{u128(0x0, 0x1), 1, u128(0x0, 0x2), u128(0x0, 0x0)},
		{u128(0x0, 0x1), 4, u128(0x0, 0x10), u128(0x0, 0x0)},
		{u128(0x0, 0x1), 63, u128(0x0, 0x8000000000000000), u128(0x0, 0x0)},
		{u128(0x0, 0x1), 64, u128(0x1, 0x0), u128(0x0, 0x0)},
		{u128(0x0, 0x1), 65, u128(0x2, 0x0), u128(0x0, 0x0)},
		{u128(0x0, 0x1), 100, u128(0x1000000000, 0x0), u128(0x0, 0x0)},
		{u128(0x0, 0x1), 127, u128(0x8000000000000000, 0x0), u128(0x0, 0x0)},
	}
	for _, tc := range cases {
		if r := shiftLeft128(tc.v, tc.n); r != tc.left {
			t.Errorf("shiftLeft128(%s, %d): expected %s, got %s", hex(tc.v), tc.n, hex(tc.left), hex(r))
		}
		if r := shiftRight128(tc.v, tc.n); r != tc.right {
			t.Errorf("shiftRight128(%s, %d): expected %s, got %s", hex(tc.v), tc.n, hex(tc.right), hex(r))
		}
	}
}

func TestShiftRight192(t *testing.T) {
	t.Parallel()

	cases := []struct {
		hi uint64
		lo Number
		n  uint
		e  uint64
	}{
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 0, 0xfedcba9876543210},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 1, 0xff6e5d4c3b2a1908},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 4, 0xffedcba987654321},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 63, 0x2468acf13579bdf},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 64, 0x123456789abcdef},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 65, 0x91a2b3c4d5e6f7},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 100, 0x4556677880123456},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 127, 0x22446688aaccef10},
		{0x1122334455667788, u128(0x123456789abcdef, 0xfedcba9876543210), 128, 0x1122334455667788},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 0, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 1, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 4, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 63, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 64, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 65, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 100, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 127, 0xffffffffffffffff},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 128, 0xffffffffffffffff},
		{0x0, u128(0x0, 0x1), 0, 0x1},
		{0x0, u128(0x0, 0x1), 1, 0x0},
		{0x0, u128(0x0, 0x1), 4, 0x0},
		{0x0, u128(0x0, 0x1), 63, 0x0},
		{0x0, u128(0x0, 0x1), 64, 0x0},
		{0x0, u128(0x0, 0x1), 65, 0x0},
		{0x0, u128(0x0, 0x1), 100, 0x0},
		{0x0, u128(0x0, 0x1), 127, 0x0},
		{0x0, u128(0x0, 0x1), 128, 0x0},
	}
	for _, tc := range cases {
		if r := shiftRight192(tc.hi, tc.lo, tc.n); r != tc.e {
			t.Errorf("shiftRight192(%#x, %s, %d): expected %#x, got %#x", tc.hi, hex(tc.lo), tc.n, tc.e, r)
		}
	}
}

func TestAdd64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v   Number
		a   uint64
		e   Number
		ovf bool
	}{
		{u128(0x0, 0x0), 0x0, u128(0x0, 0x0), false},
		{u128(0x0, 0x1), 0x2, u128(0x0, 0x3), false},
		{u128(0x0, 0xffffffffffffffff), 0x1, u128(0x1, 0x0), false},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0xffffffffffffffff, u128(0x123456789abcdf0, 0xfedcba987654320f), false},
		{u128(0xffffffffffffffff, 0xfffffffffffffffa), 0x5, u128(0xffffffffffffffff, 0xffffffffffffffff), false},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0x1, u128(0x0, 0x0), true},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, u128(0x0, 0xfffffffffffffffe), true},
		{u128(0xffffffffffffffff, 0x1), 0xffffffffffffffff, u128(0x0, 0x0), true},
	}
	for _, tc := range cases {
		if r := add64(tc.v, tc.a); r != tc.e {
			t.Errorf("add64(%s, %#x): expected %s, got %s", hex(tc.v), tc.a, hex(tc.e), hex(r))
		}
		r, ovf := add64Checked(tc.v, tc.a)
		if ovf != tc.ovf || (!ovf && r != tc.e) || (ovf && r != Number{}) {
			t.Errorf("add64Checked(%s, %#x): expected %s %v, got %s %v", hex(tc.v), tc.a, hex(tc.e), tc.ovf, hex(r), ovf)
		}
	}
}

func TestAdd128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v, a Number
		e    Number
		ovf  bool
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0), false},
		{u128(0x0, 0x1), u128(0x0, 0x2), u128(0x0, 0x3), false},
		{u128(0x0, 0xffffffffffffffff), u128(0x0, 0x1), u128(0x1, 0x0), false},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x2468acf13579bdf, 0xfdb97530eca86420), false},
		{u128(0x0, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff), u128(0x1, 0xfffffffffffffffe), false},
		{u128(0x8000000000000000, 0x0), u128(0x7fffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), false},
		{u128(0x8000000000000000, 0x0), u128(0x8000000000000000, 0x0), u128(0x0, 0x0), true},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x1), u128(0x0, 0x0), true},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xfffffffffffffffe), true},
		{u128(0xffffffffffffffff, 0x0), u128(0x1, 0x0), u128(0x0, 0x0), true},
	}
	for _, tc := range cases {
		if r := add128(tc.v, tc.a); r != tc.e {
			t.Errorf("add128(%s, %s): expected %s, got %s", hex(tc.v), hex(tc.a), hex(tc.e), hex(r))
		}
		r, ovf := add128Checked(tc.v, tc.a)
		if ovf != tc.ovf || (!ovf && r != tc.e) || (ovf && r != Number{}) {
			t.Errorf("add128Checked(%s, %s): expected %s %v, got %s %v", hex(tc.v), hex(tc.a), hex(tc.e), tc.ovf, hex(r), ovf)
		}
	}
}

func TestSub64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v Number
		a uint64
		e Number
	}{
		{u128(0x0, 0x0), 0x0, u128(0x0, 0x0)},
		{u128(0x0, 0x3), 0x2, u128(0x0, 0x1)},
		{u128(0x1, 0x0), 0x1, u128(0x0, 0xffffffffffffffff)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0xffffffffffffffff, u128(0x123456789abcdee, 0xfedcba9876543211)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, u128(0xffffffffffffffff, 0x0)},
		{u128(0x0, 0x0), 0x1, u128(0xffffffffffffffff, 0xffffffffffffffff)},
		{u128(0x0, 0x5), 0xffffffffffffffff, u128(0xffffffffffffffff, 0x6)},
	}
	for _, tc := range cases {
		if r := sub64(tc.v, tc.a); r != tc.e {
			t.Errorf("sub64(%s, %#x): expected %s, got %s", hex(tc.v), tc.a, hex(tc.e), hex(r))
		}
	}
}

func TestSub128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b Number
		e    Number
		ovf  bool
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0), false},
		{u128(0x0, 0x3), u128(0x0, 0x2), u128(0x0, 0x1), false},
		{u128(0x1, 0x0), u128(0x0, 0x1), u128(0x0, 0xffffffffffffffff), false},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0x0), false},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0xffffffffffffffff), u128(0x123456789abcdee, 0xfedcba9876543211), false},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0), false},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x1), u128(0xffffffffffffffff, 0xfffffffffffffffe), false},
		{u128(0x0, 0x0), u128(0x0, 0x1), u128(0xffffffffffffffff, 0xffffffffffffffff), true},
		{u128(0x1, 0x0), u128(0x1, 0x1), u128(0xffffffffffffffff, 0xffffffffffffffff), true},
		{u128(0x0, 0x5), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x6), true},
		{u128(0x8000000000000000, 0x0), u128(0x8000000000000000, 0x1), u128(0xffffffffffffffff, 0xffffffffffffffff), true},
	}
	for _, tc := range cases {
		if r := sub128(tc.a, tc.b); r != tc.e {
			t.Errorf("sub128(%s, %s): expected %s, got %s", hex(tc.a), hex(tc.b), hex(tc.e), hex(r))
		}
		r, ovf := sub128Checked(tc.a, tc.b)
		if ovf != tc.ovf || (!ovf && r != tc.e) || (ovf && r != Number{}) {
			t.Errorf("sub128Checked(%s, %s): expected %s %v, got %s %v", hex(tc.a), hex(tc.b), hex(tc.e), tc.ovf, hex(r), ovf)
		}
	}
}

func TestNeg256(t *testing.T) {
	t.Parallel()

	cases := []struct {
		hi, lo   Number
		eHi, eLo Number
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0)},
		{u128(0x0, 0x0), u128(0x0, 0x1), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff)},
		{u128(0x0, 0x0), u128(0x0, 0x2), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xfffffffffffffffe)},
		{u128(0x0, 0x0), u128(0x0, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0x1)},
		{u128(0x0, 0x0), u128(0x1, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0x0)},
		{u128(0x0, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x1)},
		{u128(0x0, 0x1), u128(0x0, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0xfedcba9876543210, 0x123456789abcdef), u128(0xfedcba9876543210, 0x123456789abcdf0)},
		{u128(0x8000000000000000, 0x0), u128(0x0, 0x0), u128(0x8000000000000000, 0x0), u128(0x0, 0x0)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0), u128(0x0, 0x1)},
	}
	for _, tc := range cases {
		hi, lo := neg256(tc.hi, tc.lo)
		if hi != tc.eHi || lo != tc.eLo {
			t.Errorf("neg256(%s, %s): expected %s %s, got %s %s", hex(tc.hi), hex(tc.lo), hex(tc.eHi), hex(tc.eLo), hex(hi), hex(lo))
		}
	}
}

func TestMul64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v   Number
		m   uint64
		e   Number
		ovf bool
	}{
		{u128(0x0, 0x0), 0xffffffffffffffff, u128(0x0, 0x0), false},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x0, u128(0x0, 0x0), false},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x1, u128(0x123456789abcdef, 0xfedcba9876543210), false},
		{u128(0x0, 0x3), 0x7, u128(0x0, 0x15), false},
		{u128(0x0, 0xffffffffffffffff), 0xffffffffffffffff, u128(0xfffffffffffffffe, 0x1), false},
		{u128(0x1, 0x0), 0xffffffffffffffff, u128(0xffffffffffffffff, 0x0), false},
		{u128(0x1, 0xffffffffffffffff), 0xffffffffffffffff, u128(0xfffffffffffffffd, 0x1), true},
		{u128(0x0, 0x8000000000000000), 0x2, u128(0x1, 0x0), false},
		{u128(0x8000000000000000, 0x0), 0x2, u128(0x0, 0x0), true},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x8ac7230489e80000, u128(0x674ad6ab762638e3, 0x98174357ee800000), true},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, u128(0xffffffffffffffff, 0x1), true},
	}
	for _, tc := range cases {
		if r := mul64(tc.v, tc.m); r != tc.e {
			t.Errorf("mul64(%s, %#x): expected %s, got %s", hex(tc.v), tc.m, hex(tc.e), hex(r))
		}
		r, ovf := mul64Checked(tc.v, tc.m)
		if ovf != tc.ovf || (!ovf && r != tc.e) || (ovf && r != Number{}) {
			t.Errorf("mul64Checked(%s, %#x): expected %s %v, got %s %v", hex(tc.v), tc.m, hex(tc.e), tc.ovf, hex(r), ovf)
		}
	}
}

func TestMul128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b     Number
		eLo, eHi Number
	}{
		{u128(0x0, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0), u128(0x0, 0x0)},
		{u128(0x0, 0x1), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0x0)},
		{u128(0x0, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff), u128(0xfffffffffffffffe, 0x1), u128(0x0, 0x0)},
		{u128(0x1, 0x0), u128(0x1, 0x0), u128(0x0, 0x0), u128(0x0, 0x1)},
		{u128(0x3, 0xffffffffffffffff), u128(0x3, 0xffffffffffffffff), u128(0xfffffffffffffff8, 0x1), u128(0x0, 0xf)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x1), u128(0xffffffffffffffff, 0xfffffffffffffffe)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x1), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x422871b7939f74ac, 0xdeec6cd7a44a4100), u128(0x14b66dc33f6ac, 0xdee9d609ebe253a6)},
		{u128(0x8000000000000000, 0x0), u128(0x0, 0x2), u128(0x0, 0x0), u128(0x0, 0x1)},
	}
	for _, tc := range cases {
		lo, hi := mul128(tc.a, tc.b)
		if lo != tc.eLo || hi != tc.eHi {
			t.Errorf("mul128(%s, %s): expected %s %s, got %s %s", hex(tc.a), hex(tc.b), hex(tc.eHi), hex(tc.eLo), hex(hi), hex(lo))
		}
	}
}

func TestMul128By64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a   Number
		b   uint64
		eHi uint64
		eLo Number
	}{
		{u128(0x0, 0x0), 0xffffffffffffffff, 0x0, u128(0x0, 0x0)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x1, 0x0, u128(0x123456789abcdef, 0xfedcba9876543210)},
		{u128(0x0, 0xffffffffffffffff), 0xffffffffffffffff, 0x0, u128(0xfffffffffffffffe, 0x1)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, 0xfffffffffffffffe, u128(0xffffffffffffffff, 0x1)},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0x2, 0x1, u128(0xffffffffffffffff, 0xfffffffffffffffe)},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x8ac7230489e80000, 0x9de5fc9b59c71c, u128(0x674ad6ab762638e3, 0x98174357ee800000)},
		{u128(0x8000000000000000, 0x0), 0x8000000000000000, 0x4000000000000000, u128(0x0, 0x0)},
	}
	for _, tc := range cases {
		hi, lo := mul128By64(tc.a, tc.b)
		if hi != tc.eHi || lo != tc.eLo {
			t.Errorf("mul128By64(%s, %#x): expected %#x %s, got %#x %s", hex(tc.a), tc.b, tc.eHi, hex(tc.eLo), hi, hex(lo))
		}
	}
}

func TestQuoRem64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v Number
		d uint64
		q Number
		r uint64
	}{
		{u128(0x0, 0x0), 0x7, u128(0x0, 0x0), 0x0},
		{u128(0x0, 0x64), 0x7, u128(0x0, 0xe), 0x2},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x1, u128(0x123456789abcdef, 0xfedcba9876543210), 0x0},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0x7, u128(0x299c335ccf668f, 0xffd663cca3309970), 0x0},
		{u128(0x123456789abcdef, 0xfedcba9876543210), 0xffffffffffffffff, u128(0x0, 0x123456789abcdf0), 0x0},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0x3, u128(0x5555555555555555, 0x5555555555555555), 0x0},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, u128(0x1, 0x1), 0x0},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), 0x8ac7230489e80000, u128(0x1, 0xd83c94fb6d2ac34a), 0x2ed503946aefffff},
		{u128(0x4b3b4ca85a86c47a, 0x98a223fffffffff), 0x8ac7230489e80000, u128(0x0, 0x8ac7230489e7ffff), 0x8ac7230489e7ffff},
		{u128(0x0, 0xffffffffffffffff), 0x8000000000000000, u128(0x0, 0x1), 0x7fffffffffffffff},
	}
	for _, tc := range cases {
		q, r := quoRem64(tc.v, tc.d)
		if q != tc.q || r != tc.r {
			t.Errorf("quoRem64(%s, %#x): expected %s %#x, got %s %#x", hex(tc.v), tc.d, hex(tc.q), tc.r, hex(q), r)
		}
	}
}

func TestQuoRem192By64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		hi, mid, lo, d uint64
		q0, q1, r      uint64
	}{
		{0x0, 0x0, 0x0, 0x7, 0x0, 0x0, 0x0},
		{0x0, 0x0, 0x64, 0x7, 0xe, 0x0, 0x2},
		{0x0, 0xffffffffffffffff, 0xffffffffffffffff, 0x3, 0x5555555555555555, 0x5555555555555555, 0x0},
		{0x6, 0xffffffffffffffff, 0xffffffffffffffff, 0x7, 0xffffffffffffffff, 0xffffffffffffffff, 0x6},
		{0xfffffffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xfffffffffffffffe},
		{0x8ac7230489e7ffff, 0xffffffffffffffff, 0xffffffffffffffff, 0x8ac7230489e80000, 0xffffffffffffffff, 0xffffffffffffffff, 0x8ac7230489e7ffff},
		{0x1, 0x2, 0x3, 0x8000000000000000, 0x4, 0x2, 0x3},
	}
	for _, tc := range cases {
		q0, q1, r := quoRem192By64(tc.hi, tc.mid, tc.lo, tc.d)
		if q0 != tc.q0 || q1 != tc.q1 || r != tc.r {
			t.Errorf("quoRem192By64(%#x, %#x, %#x, %#x): expected %#x %#x %#x, got %#x %#x %#x", tc.hi, tc.mid, tc.lo, tc.d, tc.q1, tc.q0, tc.r, q1, q0, r)
		}
	}
}

func TestQuoRem256By64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		hi, lo Number
		d      uint64
		q      Number
		r      uint64
	}{
		{u128(0x0, 0x0), u128(0x123456789abcdef, 0xfedcba9876543210), 0x7, u128(0x299c335ccf668f, 0xffd663cca3309970), 0x0},
		{u128(0x0, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, u128(0x1, 0x1), 0x0},
		{u128(0x0, 0x5), u128(0x123456789abcdef, 0xfedcba9876543210), 0x7, u128(0xb70509ea383d1d6b, 0x6d8d3f3a5a0c0726), 0x6},
		{u128(0x0, 0x6), u128(0xffffffffffffffff, 0xffffffffffffffff), 0x7, u128(0xffffffffffffffff, 0xffffffffffffffff), 0x6},
		{u128(0x0, 0x7), u128(0x0, 0x0), 0x7, Overflow, 0x0},
		{u128(0x0, 0x8), u128(0x0, 0x0), 0x7, Overflow, 0x0},
		{u128(0x0, 0xfffffffffffffffe), u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), 0xfffffffffffffffe},
		{u128(0x0, 0xffffffffffffffff), u128(0x0, 0x0), 0xffffffffffffffff, Overflow, 0x0},
		{u128(0x1, 0x0), u128(0x0, 0x0), 0xffffffffffffffff, Overflow, 0x0},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), 0xffffffffffffffff, Overflow, 0x0},
	}
	for _, tc := range cases {
		q, r := quoRem256By64(tc.hi, tc.lo, tc.d)
		if q != tc.q || r != tc.r {
			t.Errorf("quoRem256By64(%s, %s, %#x): expected %s %#x, got %s %#x", hex(tc.hi), hex(tc.lo), tc.d, hex(tc.q), tc.r, hex(q), r)
		}
	}
}

func TestQuoRem192By128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		hi   uint64
		lo   Number
		d    Number
		q, r Number
	}{
		{0x0, u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0x7), u128(0x299c335ccf668f, 0xffd663cca3309970), u128(0x0, 0x0)},
		{0x6, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x7), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x6)},
		{0x7, u128(0x0, 0x0), u128(0x0, 0x7), Overflow, u128(0x0, 0x0)},
		{0xfffffffffffffffe, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0xfffffffffffffffe)},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff), Overflow, u128(0x0, 0x0)},
		{0x0, u128(0x0, 0x5), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0), u128(0x0, 0x5)},
		{0x0, u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0x1), u128(0x0, 0x0)},
		// An exact quotient whose trial digit overshoots by one, so the overshoot equals the divisor.
		{0x12b92, u128(0x6012ed4f2a8f4da1, 0x2cc4087758207ba0), u128(0x4ee7, 0xc4f5a58656eef510), u128(0x3, 0xcbedc0dcb63ac19a), u128(0x0, 0x0)},
		{0x0, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x1, 0x0), u128(0x0, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff)},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x1, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff)},
		{0x414c343c1027c4d1, u128(0xec78b56f91b7584a, 0x0), u128(0xec78b56f91b7584a, 0xc386bbc4cd613e30), u128(0x0, 0x46b0acd1b72090e7), u128(0xeb83606269e18e41, 0xbd37a917e445e2b0)},
		{0x1767cf0a39cc4b1, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x1767cf0a39cc4b2, 0xafbf5310ee1b8cc4), u128(0x0, 0xffffffffffffff87), u128(0x141baac66fd6bb1, 0x116e43008b0588a3)},
		{0xf37e62cdc962efe0, u128(0xf37e62cdc962efe0, 0x0), u128(0xf37e62cdc962efe0, 0xfe3aeb11da80a3bb), u128(0x0, 0xffffffffffffffff), u128(0xe8c1da89b8453c05, 0xfe3aeb11da80a3bb)},
		{0xf4ecaf4a130865e4, u128(0x0, 0x0), u128(0xf4ecaf4a130865e4, 0x24554c27944be91a), u128(0x0, 0xffffffffffffffff), u128(0xd09763227ebc7cca, 0x24554c27944be91a)},
		{0xc324c9859b810e76, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x80000000000000, 0x0), u128(0x186, 0x49930b37021cedff), u128(0x7fffffffffffff, 0xffffffffffffffff)},
		{0xffffffffffffffff, u128(0x0, 0x0), u128(0x7fffffffffffff, 0xffffffffffffffff), u128(0x1ff, 0xfffffffffffffe00), u128(0x1ff, 0xfffffffffffffe00)},
		{0xffffffffffffffff, u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x1, 0x0), u128(0x0, 0xffffffffffffffff)},
	}
	for _, tc := range cases {
		q, r := quoRem192By128(tc.hi, tc.lo, tc.d)
		if q != tc.q || r != tc.r {
			t.Errorf("quoRem192By128(%#x, %s, %s): expected %s %s, got %s %s", tc.hi, hex(tc.lo), hex(tc.d), hex(tc.q), hex(tc.r), hex(q), hex(r))
		}
	}
}

func TestDiv2by1(t *testing.T) {
	t.Parallel()

	cases := []struct {
		u1, u0, d uint64
		q, r      uint64
	}{
		{0x0, 0x0, 0x8ac7230489e80000, 0x0, 0x0},
		{0x0, 0x8ac7230489e7ffff, 0x8ac7230489e80000, 0x0, 0x8ac7230489e7ffff},
		{0x0, 0xffffffffffffffff, 0x8ac7230489e80000, 0x1, 0x7538dcfb7617ffff},
		{0x8ac7230489e7ffff, 0xffffffffffffffff, 0x8ac7230489e80000, 0xffffffffffffffff, 0x8ac7230489e7ffff},
		{0x4563918244f40000, 0x0, 0x8ac7230489e80000, 0x8000000000000000, 0x0},
		{0x4ad8734b839d80f4, 0x396e468230b18c15, 0x8ac7230489e80000, 0x8a10cadce5289b80, 0x2836f07dc4b18c15},
		{0x890d4ed065a659b5, 0xff3e3d7e87ba38eb, 0x8ac7230489e80000, 0xfcd0f8310050d8c7, 0x213181b946238eb},
		{0x851685d1109dd974, 0xff1d4de186c00000, 0x8ac7230489e80000, 0xf58105748ed5d1b8, 0x0},
		{0x0, 0x0, 0x8000000000000000, 0x0, 0x0},
		{0x0, 0x7fffffffffffffff, 0x8000000000000000, 0x0, 0x7fffffffffffffff},
		{0x0, 0xffffffffffffffff, 0x8000000000000000, 0x1, 0x7fffffffffffffff},
		{0x7fffffffffffffff, 0xffffffffffffffff, 0x8000000000000000, 0xffffffffffffffff, 0x7fffffffffffffff},
		{0x4000000000000000, 0x0, 0x8000000000000000, 0x8000000000000000, 0x0},
		{0x4ad8734b839d80f4, 0x396e468230b18c15, 0x8000000000000000, 0x95b0e697073b01e8, 0x396e468230b18c15},
		{0x0, 0x0, 0xffffffffffffffff, 0x0, 0x0},
		{0x0, 0xfffffffffffffffe, 0xffffffffffffffff, 0x0, 0xfffffffffffffffe},
		{0x0, 0xffffffffffffffff, 0xffffffffffffffff, 0x1, 0x0},
		{0xfffffffffffffffe, 0xffffffffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xfffffffffffffffe},
		{0x7fffffffffffffff, 0x0, 0xffffffffffffffff, 0x7fffffffffffffff, 0x7fffffffffffffff},
		{0x4ad8734b839d80f4, 0x396e468230b18c15, 0xffffffffffffffff, 0x4ad8734b839d80f4, 0x8446b9cdb44f0d09},
		{0x890d4ed065a659b5, 0xff3e3d7e87ba38eb, 0xffffffffffffffff, 0x890d4ed065a659b6, 0x884b8c4eed6092a1},
		{0x851685d1109dd974, 0xff1d4de186c00000, 0xffffffffffffffff, 0x851685d1109dd975, 0x8433d3b2975dd975},
	}
	reciprocal64 := func(d uint64) uint64 {
		v, _ := bits.Div64(^d, ^uint64(0), d)
		return v
	}
	for _, tc := range cases {
		q, r := divMollerGranlund(tc.u1, tc.u0, tc.d, reciprocal64(tc.d))
		if q != tc.q || r != tc.r {
			t.Errorf("div2by1(%#x, %#x, %#x): expected %#x %#x, got %#x %#x", tc.u1, tc.u0, tc.d, tc.q, tc.r, q, r)
		}
	}
}

func TestReciprocal64(t *testing.T) {
	t.Parallel()

	ds := []uint64{
		0x8000000000000000,
		0x8000000000000001,
		0xffffffffffffffff,
		0xfffffffffffffffe,
		0xc000000000000000,
		0x8ac7230489e80000,
		0x4ad8734b839d80f4 | 1<<63,
		0x890d4ed065a659b5,
	}
	// Both ends of every table seed.
	for i := range uint64(256) {
		ds = append(ds, (256+i)<<55, (256+i)<<55|(1<<55-1))
	}
	for _, p := range pow10Reciprocals {
		if v := reciprocal64(p.d); v != p.v {
			t.Errorf("reciprocal64(%#x): expected %#x, got %#x", p.d, p.v, v)
		}
	}
	for _, d := range ds {
		want, _ := bits.Div64(^d, ^uint64(0), d)
		if v := reciprocal64(d); v != want {
			t.Errorf("reciprocal64(%#x): expected %#x, got %#x", d, want, v)
		}
	}
}

func TestQuoRem128ByScale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		v Number
		q Number
		r uint64
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), 0x0},
		{u128(0x0, 0x1), u128(0x0, 0x0), 0x1},
		{u128(0x0, 0x8ac7230489e7ffff), u128(0x0, 0x0), 0x8ac7230489e7ffff},
		{u128(0x0, 0x8ac7230489e80000), u128(0x0, 0x1), 0x0},
		{u128(0x0, 0x8ac7230489e80001), u128(0x0, 0x1), 0x1},
		{u128(0x0, 0xffffffffffffffff), u128(0x0, 0x1), 0x7538dcfb7617ffff},
		{u128(0x8ac7230489e7ffff, 0xffffffffffffffff), u128(0x0, 0xffffffffffffffff), 0x8ac7230489e7ffff},
		{u128(0x8ac7230489e80000, 0x0), u128(0x1, 0x0), 0x0},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0x2194ce4ac4a2546), 0x65c1c25c38e43210},
		{u128(0x4b3b4ca85a86c47a, 0x98a223fffffffff), u128(0x0, 0x8ac7230489e7ffff), 0x8ac7230489e7ffff},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x1, 0xd83c94fb6d2ac34a), 0x2ed503946aefffff},
		{u128(0x851685d1109dd974, 0xff1d4de186c00000), u128(0x0, 0xf58105748ed5d1b8), 0x0},
	}
	for _, tc := range cases {
		q, r := quoRem128ByScale(tc.v)
		if q != tc.q || r != tc.r {
			t.Errorf("quoRem128ByScale(%s): expected %s %#x, got %s %#x", hex(tc.v), hex(tc.q), tc.r, hex(q), r)
		}
	}
}

func TestQuoRem128ByPow10(t *testing.T) {
	t.Parallel()

	values := []Number{
		u128(0x0, 0x0),
		u128(0x0, 0x1),
		u128(0x0, 0x8ac7230489e7ffff),
		u128(0x0, 0x8ac7230489e80000),
		u128(0x0, 0xffffffffffffffff),
		u128(0x1, 0x0),
		u128(0x123456789abcdef, 0xfedcba9876543210),
		u128(0x4b3b4ca85a86c47a, 0x98a223fffffffff),
		u128(0x8ac7230489e7ffff, 0xffffffffffffffff),
		u128(0xffffffffffffffff, 0xffffffffffffffff),
	}
	for k := range len(pow10u64) {
		for _, v := range values {
			for _, v := range []Number{v, sub64(v, 1), add64(v, 1)} {
				wq, wr := quoRem64(v, pow10u64[k])
				if q, r := quoRem128ByPow10(v, k); q != wq || r != wr {
					t.Errorf("quoRem128ByPow10(%s, %d): expected %s %#x, got %s %#x", hex(v), k, hex(wq), wr, hex(q), r)
				}
			}
		}
	}
}

func TestQuoRem256ByScale(t *testing.T) {
	t.Parallel()

	cases := []struct {
		hi, lo Number
		q      Number
		r      uint64
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0), 0x0},
		{u128(0x0, 0x0), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x0, 0x2194ce4ac4a2546), 0x65c1c25c38e43210},
		{u128(0x0, 0x0), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x1, 0xd83c94fb6d2ac34a), 0x2ed503946aefffff},
		{u128(0x0, 0x1), u128(0x0, 0x0), u128(0x1, 0xd83c94fb6d2ac34a), 0x2ed503946af00000},
		{u128(0x161bcca7119915b5, 0x764b4abe8652978), u128(0xe0ff0ca0bc87870b, 0xecebbb8000000001), Overflow, 0x0},
		{u128(0x0, 0x8ac7230489e7ffff), u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), 0x8ac7230489e7ffff},
		{u128(0x0, 0x8ac7230489e80000), u128(0x0, 0x0), Overflow, 0x0},
		{u128(0x0, 0xffffffffffffffff), u128(0x0, 0x0), Overflow, 0x0},
		{u128(0x1, 0x0), u128(0x0, 0x0), Overflow, 0x0},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0xffffffffffffffff, 0xffffffffffffffff), Overflow, 0x0},
	}
	for _, tc := range cases {
		q, r := quoRem256ByScale(tc.hi, tc.lo)
		if q != tc.q || r != tc.r {
			t.Errorf("quoRem256ByScale(%s, %s): expected %s %#x, got %s %#x", hex(tc.hi), hex(tc.lo), hex(tc.q), tc.r, hex(q), r)
		}
	}
}

func TestCmpUnsigned128(t *testing.T) {
	t.Parallel()

	cases := []struct {
		a, b Number
		e    int
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), 0},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), 0},
		{u128(0x0, 0x1), u128(0x0, 0x2), -1},
		{u128(0x0, 0x2), u128(0x0, 0x1), 1},
		{u128(0x1, 0x0), u128(0x0, 0xffffffffffffffff), 1},
		{u128(0x0, 0xffffffffffffffff), u128(0x1, 0x0), -1},
		{u128(0x5, 0x1), u128(0x5, 0x2), -1},
		{u128(0x5, 0x2), u128(0x5, 0x1), 1},
		{u128(0xffffffffffffffff, 0xffffffffffffffff), u128(0x0, 0x0), 1},
	}
	for _, tc := range cases {
		if r := cmpUnsigned128(tc.a, tc.b); r != tc.e {
			t.Errorf("cmpUnsigned128(%s, %s): expected %d, got %d", hex(tc.a), hex(tc.b), tc.e, r)
		}
	}
}

func TestCmpUnsigned256(t *testing.T) {
	t.Parallel()

	cases := []struct {
		aHi, aLo, bHi, bLo Number
		e                  int
	}{
		{u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0), u128(0x0, 0x0), 0},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), 0},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba987654320f), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), -1},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba987654320f), 1},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdee, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), -1},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdee, 0xfedcba9876543210), 1},
		{u128(0x123456789abcdef, 0xfedcba987654320f), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), -1},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba987654320f), u128(0x123456789abcdef, 0xfedcba9876543210), 1},
		{u128(0x123456789abcdee, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), -1},
		{u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), u128(0x123456789abcdee, 0xfedcba9876543210), u128(0x123456789abcdef, 0xfedcba9876543210), 1},
	}
	for _, tc := range cases {
		if r := cmpUnsigned256(tc.aHi, tc.aLo, tc.bHi, tc.bLo); r != tc.e {
			t.Errorf("cmpUnsigned256(%s %s, %s %s): expected %d, got %d", hex(tc.aHi), hex(tc.aLo), hex(tc.bHi), hex(tc.bLo), tc.e, r)
		}
	}
}

func u128(hi, lo uint64) Number {
	return Number{Lo: lo, Hi: hi}
}

func hex(v Number) string {
	return fmt.Sprintf("%#016x_%016x", v.Hi, v.Lo)
}
