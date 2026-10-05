package benchmarks

import (
	"testing"

	cockroachdb_apd "github.com/cockroachdb/apd/v3"
	jokruger_dec128 "github.com/jokruger/dec128"
	quagmt_udecimal "github.com/quagmt/udecimal"
	shopspring_decimal "github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func BenchmarkCmp(b *testing.B) {
	s1, s2 := randomNumbers(4), randomNumbers(2)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n1, n2 := parseAll(s1, fixed.MustFromString), parseAll(s2, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Cmp(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n1, n2 := parseAll(s1, parse), parseAll(s2, parse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Compare(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, quagmt_udecimal.MustParse), parseAll(s2, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Cmp(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n1, n2 := parseAll(s1, shopspring_decimal.RequireFromString), parseAll(s2, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Cmp(n2[i&(numRandomOperands-1)])
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		n1, n2 := parseAll(s1, apdParse), parseAll(s2, apdParse)
		i := 0
		for b.Loop() {
			_ = n1[i&(numRandomOperands-1)].Cmp(n2[i&(numRandomOperands-1)])
			i++
		}
	})
}

func BenchmarkSign(b *testing.B) {
	s := randomNumbers(11)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n := parseAll(s, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Sign()
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n := parseAll(s, parse)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Sign()
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n := parseAll(s, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Sign()
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n := parseAll(s, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Sign()
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(s, apdParse)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].Sign()
			i++
		}
	})
}

func BenchmarkAbs(b *testing.B) {
	s := randomNumbers(12)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		n := parseAll(s, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Abs()
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		parse := jokruger_dec128.FromString[string]
		n := parseAll(s, parse)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Abs()
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		n := parseAll(s, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Abs()
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		n := parseAll(s, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = n[i&(numRandomOperands-1)].Abs()
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(s, apdParse)
		var d cockroachdb_apd.Decimal
		i := 0
		for b.Loop() {
			_ = d.Abs(ns[i&(numRandomOperands-1)])
			i++
		}
	})
}
