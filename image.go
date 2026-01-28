package moviego

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ============================================================================
// ImageClip Type
// ============================================================================

// ImageClip represents an image overlay or standalone image video clip
type ImageClip struct {
	imagePath      string
	width          uint64 // Override width (0 = use image width)
	height         uint64 // Override height (0 = use image height)
	duration       float64
	fps            uint64
	x              int     // X position for overlay mode
	y              int     // Y position for overlay mode
	startTime      float64 // When overlay appears
	overlayDuration float64 // How long overlay appears (0 = full video duration)
	opacity        float64 // Opacity (0.0-1.0)
	layer          int     // Z-order for overlays
	isOverlay      bool    // Whether used as overlay vs standalone

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
	color          string
	width          uint64
	height         uint64
	duration       float64
	fps            uint64
	x              int     // X position for overlay mode
	y              int     // Y position for overlay mode
	startTime      float64 // When overlay appears
	overlayDuration float64 // How long overlay appears (0 = full video duration)
	opacity        float64 // Opacity (0.0-1.0)
	layer          int     // Z-order for overlays
	isOverlay      bool    // Whether used as overlay vs standalone

	// Animations (optional, nil = no animation)
	positionAnim *PositionAnimParams
	rotationAnim *RotationAnimParams
	scaleAnim    *ScaleAnimParams
}

// ============================================================================
// ImageClip Constructor
// ============================================================================

// NewImageClip creates a new ImageClip from an image file path
func NewImageClip(imagePath string) *ImageClip {
	// Detect image dimensions
	width, height := getImageDimensions(imagePath)
	
	return &ImageClip{
		imagePath:      imagePath,
		width:          width,
		height:         height,
		duration:       0.0, // 0 means full video duration for overlays
		fps:            30,
		x:              0,
		y:              0,
		startTime:      0.0,
		overlayDuration: 0.0,
		opacity:        1.0,
		layer:          0,
		isOverlay:      false,
	}
}

// ============================================================================
// ColorClip Constructor
// ============================================================================

// NewColorClip creates a new ColorClip with specified color and dimensions
func NewColorClip(color string, width, height uint64) *ColorClip {
	return &ColorClip{
		color:          color,
		width:          width,
		height:         height,
		duration:       0.0, // 0 means full video duration for overlays
		fps:            30,
		x:              0,
		y:              0,
		startTime:      0.0,
		overlayDuration: 0.0,
		opacity:        1.0,
		layer:          0,
		isOverlay:      false,
	}
}

// ============================================================================
// ImageClip Builder Methods
// ============================================================================

// SetDuration sets the clip duration
func (ic *ImageClip) SetDuration(duration float64) *ImageClip {
	ic.duration = duration
	return ic
}

// SetFps sets the frames per second
func (ic *ImageClip) SetFps(fps uint64) *ImageClip {
	ic.fps = fps
	return ic
}

// SetSize sets the override size (0 = use image dimensions)
func (ic *ImageClip) SetSize(width, height uint64) *ImageClip {
	ic.width = width
	ic.height = height
	return ic
}

// SetPosition sets the position for overlay mode
func (ic *ImageClip) SetPosition(x, y int) *ImageClip {
	ic.x = x
	ic.y = y
	ic.isOverlay = true
	return ic
}

// SetStartTime sets when the overlay appears
func (ic *ImageClip) SetStartTime(startTime float64) *ImageClip {
	ic.startTime = startTime
	ic.isOverlay = true
	return ic
}

// SetOverlayDuration sets how long the overlay appears (0 = full video duration)
func (ic *ImageClip) SetOverlayDuration(duration float64) *ImageClip {
	ic.overlayDuration = duration
	ic.isOverlay = true
	return ic
}

// SetOpacity sets the opacity (0.0-1.0)
func (ic *ImageClip) SetOpacity(opacity float64) *ImageClip {
	if opacity < 0.0 {
		opacity = 0.0
	}
	if opacity > 1.0 {
		opacity = 1.0
	}
	ic.opacity = opacity
	ic.isOverlay = true
	return ic
}

// SetLayer sets the z-order/layer (higher = on top)
func (ic *ImageClip) SetLayer(layer int) *ImageClip {
	ic.layer = layer
	ic.isOverlay = true
	return ic
}

