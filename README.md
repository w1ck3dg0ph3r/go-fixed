# fixed

Package fixed provides a fast, allocation-free 128-bit fixed-point decimal type for Go, designed for
monetary and other values that must be exact in base 10.

```
go get go.w1ck3dg0ph3r.dev/fixed
```

Requires Go 1.27.

## Why

- **Exact decimals** with 19 fractional and 19 integer digits (a representable range just under ±10^19).
- **Fast and allocation-free:** a number is a plain, immutable 16-byte value that is safe to copy and compare bytewise.
- **Explicit rounding**, operations that can round take an explicit rounding mode (with short aliases that truncate).
- **No panics, no error plumbing:** failed operations return sentinel values that propagate through the operation chain.
- **Thoroughly tested**, including differential fuzzing against `shopspring/decimal` and `math/big`.
- **Zero dependencies.**

## Why not

The fixed scale that makes the package fast also limits what it can hold. Look elsewhere when:

- **Values can exceed 19 integer digits.** That rules out amounts kept in tiny units (an ETH balance
  in wei passes 10^19 at 10 ETH) and running totals in high-denomination currencies. Intermediate
  results must fit too: `10000000000 * 1000000000` overflows, even if you divide right after.
- **You need more than 19 decimal places.** Anything smaller rounds away: `0.00000000001 * 0.00000000001`
  is `0`. Scientific values, very small probabilities and per-unit rates of tiny units don't fit.
- **Intermediate rounding is a problem.** Every `Mul` and `Div` truncates its result to 19 places,
  so a chain of them can drift: `1 / 3 * 3` is `0.9999999999999999999`. Long chains, such as
  compounding over many periods, can lose the last digit or two. Multiplying before dividing helps,
  as long as the product stays in range.

