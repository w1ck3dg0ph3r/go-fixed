package pgxfixed_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"go.w1ck3dg0ph3r.dev/fixed"
)

func BenchmarkEncode(b *testing.B) {
	ns, ps := benchNumbers(b)
	m := newMap()

	// Values are boxed up front, as pgx receives them in []any, so that
	// converting them to any is not measured as part of encoding.
	nv, pv := make([]any, benchValues), make([]any, benchValues)
	for i := range ns {
		nv[i], pv[i] = ns[i], ps[i]
	}

	for _, f := range formats {
		b.Run(f.name, func(b *testing.B) {
			b.Run("pgxfixed", func(b *testing.B) {
				plan := m.PlanEncode(pgtype.NumericOID, f.code, nv[0])
				buf := make([]byte, 0, 64)
				i := 0
				for b.Loop() {
					buf, _ = plan.Encode(nv[i&(benchValues-1)], buf[:0])
					i++
				}
			})

			b.Run("pgtype", func(b *testing.B) {
				plan := m.PlanEncode(pgtype.NumericOID, f.code, pv[0])
				buf := make([]byte, 0, 64)
				i := 0
				for b.Loop() {
					buf, _ = plan.Encode(pv[i&(benchValues-1)], buf[:0])
					i++
				}
			})
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	_, ps := benchNumbers(b)
	m := newMap()

	for _, f := range formats {
		// The wire data comes from the stock codec, as a server would send it.
		srcs := make([][]byte, benchValues)
		for i, p := range ps {
			var err error
			if srcs[i], err = pgtype.NewMap().Encode(pgtype.NumericOID, f.code, p, nil); err != nil {
				b.Fatal(err)
			}
		}

		b.Run(f.name, func(b *testing.B) {
			b.Run("pgxfixed", func(b *testing.B) {
				var r fixed.Number
				plan := m.PlanScan(pgtype.NumericOID, f.code, &r)
				i := 0
				for b.Loop() {
					_ = plan.Scan(srcs[i&(benchValues-1)], &r)
					i++
				}
			})

			b.Run("pgtype", func(b *testing.B) {
				var r pgtype.Numeric
				plan := m.PlanScan(pgtype.NumericOID, f.code, &r)
				i := 0
				for b.Loop() {
					_ = plan.Scan(srcs[i&(benchValues-1)], &r)
					i++
				}
			})
		})
	}
}

// benchValues cycles through values, padded to a power of two so that indexing
// with a mask is cheap.
const benchValues = 32

// benchNumbers returns values as Numbers and as pgtype.Numerics, the type the
// stock codec encodes and scans.
func benchNumbers(b *testing.B) ([]fixed.Number, []pgtype.Numeric) {
	b.Helper()

	ns := make([]fixed.Number, benchValues)
	ps := make([]pgtype.Numeric, benchValues)

	for i := range ns {
		s := values[i%len(values)]
		ns[i] = fixed.MustFromString(s)
		if err := ps[i].ScanScientific(s); err != nil {
			b.Fatal(err)
		}
	}

	return ns, ps
}
