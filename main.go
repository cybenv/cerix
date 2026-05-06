package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

const defaultTopN = 10

func main() {
	topN := flag.Int("n", defaultTopN, "number of top relays to print on stderr")
	outPath := flag.String("o", "", "write JSON output to <path> instead of stdout")
	country := flag.String("c", "", "filter Top-N to relays starting with <country>- (e.g. us)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: cerix [-n N] [-c COUNTRY] [-o PATH] <compact-log>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	logPath := flag.Arg(0)

	out, err := parseLog(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cerix: %v\n", err)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(&out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cerix: marshal: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')

	var w io.Writer = os.Stdout
	if *outPath != "" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cerix: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	if _, err := w.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "cerix: write: %v\n", err)
		os.Exit(1)
	}

	writeOverview(os.Stderr, &out)
	writeFilterInfo(os.Stderr, out.Servers, *country)
	writeTopN(os.Stderr, filterByCountry(out.Servers, *country), *topN)
}
