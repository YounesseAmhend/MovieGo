package moviego

// Clip is the shared interface for visual clip types (Video and ImageClip).
// Audio intentionally does not implement this interface since it has no visual dimensions.
type Clip interface {
	GetWidth() uint64
	GetHeight() uint64
	GetDuration() float64
	GetPosition() Position
}

// Compile-time interface satisfaction checks.
var _ Clip = (*Video)(nil)