// SetPositionAnim sets position animation parameters
func (ic *ImageClip) SetPositionAnim(params PositionAnimParams) *ImageClip {
	ic.positionAnim = &params
	return ic
}

// SetRotationAnim sets rotation animation parameters
func (ic *ImageClip) SetRotationAnim(params RotationAnimParams) *ImageClip {
	ic.rotationAnim = &params
	return ic
}

// SetScaleAnim sets scale animation parameters
func (ic *ImageClip) SetScaleAnim(params ScaleAnimParams) *ImageClip {
	ic.scaleAnim = &params
	return ic
}

// ToOverlay marks the clip as an overlay
func (ic *ImageClip) ToOverlay() *ImageClip {
	ic.isOverlay = true
	return ic
}

// ============================================================================
// ColorClip Builder Methods
// ============================================================================

// SetDuration sets the clip duration
func (cc *ColorClip) SetDuration(duration float64) *ColorClip {
	cc.duration = duration
	return cc
}

// SetFps sets the frames per second
func (cc *ColorClip) SetFps(fps uint64) *ColorClip {
	cc.fps = fps
	return cc
}

// SetSize sets the clip size
func (cc *ColorClip) SetSize(width, height uint64) *ColorClip {
	cc.width = width
	cc.height = height
	return cc
}

// SetPosition sets the position for overlay mode
func (cc *ColorClip) SetPosition(x, y int) *ColorClip {
	cc.x = x
	cc.y = y
	cc.isOverlay = true
	return cc
}

// SetStartTime sets when the overlay appears
func (cc *ColorClip) SetStartTime(startTime float64) *ColorClip {
	cc.startTime = startTime
	cc.isOverlay = true
	return cc
}

// SetOverlayDuration sets how long the overlay appears (0 = full video duration)
func (cc *ColorClip) SetOverlayDuration(duration float64) *ColorClip {
	cc.overlayDuration = duration
	cc.isOverlay = true
	return cc
}

// SetOpacity sets the opacity (0.0-1.0)
func (cc *ColorClip) SetOpacity(opacity float64) *ColorClip {
	if opacity < 0.0 {
		opacity = 0.0
	}
	if opacity > 1.0 {
		opacity = 1.0
	}
	cc.opacity = opacity
	cc.isOverlay = true
	return cc
}

// SetLayer sets the z-order/layer (higher = on top)
func (cc *ColorClip) SetLayer(layer int) *ColorClip {
	cc.layer = layer
	cc.isOverlay = true
	return cc
}

// SetPositionAnim sets position animation parameters
func (cc *ColorClip) SetPositionAnim(params PositionAnimParams) *ColorClip {
	cc.positionAnim = &params
	return cc
}

// SetRotationAnim sets rotation animation parameters
func (cc *ColorClip) SetRotationAnim(params RotationAnimParams) *ColorClip {
	cc.rotationAnim = &params
	return cc
}

// SetScaleAnim sets scale animation parameters
func (cc *ColorClip) SetScaleAnim(params ScaleAnimParams) *ColorClip {
	cc.scaleAnim = &params
	return cc
}

// ToOverlay marks the clip as an overlay
func (cc *ColorClip) ToOverlay() *ColorClip {
	cc.isOverlay = true
	return cc
}

// ============================================================================
// ImageClip Getters
// ============================================================================

// GetImagePath returns the image file path
func (ic *ImageClip) GetImagePath() string {
	return ic.imagePath
}

// GetWidth returns the width
func (ic *ImageClip) GetWidth() uint64 {
	return ic.width
}

// GetHeight returns the height
func (ic *ImageClip) GetHeight() uint64 {
	return ic.height
}

// GetDuration returns the duration
func (ic *ImageClip) GetDuration() float64 {
	return ic.duration
}

// GetFps returns the frames per second
func (ic *ImageClip) GetFps() uint64 {
	return ic.fps
}

// GetLayer returns the layer/z-order
func (ic *ImageClip) GetLayer() int {
	return ic.layer
}

// GetStartTime returns when the overlay starts
func (ic *ImageClip) GetStartTime() float64 {
	return ic.startTime
}

