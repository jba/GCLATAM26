//go:build goexperiment.simd && amd64

// Command countspaces counts the number of ASCII space (0x20) bytes in a
// byte slice using AVX-512: 64 bytes are compared per iteration, and the
// resulting comparison mask is popcounted.
package main

import (
	"fmt"
	"math/bits"
	"simd/archsimd"
)

// CountSpaces returns the number of ' ' bytes in b.
func CountSpaces(b []byte) int {
	spaces := archsimd.BroadcastUint8x64(' ')
	count := 0
	i := 0
	for ; i+64 <= len(b); i += 64 {
		v := archsimd.LoadUint8x64Slice(b[i : i+64])
		mask := v.Equal(spaces)
		count += bits.OnesCount64(mask.ToBits())
	}
	if rem := b[i:]; len(rem) > 0 {
		// SlicePart zero-fills unused lanes; 0x00 != ' ', so they don't
		// affect the count.
		v := archsimd.LoadUint8x64SlicePart(rem)
		mask := v.Equal(spaces)
		// Restrict to the valid lanes just to be explicit.
		valid := uint64(0xffffffffffffffff) >> (64 - len(rem))
		count += bits.OnesCount64(mask.ToBits() & valid)
	}
	return count
}

func main() {
	text := []byte("the quick brown fox jumps over the lazy dog, several times over")
	fmt.Printf("input:  %q\n", text)
	fmt.Printf("length: %d\n", len(text))
	fmt.Printf("spaces: %d\n", CountSpaces(text))
}
