package benchmarks

import (
	"strconv"
	"testing"

	cockroachdb_apd "github.com/cockroachdb/apd/v3"
	jokruger_dec128 "github.com/jokruger/dec128"
	quagmt_udecimal "github.com/quagmt/udecimal"
	shopspring_decimal "github.com/shopspring/decimal"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func BenchmarkFromString(b *testing.B) {
	variants := []struct {
		name string
		in   string
	}{
		{"short", "42"},
		{"medium", "123.45678901"},
		{"long", "2398475923874510023.1203812094581294120"},
	}

	random := randomDecimals(7)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				for b.Loop() {
					_ = fixed.MustFromString(v.in)
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = fixed.MustFromString(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				for b.Loop() {
					_ = jokruger_dec128.FromString(v.in)
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = jokruger_dec128.FromString(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				for b.Loop() {
					_ = quagmt_udecimal.MustParse(v.in)
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = quagmt_udecimal.MustParse(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				for b.Loop() {
					_ = shopspring_decimal.RequireFromString(v.in)
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = shopspring_decimal.RequireFromString(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		var d cockroachdb_apd.Decimal
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				for b.Loop() {
					_, _, _ = d.SetString(v.in)
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_, _, _ = d.SetString(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})
}

func BenchmarkString(b *testing.B) {
	variants := []struct {
		name string
		in   string
	}{
		{"short", "42"},
		{"medium", "123.45678901"},
		{"long", "2398475923874510023.1203812094581294120"},
	}

	random := randomDecimals(7)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				n := fixed.MustFromString(v.in)
				for b.Loop() {
					_ = n.String()
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			ns := parseAll(random, fixed.MustFromString)
			i := 0
			for b.Loop() {
				_ = ns[i&(numRandomOperands-1)].String()
				i++
			}
		})
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				n := jokruger_dec128.FromString(v.in)
				for b.Loop() {
					_ = n.String()
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			ns := parseAll(random, jokruger_dec128.FromString[string])
			i := 0
			for b.Loop() {
				_ = ns[i&(numRandomOperands-1)].String()
				i++
			}
		})
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				n := quagmt_udecimal.MustParse(v.in)
				for b.Loop() {
					_ = n.String()
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			ns := parseAll(random, quagmt_udecimal.MustParse)
			i := 0
			for b.Loop() {
				_ = ns[i&(numRandomOperands-1)].String()
				i++
			}
		})
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				n := shopspring_decimal.RequireFromString(v.in)
				for b.Loop() {
					_ = n.String()
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			ns := parseAll(random, shopspring_decimal.RequireFromString)
			i := 0
			for b.Loop() {
				_ = ns[i&(numRandomOperands-1)].String()
				i++
			}
		})
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		for _, v := range variants {
			b.Run(v.name, func(b *testing.B) {
				n := apdParse(v.in)
				for b.Loop() {
					_ = n.String()
				}
			})
		}
		b.Run("random", func(b *testing.B) {
			ns := parseAll(random, apdParse)
			i := 0
			for b.Loop() {
				_ = ns[i&(numRandomOperands-1)].String()
				i++
			}
		})
	})
}

func BenchmarkFromFloat64(b *testing.B) {
	small := 12345.67891

	random := parseAll(randomNumbers(8), func(s string) float64 {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	})

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		b.Run("small", func(b *testing.B) {
			for b.Loop() {
				_ = fixed.FromFloat64(small)
			}
		})
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = fixed.FromFloat64(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		b.Run("small", func(b *testing.B) {
			for b.Loop() {
				_ = jokruger_dec128.FromFloat64(small)
			}
		})
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = jokruger_dec128.FromFloat64(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		b.Run("small", func(b *testing.B) {
			for b.Loop() {
				_, _ = quagmt_udecimal.NewFromFloat64(small)
			}
		})
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_, _ = quagmt_udecimal.NewFromFloat64(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		b.Run("small", func(b *testing.B) {
			for b.Loop() {
				_ = shopspring_decimal.NewFromFloat(small)
			}
		})
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_ = shopspring_decimal.NewFromFloat(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		var d cockroachdb_apd.Decimal
		b.Run("small", func(b *testing.B) {
			for b.Loop() {
				_, _ = d.SetFloat64(small)
			}
		})
		b.Run("random", func(b *testing.B) {
			i := 0
			for b.Loop() {
				_, _ = d.SetFloat64(random[i&(numRandomOperands-1)])
				i++
			}
		})
	})
}

func BenchmarkFloat64(b *testing.B) {
	random := randomNumbers(9)

	b.Run("w1ck3dg0ph3r/fixed", func(b *testing.B) {
		ns := parseAll(random, fixed.MustFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].Float64()
			i++
		}
	})

	b.Run("jokruger/dec128", func(b *testing.B) {
		ns := parseAll(random, jokruger_dec128.FromString[string])
		i := 0
		for b.Loop() {
			_, _ = ns[i&(numRandomOperands-1)].InexactFloat64()
			i++
		}
	})

	b.Run("quagmt/udecimal", func(b *testing.B) {
		ns := parseAll(random, quagmt_udecimal.MustParse)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].InexactFloat64()
			i++
		}
	})

	b.Run("shopspring/decimal", func(b *testing.B) {
		ns := parseAll(random, shopspring_decimal.RequireFromString)
		i := 0
		for b.Loop() {
			_ = ns[i&(numRandomOperands-1)].InexactFloat64()
			i++
		}
	})

	b.Run("cockroachdb/apd", func(b *testing.B) {
		ns := parseAll(random, apdParse)
		i := 0
		for b.Loop() {
			_, _ = ns[i&(numRandomOperands-1)].Float64()
			i++
		}
	})
}
