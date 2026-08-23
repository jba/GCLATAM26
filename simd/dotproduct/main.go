//go:build goexperiment.simd

// Command dotproduct computes the dot product of two []float32 slices three
// ways:
//
//   - SeqDot is a plain sequential (scalar) loop, the reference version.
//   - ArchDot uses the amd64-specific simd/archsimd package (AVX-512,
//     16 floats/iteration) with a fused multiply-add and a pairwise-add
//     horizontal reduction.
//   - Dot uses the portable, vector-size-agnostic simd package new in
//     Go 1.27, so it runs on any supported architecture.
package main

import (
	"fmt"
	"runtime"
	"simd"
)

// Dot returns the dot product of a and b using the portable simd package.
// The vector width is chosen by the hardware at runtime.
func Dot(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	var acc simd.Float32s // running per-lane sums of a[i]*b[i]
	width := acc.Len()    // vector width chosen by the hardware at runtime
	i := 0
	for ; i+width <= len(a); i += width {
		va := simd.LoadFloat32s(a[i:])
		vb := simd.LoadFloat32s(b[i:])
		acc = va.MulAdd(vb, acc) // acc += va*vb
	}

	// Tail: zero-filled lanes multiply to zero, so they don't affect acc.
	if i < len(a) {
		va, _ := simd.LoadFloat32sPart(a[i:])
		vb, _ := simd.LoadFloat32sPart(b[i:])
		acc = va.MulAdd(vb, acc)
	}

	// The portable API has no horizontal reduction; store and sum lanes.
	lanes := make([]float32, width)
	acc.Store(lanes)
	var sum float32
	for _, v := range lanes {
		sum += v
	}
	return sum
}

// SeqDot returns the dot product of a and b with a plain sequential loop.
// It serves as the scalar reference for the SIMD versions.
func SeqDot(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func main() {
	a := []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	b := []float32{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	fmt.Printf("a: %v\nb: %v\n", a, b)
	fmt.Printf("arch: %s\n", runtime.GOARCH)
	fmt.Printf("SeqDot (scalar):  %g\n", SeqDot(a, b))
	fmt.Printf("Dot (portable):   %g\n", Dot(a, b))
	fmt.Printf("ArchDot (amd64):  %g\n", ArchDot(a, b))
}
