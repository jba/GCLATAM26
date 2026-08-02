package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
)

// makeInput builds a moderately large JSON document with a mix of string
// values (some matching, some not), numbers, nested objects, and arrays.
func makeInput(n int) string {
	var b strings.Builder
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b,
			`{"name":"Alice%d","city":"Boston","tags":["apple","banana"],`+
				`"nested":{"place":"Austin","n":%d,"ok":true},"other":"Zeta"}`,
			i, i)
	}
	b.WriteByte(']')
	return b.String()
}

func benchmarkRun(b *testing.B, run func(io.Writer, io.Reader, *regexp.Regexp) error) {
	input := makeInput(1000)
	re := regexp.MustCompile("^A")
	b.SetBytes(int64(len(input)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := run(io.Discard, strings.NewReader(input), re); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRunV1(b *testing.B)      { benchmarkRun(b, runV1) }
func BenchmarkRunV2(b *testing.B)      { benchmarkRun(b, runV2) }
func BenchmarkRunV2Bytes(b *testing.B) { benchmarkRun(b, runV2Bytes) }
