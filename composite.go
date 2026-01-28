package moviego

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/fatih/color"
)

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
	PixelFormat string
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

// ============================================================================
// CompositeClipItem Constructor and Builder Methods
// ============================================================================

// NewCompositeClipItem creates a new CompositeClipItem from a Video
func NewCompositeClipItem(video *Video) *CompositeClipItem {
	return &CompositeClipItem{
		video:     video,
		x:         0,
		y:         0,
		width:     0, // 0 means use video's original width
		height:    0, // 0 means use video's original height
		startTime: 0.0,
		duration:  0.0, // 0 means use full video duration
		opacity:   1.0,
		layer:     0,
	}
}

// SetPosition sets the absolute position of the clip in pixels
func (item *CompositeClipItem) SetPosition(x, y int64) *CompositeClipItem {
	item.x = x
	item.y = y
	return item
}

// SetPositionPercent sets the position as a percentage of the canvas size (0.0-1.0)
// Note: This requires canvas size to be known, so it's applied during rendering
func (item *CompositeClipItem) SetPositionPercent(x, y float64) *CompositeClipItem {
	// Store as negative values to indicate percentage positioning
	// This will be converted to absolute pixels during rendering
	item.x = int64(-x * 10000) // Multiply by 10000 to preserve precision
	item.y = int64(-y * 10000)
	return item
}

// SetSize overrides the clip size (0 = use original size)
func (item *CompositeClipItem) SetSize(width, height uint64) *CompositeClipItem {
	item.width = width
	item.height = height
	return item
}

// SetStartTime sets when this clip appears in the composite timeline
func (item *CompositeClipItem) SetStartTime(start float64) *CompositeClipItem {
	item.startTime = start
	return item
}

// SetDuration sets how long this clip appears (0 = full clip duration)
func (item *CompositeClipItem) SetDuration(duration float64) *CompositeClipItem {
	item.duration = duration
	return item
}

// SetOpacity sets the clip opacity (0.0-1.0, default: 1.0)
func (item *CompositeClipItem) SetOpacity(opacity float64) *CompositeClipItem {
	if opacity < 0.0 {
		opacity = 0.0
	}
	if opacity > 1.0 {
		opacity = 1.0
	}
	item.opacity = opacity
	return item
}

// SetLayer sets the z-order/layer (higher = on top, default: 0)
func (item *CompositeClipItem) SetLayer(layer int) *CompositeClipItem {
	item.layer = layer
	return item
}

// SetPositionAnim sets position animation parameters
func (item *CompositeClipItem) SetPositionAnim(params PositionAnimParams) *CompositeClipItem {
	item.positionAnim = &params
	return item
}

// SetRotationAnim sets rotation animation parameters
func (item *CompositeClipItem) SetRotationAnim(params RotationAnimParams) *CompositeClipItem {
	item.rotationAnim = &params
	return item
}

// SetScaleAnim sets scale animation parameters
func (item *CompositeClipItem) SetScaleAnim(params ScaleAnimParams) *CompositeClipItem {
	item.scaleAnim = &params
	return item
}

// ============================================================================
// CompositeClipItem Getters
// ============================================================================

// GetVideo returns the video for this composite item
func (item *CompositeClipItem) GetVideo() *Video {
	return item.video
}

// GetX returns the X position
func (item *CompositeClipItem) GetX() int64 {
	return item.x
}

// GetY returns the Y position
func (item *CompositeClipItem) GetY() int64 {
	return item.y
}

// GetWidth returns the override width (0 = use video width)
func (item *CompositeClipItem) GetWidth() uint64 {
	return item.width
}

// GetHeight returns the override height (0 = use video height)
func (item *CompositeClipItem) GetHeight() uint64 {
	return item.height
}

// GetStartTime returns when this clip starts in the composite
func (item *CompositeClipItem) GetStartTime() float64 {
	return item.startTime
}

// GetDuration returns how long this clip appears
func (item *CompositeClipItem) GetDuration() float64 {
	return item.duration
}

// GetOpacity returns the opacity level
func (item *CompositeClipItem) GetOpacity() float64 {
	return item.opacity
}

