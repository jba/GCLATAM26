//go:build goexperiment.simd && amd64

package main

import (
	"fmt"
	"simd/archsimd"
)

// AddSlices adds elements of slice 'b' into slice 'a' in place (a[i] += b[i]).
//
// It dispatches to the widest SIMD implementation the CPU supports at
// runtime, falling back to a plain scalar loop when no vector unit is
// available (e.g. AVX is missing).
func AddSlices(a, b []int64) {
	if len(a) != len(b) {
		panic("slices must have the same length")
	}
	switch {
	case archsimd.X86.AVX512():
		addSlicesAVX512(a, b)
	case archsimd.X86.AVX2():
		addSlicesAVX2(a, b)
	default:
		addSlicesScalar(a, b)
	}
}

// addSlicesAVX512 processes 8 int64s (512 bits) per iteration and uses a
// masked load/store to handle the final partial block, so there is no
// separate scalar tail loop.
func addSlicesAVX512(a, b []int64) {
	const w = 8 // Int64x8 == 512 bits
	i := 0
	for ; i+w <= len(a); i += w {
		va := archsimd.LoadInt64x8Slice(a[i : i+w])
		vb := archsimd.LoadInt64x8Slice(b[i : i+w])
		va.Add(vb).StoreSlice(a[i : i+w])
	}
	if rem := a[i:]; len(rem) > 0 {
		// Masked load/store: only the len(rem) valid lanes are touched.
		va := archsimd.LoadInt64x8SlicePart(rem)
		vb := archsimd.LoadInt64x8SlicePart(b[i:])
		va.Add(vb).StoreSlicePart(rem)
	}
}

// addSlicesAVX2 processes 4 int64s (256 bits) per iteration and uses a
// masked load/store for the tail.
func addSlicesAVX2(a, b []int64) {
	const w = 4 // Int64x4 == 256 bits
	i := 0
	for ; i+w <= len(a); i += w {
		va := archsimd.LoadInt64x4Slice(a[i : i+w])
		vb := archsimd.LoadInt64x4Slice(b[i : i+w])
		va.Add(vb).StoreSlice(a[i : i+w])
	}
	if rem := a[i:]; len(rem) > 0 {
		va := archsimd.LoadInt64x4SlicePart(rem)
		vb := archsimd.LoadInt64x4SlicePart(b[i:])
		va.Add(vb).StoreSlicePart(rem)
	}
}

func addSlicesScalar(a, b []int64) {
	for i := range a {
		a[i] += b[i]
	}
}

func main() {
	srcA := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	srcB := []int64{10, 20, 30, 40, 50, 60, 70, 80, 90}

	AddSlices(srcA, srcB)
	fmt.Println("Result:", srcA)
	// Output: Result: [11 22 33 44 55 66 77 88 99]

	fmt.Print("\nDispatch: ")
	switch {
	case archsimd.X86.AVX512():
		fmt.Println("AVX-512 (8 int64/iter)")
	case archsimd.X86.AVX2():
		fmt.Println("AVX2 (4 int64/iter)")
	default:
		fmt.Println("scalar")
	}
}
