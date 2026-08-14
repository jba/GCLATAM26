// Command match reads JSON from stdin (or a file) and prints the number of
// JSON strings (both object member names and string values) equal to the
// string given on the command line.
//
// Usage:
//
//	match [-v1] <string> [file]
//
// By default it uses encoding/json/jsontext. With -v1 it uses the original
// encoding/json (v1) token API.
package main

import (
	jsonv1 "encoding/json"
	"encoding/json/jsontext"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	v1 := flag.Bool("v1", false, "parse using the encoding/json v1 API")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: match [-v1] <string> [file]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}
	target := flag.Arg(0)

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

	count := countV2
	if *v1 {
		count = countV1
	}
	n, err := count(r, target)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(n)
}

// countV1 counts JSON strings (both object member names and string values)
// equal to target using the encoding/json v1 token API. Each string token is
// materialized as a Go string.
func countV1(r io.Reader, target string) (int, error) {
	dec := jsonv1.NewDecoder(r)
	n := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
		if s, ok := tok.(string); ok && s == target {
			n++
		}
	}
}

// countV2 counts JSON strings (both object member names and string values)
// equal to target using jsontext's token API. It materializes each string
// token as a Go string via Token.String(); because the resulting string is
// used only in a comparison and does not escape, the compiler avoids
// allocating it.
func countV2(r io.Reader, target string) (int, error) {
	dec := jsontext.NewDecoder(r)
	n := 0
	for {
		tok, err := dec.ReadToken()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
		if tok.Kind() == '"' && tok.String() == target {
			n++
		}
	}
}

// findV2 returns a JSON Pointer (RFC 6901) for each JSON string (object member
// name or string value) equal to target, using jsontext's token API.
func findV2(r io.Reader, target string) ([]string, error) {
	dec := jsontext.NewDecoder(r)
	var ptrs []string
	for {
		tok, err := dec.ReadToken()
		if err == io.EOF {
			return ptrs, nil
		}
		if err != nil {
			return ptrs, err
		}
		if tok.Kind() == '"' && tok.String() == target {
			ptrs = append(ptrs, string(dec.StackPointer()))
		}
	}
}
