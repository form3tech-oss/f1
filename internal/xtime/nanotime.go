package xtime

import "unsafe"

// Make goimports import the unsafe package, which is required to be able
// to use //go:linkname
var _ = unsafe.Sizeof(0)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

// NanoTime returns the current time in nanoseconds from a monotonic clock.
// The time returned is based on some arbitrary platform-specific point in the
// past. The time returned is guaranteed to increase monotonically at a
// constant rate
//
// This can be use for performance critical code where getting the wall time
// can slow down execution.
//
// https://github.com/golang/go/issues/12914
func NanoTime() int64 {
	return nanotime()
}
