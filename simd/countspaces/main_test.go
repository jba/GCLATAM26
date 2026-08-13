//go:build goexperiment.simd && amd64

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestCountSpaces(t *testing.T) {
	// Exercise lengths across several 64-byte blocks plus partial tails.
	for n := 0; n <= 200; n++ {
		b := make([]byte, n)
		for i := range b {
			if i%3 == 0 {
				b[i] = ' '
			} else {
				b[i] = 'x'
			}
		}
		want := bytes.Count(b, []byte{' '})
		if got := CountSpaces(b); got != want {
			t.Fatalf("n=%d: got %d want %d", n, got, want)
		}
	}
}

func BenchmarkCountSpaces(b *testing.B) {
	data := []byte(strings.Repeat("the quick brown fox ", 500))
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CountSpaces(data)
	}
}
