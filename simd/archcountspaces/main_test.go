//go:build goexperiment.simd

package main

import (
	"bytes"
	"strings"
	"testing"
)

var impls = map[string]func([]byte) int{
	"CountSpaces":     CountSpaces,
	"ArchCountSpaces": ArchCountSpaces,
}

func TestCountSpaces(t *testing.T) {
	for name, fn := range impls {
		t.Run(name, func(t *testing.T) {
			// Lengths spanning several vector widths plus partial tails.
			for n := 0; n <= 300; n++ {
				b := make([]byte, n)
				for i := range b {
					if i%3 == 0 {
						b[i] = ' '
					} else {
						b[i] = 'x'
					}
				}
				want := bytes.Count(b, []byte{' '})
				if got := fn(b); got != want {
					t.Fatalf("n=%d: got %d want %d", n, got, want)
				}
			}
		})
	}
}

func BenchmarkCountSpaces(b *testing.B) {
	data := []byte(strings.Repeat("the quick brown fox ", 500))
	for name, fn := range impls {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				fn(data)
			}
		})
	}
}
