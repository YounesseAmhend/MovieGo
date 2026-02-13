package moviego

// ============================================================================
// CompositeClip Types
// ============================================================================

// CompositeClipItem represents a single video clip in a composite with positioning and timing
type CompositeClipItem struct {
	video     *Video  // The video clip to composite
	x         int64   // X position (can be negative for centering)
	y         int64   // Y position (can be negative for centering)
	width     uint64  // Override width (0 = use video width)
	height    uint64  // Override height (0 = use video height)
	startTime float64 // When this clip starts in the composite (default: 0)
	duration  float64 // How long this clip appears (0 = full clip duration)
	opacity   float64 // Opacity (0.0-1.0, default: 1.0)
	layer     int     // Z-order/layer (higher = on top, default: 0)

	// Animations (optional, nil = no animation)
	positionAnim *PositionAnimParams
	rotationAnim *RotationAnimParams
	scaleAnim    *ScaleAnimParams
}

// CompositeClipParameters holds configuration for composite video processing
type CompositeClipParameters struct {
	OutputPath  string
	Width       uint64 // Canvas width (0 = auto from clips)
	Height      uint64 // Canvas height (0 = auto from clips)
	Fps         uint64 // Output FPS (0 = use first clip's FPS)
	BgColor     string // Background color (default: "black")
	Threads     uint16
	Codec       Codec
	Preset      preset
	Bitrate     string
	PixelFormat PixelFormat
}

// ============================================================================
// CompositeAudio Types (for future audio mixing)
// ============================================================================

// CompositeAudioItem represents a single audio clip in a composite
type CompositeAudioItem struct {
	audio     *Audio  // The audio to composite
	startTime float64 // When this audio starts
	duration  float64 // How long this audio plays
	volume    float64 // Volume level (0.0-1.0, default: 1.0)
}
