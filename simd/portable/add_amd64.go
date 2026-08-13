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
		va := archsimd.LoadInt64x8Slice(a[i : i+w])
		vb := archsimd.LoadInt64x8Slice(b[i : i+w])
		va.Add(vb).StoreSlice(a[i : i+w])
	}
	if rem := a[i:]; len(rem) > 0 {
		va := archsimd.LoadInt64x8SlicePart(rem)
		vb := archsimd.LoadInt64x8SlicePart(b[i:])
		va.Add(vb).StoreSlicePart(rem)
	}
}
