//go:build goexperiment.simd && !amd64

package main

// ArchCountSpaces has no architecture-specific implementation off amd64,
// so it falls back to the portable version.
func ArchCountSpaces(b []byte) int { return CountSpaces(b) }
