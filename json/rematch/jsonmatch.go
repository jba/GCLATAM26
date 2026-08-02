// Command jsonmatch reads a file of JSON and prints all string values that
// match a regexp.
//
// Usage:
//
//	jsonmatch [-v1] <regexp> [file]
//
// By default it parses using the encoding/json/jsontext streaming decoder.
// With -v1 it parses using the original encoding/json (v1) token API,
// producing the same output.
//
// If no file is given, JSON is read from standard input.
package main

import (
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main() {
	v1 := flag.Bool("v1", false, "parse using the encoding/json v1 API")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: jsonmatch [-v1] <regexp> [file]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	re, err := regexp.Compile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid regexp: %v\n", err)
		os.Exit(1)
	}

	r := io.Reader(os.Stdin)
	if flag.NArg() > 1 {
		f, err := os.Open(flag.Arg(1))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	run := runV2
	if *v1 {
		run = runV1
	}
	if err := run(os.Stdout, r, re); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runV2 reads JSON tokens with jsontext and writes every string token that
// matches re to w.
func runV2(w io.Writer, r io.Reader, re *regexp.Regexp) error {
	dec := jsontext.NewDecoder(r)
	for {
		tok, err := dec.ReadToken()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// Kind '"' denotes a JSON string. This includes object member names.
		if tok.Kind() == '"' {
			if s := tok.String(); re.MatchString(s) {
				fmt.Fprintln(w, s)
			}
		}
	}
}

// runV1 reads JSON tokens with the encoding/json v1 API and writes every
// string token that matches re to w. It produces the same output as runV2.
func runV1(w io.Writer, r io.Reader, re *regexp.Regexp) error {
	dec := jsonv1.NewDecoder(r)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// A JSON string (both object member names and string values) is
		// reported as a Go string. Delimiters come back as json.Delim.
		if s, ok := tok.(string); ok {
			if re.MatchString(s) {
				fmt.Fprintln(w, s)
			}
		}
	}
}
// runV2Bytes is an allocation-minimized variant of runV2. Instead of
// converting each string token to a Go string (which copies) and printing it
// through fmt (which boxes into interface{}), it reads string tokens as raw
// jsontext.Values that alias the decoder's buffer, matches on those bytes, and
// writes bytes directly.
//
// Note: it matches/writes the raw JSON bytes with the surrounding quotes
// stripped, without unescaping. For input strings containing JSON escape
// sequences the output would differ from runV2; for escape-free input it is
// identical.
func runV2Bytes(w io.Writer, r io.Reader, re *regexp.Regexp) error {
	dec := jsontext.NewDecoder(r)
	nl := []byte{'\n'}
	for {
		if dec.PeekKind() == '"' {
			val, err := dec.ReadValue()
			if err != nil {
				return err
			}
			// val is `"..."`; strip the surrounding quotes.
			inner := val[1 : len(val)-1]
			if re.Match(inner) {
				w.Write(inner)
				w.Write(nl)
			}
			continue
		}
		_, err := dec.ReadToken()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
