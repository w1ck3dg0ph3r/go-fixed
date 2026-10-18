// Benchtable runs the comparison benchmarks and prints the README performance
// table, with the time per operation averaged over several runs.
//
//	go run ./cmd/benchtable [-count 3]
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var columns = []struct{ title, lib string }{
	{"fixed", "w1ck3dg0ph3r/fixed"},
	{"dec128", "jokruger/dec128"},
	{"udecimal", "quagmt/udecimal"},
	{"apd", "cockroachdb/apd"},
	{"shopspring", "shopspring/decimal"},
}

// rows names the benchmark of each operation, where %s is the library.
var rows = []struct{ title, bench string }{
	{"Add", "Add/%s"},
	{"Mul", "Mul/%s"},
	{"Div", "Div/%s"},
	{"Mod", "Mod/%s"},
	{"Sqrt", "Sqrt/%s"},
	{"PowInt", "PowInt/%s"},
	{"Exp", "Exp/%s"},
	{"Ln", "Ln/%s"},
	{"Pow", "Pow/%s"},
	{"Round", "Round/%s"},
	{"FromString", "FromString/%s/random"},
	{"String", "String/%s/random"},
	{"FromFloat64", "FromFloat64/%s/random"},
	{"Float64", "Float64/%s"},
}

func main() {
	count := flag.Int("count", 3, "number of runs to average")
	benchtime := flag.String("benchtime", "1s", "go test -benchtime")
	flag.Parse()

	var out bytes.Buffer
	for _, p := range patterns() {
		cmd := exec.Command("go", "test", "benchmarks", "-run=^$", "-bench="+p,
			"-count="+strconv.Itoa(*count), "-benchtime="+*benchtime)
		cmd.Stdout = io.MultiWriter(&out, os.Stderr)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	ns, err := average(&out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	printTable(os.Stdout, ns)
}

// patterns returns a -bench pattern for each sub-benchmark suffix in rows. A
// benchmark with fewer levels than the pattern does not report, so they need
// separate runs. Library names hold a slash, so they take two levels.
func patterns() []string {
	var suffixes []string
	names := map[string][]string{}
	for _, r := range rows {
		name, rest, _ := strings.Cut(r.bench, "/%s")
		suffix := strings.TrimPrefix(rest, "/")
		if _, ok := names[suffix]; !ok {
			suffixes = append(suffixes, suffix)
		}
		names[suffix] = append(names[suffix], name)
	}

	var ps []string
	for _, suffix := range suffixes {
		p := "^Benchmark(" + strings.Join(names[suffix], "|") + ")$"
		if suffix != "" {
			p += "///^" + suffix + "$"
		}
		ps = append(ps, p)
	}
	return ps
}

var procSuffix = regexp.MustCompile(`-\d+$`)

// average returns the mean ns/op of each benchmark in the go test output r,
// keyed by name without the Benchmark prefix and GOMAXPROCS suffix.
func average(r io.Reader) (map[string]float64, error) {
	sum, n := map[string]float64{}, map[string]int{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		f := strings.Fields(s.Text())
		if len(f) < 4 || !strings.HasPrefix(f[0], "Benchmark") || f[3] != "ns/op" {
			continue
		}
		v, err := strconv.ParseFloat(f[2], 64)
		if err != nil {
			continue
		}
		name := procSuffix.ReplaceAllString(strings.TrimPrefix(f[0], "Benchmark"), "")
		sum[name] += v
		n[name]++
	}
	for k := range sum {
		sum[k] /= float64(n[k])
	}
	return sum, s.Err()
}

func printTable(w io.Writer, ns map[string]float64) {
	header := []string{"Operation"}
	for _, c := range columns {
		header = append(header, c.title)
	}
	table := [][]string{header}
	for _, r := range rows {
		line := []string{r.title}
		for _, c := range columns {
			cell := "—"
			if v, ok := ns[fmt.Sprintf(r.bench, c.lib)]; ok {
				cell = formatTime(v)
			}
			line = append(line, cell)
		}
		table = append(table, line)
	}

	widths := make([]int, len(header))
	for _, line := range table {
		for i, cell := range line {
			widths[i] = max(widths[i], utf8.RuneCountInString(cell))
		}
	}

	// The first column is aligned left and the times right.
	writeLine := func(line []string) {
		for i, cell := range line {
			pad := strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell))
			if i == 0 {
				fmt.Fprintf(w, "| %s%s ", cell, pad)
			} else {
				fmt.Fprintf(w, "| %s%s ", pad, cell)
			}
		}
		fmt.Fprintln(w, "|")
	}
	writeLine(table[0])
	sep := []string{strings.Repeat("-", widths[0])}
	for _, wd := range widths[1:] {
		sep = append(sep, strings.Repeat("-", wd-1)+":")
	}
	writeLine(sep)
	for _, line := range table[1:] {
		writeLine(line)
	}
}

// formatTime formats ns with three significant digits, but at least one
// decimal place below 100 ns.
func formatTime(ns float64) string {
	switch {
	case ns < 99.95:
		return fmt.Sprintf("%.1f ns", ns)
	case ns < 999.5:
		return fmt.Sprintf("%.0f ns", ns)
	}
	v, unit := ns/1e3, "µs"
	if v >= 999.5 {
		v, unit = v/1e3, "ms"
	}
	switch {
	case v < 9.995:
		return fmt.Sprintf("%.2f %s", v, unit)
	case v < 99.95:
		return fmt.Sprintf("%.1f %s", v, unit)
	}
	return fmt.Sprintf("%.0f %s", v, unit)
}
