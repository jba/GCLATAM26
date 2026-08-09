package main

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

// makeInput builds a JSON array of n records with a mix of string values,
// some equal to the search target "foo" and some not.
func makeInput(n int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b,
			`{"a":"foo","b":["foo","bar"],"nested":{"foo":"baz","k":%d},"other":"Zeta%d"}`,
			i, i)
	}
	b.WriteByte(']')
	return b.String()
}

func benchmarkCount(b *testing.B, count func(io.Reader, string) (int, error)) {
	input := makeInput(1000)
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := count(strings.NewReader(input), "foo"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCountV1(b *testing.B) { benchmarkCount(b, countV1) }
func BenchmarkCountV2(b *testing.B) { benchmarkCount(b, countV2) }

// TestSameCount verifies both implementations agree on escape-free input.
func TestSameCount(t *testing.T) {
	in := makeInput(200)
	n1, _ := countV1(strings.NewReader(in), "foo")
	n2, _ := countV2(strings.NewReader(in), "foo")
	if n1 != n2 {
		t.Fatalf("v1=%d v2=%d", n1, n2)
	}
}
