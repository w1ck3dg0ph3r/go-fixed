package benchmarks

import (
	"fmt"
	"math/rand/v2"
)

// numRandomOperands is the number of random operands, a power of two so that
// indexing with a mask is cheap.
//
//	i&(numRandomOperands-1)
const numRandomOperands = 1024

// randomPositiveNumbers returns random positive numbers with an integer part
// in [1, 1e6) and all Scale fractional digits.
func randomPositiveNumbers(seed uint64) []string {
	r := rand.New(rand.NewPCG(seed, seed+1))
	ss := make([]string, numRandomOperands)
	for i := range ss {
		ss[i] = fmt.Sprintf("%d.%019d", 1+r.IntN(999999), r.Uint64N(1e19))
	}
	return ss
}

// randomNumbers returns the numbers of randomPositiveNumbers, each negated
// or not at random, so that branches on the sign cannot be predicted.
func randomNumbers(seed uint64) []string {
	r := rand.New(rand.NewPCG(seed, seed+2))
	ss := randomPositiveNumbers(seed)
	for i := range ss {
		if r.IntN(2) == 0 {
			ss[i] = "-" + ss[i]
		}
	}
	return ss
}

// randomDecimals returns random numbers of varying length: a sign, 1 to 19
// integer digits and 0 to Scale fractional digits, so that loops over the
// digits run a different number of times from one string to the next.
func randomDecimals(seed uint64) []string {
	r := rand.New(rand.NewPCG(seed, seed+1))
	ss := make([]string, numRandomOperands)
	for i := range ss {
		var b []byte
		if r.IntN(2) == 0 {
			b = append(b, '-')
		}
		b = append(b, byte('1'+r.IntN(9)))
		for range r.IntN(19) {
			b = append(b, byte('0'+r.IntN(10)))
		}
		if k := r.IntN(20); k > 0 {
			b = append(b, '.')
			for range k {
				b = append(b, byte('0'+r.IntN(10)))
			}
		}
		ss[i] = string(b)
	}
	return ss
}

func parseAll[T any](ss []string, parse func(string) T) []T {
	ts := make([]T, len(ss))
	for i, s := range ss {
		ts[i] = parse(s)
	}
	return ts
}
