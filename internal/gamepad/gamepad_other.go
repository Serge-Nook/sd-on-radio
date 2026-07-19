//go:build !linux

package gamepad

// Reader is a no-op controller reader on non-Linux platforms.
type Reader struct{}

// Start returns nil because controller support is only implemented on Linux.
func Start(handler Handler) *Reader { return nil }

// Stop does nothing.
func (r *Reader) Stop() {}
