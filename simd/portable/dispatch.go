//go:build goexperiment.simd

// Command portable demonstrates using archsimd.X86 feature checks in code
// that compiles on every GOARCH. The SIMD kernel is amd64-only, reached
// through a function variable, so this package cross-compiles to arm64
// (and others) with an automatic scalar fallback.
package main

import (
	"fmt"
	"runtime"
	"simd/archsimd"
)

// simdAdd is wired up only by the amd64 build (see add_amd64.go); it stays
// nil on other architectures.
var simdAdd func(a, b []int64)

// Add computes a[i] += b[i]. The dispatch logic is fully portable: on
// non-amd64, AVX512() is false and simdAdd is nil, so it falls to scalar.
func Add(a, b []int64) {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	if archsimd.X86.AVX512() && simdAdd != nil {
		simdAdd(a, b)
		return
	}
	for i := range a {
		a[i] += b[i]
	}
}

func main() {
	a := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	b := []int64{10, 20, 30, 40, 50, 60, 70, 80, 90}
	Add(a, b)
	fmt.Printf("arch=%s AVX512=%v simdPath=%v\nResult: %v\n",
		runtime.GOARCH, archsimd.X86.AVX512(), simdAdd != nil, a)
}
