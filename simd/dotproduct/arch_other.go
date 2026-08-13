//go:build goexperiment.simd && !amd64

package main

// ArchDot has no architecture-specific implementation off amd64, so it
// falls back to the portable version.
func ArchDot(a, b []float32) float32 { return Dot(a, b) }
