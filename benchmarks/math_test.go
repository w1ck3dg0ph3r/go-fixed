package benchmarks

import (
	"testing"

	cockroachdb_apd "github.com/cockroachdb/apd/v3"
	jokruger_dec128 "github.com/jokruger/dec128"
	quagmt_udecimal "github.com/quagmt/udecimal"
	shopspring_decimal "github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func BenchmarkAdd(b *testing.B) {
	s1, s2 := randomNumbers(1), randomNumbers(2)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Add(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Add(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Add(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Add(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Add(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			i++
		}
	})
}

func BenchmarkMul(b *testing.B) {
	s1, s2 := randomNumbers(1), randomNumbers(2)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].MulRound(n2[i&(numRandomOperands-1)], fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].MulRound(n2[i&(numRandomOperands-1)], fixed.Scale, jokruger_dec128.ROUND_BANK)
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mul(n2[i&(numRandomOperands-1)]).RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mul(n2[i&(numRandomOperands-1)]).RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Mul(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkMulPositive(b *testing.B) {
	s1, s2 := randomPositiveNumbers(1), randomPositiveNumbers(2)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].MulRound(n2[i&(numRandomOperands-1)], fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].MulRound(n2[i&(numRandomOperands-1)], fixed.Scale, jokruger_dec128.ROUND_BANK)
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mul(n2[i&(numRandomOperands-1)]).RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mul(n2[i&(numRandomOperands-1)]).RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Mul(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkDiv(b *testing.B) {
	s1, s2 := randomNumbers(3), randomNumbers(4)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].DivRound(n2[i&(numRandomOperands-1)], fixed.RoundHalfInf)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].DivRound(n2[i&(numRandomOperands-1)], fixed.Scale, jokruger_dec128.ROUND_HALF_AWAY_FROM_ZERO)
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			r, _ := n1[i&(numRandomOperands-1)].Div(n2[i&(numRandomOperands-1)])
			_ = r.RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].DivRound(n2[i&(numRandomOperands-1)], fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		// Quo rounds to significant digits, so truncate far below the last
		// place first.
		quo, ctx := apdContext(cockroachdb_apd.RoundDown), apdContext(cockroachdb_apd.RoundHalfUp)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = quo.Quo(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkDivPositive(b *testing.B) {
	s1, s2 := randomPositiveNumbers(3), randomPositiveNumbers(4)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].DivRound(n2[i&(numRandomOperands-1)], fixed.RoundHalfInf)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].DivRound(n2[i&(numRandomOperands-1)], fixed.Scale, jokruger_dec128.ROUND_HALF_AWAY_FROM_ZERO)
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			r, _ := n1[i&(numRandomOperands-1)].Div(n2[i&(numRandomOperands-1)])
			_ = r.RoundBank(fixed.Scale)
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].DivRound(n2[i&(numRandomOperands-1)], fixed.Scale)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		quo, ctx := apdContext(cockroachdb_apd.RoundDown), apdContext(cockroachdb_apd.RoundHalfUp)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = quo.Quo(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			_, _ = ctx.Quantize(&d, &d, -fixed.Scale)
			i++
		}
	})
}

func BenchmarkMod(b *testing.B) {
	s1, s2 := randomNumbers(1), randomNumbers(2)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_, _ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Rem(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			i++
		}
	})
}

func BenchmarkModPositive(b *testing.B) {
	s1, s2 := randomPositiveNumbers(1), randomPositiveNumbers(2)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_, _ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Mod(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Rem(&d, n1[i&(numRandomOperands-1)], n2[i&(numRandomOperands-1)])
			i++
		}
	})
}
