package benchmarks

import (
	"fmt"
	"math/rand/v2"
	"testing"

	cockroachdb_apd "github.com/cockroachdb/apd/v3"
	jokruger_dec128 "github.com/jokruger/dec128"
	quagmt_udecimal "github.com/quagmt/udecimal"
	shopspring_decimal "github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func BenchmarkSqrt(b *testing.B) {
	s := randomPositiveNumbers(5)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		ns := parseAll(s, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].Sqrt()
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		ns := parseAll(s, jokruger_dec128.FromString[string])
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].SqrtRound(fixed.Scale, jokruger_dec128.ROUND_TOWARD_ZERO)
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		ns := parseAll(s, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_, _ = ns[i&(numRandomOperands-1)].Sqrt()
			i++
		}
	})

	// shopspring/decimal has no square root

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(s, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundDown)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Sqrt(&d, ns[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkPowInt(b *testing.B) {
	bases, exps := powOperands(13)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		ns := parseAll(bases, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].PowIntRound(exps[i&(numRandomOperands-1)], fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		ns := parseAll(bases, jokruger_dec128.FromString[string])
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].PowIntRound(int64(exps[i&(numRandomOperands-1)]), fixed.Scale, jokruger_dec128.ROUND_BANK)
			i++
		}
	})

	// udecimal's PowInt32 truncates and has no rounding mode.
	b.Run("quagmt/udecimal", func(b *testing.B) {
		ns := parseAll(bases, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_, _ = ns[i&(numRandomOperands-1)].PowInt32(int32(exps[i&(numRandomOperands-1)]))
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		ns := parseAll(bases, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			p, _ := ns[i&(numRandomOperands-1)].PowInt32(int32(exps[i&(numRandomOperands-1)]))
			_ = p.RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(bases, apdParse)
		ms := make([]*cockroachdb_apd.Decimal, len(exps))
		for j, m := range exps {
			ms[j] = cockroachdb_apd.New(int64(m), 0)
		}
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Pow(&d, ns[i&(numRandomOperands-1)], ms[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkPow(b *testing.B) {
	bases, exps := powRealOperands(16)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		xs, ys := parseAll(bases, fixed.MustFromString), parseAll(exps, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = xs[i&(numRandomOperands-1)].PowRound(ys[i&(numRandomOperands-1)], fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		xs, ys := parseAll(bases, parse), parseAll(exps, parse)
		i := 0
		for b.Loop() {
			_ = xs[i&(numRandomOperands-1)].Pow(ys[i&(numRandomOperands-1)], fixed.Scale, jokruger_dec128.ROUND_BANK)
			i++
		}
	})

	// udecimal only raises to the integer part of an exponent, so it is absent
	// here.

	b.Run("shopspring/decimal", func(b *testing.B) {
		parse := shopspring_decimal.RequireFromString
		xs, ys := parseAll(bases, parse), parseAll(exps, parse)
		i := 0
		for b.Loop() {
			_, _ = xs[i&(numRandomOperands-1)].PowWithPrecision(ys[i&(numRandomOperands-1)], fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		xs, ys := parseAll(bases, apdParse), parseAll(exps, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Pow(&d, xs[i&(numRandomOperands-1)], ys[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkExp(b *testing.B) {
	s := expOperands(14)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		ns := parseAll(s, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].ExpRound(fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		ns := parseAll(s, jokruger_dec128.FromString[string])
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].Exp(fixed.Scale, jokruger_dec128.ROUND_BANK)
			i++
		}
	})

	// udecimal has no exponential, so it is absent here.

	b.Run("shopspring/decimal", func(b *testing.B) {
		ns := parseAll(s, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_, _ = ns[i&(numRandomOperands-1)].ExpTaylor(fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(s, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Exp(&d, ns[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkLn(b *testing.B) {
	s := randomPositiveNumbers(15)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		ns := parseAll(s, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].LnRound(fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		ns := parseAll(s, jokruger_dec128.FromString[string])
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].Ln(fixed.Scale, jokruger_dec128.ROUND_BANK)
			i++
		}
	})

	// udecimal has no logarithm, so it is absent here.

	b.Run("shopspring/decimal", func(b *testing.B) {
		ns := parseAll(s, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_, _ = ns[i&(numRandomOperands-1)].Ln(fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(s, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Ln(&d, ns[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

// powOperands returns bases in [1, 1.1) with all Scale fractional digits and
// exponents in [1, 365], as for compounding over up to a year of days.
func powOperands(seed uint64) ([]string, []int) {
	r := rand.New(rand.NewPCG(seed, seed+1))
	bases := make([]string, numRandomOperands)
	exps := make([]int, numRandomOperands)
	for i := range bases {
		bases[i] = fmt.Sprintf("1.0%018d", r.Uint64N(1e18))
		exps[i] = 1 + r.IntN(365)
	}
	return bases, exps
}

// expOperands returns numbers in [-40, 40) with all Scale fractional digits,
// where e^x stays in range.
func expOperands(seed uint64) []string {
	r := rand.New(rand.NewPCG(seed, seed+1))
	ss := make([]string, numRandomOperands)
	for i := range ss {
		ss[i] = fmt.Sprintf("%d.%019d", r.IntN(80)-40, r.Uint64N(1e19))
	}
	return ss
}

// powRealOperands returns bases in [1, 1000) and exponents in [-3, 3), both
// with all Scale fractional digits.
func powRealOperands(seed uint64) ([]string, []string) {
	r := rand.New(rand.NewPCG(seed, seed+1))
	bases := make([]string, numRandomOperands)
	exps := make([]string, numRandomOperands)
	for i := range bases {
		bases[i] = fmt.Sprintf("%d.%019d", 1+r.IntN(999), r.Uint64N(1e19))
		e := r.IntN(6) - 3
		exps[i] = fmt.Sprintf("%d.%019d", e, r.Uint64N(1e19))
		if e < 0 {
			// Keep the fraction pointing the same way as the whole part.
			exps[i] = fmt.Sprintf("-%d.%019d", -e-1, r.Uint64N(1e19))
		}
	}
	return bases, exps
}
