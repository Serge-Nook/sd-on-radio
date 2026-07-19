// Package gamepad provides optional game controller navigation. On unsupported
// platforms or when no controller is present it does nothing.
package gamepad

// Action is a high-level navigation command produced from controller input.
type Action int

const (
	// ActionUp moves the selection up.
	ActionUp Action = iota
	// ActionDown moves the selection down.
	ActionDown
	// ActionSelect activates the current selection.
	ActionSelect
	// ActionBack cancels or stops.
	ActionBack
	// ActionPlayPause toggles playback.
	ActionPlayPause
	// ActionVolumeUp raises the volume.
	ActionVolumeUp
	// ActionVolumeDown lowers the volume.
	ActionVolumeDown
)

// Handler receives controller actions. It is called from a background
// goroutine.
type Handler func(Action)
