package benchmarks

import (
	"testing"

	cockroachdb_apd "github.com/cockroachdb/apd/v3"
	jokruger_dec128 "github.com/jokruger/dec128"
	quagmt_udecimal "github.com/quagmt/udecimal"
	shopspring_decimal "github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func BenchmarkRound(b *testing.B) {
	s := randomNumbers(5)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		ns := parseAll(s, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].Round(2, fixed.RoundBank)
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		ns := parseAll(s, jokruger_dec128.FromString[string])
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].RoundBank(2)
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		ns := parseAll(s, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].RoundBank(2)
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		ns := parseAll(s, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].RoundBank(2)
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(s, apdParse)
		ctx := apdContext(cockroachdb_apd.RoundHalfEven)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_, _ = ctx.Quantize(&d, ns[i&(numRandomOperands-1)], -2)
			i++
		}
	})
}
