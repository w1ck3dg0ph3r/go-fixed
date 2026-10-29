package pgxfixed_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.w1ck3dg0ph3r.dev/fixed"
	"go.w1ck3dg0ph3r.dev/fixed/codecs/pgxfixed"
)

var formats = []struct {
	name string
	code int16
}{
	{"binary", pgtype.BinaryFormatCode},
	{"text", pgtype.TextFormatCode},
}

var values = []string{
	"0",
	"1",
	"-1",
	"0.5",
	"-0.5",
	"123.45",
	"-123.45",
	"10000",
	"100000000",
	"1000000000000000000",
	"0.0001",
	"0.00001",
	"0.0000000000000000001",
	"-0.0000000000000000001",
	"0.0000000000000000010",
	"0.1234567890123456789",
	"1234567890123456789",
	"1234567890.0000000001",
	"9999999999999999999.9999999999999999999",
	"-9999999999999999999.9999999999999999999",
}

func newMap() *pgtype.Map {
	m := pgtype.NewMap()
	pgxfixed.Register(m)
	return m
}

// TestEncode encodes a Number and decodes it with the default numeric codec.
func TestEncode(t *testing.T) {
	t.Parallel()

	m := newMap()
	for _, f := range formats {
		for _, s := range values {
			t.Run(f.name+"/"+s, func(t *testing.T) {
				buf, err := m.Encode(pgtype.NumericOID, f.code, fixed.MustFromString(s), nil)
				if err != nil {
					t.Fatal(err)
				}
				var r pgtype.Numeric
				if err := pgtype.NewMap().Scan(pgtype.NumericOID, f.code, buf, &r); err != nil {
					t.Fatal(err)
				}
				got, _ := r.MarshalJSON()
				if exp := fixed.MustFromString(s).String(); string(got) != exp {
					t.Errorf("expected %s, got %s", exp, got)
				}
			})
		}
	}
}

// TestScan encodes a value with the default numeric codec and scans it into a
// Number.
func TestScan(t *testing.T) {
	t.Parallel()

	m := newMap()
	for _, f := range formats {
		for _, s := range values {
			t.Run(f.name+"/"+s, func(t *testing.T) {
				var src pgtype.Numeric
				if err := src.ScanScientific(s); err != nil {
					t.Fatal(err)
				}
				buf, err := pgtype.NewMap().Encode(pgtype.NumericOID, f.code, src, nil)
				if err != nil {
					t.Fatal(err)
				}
				var r fixed.Number
				if err := m.Scan(pgtype.NumericOID, f.code, buf, &r); err != nil {
					t.Fatal(err)
				}
				if exp := fixed.MustFromString(s); r != exp {
					t.Errorf("expected %q, got %q", exp, r)
				}
			})
		}
	}
}

// TestScanNull scans NULL as zero, while pointer and sql.Null targets keep it
// distinct.
func TestScanNull(t *testing.T) {
	t.Parallel()

	m := newMap()
	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			r := fixed.MustFromString("42")
			if err := m.Scan(pgtype.NumericOID, f.code, nil, &r); err != nil || r != fixed.Zero {
				t.Errorf("Number: expected 0, got %q (%v)", r, err)
			}

			p := new(fixed.Number)
			if err := m.Scan(pgtype.NumericOID, f.code, nil, &p); err != nil || p != nil {
				t.Errorf("*Number: expected nil, got %v (%v)", p, err)
			}

			n := sql.Null[fixed.Number]{V: fixed.One, Valid: true}
			if err := m.Scan(pgtype.NumericOID, f.code, nil, &n); err != nil || n.Valid {
				t.Errorf("sql.Null: expected invalid, got %v (%v)", n, err)
			}
		})
	}
}

func TestScanErrors(t *testing.T) {
	t.Parallel()

	m := newMap()
	cases := []struct {
		name string
		in   string
		err  error
	}{
		{"NaN", "NaN", fixed.ErrNaN},
		{"Infinity", "Infinity", fixed.ErrNaN},
		{"-Infinity", "-Infinity", fixed.ErrNaN},
		{"too big", "10000000000000000000", fixed.ErrOverflow},
		{"too big negative", "-10000000000000000000", fixed.ErrOverflow},
		{"way too big", "1e30", fixed.ErrOverflow},
		{"too precise", "0.00000000000000000001", fixed.ErrInexact},
		{"too precise digit", "0.12345678901234567891", fixed.ErrInexact},
		{"too big and too precise", "10000000000000000000.00000000000000000001", fixed.ErrOverflow},
	}
	for _, f := range formats {
		for _, tc := range cases {
			t.Run(f.name+"/"+tc.name, func(t *testing.T) {
				var src pgtype.Numeric
				if err := src.ScanScientific(tc.in); err != nil {
					t.Fatal(err)
				}
				buf, err := pgtype.NewMap().Encode(pgtype.NumericOID, f.code, src, nil)
				if err != nil {
					t.Fatal(err)
				}
				r := fixed.MustFromString("42")
				if err := m.Scan(pgtype.NumericOID, f.code, buf, &r); !errors.Is(err, tc.err) {
					t.Errorf("expected err %v, got %v", tc.err, err)
				}
				if r != fixed.MustFromString("42") {
					t.Errorf("value changed on error: %q", r)
				}
			})
		}
	}

	t.Run("binary/malformed", func(t *testing.T) {
		for _, buf := range [][]byte{
			{0, 1},
			{0, 1, 0, 0, 0, 0, 0, 0},
			{0, 1, 0, 0, 0, 0, 0, 0, 0x27, 0x10},
			{0, 0, 0, 0, 0x12, 0x34, 0, 0},
		} {
			var r fixed.Number
			err := m.Scan(pgtype.NumericOID, pgtype.BinaryFormatCode, buf, &r)
			if err == nil || errors.Is(err, fixed.ErrNaN) {
				t.Errorf("%x: expected malformed data error, got %q (%v)", buf, r, err)
			}
		}
	})
}

