//go:build goexperiment.simd && amd64

// Command avx512fault demonstrates what happens when you use a 512-bit
// simd/archsimd operation on hardware that does not support AVX-512.
//
// The archsimd intrinsics are UNCONDITIONAL: a Float32x16 op compiles
// straight to an EVEX-encoded AVX-512 instruction with no runtime
// capability check. On a CPU without AVX-512, the first such instruction
// raises an undefined-opcode fault (#UD), which the Go runtime reports as
// SIGILL, killing the process. There is no automatic narrowing to smaller
// registers and no graceful fallback -- that is what archsimd.X86.AVX512()
// is for.
//
// On a machine WITH AVX-512 this program runs fine and prints a result.
// To see the crash, run it on a CPU without AVX-512, e.g. under qemu:
//
//	GOEXPERIMENT=simd go build -o demo .
//	qemu-x86_64-static -cpu Haswell ./demo   # Haswell = AVX2 only
//
// which fails with:
//
//	SIGILL: illegal instruction
//	...
//	instruction bytes: 0x62 0xf2 ...   # 0x62 is the AVX-512 EVEX prefix
//	simd/archsimd.BroadcastFloat32x16(...)
package main

import (
	"fmt"
	"simd/archsimd"
)

func main() {
	fmt.Printf("CPU reports AVX512 = %v\n", archsimd.X86.AVX512())
	fmt.Println("about to execute a 512-bit AVX-512 op (Float32x16.Add)...")

	a := archsimd.BroadcastFloat32x16(2) // uses a 512-bit ZMM register
	b := archsimd.BroadcastFloat32x16(3)
	c := a.Add(b)                        // VADDPS zmm -- AVX-512 instruction

	out := make([]float32, 16)
	c.Store(out)
	fmt.Println("result:", out[:4], "...")
	fmt.Println("survived: hardware supports AVX-512")
}
