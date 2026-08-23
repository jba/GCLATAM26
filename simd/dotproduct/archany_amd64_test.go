//go:build goexperiment.simd && amd64

package main

import (
	"math"
	"testing"

	"simd/archsimd"
)

// TestArchDotAny checks the register-size-dispatching ArchDotAny against the
// scalar reference. It runs whatever path the current CPU selects (see the
// log line), and by construction never faults regardless of hardware.
func TestArchDotAny(t *testing.T) {
	t.Logf("dispatch inputs: AVX=%v AVX2=%v AVX512=%v",
		archsimd.X86.AVX(), archsimd.X86.AVX2(), archsimd.X86.AVX512())
	for n := 0; n <= 200; n++ {
		a := make([]float32, n)
		b := make([]float32, n)
		for i := 0; i < n; i++ {
			a[i] = float32(i%7) - 3
			b[i] = float32(i%5) - 2
		}
		want := boringDot(a, b)
		got := ArchDotAny(a, b)
		if d := math.Abs(float64(got - want)); d > 1e-3*(1+math.Abs(float64(want))) {
			t.Fatalf("n=%d: got %v want %v", n, got, want)
		}
	}
}
