//go:build goexperiment.simd && amd64

package main

import (
	"testing"
)

func ref(a, b []int64) []int64 {
	out := make([]int64, len(a))
	for i := range a {
		out[i] = a[i] + b[i]
	}
	return out
}

func TestAddSlices(t *testing.T) {
	// Cover lengths spanning several vector widths and partial tails.
	for n := 0; n <= 40; n++ {
		a := make([]int64, n)
		b := make([]int64, n)
		for i := 0; i < n; i++ {
			a[i] = int64(i)
			b[i] = int64(i * 10)
		}
		want := ref(a, b)
		AddSlices(a, b)
		for i := 0; i < n; i++ {
			if a[i] != want[i] {
				t.Fatalf("n=%d idx=%d: got %d want %d", n, i, a[i], want[i])
			}
		}
	}
}

func BenchmarkAddSlices(b *testing.B) {
	x := make([]int64, 1024)
	y := make([]int64, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddSlices(x, y)
	}
}