// GetLayer returns the layer/z-order
func (item *CompositeClipItem) GetLayer() int {
	return item.layer
}

// ============================================================================
// CompositeClip Creation Functions
// ============================================================================

// NewCompositeClip creates a new composite video from a list of clips
// The canvas size is automatically calculated from the clips
func NewCompositeClip(clips []*Video) *Video {
	if len(clips) == 0 {
		// Return empty composite with default size
		return &Video{
			width:          1920,
			height:         1080,
			fps:            30,
			duration:       0,
			isComposited:   true,
			compositeItems: []*CompositeClipItem{},
		}
	}

	// Convert clips to CompositeClipItems
	items := make([]*CompositeClipItem, len(clips))
	for i, clip := range clips {
		items[i] = NewCompositeClipItem(clip)
	}

	// Calculate canvas size and duration
	width, height := calculateCanvasSize(items)
	duration := calculateCompositeDuration(items)

	// Use first clip's FPS or default to 30
	fps := uint64(30)
	if len(clips) > 0 && clips[0].GetFps() > 0 {
		fps = clips[0].GetFps()
	}

	return &Video{
		width:          width,
		height:         height,
		fps:            fps,
		duration:       duration,
		frames:         uint64(float64(fps) * duration),
		isComposited:   true,
		compositeItems: items,
		codec:          CodecLibx264,
		preset:         Medium,
		pixelFormat:    "yuv420p",
	}
}

// NewCompositeClipWithSize creates a composite video with explicit canvas size
func NewCompositeClipWithSize(width, height uint64) *Video {
	return &Video{
		width:          width,
		height:         height,
		fps:            30,
		duration:       0,
		isComposited:   true,
		compositeItems: []*CompositeClipItem{},
		codec:          CodecLibx264,
		preset:         Medium,
		pixelFormat:    "yuv420p",
	}
}

// NewCompositeClipWithParams creates a composite video from items and parameters
func NewCompositeClipWithParams(items []*CompositeClipItem, params CompositeClipParameters) *Video {
	if len(items) == 0 {
		return &Video{
			width:          params.Width,
			height:         params.Height,
			fps:            params.Fps,
			duration:       0,
			isComposited:   true,
			compositeItems: []*CompositeClipItem{},
		}
	}

	// Calculate canvas size if not provided
	width := params.Width
	height := params.Height
	if width == 0 || height == 0 {
		width, height = calculateCanvasSize(items)
	}

	// Calculate duration
	duration := calculateCompositeDuration(items)

	// Use provided FPS or first clip's FPS
	fps := params.Fps
	if fps == 0 && len(items) > 0 && items[0].GetVideo() != nil {
		fps = items[0].GetVideo().GetFps()
	}
	if fps == 0 {
		fps = 30
	}

	// Set codec from params or default
	codec := params.Codec
	if string(codec) == "" {
		codec = CodecLibx264
	}

	// Set preset from params or default
	presetValue := params.Preset
	if presetValue == "" {
		presetValue = Medium
	}

	return &Video{
		width:          width,
		height:         height,
		fps:            fps,
		duration:       duration,
		frames:         uint64(float64(fps) * duration),
		isComposited:   true,
		compositeItems: items,
		codec:          codec,
		preset:         presetValue,
		bitRate:        params.Bitrate,
		pixelFormat:    params.PixelFormat,
	}
}

// ============================================================================
// CompositeClip Builder Methods (on Video)
// ============================================================================

// AddClip adds a CompositeClipItem to the composite
func (v *Video) AddClip(item *CompositeClipItem) *Video {
	if !v.isComposited {
		v.isComposited = true
	}
	v.compositeItems = append(v.compositeItems, item)

	// Recalculate duration
	v.duration = calculateCompositeDuration(v.compositeItems)
	v.frames = uint64(float64(v.fps) * v.duration)

	return v
}

// SetBgColor sets the background color for the composite
// Note: This is stored as a custom property and used during rendering
func (v *Video) SetBgColor(color string) *Video {
	// Store in ffmpegArgs for now
	if v.ffmpegArgs == nil {
		v.ffmpegArgs = make(map[string][]string)
	}
	v.ffmpegArgs["bgcolor"] = []string{color}
	return v
}