If any of those describes your case, reach instead for a type with a flexible scale, such as
[`jokruger/dec128`](https://github.com/jokruger/dec128), which trades precision for magnitude;
an arbitrary-precision decimal, such as [`shopspring/decimal`](https://github.com/shopspring/decimal)
or [`cockroachdb/apd`](https://github.com/cockroachdb/apd); or exact rationals with
[`math/big.Rat`](https://pkg.go.dev/math/big#Rat).

## Usage

```go
price := fixed.MustFromString("19.99")
rate := fixed.MustFromString("0.0825")
quantity := 3

subtotal := price.MulInt(quantity)                  // 59.97
tax := subtotal.Mul(rate).Round(2, fixed.RoundBank) // 4.947525 -> 4.95
total := subtotal.Add(tax)                          // 64.92

share := total.DivInt(7) // 9.2742857142857142857
if err := share.Err(); err != nil {
    // NaN, overflow or division by zero anywhere above.
}
fmt.Println(share.Round(2, fixed.RoundHalfInf)) // 9.27
```

### Rounding

`Mul` and `Div` truncate their result at the 19th place, which suits intermediate steps: round once
at the end with `Round(scale, mode)`. Where the 19th place matters, `MulRound(v, mode)` and
`DivRound(v, mode)` round the result with one of:

| Mode            | Rounds                          |
| --------------- | ------------------------------- |
| `RoundZero`     | towards zero                    |
| `RoundInf`      | away from zero                  |
| `RoundFloor`    | towards negative infinity       |
| `RoundCeil`     | towards positive infinity       |
| `RoundHalfZero` | to nearest, ties towards zero   |
| `RoundHalfInf`  | to nearest, ties away from zero |
| `RoundBank`     | to nearest, ties to even        |

A negative scale rounds to tens, hundreds and so on:
`FromInt(1250).Round(-2, RoundBank)` is `1200`. `MulExact` returns `ErrInexact` instead of rounding.

### Integer operands

`AddInt`, `SubInt`, `MulInt`, `DivInt`, `DivRoundInt`, `QuoRemInt` and `ModInt` take a plain `int`,
as in `price.MulInt(quantity)`. They are also faster than going through `FromInt`: `MulInt` is exact
and twice as fast as `Mul`, and the division-based ones save about a third.

### Powers and logarithms

`Sqrt`, `PowInt(m)`, `Exp()`, `Ln()` and `Pow(y)` truncate at the 19th place, and `PowIntRound`,
`ExpRound`, `LnRound` and `PowRound` take a rounding mode. They work with 42 to 57 significant
digits and round once, so results are correctly rounded except when the exact value lies extremely
close to a rounding boundary. `Sqrt` truncates exactly and has no rounding variant.

```go
// 5% a year, compounded monthly for 10 years.
growth := fixed.One.Add(fixed.New(5, -2).DivInt(12)).PowIntRound(120, fixed.RoundBank)
```

`Pow` recognizes exact results, so `4 ^ 0.5` is exactly `2` and `1.21 ^ 1.5` is `1.331` with every
rounding mode. `Ln` returns `NaN` for zero and negative numbers, and `Pow` for a negative base with
a fractional exponent.

### Conversion and encoding

- `New(coef, exp)`, `FromString`, `FromInt`, `FromUint`, `FromFloat64`, and back with `String`,
  `Append`, `Int`, `Uint`, `Float64`.
- `encoding.TextMarshaler` and `BinaryMarshaler`. The 16-byte binary form sorts in numeric order as
  raw bytes.
- JSON, as a string, with both `encoding/json` and `encoding/json/v2`.
- `database/sql` `Scanner` and `Valuer`. NULL scans as zero; use `sql.Null[fixed.Number]` to tell
  NULL apart from zero.
- [`codecs/pgxfixed`](codecs/pgxfixed): native PostgreSQL `numeric` support for
  [jackc/pgx](https://github.com/jackc/pgx), binary format included: `pgxfixed.Register(conn.TypeMap())`.
- [`codecs/jxfixed`](codecs/jxfixed): `go-faster/jx` encoding, and a type for [ogen](https://ogen.dev/)'s `x-ogen-type`.

### Errors

Failed operations return sentinels, and the first one wins: `Overflow.Add(NaN)` is `Overflow`.
`Number.Err()` maps a number to `ErrNaN`, `ErrOverflow` or `ErrDivisionByZero`, or to nil for a
valid number. Sentinels compare greater than every valid number.

Conversions return errors directly. `FromString` returns `ErrNaN` for bad syntax, `ErrOverflow` for
too many integer digits and `ErrInexact` for nonzero digits past the 19th decimal place.

## Performance

Benchmarked on i7-9700K, each operation cycles through 1024 random operands, average of 3 runs:

| Operation   |   fixed |  dec128 | udecimal |     apd | shopspring |
| ----------- | ------: | ------: | -------: | ------: | ---------: |
| Add         |  2.9 ns |  9.2 ns |  13.0 ns | 59.5 ns |    67.1 ns |
| Mul         | 18.0 ns | 25.2 ns |  54.4 ns |  422 ns |    1.34 µs |
| Div         | 32.1 ns | 55.1 ns |  61.3 ns | 1.35 µs |     852 ns |
| Mod         | 23.5 ns | 38.8 ns |  46.0 ns |  153 ns |     320 ns |
| Sqrt        | 51.9 ns |  210 ns |   385 ns | 7.29 µs |          — |
| PowInt      | 1.21 µs | 1.28 µs |  18.2 µs | 7.96 µs |    33.6 µs |
| Exp         | 5.29 µs | 13.9 µs |        — |  103 µs |     727 µs |
| Ln          | 6.22 µs | 24.0 µs |        — |  178 µs |     401 µs |
| Pow         | 12.2 µs | 39.9 µs |        — |  344 µs |     478 µs |
| Round       | 14.0 ns | 18.0 ns |  30.4 ns |  201 ns |    1.15 µs |
| FromString  | 34.6 ns | 40.7 ns |  49.6 ns |  326 ns |     329 ns |
| String      | 53.1 ns | 54.9 ns |  99.1 ns |  233 ns |     340 ns |
| FromFloat64 | 61.9 ns | 61.8 ns |   101 ns |  299 ns |     444 ns |
| Float64     | 13.2 ns |  139 ns |   187 ns |  358 ns |    1.06 µs |

Run these benchmarks on your machine with `make bench-table`.

## References

- [Division by Invariant Integers using Multiplication](https://doi.org/10.1145/178243.178249), Granlund, Montgomery, 1994.
- [Improved Division by Invariant Integers](https://doi.org/10.1109/TC.2010.143), Möller, Granlund, 2011.
- [Modern Computer Arithmetic](https://doi.org/10.48550/arXiv.1004.4710), Brent, Zimmermann, 2010.
- [jokruger/dec128](https://github.com/jokruger/dec128)
- [lukechampine/uint128](https://github.com/lukechampine/uint128)
- [quagmt/udecimal](https://github.com/quagmt/udecimal)
