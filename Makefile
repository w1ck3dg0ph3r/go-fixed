BENCH    ?= .
FUZZ     ?= .
FUZZTIME ?= 1000000x

default:
	@echo "Usage:"
	@echo "  make test            - run all tests and record coverage"
	@echo "  make lint            - run go vet and check formatting"
	@echo "  make bench [BENCH=.] - run some or all benchmarks"
	@echo "  make fuzz  [FUZZ=.]  - run one or all fuzz tests"
.PHONY: default

TESTFLAGS := -race -coverprofile=coverage.out
MODULES   := . codecs/pgxfixed codecs/jxfixed benchmarks fuzzing

test: test-fixed test-codecs test-benchmarks test-fuzzing
.PHONY: test

test-codecs: test-pgxfixed test-jxfixed
.PHONY: test-codecs

test-fixed:
	@go test ${TESTFLAGS} .
.PHONY: test-fixed

test-pgxfixed:
	@cd codecs/pgxfixed &&\
	go test ${TESTFLAGS} .
.PHONY: test-pgxfixed

test-jxfixed:
	@cd codecs/jxfixed &&\
	go test ${TESTFLAGS} .
.PHONY: test-jxfixed

test-benchmarks:
	@cd benchmarks &&\
	go test .
.PHONY: test-benchmarks

test-fuzzing:
	@cd fuzzing &&\
	go test .
.PHONY: test-fuzzing

lint:
	@unformatted=$$(gofmt -l .) ;\
	if [ -n "$$unformatted" ]; then \
		echo "not formatted with gofmt:" ;\
		echo "$$unformatted" ;\
		exit 1 ;\
	fi
	@for m in $(MODULES); do \
		(cd $$m && go vet ./...) || exit 1 ;\
	done
.PHONY: lint

bench:
	@cd benchmarks &&\
	go test . -v -run=^$$ -benchmem -bench=Benchmark$(BENCH)
.PHONY: bench

bench-table:
	@cd benchmarks &&\
	go run ./cmd/benchtable
.PHONY: bench-table

fuzz:
	@cd fuzzing &&\
	if [ "$(FUZZ)" = "." ]; then \
		for name in $$(go test -list='^Fuzz' | grep '^Fuzz'); do \
			echo "running $$name" ;\
			go test . -fuzz="^$${name}$$" -fuzztime=$(FUZZTIME) || exit 1 ;\
		done ;\
	else \
		go test . -fuzz='Fuzz$(FUZZ)$$' ;\
	fi
.PHONY: fuzz