// GetOverlayDuration returns how long the overlay appears
func (ic *ImageClip) GetOverlayDuration() float64 {
	return ic.overlayDuration
}

// GetOpacity returns the opacity
func (ic *ImageClip) GetOpacity() float64 {
	return ic.opacity
}

// IsOverlay returns whether this is used as an overlay
func (ic *ImageClip) IsOverlay() bool {
	return ic.isOverlay
}

// HasAnimation returns true if the image clip has any animation
func (ic *ImageClip) HasAnimation() bool {
	return ic.positionAnim != nil || ic.rotationAnim != nil || ic.scaleAnim != nil
}

// GetPositionAnim returns the position animation parameters
func (ic *ImageClip) GetPositionAnim() *PositionAnimParams {
	return ic.positionAnim
}

// GetRotationAnim returns the rotation animation parameters
func (ic *ImageClip) GetRotationAnim() *RotationAnimParams {
	return ic.rotationAnim
}

// GetScaleAnim returns the scale animation parameters
func (ic *ImageClip) GetScaleAnim() *ScaleAnimParams {
	return ic.scaleAnim
}

// ============================================================================
// ColorClip Getters
// ============================================================================

// GetColor returns the color
func (cc *ColorClip) GetColor() string {
	return cc.color
}

// GetWidth returns the width
func (cc *ColorClip) GetWidth() uint64 {
	return cc.width
}

// GetHeight returns the height
func (cc *ColorClip) GetHeight() uint64 {
	return cc.height
}

// GetDuration returns the duration
func (cc *ColorClip) GetDuration() float64 {
	return cc.duration
}

// GetFps returns the frames per second
func (cc *ColorClip) GetFps() uint64 {
	return cc.fps
}

// GetLayer returns the layer/z-order
func (cc *ColorClip) GetLayer() int {
	return cc.layer
}

// GetStartTime returns when the overlay starts
func (cc *ColorClip) GetStartTime() float64 {
	return cc.startTime
}

// GetOverlayDuration returns how long the overlay appears
func (cc *ColorClip) GetOverlayDuration() float64 {
	return cc.overlayDuration
}

// GetOpacity returns the opacity
func (cc *ColorClip) GetOpacity() float64 {
	return cc.opacity
}

// IsOverlay returns whether this is used as an overlay
func (cc *ColorClip) IsOverlay() bool {
	return cc.isOverlay
}

// HasAnimation returns true if the color clip has any animation
func (cc *ColorClip) HasAnimation() bool {
	return cc.positionAnim != nil || cc.rotationAnim != nil || cc.scaleAnim != nil
}

// GetPositionAnim returns the position animation parameters
func (cc *ColorClip) GetPositionAnim() *PositionAnimParams {
	return cc.positionAnim
}

// GetRotationAnim returns the rotation animation parameters
func (cc *ColorClip) GetRotationAnim() *RotationAnimParams {
	return cc.rotationAnim
}

// GetScaleAnim returns the scale animation parameters
func (cc *ColorClip) GetScaleAnim() *ScaleAnimParams {
	return cc.scaleAnim
}

// ============================================================================
// Conversion to Video
// ============================================================================

// ToVideo converts ImageClip to a standalone Video
func (ic *ImageClip) ToVideo() *Video {
	// Ensure we have valid dimensions
	width := ic.width
	height := ic.height
	if width == 0 || height == 0 {
		// Try to get dimensions from image
		width, height = getImageDimensions(ic.imagePath)
		if width == 0 || height == 0 {
			// Fallback to default
			width = 1920
			height = 1080
		}
	}

	// Ensure we have a valid duration
	duration := ic.duration
	if duration <= 0 {
		duration = 5.0 // Default duration
	}

	// Ensure we have valid FPS
	fps := ic.fps
	if fps == 0 {
		fps = 30
	}

	return &Video{
		filename:    ic.imagePath,
		width:       width,
		height:      height,
		fps:         fps,
		duration:    duration,
		frames:      uint64(float64(fps) * duration),
		codec:       CodecLibx264,
		preset:      Medium,
		pixelFormat: "yuv420p",
		isTemp:      false,
	}
}

