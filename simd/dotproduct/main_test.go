//go:build goexperiment.simd

package main

import (
	"math"
	"testing"
)

func boringDot(a, b []float32) float32 {
	var s float32
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

var impls = map[string]func(a, b []float32) float32{
	"Dot":     Dot,
	"ArchDot": ArchDot,
}

func TestDot(t *testing.T) {
	for name, fn := range impls {
		t.Run(name, func(t *testing.T) {
			for n := 0; n <= 200; n++ {
				a := make([]float32, n)
				b := make([]float32, n)
				for i := 0; i < n; i++ {
					a[i] = float32(i%7) - 3
					b[i] = float32(i%5) - 2
				}
				want := boringDot(a, b)
				got := fn(a, b)
				// Allow tiny FP reassociation error.
				if d := math.Abs(float64(got - want)); d > 1e-3*(1+math.Abs(float64(want))) {
					t.Fatalf("n=%d: got %v want %v", n, got, want)
				}
			}
		})
	}
}

func BenchmarkDot(b *testing.B) {
	const n = 4096
	x := make([]float32, n)
	y := make([]float32, n)
	for i := range x {
		x[i] = float32(i) * 0.5
		y[i] = float32(n - i)
	}
	for name, fn := range impls {
		b.Run(name, func(b *testing.B) {
			b.SetBytes(int64(n * 4 * 2))
			for i := 0; i < b.N; i++ {
				fn(x, y)
			}
		})
	}
}