// ============================================================================
// Helper Functions
// ============================================================================

// calculateCanvasSize automatically calculates canvas size from clips
func calculateCanvasSize(items []*CompositeClipItem) (width, height uint64) {
	if len(items) == 0 {
		return 1920, 1080 // Default HD size
	}

	maxWidth := uint64(0)
	maxHeight := uint64(0)

	for _, item := range items {
		if item.video == nil {
			continue
		}

		// Calculate clip dimensions
		clipWidth := item.width
		if clipWidth == 0 {
			clipWidth = item.video.GetWidth()
		}

		clipHeight := item.height
		if clipHeight == 0 {
			clipHeight = item.video.GetHeight()
		}

		// Calculate the rightmost and bottommost points
		rightEdge := uint64(item.x) + clipWidth
		bottomEdge := uint64(item.y) + clipHeight

		if rightEdge > maxWidth {
			maxWidth = rightEdge
		}
		if bottomEdge > maxHeight {
			maxHeight = bottomEdge
		}
	}

	// If no clips have valid dimensions, use first clip's size
	if maxWidth == 0 || maxHeight == 0 {
		for _, item := range items {
			if item.video != nil {
				return item.video.GetWidth(), item.video.GetHeight()
			}
		}
	}

	return maxWidth, maxHeight
}

// calculateCompositeDuration calculates the total duration of the composite
func calculateCompositeDuration(items []*CompositeClipItem) float64 {
	maxDuration := 0.0

	for _, item := range items {
		if item.video == nil {
			continue
		}

		clipDuration := item.duration
		if clipDuration == 0 {
			clipDuration = item.video.GetDuration()
		}

		endTime := item.startTime + clipDuration
		if endTime > maxDuration {
			maxDuration = endTime
		}
	}

	return maxDuration
}

// sortClipsByLayer sorts clips by layer (lower layers first)
func sortClipsByLayer(items []*CompositeClipItem) []*CompositeClipItem {
	sorted := make([]*CompositeClipItem, len(items))
	copy(sorted, items)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].layer < sorted[j].layer
	})

	return sorted
}



