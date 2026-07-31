// Command jsonmatch reads a file of JSON and prints all string values that
// match a regexp.
//
// Usage:
//
//	GOEXPERIMENT=jsonv2 go run . <regexp> [file]
//
// If no file is given, JSON is read from standard input.
package main

import (
	"encoding/json/jsontext"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: jsonmatch <regexp> [file]")
		os.Exit(2)
	}

	re, err := regexp.Compile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid regexp: %v\n", err)
		os.Exit(1)
	}

	r := io.Reader(os.Stdin)
	if len(os.Args) > 2 {
		f, err := os.Open(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	if err := run(r, re); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run reads JSON tokens from r and prints every string token that matches re.
func run(r io.Reader, re *regexp.Regexp) error {
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
				fmt.Println(s)
			}
		}
	}
}