// ToVideo converts ColorClip to a standalone Video
func (cc *ColorClip) ToVideo() *Video {
	// Ensure we have valid dimensions
	width := cc.width
	height := cc.height
	if width == 0 || height == 0 {
		width = 1920
		height = 1080
	}

	// Ensure we have a valid duration
	duration := cc.duration
	if duration <= 0 {
		duration = 5.0 // Default duration
	}

	// Ensure we have valid FPS
	fps := cc.fps
	if fps == 0 {
		fps = 30
	}

	// Create a copy of the color clip for standalone video
	// This preserves the color information for standalone generation
	standaloneClip := &ColorClip{
		color:          cc.color,
		width:          width,
		height:         height,
		duration:       duration,
		fps:            fps,
		x:              0,
		y:              0,
		startTime:      0.0,
		overlayDuration: 0.0,
		opacity:        1.0,
		layer:          0,
		isOverlay:      false,
	}

	return &Video{
		filename:    "", // Color clips don't have a file
		width:       width,
		height:      height,
		fps:         fps,
		duration:    duration,
		frames:      uint64(float64(fps) * duration),
		codec:       CodecLibx264,
		preset:      Medium,
		pixelFormat: "yuv420p",
		isTemp:      false,
		colorClips:  []*ColorClip{standaloneClip}, // Store color clip for standalone generation
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// getImageDimensions gets the width and height of an image file using ffprobe
func getImageDimensions(imagePath string) (uint64, uint64) {
	if imagePath == "" {
		return 0, 0
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return 0, 0
	}

	// Use ffprobe to get image dimensions
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "json", imagePath)
	output, err := cmd.Output()
	if err != nil {
		// If ffprobe fails, try alternative method
		return getImageDimensionsAlternative(imagePath)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return getImageDimensionsAlternative(imagePath)
	}

	if streams, ok := result["streams"].([]interface{}); ok && len(streams) > 0 {
		if streamMap, ok := streams[0].(map[string]interface{}); ok {
			var width, height uint64
			if w, ok := streamMap["width"].(float64); ok {
				width = uint64(w)
			}
			if h, ok := streamMap["height"].(float64); ok {
				height = uint64(h)
			}
			if width > 0 && height > 0 {
				return width, height
			}
		}
	}

	return getImageDimensionsAlternative(imagePath)
}

// getImageDimensionsAlternative tries to get dimensions using ffprobe with different options
func getImageDimensionsAlternative(imagePath string) (uint64, uint64) {
	// Try with format probe
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "stream=width,height", "-of", "json", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return 0, 0
	}

	if streams, ok := result["streams"].([]interface{}); ok && len(streams) > 0 {
		if streamMap, ok := streams[0].(map[string]interface{}); ok {
			var width, height uint64
			if w, ok := streamMap["width"].(float64); ok {
				width = uint64(w)
			} else if w, ok := streamMap["width"].(string); ok {
				if wInt, err := strconv.ParseUint(w, 10, 64); err == nil {
					width = wInt
				}
			}
			if h, ok := streamMap["height"].(float64); ok {
				height = uint64(h)
			} else if h, ok := streamMap["height"].(string); ok {
				if hInt, err := strconv.ParseUint(h, 10, 64); err == nil {
					height = hInt
				}
			}
			if width > 0 && height > 0 {
				return width, height
			}
		}
	}

	return 0, 0
}

// ============================================================================
// Sorting Helper Functions
// ============================================================================

// sortImageClipsByLayer sorts image clips by layer (lower layers first)
func sortImageClipsByLayer(items []*ImageClip) []*ImageClip {
	sorted := make([]*ImageClip, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].layer < sorted[j].layer
	})

	return sorted
}

// sortColorClipsByLayer sorts color clips by layer (lower layers first)
func sortColorClipsByLayer(items []*ColorClip) []*ColorClip {
	sorted := make([]*ColorClip, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].layer < sorted[j].layer
	})

	return sorted
}

// ============================================================================
// FFmpeg Filter Generation - Image Overlay
// ============================================================================

