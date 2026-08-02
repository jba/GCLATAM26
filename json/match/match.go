// Command match reads JSON from stdin (or a file) and prints the number of
// JSON string values equal to the string given on the command line.
//
// Usage:
//
//	match [-v1] <string> [file]
//
// By default it uses encoding/json/jsontext and compares bytes, avoiding the
// allocation of converting each token to a Go string. With -v1 it uses the
// original encoding/json (v1) token API, which must materialize each string.
package main

import (
	"bytes"
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

// countV1 counts JSON string values equal to target using the encoding/json
// v1 token API. Each string token is materialized as a Go string.
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

// countV2 counts JSON string values equal to target using jsontext. It reads
// string tokens as raw jsontext.Values that alias the decoder's buffer and
// compares bytes, so no per-token string is allocated.
//
// Comparison is on the raw JSON bytes with the surrounding quotes stripped and
// without unescaping; for escape-free input this matches countV1.
func countV2(r io.Reader, target string) (int, error) {
	dec := jsontext.NewDecoder(r)
	tbytes := []byte(target)
	n := 0
	for {
		if dec.PeekKind() == '"' {
			val, err := dec.ReadValue()
			if err != nil {
				return n, err
			}
			// val is `"..."`; strip the surrounding quotes and compare bytes.
			if bytes.Equal(val[1:len(val)-1], tbytes) {
				n++
			}
			continue
		}
		_, err := dec.ReadToken()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
	}
}

// countV2Token counts JSON string values equal to target using jsontext's
// token API. Unlike countV2, it materializes each string token as a Go string
// via Token.String(), which allocates — mirroring the v1 approach but on the
// v2 decoder.
func countV2Token(r io.Reader, target string) (int, error) {
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
