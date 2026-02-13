package moviego

// ============================================================================
// ImageClip Type
// ============================================================================

// ImageClip represents an image overlay or standalone image video clip
type ImageClip struct {
	imagePath       string
	width           uint64  // Override width (0 = use image width)
	height          uint64  // Override height (0 = use image height)
	duration        float64
	fps             uint64
	x               int     // X position for overlay mode
	y               int     // Y position for overlay mode
	startTime       float64 // When overlay appears
	overlayDuration float64 // How long overlay appears (0 = full video duration)
	opacity         float64 // Opacity (0.0-1.0)
	layer           int     // Z-order for overlays
	isOverlay       bool    // Whether used as overlay vs standalone

	// Animations (optional, nil = no animation)
	positionAnim *PositionAnimParams
	rotationAnim *RotationAnimParams
	scaleAnim    *ScaleAnimParams
}

// ============================================================================
// ColorClip Type
// ============================================================================

// ColorClip represents a solid color overlay or standalone color video clip
type ColorClip struct {
	color           string
	width           uint64
	height          uint64
	duration        float64
	fps             uint64
	x               int     // X position for overlay mode
	y               int     // Y position for overlay mode
	startTime       float64 // When overlay appears
	overlayDuration float64 // How long overlay appears (0 = full video duration)
	opacity         float64 // Opacity (0.0-1.0)
	layer           int     // Z-order for overlays
	isOverlay       bool    // Whether used as overlay vs standalone

	// Animations (optional, nil = no animation)
	positionAnim *PositionAnimParams
	rotationAnim *RotationAnimParams
	scaleAnim    *ScaleAnimParams
}
