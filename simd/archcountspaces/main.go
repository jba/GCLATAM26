//go:build goexperiment.simd

// Command countspaces counts the number of ASCII space (0x20) bytes in a
// byte slice, two ways:
//
//   - ArchCountSpaces uses the amd64-specific simd/archsimd package
//     (AVX-512): compare 64 bytes/iteration, then popcount the mask.
//   - CountSpaces uses the portable, vector-size-agnostic simd package
//     new in Go 1.27, so it works on any supported architecture.
package main

import (
	"fmt"
	"runtime"
	"simd"
	"slices"
)

// CountSpaces returns the number of ' ' bytes in b using the portable
// simd package. The vector width is chosen by the hardware at runtime, so
// the same code compiles and runs on amd64, arm64, etc.
func CountSpaces(b []byte) int {
	spaces := simd.BroadcastUint8s(' ')
	width := spaces.Len()
	scratch := make([]int8, width)

	// The portable API has no mask-popcount, so we accumulate matches in a
	// per-lane int8 vector. Mask8s.ToInt8s yields -1 for true lanes, so
	// subtracting it adds 1 to each matching lane's counter. int8 counters
	// saturate at 127, so flush to the scalar total periodically.
	var acc simd.Int8s
	count, since := 0, 0
	flush := func() {
		acc.Store(scratch)
		for _, v := range scratch {
			count += int(v) // acc holds +1 per match
		}
		acc = simd.Int8s{}
		since = 0
	}

	// slices.Chunk yields width-sized chunks, with a possibly-short final
	// chunk that folds the scalar tail into the same loop.
	for chunk := range slices.Chunk(b, width) {
		var v simd.Uint8s
		if len(chunk) == width {
			v = simd.LoadUint8s(chunk)
		} else {
			// Short final chunk: zero-fill unused lanes (0x00 != ' ').
			v, _ = simd.LoadUint8sPart(chunk)
		}
		acc = acc.Sub(v.Equal(spaces).ToInt8s())
		if since++; since == 120 {
			flush()
		}
	}
	flush()
	return count
}

func main() {
	text := []byte("the quick brown fox jumps over the lazy dog, several times over")
	fmt.Printf("input:  %q\n", text)
	fmt.Printf("length: %d\n", len(text))
	fmt.Printf("arch:   %s\n", runtime.GOARCH)
	fmt.Printf("CountSpaces (portable): %d\n", CountSpaces(text))
	fmt.Printf("ArchCountSpaces (amd64):  %d\n", ArchCountSpaces(text))
}
