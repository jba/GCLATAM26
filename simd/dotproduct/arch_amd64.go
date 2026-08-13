//go:build goexperiment.simd && amd64

package main

import "simd/archsimd"

// ArchDot returns the dot product of a and b using the amd64-specific
// simd/archsimd package: AVX-512 processes 16 float32s per iteration with
// a fused multiply-add.
func ArchDot(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	var acc archsimd.Float32x16
	i := 0
	for ; i+16 <= len(a); i += 16 {
		va := archsimd.LoadFloat32x16(a[i:])
		vb := archsimd.LoadFloat32x16(b[i:])
		acc = va.MulAdd(vb, acc) // acc += va*vb
	}
	if i < len(a) {
		// Part zero-fills unused lanes; 0*0 == 0, so no effect on the sum.
		va, _ := archsimd.LoadFloat32x16Part(a[i:])
		vb, _ := archsimd.LoadFloat32x16Part(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	return horizontalSum16(acc)
}

// horizontalSum16 sums the 16 lanes of a Float32x16. It folds the high
// 256 bits into the low, then uses grouped pairwise adds to reduce the
// two 128-bit halves down to a single element each.
func horizontalSum16(v archsimd.Float32x16) float32 {
	w := v.GetLo().Add(v.GetHi()) // Float32x8: 8 partial sums
	w = w.ConcatAddPairsGrouped(w)
	w = w.ConcatAddPairsGrouped(w)
	return w.GetLo().GetElem(0) + w.GetHi().GetElem(0)
}