// buildOverlayFilterString builds the FFmpeg filter_complex string for overlay composition
func buildOverlayFilterString(video *Video, sourceMap map[string]int) (string, bool) {
	if len(video.compositeItems) == 0 {
		return "", false
	}

	// Sort clips by layer
	sortedItems := sortClipsByLayer(video.compositeItems)

	// Get background color
	bgColor := "black"
	if video.ffmpegArgs != nil {
		if colors, ok := video.ffmpegArgs["bgcolor"]; ok && len(colors) > 0 {
			bgColor = colors[0]
		}
	}

	var filterParts []string
	hasAudio := false

	// Create base canvas
	canvasFilter := fmt.Sprintf("color=c=%s:s=%dx%d:r=%d:d=%.3f[base]",
		bgColor, video.GetWidth(), video.GetHeight(), video.GetFps(), video.GetDuration())
	filterParts = append(filterParts, canvasFilter)

	// Check if any clip has audio (before filtering valid items)
	for _, item := range sortedItems {
		if item.video != nil && item.video.HasAudio() {
			hasAudio = true
			break
		}
	}

	// Process each clip - first collect valid clips
	validItems := []*CompositeClipItem{}
	for _, item := range sortedItems {
		if item.video == nil {
			continue
		}
		filename := item.video.GetFilename()
		if filename == "" {
			continue
		}
		if _, exists := sourceMap[filename]; !exists {
			continue
		}
		validItems = append(validItems, item)
	}

	if len(validItems) == 0 {
		return "", false
	}

	// Process valid clips
	currentOutput := "base"
	for i, item := range validItems {
		filename := item.video.GetFilename()
		sourceIndex := sourceMap[filename]
		clipLabel := fmt.Sprintf("clip%d", i)

		// Build video filter chain for this clip
		videoFilter := fmt.Sprintf("[%d:v]", sourceIndex)

		// Handle trimming - account for subclip timing and composite item duration
		clipStart := item.video.GetStartTime()
		clipEnd := item.video.GetEndTime()
		clipDuration := item.video.GetDuration()

		// Determine trim parameters
		needsTrim := false
		trimStart := 0.0
		trimEnd := 0.0
		trimDuration := 0.0

		// If video is a subclip, we need to trim from source
		if clipStart > 0 || clipEnd > 0 {
			needsTrim = true
			trimStart = clipStart
			if clipEnd > 0 {
				trimEnd = clipEnd
				trimDuration = clipEnd - clipStart
			} else {
				trimDuration = clipDuration
			}
		}

		// If composite item has duration override, limit the duration
		if item.duration > 0 {
			if needsTrim {
				// Limit the duration from the trim start
				trimDuration = item.duration
				if trimEnd > 0 {
					trimEnd = trimStart + item.duration
				}
			} else {
				// Trim from beginning of video
				needsTrim = true
				trimStart = 0.0
				trimDuration = item.duration
			}
		}

		// Apply trim if needed
		if needsTrim {
			if trimEnd > 0 {
				videoFilter += fmt.Sprintf("trim=start=%.3f:end=%.3f,setpts=PTS-STARTPTS", trimStart, trimEnd)
			} else if trimDuration > 0 {
				videoFilter += fmt.Sprintf("trim=start=%.3f:duration=%.3f,setpts=PTS-STARTPTS", trimStart, trimDuration)
			} else {
				videoFilter += "setpts=PTS-STARTPTS"
			}
		} else {
			videoFilter += "setpts=PTS-STARTPTS"
		}

		// Apply scale animation if specified
		if item.scaleAnim != nil {
			// Use clip-local time (t) since we already reset PTS
			scaleExpr := linearExpr("t", item.scaleAnim.Start, item.scaleAnim.Duration,
				item.scaleAnim.From, item.scaleAnim.To)
			videoFilter += "," + buildScaleAnimationFilter(scaleExpr)
		} else if item.width > 0 || item.height > 0 {
			// Apply static scale if size override specified
			w := item.width
			h := item.height
			if w == 0 {
				w = item.video.GetWidth()
			}
			if h == 0 {
				h = item.video.GetHeight()
			}
			videoFilter += fmt.Sprintf(",scale=%d:%d", w, h)
		}

		// Apply rotation animation if specified
		if item.rotationAnim != nil {
			// Convert degrees to radians and use clip-local time
			fromRad := item.rotationAnim.FromDeg * math.Pi / 180.0
			toRad := item.rotationAnim.ToDeg * math.Pi / 180.0
			angleExpr := linearExpr("t", item.rotationAnim.Start, item.rotationAnim.Duration, fromRad, toRad)
			videoFilter += "," + buildRotationAnimationFilter(angleExpr)
		}

		// Apply opacity if needed
		if item.opacity < 1.0 {
			videoFilter += fmt.Sprintf(",format=yuva420p,colorchannelmixer=aa=%.2f", item.opacity)
		}

		// Apply filters from the video
		filterString := translateFiltersToFFmpeg(item.video.filters)
		if filterString != "" {
			videoFilter += "," + filterString
		}

		videoFilter += fmt.Sprintf("[%s]", clipLabel)
		filterParts = append(filterParts, videoFilter)

		// Build overlay
		nextOutput := fmt.Sprintf("out%d", i)
		if i == len(validItems)-1 {
			nextOutput = "outv"
		}

		// Build overlay with position animation if specified
		if item.positionAnim != nil {
			// Use timeline time offset by startTime to get clip-local animation
			tOffsetExpr := fmt.Sprintf("(t-%.3f)", item.startTime)
			xExpr := linearExpr(tOffsetExpr, item.positionAnim.Start, item.positionAnim.Duration,
				item.positionAnim.FromX, item.positionAnim.ToX)
			yExpr := linearExpr(tOffsetExpr, item.positionAnim.Start, item.positionAnim.Duration,
				item.positionAnim.FromY, item.positionAnim.ToY)
			overlayFilter := fmt.Sprintf("[%s][%s]overlay=x='%s':y='%s'",
				currentOutput, clipLabel, xExpr, yExpr)
			if item.startTime > 0 {
				overlayFilter += fmt.Sprintf(":enable='gte(t,%.3f)'", item.startTime)
			}
			overlayFilter += fmt.Sprintf("[%s]", nextOutput)
			filterParts = append(filterParts, overlayFilter)
		} else {
			// Static position
			overlayFilter := fmt.Sprintf("[%s][%s]overlay=x=%d:y=%d",
				currentOutput, clipLabel, item.x, item.y)
			if item.startTime > 0 {
				overlayFilter += fmt.Sprintf(":enable='gte(t,%.3f)'", item.startTime)
			}
			overlayFilter += fmt.Sprintf("[%s]", nextOutput)
			filterParts = append(filterParts, overlayFilter)
		}

		currentOutput = nextOutput
	}

	// Handle audio mixing if needed
	if hasAudio {
		audioFilters := []string{}
		audioInputs := []string{}

		for i, item := range validItems {
			if item.video == nil || !item.video.HasAudio() {
				continue
			}

			filename := item.video.GetFilename()
			sourceIndex := sourceMap[filename]
			audioLabel := fmt.Sprintf("a%d", i)

			audioFilter := fmt.Sprintf("[%d:a]", sourceIndex)

			// Apply audio trim - account for subclip timing and composite item duration
			clipStart := item.video.GetStartTime()
			clipEnd := item.video.GetEndTime()
			clipDuration := item.video.GetDuration()

			needsTrim := false
			trimStart := 0.0
			trimDuration := clipDuration

			if clipStart > 0 || clipEnd > 0 {
				needsTrim = true
				trimStart = clipStart
				if clipEnd > 0 {
					trimDuration = clipEnd - clipStart
				}
			}

			if item.duration > 0 {
				if needsTrim {
					trimDuration = item.duration
				} else {
					needsTrim = true
					trimStart = 0.0
					trimDuration = item.duration
				}
			}

			if needsTrim {
				audioFilter += fmt.Sprintf("atrim=start=%.3f:duration=%.3f,asetpts=PTS-STARTPTS", trimStart, trimDuration)
			} else {
				audioFilter += "asetpts=PTS-STARTPTS"
			}

			// Delay audio if startTime is set
			if item.startTime > 0 {
				audioFilter += fmt.Sprintf(",adelay=%.0f|%.0f", item.startTime*1000, item.startTime*1000)
			}

			audioFilter += fmt.Sprintf("[%s]", audioLabel)
			audioFilters = append(audioFilters, audioFilter)
			audioInputs = append(audioInputs, fmt.Sprintf("[%s]", audioLabel))
		}

		// Mix all audio streams
		if len(audioInputs) > 0 {
			filterParts = append(filterParts, audioFilters...)
			mixFilter := strings.Join(audioInputs, "") + fmt.Sprintf("amix=inputs=%d[outa]", len(audioInputs))
			filterParts = append(filterParts, mixFilter)
		}
	}

	// Collect all overlay clips and sort by layer
	type overlayItem struct {
		layer    int
		clipType string // "text", "image", "color", "subtitle"
		index    int
		clip     interface{}
	}
	var overlays []overlayItem

	// Add text clips
	if len(video.textClips) > 0 {
		for i, textClip := range video.textClips {
			overlays = append(overlays, overlayItem{
				layer:    textClip.GetLayer(),
				clipType: "text",
				index:    i,
				clip:     textClip,
			})
		}
	}

	// Add image clips
	if len(video.imageClips) > 0 {
		for i, imageClip := range video.imageClips {
			overlays = append(overlays, overlayItem{
				layer:    imageClip.GetLayer(),
				clipType: "image",
				index:    i,
				clip:     imageClip,
			})
		}
	}

	// Add color clips
	if len(video.colorClips) > 0 {
		for i, colorClip := range video.colorClips {
			overlays = append(overlays, overlayItem{
				layer:    colorClip.GetLayer(),
				clipType: "color",
				index:    i,
				clip:     colorClip,
			})
		}
	}

	// Add subtitle clips
	if len(video.subtitleClips) > 0 {
		for i, subtitleClip := range video.subtitleClips {
			overlays = append(overlays, overlayItem{
				layer:    0, // Subtitles typically go on top
				clipType: "subtitle",
				index:    i,
				clip:     subtitleClip,
			})
		}
	}

	// Sort overlays by layer
	sort.Slice(overlays, func(i, j int) bool {
		return overlays[i].layer < overlays[j].layer
	})

	// Process overlays in layer order
	if len(overlays) > 0 {
		currentOutput = "outv"

		// Track image input indices - calculate starting index after all video inputs
		maxVideoInputIdx := -1
		for _, idx := range sourceMap {
			if idx > maxVideoInputIdx {
				maxVideoInputIdx = idx
			}
		}
		imageInputIdx := maxVideoInputIdx + 1 // Start after all video inputs
		imageClipToInput := make(map[*ImageClip]int)

		// First pass: collect image clips and add inputs
		for _, overlay := range overlays {
			if overlay.clipType == "image" {
				imageClip := overlay.clip.(*ImageClip)
				if imageClip.imagePath != "" {
					if _, exists := imageClipToInput[imageClip]; !exists {
						imageClipToInput[imageClip] = imageInputIdx
						imageInputIdx++
					}
				}
			}
		}

		// Process overlays
		for i, overlay := range overlays {
			isLast := i == len(overlays)-1
			var nextOutput string
			if isLast {
				nextOutput = "outv_final"
			} else {
				nextOutput = fmt.Sprintf("overlay%d", i)
			}

			switch overlay.clipType {
			case "text":
				textClip := overlay.clip.(*TextClip)
				// Check if text needs rotation/scale (requires special handling)
				if textNeedsRotationOrScale(textClip) {
					// Use special transparent layer approach for rotated/scaled text
					textFilter := buildRotatedScaledTextFilterString(textClip, video.GetWidth(), video.GetHeight(), video.GetDuration(), currentOutput, nextOutput)
					if textFilter != "" {
						// Split by semicolon and add each part
						parts := strings.Split(textFilter, ";")
						for _, part := range parts {
							part = strings.TrimSpace(part)
							if part != "" {
								filterParts = append(filterParts, part)
							}
						}
						currentOutput = nextOutput
					}
				} else {
					// Normal text overlay (position animation only, no rotation/scale)
					textFilter := buildTextFilterString(textClip, video.GetWidth(), video.GetHeight(), video.GetDuration())
					filterParts = append(filterParts, fmt.Sprintf("[%s]%s[%s]", currentOutput, textFilter, nextOutput))
					currentOutput = nextOutput
				}

			case "image":
				imageClip := overlay.clip.(*ImageClip)
				if imageClip.imagePath != "" {
					if inputIdx, exists := imageClipToInput[imageClip]; exists {
						imageFilter, err := buildImageOverlayFilterString(imageClip, video.GetWidth(), video.GetHeight(), video.GetDuration())
						if err != nil {
							// Skip image on error
							continue
						}
						// Replace [1:v] with actual input index and [0:v] with current label
						imageInputLabel := fmt.Sprintf("%d:v", inputIdx)
						imageFilter = strings.ReplaceAll(imageFilter, "[1:v]", "["+imageInputLabel+"]")
						imageFilter = strings.ReplaceAll(imageFilter, "[0:v]", "["+currentOutput+"]")
						imageFilter = strings.ReplaceAll(imageFilter, "[out]", "["+nextOutput+"]")
						filterParts = append(filterParts, imageFilter)
						currentOutput = nextOutput
					}
				}

			case "color":
				colorClip := overlay.clip.(*ColorClip)
				colorFilter := buildColorOverlayFilterString(colorClip, video.GetWidth(), video.GetHeight(), video.GetDuration())
				// Replace [0:v] with current label and [out] with next label
				colorFilter = strings.ReplaceAll(colorFilter, "[0:v]", "["+currentOutput+"]")
				colorFilter = strings.ReplaceAll(colorFilter, "[out]", "["+nextOutput+"]")
				filterParts = append(filterParts, colorFilter)
				currentOutput = nextOutput

			case "subtitle":
				subtitleClip := overlay.clip.(*SubtitleClip)
				subtitleFilter, err := buildSubtitleFilterString(subtitleClip)
				if err != nil {
					// Skip subtitle on error
					continue
				}
				filterParts = append(filterParts, fmt.Sprintf("[%s]%s[%s]", currentOutput, subtitleFilter, nextOutput))
				currentOutput = nextOutput
			}
		}

		// Rename final output back to outv
		if currentOutput == "outv_final" {
			filterParts = append(filterParts, fmt.Sprintf("[outv_final]null[outv]"))
		}
	}

	return strings.Join(filterParts, ";"), hasAudio
}