func TestEncodeErrors(t *testing.T) {
	t.Parallel()

	m := newMap()
	for _, f := range formats {
		for _, v := range []fixed.Number{fixed.NaN, fixed.Overflow, fixed.DivisionByZero} {
			t.Run(f.name+"/"+v.String(), func(t *testing.T) {
				if _, err := m.Encode(pgtype.NumericOID, f.code, v, nil); !errors.Is(err, v.Err()) {
					t.Errorf("expected err %v, got %v", v.Err(), err)
				}
			})
		}
	}
}

// TestOtherTypes checks that Register keeps the default handling of other Go
// types.
func TestOtherTypes(t *testing.T) {
	t.Parallel()

	m := newMap()
	buf, err := m.Encode(pgtype.NumericOID, pgtype.BinaryFormatCode, 1.5, nil)
	if err != nil {
		t.Fatal(err)
	}
	var f float64
	if err := m.Scan(pgtype.NumericOID, pgtype.BinaryFormatCode, buf, &f); err != nil || f != 1.5 {
		t.Errorf("expected 1.5, got %v (%v)", f, err)
	}
	var n fixed.Number
	if err := m.Scan(pgtype.NumericOID, pgtype.BinaryFormatCode, buf, &n); err != nil || n != fixed.MustFromString("1.5") {
		t.Errorf("expected 1.5, got %q (%v)", n, err)
	}
}

func TestArray(t *testing.T) {
	t.Parallel()

	m := newMap()
	in := []fixed.Number{fixed.MustFromString("1.5"), fixed.MustFromString("-0.0000000000000000001")}
	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			buf, err := m.Encode(pgtype.NumericArrayOID, f.code, in, nil)
			if err != nil {
				t.Fatal(err)
			}
			var r []fixed.Number
			if err := m.Scan(pgtype.NumericArrayOID, f.code, buf, &r); err != nil {
				t.Fatal(err)
			}
			if len(r) != len(in) || r[0] != in[0] || r[1] != in[1] {
				t.Errorf("expected %v, got %v", in, r)
			}
		})
	}
}

func TestAllocs(t *testing.T) {
	m := newMap()
	var v any = fixed.MustFromString("-1234567890.1234567890123456789")
	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			ep := m.PlanEncode(pgtype.NumericOID, f.code, v)
			buf := make([]byte, 0, 64)
			src, _ := ep.Encode(v, nil)
			var r fixed.Number
			sp := m.PlanScan(pgtype.NumericOID, f.code, &r)
			if a := testing.AllocsPerRun(100, func() { buf, _ = ep.Encode(v, buf[:0]) }); a != 0 {
				t.Errorf("encode: %v allocs", a)
			}
			if a := testing.AllocsPerRun(100, func() { _ = sp.Scan(src, &r) }); a != 0 {
				t.Errorf("scan: %v allocs", a)
			}
		})
	}
}

// TestDatabase round-trips values through a real server when PGX_TEST_DATABASE
// is set to its connection string.
//
//	podman run --rm -d --name pgxfixed-test -e POSTGRES_PASSWORD=postgres -p 5432:5432 docker.io/library/postgres:18
//	PGX_TEST_DATABASE='postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable'
func TestDatabase(t *testing.T) {
	dsn := os.Getenv("PGX_TEST_DATABASE")
	if dsn == "" {
		t.Skip("PGX_TEST_DATABASE is not set")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	pgxfixed.Register(conn.TypeMap())

	for _, s := range values {
		t.Run(s, func(t *testing.T) {
			in := fixed.MustFromString(s)
			var r fixed.Number
			var text string
			err := conn.QueryRow(ctx, "select $1::numeric, $1::numeric::text", in).Scan(&r, &text)
			if err != nil {
				t.Fatal(err)
			}
			if r != in {
				t.Errorf("expected %q, got %q", in, r)
			}
			if exp := in.String(); fixed.MustFromString(text) != in {
				t.Errorf("server sees %s, expected %s", text, exp)
			}
		})
	}
}
