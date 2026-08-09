// Command match reads JSON from stdin (or a file) and prints the number of
// JSON string values equal to the string given on the command line.
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

// countV1 counts JSON string values (not object member names) equal to target
// using the encoding/json v1 token API. Each string token is materialized as a
// Go string.
func countV1(r io.Reader, target string) (int, error) {
	dec := jsonv1.NewDecoder(r)
	var t valueTracker
	n := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
		var kind byte
		var s string
		switch v := tok.(type) {
		case jsonv1.Delim:
			kind = byte(v)
		case string:
			kind, s = '"', v
		default:
			kind = '0' // number, bool, or null
		}
		if t.step(kind) && s == target {
			n++
		}
	}
}

// countV2 counts JSON string values (not object member names) equal to target
// using jsontext's token API. It materializes each string token as a Go string
// via Token.String(); because the resulting string is used only in a
// comparison and does not escape, the compiler avoids allocating it.
func countV2(r io.Reader, target string) (int, error) {
	dec := jsontext.NewDecoder(r)
	var t valueTracker
	n := 0
	for {
		tok, err := dec.ReadToken()
		if err == io.EOF {
			return n, nil
		}
		if err != nil {
			return n, err
		}
		if t.step(byte(tok.Kind())) && tok.String() == target {
			n++
		}
	}
}

// findV2 returns a JSON Pointer (RFC 6901) for each JSON string value (not
// object member name) equal to target, using jsontext's token API.
func findV2(r io.Reader, target string) ([]string, error) {
	dec := jsontext.NewDecoder(r)
	var t valueTracker
	var ptrs []string
	for {
		tok, err := dec.ReadToken()
		if err == io.EOF {
			return ptrs, nil
		}
		if err != nil {
			return ptrs, err
		}
		if t.step(byte(tok.Kind())) && tok.String() == target {
			ptrs = append(ptrs, string(dec.StackPointer()))
		}
	}
}

// valueTracker tracks nesting so that a string token can be classified as an
// object member name or a value. In a JSON object, tokens alternate between
// member names and values; in an array, every element is a value.
type valueTracker struct {
	stack []byte // container kinds: '{' or '['
	name  []bool // per container: is the next object slot a name?
}

// step processes one token of the given kind (as reported by
// jsontext.Kind or derived from a v1 token) and reports whether that token is
// a string in value position (a value, not an object member name).
func (t *valueTracker) step(kind byte) (isStringValue bool) {
	expectName := len(t.stack) > 0 && t.stack[len(t.stack)-1] == '{' && t.name[len(t.name)-1]
	switch kind {
	case '{', '[':
		t.consumeValue() // the container is itself a value in its parent
		t.stack = append(t.stack, kind)
		t.name = append(t.name, kind == '{')
	case '}', ']':
		t.stack = t.stack[:len(t.stack)-1]
		t.name = t.name[:len(t.name)-1]
	case '"':
		if expectName {
			t.name[len(t.name)-1] = false
		} else {
			t.consumeValue()
			return true
		}
	default: // number, bool, null: always a value
		t.consumeValue()
	}
	return false
}

// consumeValue records that a value slot was filled in the current container.
// In an object, the next slot then expects a member name.
func (t *valueTracker) consumeValue() {
	if n := len(t.stack); n > 0 && t.stack[n-1] == '{' {
		t.name[n-1] = true
	}
}