// writeCompositeVideo writes a composite video using FFmpeg overlay filters
func (video *Video) writeCompositeVideo(params VideoParameters) error {
	if len(video.compositeItems) == 0 {
		return fmt.Errorf("no clips in composite video")
	}

	// Build source map
	sourceMap := make(map[string]int)
	sources := []string{}

	for _, item := range video.compositeItems {
		if item.video == nil {
			continue
		}
		filename := item.video.GetFilename()
		if _, exists := sourceMap[filename]; !exists {
			sourceMap[filename] = len(sources)
			sources = append(sources, filename)
		}
	}

	// Build FFmpeg command
	args := []string{"-loglevel", "error"}

	// Add all source files as inputs
	for _, source := range sources {
		args = append(args, "-i", source)
	}

	// Add image inputs for image clips
	// Limit image inputs with duration to prevent infinite streams (MoviePy-style)
	videoDuration := video.GetDuration()
	if videoDuration <= 0 {
		// Calculate duration from composite items if not set
		maxDuration := 0.0
		for _, item := range video.compositeItems {
			if item.video != nil {
				itemEnd := item.startTime + item.video.GetDuration()
				if itemEnd > maxDuration {
					maxDuration = itemEnd
				}
			}
		}
		if maxDuration > 0 {
			videoDuration = maxDuration
		} else {
			videoDuration = 10.0 // Fallback default
		}
	}
	if video.HasImageClips() {
		for _, imageClip := range video.GetImageClips() {
			if imageClip.imagePath != "" {
				// Add loop, duration limit, and input for image overlay (MoviePy-style)
				args = append(args, "-loop", "1", "-t", fmt.Sprintf("%.3f", videoDuration), "-i", imageClip.imagePath)
			}
		}
	}

	// Build filter_complex
	filterComplex, hasAudio := buildOverlayFilterString(video, sourceMap)

	args = append(args, "-filter_complex", filterComplex)
	args = append(args, "-map", "[outv]")

	if hasAudio {
		args = append(args, "-map", "[outa]")
	}

	// Apply encoding parameters
	selectedCodec := resolveCodec(string(params.Codec), video.GetCodec())
	args = append(args, "-c:v", selectedCodec)

	selectedPreset := mapPresetForCodec(selectedCodec, resolvePreset(params.Preset, video.GetPreset()))
	if selectedPreset != "" {
		args = append(args, "-preset", selectedPreset)
	}

	if bitrate := resolveBitrate(params.Bitrate, video.GetBitRate()); bitrate != "" {
		args = append(args, "-b:v", bitrate)
	}

	if fps := resolveFps(params.Fps, video.GetFps()); fps > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", fps))
	}

	if params.PixelFormat != "" {
		args = append(args, "-pix_fmt", params.PixelFormat)
	} else if video.GetPixelFormat() != "" {
		args = append(args, "-pix_fmt", video.GetPixelFormat())
	}

	if params.Threads != 0 {
		args = append(args, "-threads", fmt.Sprintf("%d", params.Threads))
	}

	// Audio encoding
	if hasAudio {
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	}

	// Output
	args = append(args, "-y", params.OutputPath)

	// Use progress parser for composite video
	config := FFmpegProgressConfig{
		Args:          args,
		TotalDuration: videoDuration,
		OperationName: "Processing composite video",
		OutputPath:    params.OutputPath,
		Bitrate:       resolveBitrate(params.Bitrate, video.GetBitRate()),
	}

	if err := runFFmpegWithProgress(config); err != nil {
		return fmt.Errorf("ffmpeg composite failed: %w", err)
	}

	fmt.Printf("%s %s\n",
		color.GreenString("Composite video written successfully:"),
		color.MagentaString(params.OutputPath))
	return nil
}
