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
	w := acc.Len() // lanes per iteration (16 for Float32x16)
	i := 0
	for ; i+w <= len(a); i += w {
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

// ArchDotAny returns the dot product of a and b using the widest SIMD
// register the CPU supports at runtime, switching on the register size:
// 512-bit (AVX-512), then 256-bit (AVX2), then 128-bit (AVX), and finally a
// scalar loop. Unlike ArchDot -- which unconditionally emits 512-bit AVX-512
// instructions and faults with SIGILL on hardware that lacks them -- this
// version runs correctly on any amd64.
func ArchDotAny(a, b []float32) float32 {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	switch {
	case archsimd.X86.AVX512():
		return archDot512(a, b) // 16 float32/iter
	case archsimd.X86.AVX2():
		return archDot256(a, b) // 8 float32/iter
	case archsimd.X86.AVX():
		return archDot128(a, b) // 4 float32/iter
	default:
		return SeqDot(a, b) // no vector unit: scalar fallback
	}
}

// archDot512 accumulates in a 512-bit Float32x16 (AVX-512).
func archDot512(a, b []float32) float32 {
	var acc archsimd.Float32x16
	w := acc.Len()
	i := 0
	for ; i+w <= len(a); i += w {
		va := archsimd.LoadFloat32x16(a[i:])
		vb := archsimd.LoadFloat32x16(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	if i < len(a) {
		va, _ := archsimd.LoadFloat32x16Part(a[i:])
		vb, _ := archsimd.LoadFloat32x16Part(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	return horizontalSum16(acc)
}

// archDot256 accumulates in a 256-bit Float32x8 (AVX2).
func archDot256(a, b []float32) float32 {
	var acc archsimd.Float32x8
	w := acc.Len()
	i := 0
	for ; i+w <= len(a); i += w {
		va := archsimd.LoadFloat32x8(a[i:])
		vb := archsimd.LoadFloat32x8(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	if i < len(a) {
		va, _ := archsimd.LoadFloat32x8Part(a[i:])
		vb, _ := archsimd.LoadFloat32x8Part(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	return horizontalSum8(acc)
}

// archDot128 accumulates in a 128-bit Float32x4 (SSE/AVX).
func archDot128(a, b []float32) float32 {
	var acc archsimd.Float32x4
	w := acc.Len()
	i := 0
	for ; i+w <= len(a); i += w {
		va := archsimd.LoadFloat32x4(a[i:])
		vb := archsimd.LoadFloat32x4(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	if i < len(a) {
		va, _ := archsimd.LoadFloat32x4Part(a[i:])
		vb, _ := archsimd.LoadFloat32x4Part(b[i:])
		acc = va.MulAdd(vb, acc)
	}
	return horizontalSum4(acc)
}

// horizontalSum8 folds the high 128 bits of a Float32x8 into the low half,
// then sums the remaining 4 lanes.
func horizontalSum8(v archsimd.Float32x8) float32 {
	return horizontalSum4(v.GetLo().Add(v.GetHi()))
}

// horizontalSum4 sums the 4 lanes of a Float32x4.
func horizontalSum4(v archsimd.Float32x4) float32 {
	return v.GetElem(0) + v.GetElem(1) + v.GetElem(2) + v.GetElem(3)
}
