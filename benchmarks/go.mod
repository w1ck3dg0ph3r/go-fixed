module benchmarks

go 1.27

replace go.w1ck3dg0ph3r.dev/fixed => ..

require (
	github.com/jokruger/dec128 v1.4.0
	github.com/quagmt/udecimal v1.10.0
	github.com/shopspring/decimal v1.4.0
	go.w1ck3dg0ph3r.dev/fixed v0.0.0
)

require github.com/cockroachdb/apd/v3 v3.2.3 // indirect
