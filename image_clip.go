package moviego

// ImageClip represents a still image used as a clip in compositions/timelines.
// It carries only image-relevant fields — no audio, codec, fps, or video filter chains.
type ImageClip struct {
	filename         string
	width            uint64
	height           uint64
	duration         float64
	position         Position
	animatedPosition *AnimatedPosition
	animatedOpacity  *Animation
}

// Compile-time interface satisfaction check.
var _ Clip = (*ImageClip)(nil)

// NewImageClip creates a new ImageClip with the given filename, dimensions, and duration.
func NewImageClip(filename string, width, height uint64, duration float64) *ImageClip {
	return &ImageClip{
		filename: filename,
		width:    width,
		height:   height,
		duration: duration,
	}
}

// ============================================================================
// Clip Interface Implementation
// ============================================================================

// GetWidth returns the image width.
func (ic *ImageClip) GetWidth() uint64 {
	return ic.width
}

// GetHeight returns the image height.
func (ic *ImageClip) GetHeight() uint64 {
	return ic.height
}

// GetDuration returns how long the image should be displayed.
func (ic *ImageClip) GetDuration() float64 {
	return ic.duration
}

// GetPosition returns the overlay position.
// Returns center position if none was explicitly set.
func (ic *ImageClip) GetPosition() Position {
	if ic.position.X == "" && ic.position.Y == "" {
		return CenterPosition()
	}
	return ic.position
}

// ============================================================================
// ImageClip-specific Getters and Setters
// ============================================================================

// GetFilename returns the image filename.
func (ic *ImageClip) GetFilename() string {
	return ic.filename
}

// Duration sets how long the image should be displayed.
func (ic *ImageClip) Duration(d float64) *ImageClip {
	ic.duration = d
	return ic
}

// SetPosition sets the overlay position used by CompositeClip.
func (ic *ImageClip) SetPosition(position Position) *ImageClip {
	ic.position = position
	return ic
}

// SetAnimatedPosition sets the overlay position animation for CompositeClip.
func (ic *ImageClip) SetAnimatedPosition(ap AnimatedPosition) *ImageClip {
	ic.animatedPosition = &ap
	return ic
}

// SetAnimatedOpacity sets the overlay opacity animation for CompositeClip.
func (ic *ImageClip) SetAnimatedOpacity(a Animation) *ImageClip {
	ic.animatedOpacity = &a
	return ic
}