// buildImageOverlayFilterString generates FFmpeg filter string for image overlay
// Returns filter string that processes image input and overlays it on video
func buildImageOverlayFilterString(imageClip *ImageClip, videoWidth, videoHeight uint64, videoDuration float64) (string, error) {
	// Check if image file exists
	if _, err := os.Stat(imageClip.imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("image file not found: %s", imageClip.imagePath)
	}

	// Determine overlay duration
	overlayDuration := imageClip.overlayDuration
	if overlayDuration == 0 {
		overlayDuration = videoDuration - imageClip.startTime
		if overlayDuration < 0 {
			overlayDuration = 0
		}
	}

	// Build scale animation or static scale filter
	scaleFilter := ""
	if imageClip.scaleAnim != nil {
		// Use timeline time offset by startTime for clip-local animation
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", imageClip.startTime)
		scaleExpr := linearExpr(tOffsetExpr, imageClip.scaleAnim.Start, imageClip.scaleAnim.Duration,
			imageClip.scaleAnim.From, imageClip.scaleAnim.To)
		scaleFilter = buildScaleAnimationFilter(scaleExpr)
	} else if imageClip.width > 0 && imageClip.height > 0 {
		scaleFilter = fmt.Sprintf("scale=%d:%d", imageClip.width, imageClip.height)
	}

	// Build rotation animation if specified
	rotationFilter := ""
	if imageClip.rotationAnim != nil {
		// Convert degrees to radians and use timeline time offset by startTime
		fromRad := imageClip.rotationAnim.FromDeg * math.Pi / 180.0
		toRad := imageClip.rotationAnim.ToDeg * math.Pi / 180.0
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", imageClip.startTime)
		angleExpr := linearExpr(tOffsetExpr, imageClip.rotationAnim.Start, imageClip.rotationAnim.Duration, fromRad, toRad)
		rotationFilter = buildRotationAnimationFilter(angleExpr)
	}

	// Build opacity filter if opacity < 1.0
	opacityFilter := ""
	if imageClip.opacity < 1.0 {
		opacityFilter = fmt.Sprintf("format=yuva420p,colorchannelmixer=aa=%.2f", imageClip.opacity)
	}

	// Build timing expression
	enableExpr := ""
	if imageClip.startTime > 0 || overlayDuration > 0 {
		endTime := imageClip.startTime + overlayDuration
		if endTime > videoDuration {
			endTime = videoDuration
		}
		enableExpr = fmt.Sprintf(":enable='between(t,%.3f,%.3f)'", imageClip.startTime, endTime)
	}

	// Build the complete filter chain
	// Process image: scale and opacity, then overlay on video
	// The caller will replace [1:v] with actual input label and [0:v] with current video label
	
	// Build image processing filter chain
	var imageProcessingFilters []string
	if scaleFilter != "" {
		imageProcessingFilters = append(imageProcessingFilters, scaleFilter)
	}
	if rotationFilter != "" {
		imageProcessingFilters = append(imageProcessingFilters, rotationFilter)
	}
	if opacityFilter != "" {
		imageProcessingFilters = append(imageProcessingFilters, opacityFilter)
	}
	
	// If no filters are needed, use null filter to pass through unchanged
	imageProcessingPart := ""
	if len(imageProcessingFilters) > 0 {
		imageProcessingPart = strings.Join(imageProcessingFilters, ",")
	} else {
		imageProcessingPart = "null" // Pass-through filter when no transformations needed
	}
	
	// Build overlay filter with animated or static position
	var overlayParams string
	if imageClip.positionAnim != nil {
		// Use timeline time offset by startTime for clip-local animation
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", imageClip.startTime)
		xExpr := linearExpr(tOffsetExpr, imageClip.positionAnim.Start, imageClip.positionAnim.Duration,
			imageClip.positionAnim.FromX, imageClip.positionAnim.ToX)
		yExpr := linearExpr(tOffsetExpr, imageClip.positionAnim.Start, imageClip.positionAnim.Duration,
			imageClip.positionAnim.FromY, imageClip.positionAnim.ToY)
		overlayParams = fmt.Sprintf("x='%s':y='%s':shortest=1:eof_action=pass", xExpr, yExpr)
	} else {
		overlayParams = fmt.Sprintf("x=%d:y=%d:shortest=1:eof_action=pass", imageClip.x, imageClip.y)
	}
	imageFilter := fmt.Sprintf("[1:v]%s[img];[0:v][img]overlay=%s%s",
		imageProcessingPart, overlayParams, enableExpr)
	
	return imageFilter, nil
}

