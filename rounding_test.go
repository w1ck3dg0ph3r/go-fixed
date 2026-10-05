package fixed_test

import (
	"math"
	"testing"

	"go.w1ck3dg0ph3r.dev/fixed"
)

func TestRoundingModes(t *testing.T) {
	t.Parallel()

	n := fixed.MustFromString
	tests := []struct {
		name  string
		mode  fixed.RoundingMode
		in    string
		scale int
		want  fixed.Number
	}{
		{"zero positive", fixed.RoundZero, "3.14159", 3, n("3.141")},
		{"zero negative", fixed.RoundZero, "-3.14159", 3, n("-3.141")},
		{"infinity positive", fixed.RoundInf, "3.14101", 3, n("3.142")},
		{"infinity negative", fixed.RoundInf, "-3.14101", 3, n("-3.142")},
		{"floor positive", fixed.RoundFloor, "3.14159", 3, n("3.141")},
		{"floor negative", fixed.RoundFloor, "-3.14101", 3, n("-3.142")},
		{"ceil positive", fixed.RoundCeil, "3.14101", 3, n("3.142")},
		{"ceil negative", fixed.RoundCeil, "-3.14159", 3, n("-3.141")},
		{"half zero below", fixed.RoundHalfZero, "1.2344", 3, n("1.234")},
		{"half zero above", fixed.RoundHalfZero, "1.2346", 3, n("1.235")},
		{"half zero positive tie", fixed.RoundHalfZero, "1.2345", 3, n("1.234")},
		{"half zero negative tie", fixed.RoundHalfZero, "-1.2345", 3, n("-1.234")},
		{"half infinity positive tie", fixed.RoundHalfInf, "1.2345", 3, n("1.235")},
		{"half infinity negative tie", fixed.RoundHalfInf, "-1.2345", 3, n("-1.235")},
		{"bank positive even tie", fixed.RoundBank, "1.2345", 3, n("1.234")},
		{"bank positive odd tie", fixed.RoundBank, "1.2355", 3, n("1.236")},
		{"bank negative even tie", fixed.RoundBank, "-1.2345", 3, n("-1.234")},
		{"bank negative odd tie", fixed.RoundBank, "-1.2355", 3, n("-1.236")},
		{"exact", fixed.RoundInf, "1.234", 3, n("1.234")},
		{"zero", fixed.RoundInf, "0", 0, fixed.Zero},
		{"full scale", fixed.RoundInf, "1.2345678901234567891", fixed.Scale, n("1.2345678901234567891")},
		{"negative scale zero positive", fixed.RoundZero, "123.45", -2, n("100")},
		{"negative scale zero negative", fixed.RoundZero, "-123.45", -2, n("-100")},
		{"negative scale infinity positive", fixed.RoundInf, "1.2345", -1, n("10")},
		{"negative scale infinity negative", fixed.RoundInf, "-1.2345", -1, n("-10")},
		{"negative scale floor positive", fixed.RoundFloor, "123.45", -1, n("120")},
		{"negative scale floor negative", fixed.RoundFloor, "-123.45", -1, n("-130")},
		{"negative scale ceil positive", fixed.RoundCeil, "123.45", -1, n("130")},
		{"negative scale ceil negative", fixed.RoundCeil, "-123.45", -1, n("-120")},
		{"negative scale half zero tie", fixed.RoundHalfZero, "125", -1, n("120")},
		{"negative scale half zero negative tie", fixed.RoundHalfZero, "-125", -1, n("-120")},
		{"negative scale half zero above tie by fraction", fixed.RoundHalfZero, "125.0000000000000000001", -1, n("130")},
		{"negative scale half infinity tie", fixed.RoundHalfInf, "125", -1, n("130")},
		{"negative scale half infinity whole below tie", fixed.RoundHalfInf, "1240", -2, n("1200")},
		{"negative scale bank whole below tie odd", fixed.RoundBank, "1140", -2, n("1100")},
		{"negative scale half infinity negative tie", fixed.RoundHalfInf, "-125", -1, n("-130")},
		{"negative scale half infinity below tie by fraction", fixed.RoundHalfInf, "124.9999999999999999999", -1, n("120")},
		{"negative scale bank even tie", fixed.RoundBank, "125", -1, n("120")},
		{"negative scale bank odd tie", fixed.RoundBank, "-135", -1, n("-140")},
		{"negative scale bank above tie by fraction", fixed.RoundBank, "125.0000000000000000001", -1, n("130")},
		{"negative scale exact", fixed.RoundInf, "1200", -2, n("1200")},
		{"negative scale fraction only", fixed.RoundInf, "1200.5", -2, n("1300")},
		{"negative scale fraction only floor", fixed.RoundFloor, "-1200.5", -2, n("-1300")},
		{"negative scale zero", fixed.RoundInf, "0", -20, fixed.Zero},
		{"scale -18 half infinity", fixed.RoundHalfInf, "500000000000000000", -18, n("1000000000000000000")},
		{"scale -18 zero", fixed.RoundZero, "9500000000000000000", -18, n("9000000000000000000")},
		{"scale -18 ceil overflow", fixed.RoundCeil, "9500000000000000000", -18, fixed.Overflow},
		{"scale -18 floor overflow", fixed.RoundFloor, "-9500000000000000000", -18, fixed.Overflow},
		{"scale -19 infinity overflow", fixed.RoundInf, "1", -19, fixed.Overflow},
		{"scale -19 half infinity tie overflow", fixed.RoundHalfInf, "5000000000000000000", -19, fixed.Overflow},
		{"scale -19 half infinity below tie", fixed.RoundHalfInf, "4999999999999999999.9999999999999999999", -19, fixed.Zero},
		{"scale -19 half zero tie", fixed.RoundHalfZero, "5000000000000000000", -19, fixed.Zero},
		{"scale -19 bank even tie", fixed.RoundBank, "5000000000000000000", -19, fixed.Zero},
		{"scale -19 bank above tie", fixed.RoundBank, "-5000000000000000000.0000000000000000001", -19, fixed.Overflow},
		{"scale -19 truncate max", fixed.RoundZero, "9999999999999999999.9999999999999999999", -19, fixed.Zero},
		{"scale -20 half infinity max", fixed.RoundHalfInf, "9999999999999999999.9999999999999999999", -20, fixed.Zero},
		{"scale -20 infinity overflow", fixed.RoundInf, "0.0000000000000000001", -20, fixed.Overflow},
		{"scale -20 floor positive", fixed.RoundFloor, "1", -20, fixed.Zero},
		{"scale -20 floor negative overflow", fixed.RoundFloor, "-1", -20, fixed.Overflow},
		{"scale -20 ceil negative", fixed.RoundCeil, "-1", -20, fixed.Zero},
		{"min int scale", fixed.RoundBank, "-9999999999999999999.9999999999999999999", math.MinInt, fixed.Zero},
		{"positive overflow", fixed.RoundInf, "9999999999999999999.9999999999999999999", 0, fixed.Overflow},
		{"negative overflow", fixed.RoundInf, "-9999999999999999999.9999999999999999999", 0, fixed.Overflow},
		{"propagate nan", fixed.RoundBank, "", 2, fixed.NaN},
		{"propagate overflow", fixed.RoundBank, "", 2, fixed.Overflow},
		{"propagate division by zero", fixed.RoundBank, "", 2, fixed.DivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.want
			if tt.in != "" {
				in = n(tt.in)
			}
			if got := in.Round(tt.scale, tt.mode); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
