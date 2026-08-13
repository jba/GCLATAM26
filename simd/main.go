//go:build goexperiment.simd && amd64

package main

import (
	"fmt"
	"simd/archsimd"
)

// AddSlicesVectorized adds elements of slice 'b' into slice 'a' in-place.
func AddSlicesVectorized(a, b []int64) {
	n := len(a)
	if n != len(b) {
		panic("slices must have the same length")
	}

	// Int64x4 is a 256-bit vector (4 elements * 64 bits)
	const vectorSize = 4

	// Process elements in blocks of 4
	loopEnd := n - (n % vectorSize)
	for i := 0; i < loopEnd; i += vectorSize {
		// Load 4 elements from each slice into 256-bit registers
		vecA := archsimd.LoadInt64x4Slice(a[i : i+vectorSize])
		vecB := archsimd.LoadInt64x4Slice(b[i : i+vectorSize])

		// Perform parallel addition across all 4 lanes
		resultVec := vecA.Add(vecB)

		// Store the results back to slice 'a'
		resultVec.StoreSlice(a[i : i+vectorSize])
	}

	// Clean up the scalar tail (remaining elements)
	for i := loopEnd; i < n; i++ {
		a[i] += b[i]
	}
}

func main() {
	srcA := []int64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	srcB := []int64{10, 20, 30, 40, 50, 60, 70, 80, 90}

	AddSlicesVectorized(srcA, srcB)

	fmt.Println("Result:", srcA)
	// Output: Result: [11 22 33 44 55 66 77 88 99]

	fmt.Println("\nX86 CPU features:")
	x := archsimd.X86
	features := []struct {
		name string
		val  bool
	}{
		{"AVX", x.AVX()},
		{"AVX2", x.AVX2()},
		{"AVX512", x.AVX512()},
		{"AVX512BITALG", x.AVX512BITALG()},
		{"AVX512GFNI", x.AVX512GFNI()},
		{"AVX512VAES", x.AVX512VAES()},
		{"AVX512VBMI", x.AVX512VBMI()},
		{"AVX512VBMI2", x.AVX512VBMI2()},
		{"AVX512VNNI", x.AVX512VNNI()},
		{"AVX512VPCLMULQDQ", x.AVX512VPCLMULQDQ()},
		{"AVX512VPOPCNTDQ", x.AVX512VPOPCNTDQ()},
		{"AVXAES", x.AVXAES()},
		{"AVXVNNI", x.AVXVNNI()},
		{"FMA", x.FMA()},
		{"SHA", x.SHA()},
		{"VAES", x.VAES()},
	}
	for _, f := range features {
		fmt.Printf("  %-18s %v\n", f.name, f.val)
	}
}