// ============================================================================
// FFmpeg Filter Generation - Color Overlay
// ============================================================================

// buildColorOverlayFilterString generates FFmpeg filter string for color overlay
func buildColorOverlayFilterString(colorClip *ColorClip, videoWidth, videoHeight uint64, videoDuration float64) string {
	// Normalize color
	color := normalizeColor(colorClip.color)

	// Determine overlay duration
	overlayDuration := colorClip.overlayDuration
	if overlayDuration == 0 {
		overlayDuration = videoDuration - colorClip.startTime
		if overlayDuration < 0 {
			overlayDuration = 0
		}
	}

	// Build opacity filter if opacity < 1.0
	opacityFilter := ""
	if colorClip.opacity < 1.0 {
		opacityFilter = fmt.Sprintf(",format=yuva420p,colorchannelmixer=aa=%.2f", colorClip.opacity)
	}

	// Build timing expression
	enableExpr := ""
	if colorClip.startTime > 0 || overlayDuration > 0 {
		endTime := colorClip.startTime + overlayDuration
		if endTime > videoDuration {
			endTime = videoDuration
		}
		enableExpr = fmt.Sprintf(":enable='between(t,%.3f,%.3f)'", colorClip.startTime, endTime)
	}

	// Build color source filter
	colorSource := fmt.Sprintf("color=c=%s:s=%dx%d:r=%d:d=%.3f%s[color]", 
		color, colorClip.width, colorClip.height, colorClip.fps, overlayDuration, opacityFilter)

	// Build transform filters for color source (scale and rotation)
	var transformFilters []string
	if colorClip.scaleAnim != nil {
		// Use timeline time offset by startTime for clip-local animation
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", colorClip.startTime)
		scaleExpr := linearExpr(tOffsetExpr, colorClip.scaleAnim.Start, colorClip.scaleAnim.Duration,
			colorClip.scaleAnim.From, colorClip.scaleAnim.To)
		transformFilters = append(transformFilters, buildScaleAnimationFilter(scaleExpr))
	}
	if colorClip.rotationAnim != nil {
		// Convert degrees to radians
		fromRad := colorClip.rotationAnim.FromDeg * math.Pi / 180.0
		toRad := colorClip.rotationAnim.ToDeg * math.Pi / 180.0
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", colorClip.startTime)
		angleExpr := linearExpr(tOffsetExpr, colorClip.rotationAnim.Start, colorClip.rotationAnim.Duration, fromRad, toRad)
		transformFilters = append(transformFilters, buildRotationAnimationFilter(angleExpr))
	}

	// Build overlay filter with animated or static position
	var overlayFilter string
	if colorClip.positionAnim != nil {
		// Use timeline time offset by startTime for clip-local animation
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", colorClip.startTime)
		xExpr := linearExpr(tOffsetExpr, colorClip.positionAnim.Start, colorClip.positionAnim.Duration,
			colorClip.positionAnim.FromX, colorClip.positionAnim.ToX)
		yExpr := linearExpr(tOffsetExpr, colorClip.positionAnim.Start, colorClip.positionAnim.Duration,
			colorClip.positionAnim.FromY, colorClip.positionAnim.ToY)
		overlayFilter = fmt.Sprintf("[0:v][color]overlay=x='%s':y='%s'%s", xExpr, yExpr, enableExpr)
	} else {
		overlayFilter = fmt.Sprintf("[0:v][color]overlay=x=%d:y=%d%s", colorClip.x, colorClip.y, enableExpr)
	}

	// Combine color source, transforms, and overlay
	if len(transformFilters) > 0 {
		transformPart := strings.Join(transformFilters, ",")
		// Replace [color] with [color_transformed] in overlay filter
		overlayFilterTransformed := strings.ReplaceAll(overlayFilter, "[color]", "[color_transformed]")
		return fmt.Sprintf("%s;[color]%s[color_transformed];%s", colorSource, transformPart, overlayFilterTransformed)
	}
	return colorSource + ";" + overlayFilter
}
