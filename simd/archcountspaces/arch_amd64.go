//go:build goexperiment.simd && amd64

package main

import (
	"math/bits"
	"simd/archsimd"
)

// ArchCountSpaces counts ' ' bytes using the amd64-specific simd/archsimd
// package (AVX-512): compare 64 bytes per iteration, then popcount the
// resulting comparison mask.
func ArchCountSpaces(b []byte) int {
	spaces := archsimd.BroadcastUint8x64(' ')
	count := 0
	i := 0
	for ; i+64 <= len(b); i += 64 {
		v := archsimd.LoadUint8x64(b[i:])
		count += bits.OnesCount64(v.Equal(spaces).ToBits())
	}
	if i < len(b) {
		// Part zero-fills unused lanes and returns the valid count n;
		// 0x00 != ' ', but restrict to n lanes to be explicit.
		v, n := archsimd.LoadUint8x64Part(b[i:])
		valid := uint64(0xffffffffffffffff) >> (64 - n)
		count += bits.OnesCount64(v.Equal(spaces).ToBits() & valid)
	}
	return count
}
