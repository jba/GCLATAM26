//go:build goexperiment.simd && amd64

package main

// This file demonstrates what happens when you ask the simd/archsimd package
// for a SMALLER vector register than the hardware actually supports.
//
// Short answer: nothing bad. It just works.
//
// A 128-bit (SSE/Float32x4) or 256-bit (AVX2/Float32x8) instruction is a
// strict subset of what an AVX-512 machine can execute. The narrower op runs
// natively on the corresponding XMM/YMM register; the extra width the CPU has
// is simply left unused. You process fewer lanes per iteration, so you may
// leave a little performance on the table, but the result is identical.
//
// This is the exact opposite of the avx512fault demo, where asking for a
// LARGER register than the hardware supports (a 512-bit op on an AVX2-only
// CPU) faults with SIGILL, because archsimd intrinsics are unconditional and
// there is no automatic narrowing.

import (
	"math"
	"testing"

	"simd/archsimd"
)

// dot128 computes a dot product using 128-bit registers (4 float32/iter).
// On an AVX-512 machine this deliberately under-uses the hardware.
func dot128(a, b []float32) float32 {
	var acc archsimd.Float32x4
	w := acc.Len() // 4
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
	// Horizontal sum of 4 lanes.
	return acc.GetElem(0) + acc.GetElem(1) + acc.GetElem(2) + acc.GetElem(3)
}

// dot256 computes a dot product using 256-bit registers (8 float32/iter).
func dot256(a, b []float32) float32 {
	var acc archsimd.Float32x8
	w := acc.Len() // 8
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
	// Fold the two 128-bit halves together, then sum the 4 remaining lanes.
	lo := acc.GetLo().Add(acc.GetHi())
	return lo.GetElem(0) + lo.GetElem(1) + lo.GetElem(2) + lo.GetElem(3)
}

// TestSmallerRegisterThanHardware runs 128- and 256-bit dot products and
// checks they agree with the scalar reference. On CI/dev machines with
// AVX-512 (see the log line), this proves that requesting a narrower register
// than the hardware provides is perfectly safe: no fault, correct result.
func TestSmallerRegisterThanHardware(t *testing.T) {
	t.Logf("hardware: AVX=%v AVX2=%v AVX512=%v",
		archsimd.X86.AVX(), archsimd.X86.AVX2(), archsimd.X86.AVX512())

	if !archsimd.X86.AVX512() {
		t.Skip("this machine lacks AVX-512, so 128/256-bit is not 'smaller than hardware'")
	}

	narrow := map[string]func(a, b []float32) float32{
		"dot128 (SSE, 128-bit)":  dot128,
		"dot256 (AVX2, 256-bit)": dot256,
	}
	for name, fn := range narrow {
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
				if d := math.Abs(float64(got - want)); d > 1e-3*(1+math.Abs(float64(want))) {
					t.Fatalf("n=%d: got %v want %v", n, got, want)
				}
			}
		})
	}
}
