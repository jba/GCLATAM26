//go:build amd64 && goexperiment.simd

package main

import "simd/archsimd"

// init wires the SIMD kernel into the portable dispatcher. This file (and
// therefore any reference to vector types) exists only on amd64.
func init() { simdAdd = addAVX512 }

func addAVX512(a, b []int64) {
	const w = 8 // Int64x8 == 512 bits
	i := 0
	for ; i+w <= len(a); i += w {
		va := archsimd.LoadInt64x8(a[i : i+w])
		vb := archsimd.LoadInt64x8(b[i : i+w])
		va.Add(vb).Store(a[i : i+w])
	}
	if rem := a[i:]; len(rem) > 0 {
		va, _ := archsimd.LoadInt64x8Part(rem)
		vb, _ := archsimd.LoadInt64x8Part(b[i:])
		va.Add(vb).StorePart(rem)
	}
}
