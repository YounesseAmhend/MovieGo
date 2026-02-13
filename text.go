package moviego

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type PositionAnimParams struct {
	FromX    float64
	FromY    float64
	ToX      float64
	ToY      float64
	Start    float64
	Duration float64
}

type RotationAnimParams struct {
	FromDeg  float64
	ToDeg    float64
	Start    float64
	Duration float64
}

type ScaleAnimParams struct {
	From     float64
	To       float64
	Start    float64
	Duration float64
}

type TextAlignment int

const (
	AlignTopLeft TextAlignment = iota
	AlignTopCenter
	AlignTopRight
	AlignCenterLeft
	AlignCenter
	AlignCenterRight
	AlignBottomLeft
	AlignBottomCenter
	AlignBottomRight
)

type TextClip struct {
	text      string
	x         int
	y         int
	alignment TextAlignment
	useAlign  bool

	fontFamily string
	fontSize   int
	fontColor  string
	bold       bool
	italic     bool
	startTime float64
	duration  float64
	fadeIn    float64
	fadeOut   float64

	shadowX     int
	shadowY     int
	shadowColor string
	borderWidth int
	borderColor string
	boxEnabled  bool
	boxColor    string
	boxOpacity  float64

	layer int
	positionAnim *PositionAnimParams
	rotationAnim *RotationAnimParams
	scaleAnim    *ScaleAnimParams
}

type SubtitleClip struct {
	filePath string
	fileType string
	fontFamily string
	fontSize   int
	fontColor  string
	marginV int
	marginH int
	charenc string
}

func NewTextClip(text string) *TextClip {
	return &TextClip{
		text:        text,
		x:           0,
		y:           0,
		alignment:   AlignBottomCenter,
		useAlign:    false,
		fontFamily:  "Arial",
		fontSize:    24,
		fontColor:   "white",
		bold:        false,
		italic:      false,
		startTime:   0.0,
		duration:    0.0, // 0 means full video duration
		fadeIn:      0.0,
		fadeOut:     0.0,
		shadowX:     0,
		shadowY:     0,
		shadowColor: "",
		borderWidth: 0,
		borderColor: "",
		boxEnabled:  false,
		boxColor:    "black",
		boxOpacity:  0.5,
		layer:       0,
	}
}

func (tc *TextClip) SetPosition(x, y int) *TextClip {
	tc.x = x
	tc.y = y
	tc.useAlign = false
	return tc
}

func (tc *TextClip) SetAlignment(alignment TextAlignment) *TextClip {
	tc.alignment = alignment
	tc.useAlign = true
	return tc
}

func (tc *TextClip) SetFont(family string, size int) *TextClip {
	tc.fontFamily = family
	tc.fontSize = size
	return tc
}

func (tc *TextClip) SetFontFamily(family string) *TextClip {
	tc.fontFamily = family
	return tc
}

func (tc *TextClip) SetFontSize(size int) *TextClip {
	tc.fontSize = size
	return tc
}

func (tc *TextClip) SetColor(color string) *TextClip {
	tc.fontColor = color
	return tc
}

func (tc *TextClip) SetBold(bold bool) *TextClip {
	tc.bold = bold
	return tc
}

func (tc *TextClip) SetItalic(italic bool) *TextClip {
	tc.italic = italic
	return tc
}

func (tc *TextClip) SetTiming(startTime, duration float64) *TextClip {
	tc.startTime = startTime
	tc.duration = duration
	return tc
}

func (tc *TextClip) SetStartTime(startTime float64) *TextClip {
	tc.startTime = startTime
	return tc
}

func (tc *TextClip) SetDuration(duration float64) *TextClip {
	tc.duration = duration
	return tc
}

func (tc *TextClip) SetFadeIn(duration float64) *TextClip {
	tc.fadeIn = duration
	return tc
}

func (tc *TextClip) SetFadeOut(duration float64) *TextClip {
	tc.fadeOut = duration
	return tc
}

func (tc *TextClip) SetShadow(x, y int, color string) *TextClip {
	tc.shadowX = x
	tc.shadowY = y
	tc.shadowColor = color
	return tc
}

func (tc *TextClip) SetBorder(width int, color string) *TextClip {
	tc.borderWidth = width
	tc.borderColor = color
	return tc
}

func (tc *TextClip) SetBox(enabled bool, color string, opacity float64) *TextClip {
	tc.boxEnabled = enabled
	tc.boxColor = color
	tc.boxOpacity = opacity
	return tc
}

func (tc *TextClip) SetLayer(layer int) *TextClip {
	tc.layer = layer
	return tc
}

func (tc *TextClip) SetPositionAnim(params PositionAnimParams) *TextClip {
	tc.positionAnim = &params
	return tc
}

func (tc *TextClip) SetRotationAnim(params RotationAnimParams) *TextClip {
	tc.rotationAnim = &params
	return tc
}

func (tc *TextClip) SetScaleAnim(params ScaleAnimParams) *TextClip {
	tc.scaleAnim = &params
	return tc
}

func (tc *TextClip) GetText() string {
	return tc.text
}

func (tc *TextClip) GetLayer() int {
	return tc.layer
}

func (tc *TextClip) GetStartTime() float64 {
	return tc.startTime
}

func (tc *TextClip) GetDuration() float64 {
	return tc.duration
}

func (tc *TextClip) HasAnimation() bool {
	return tc.positionAnim != nil || tc.rotationAnim != nil || tc.scaleAnim != nil
}

func (tc *TextClip) GetPositionAnim() *PositionAnimParams {
	return tc.positionAnim
}

func (tc *TextClip) GetRotationAnim() *RotationAnimParams {
	return tc.rotationAnim
}

func (tc *TextClip) GetScaleAnim() *ScaleAnimParams {
	return tc.scaleAnim
}

func NewSubtitleClip(filePath string) *SubtitleClip {
	ext := strings.ToLower(filepath.Ext(filePath))
	fileType := "srt"
	switch ext {
	case ".srt":
		fileType = "srt"
	case ".ass":
		fileType = "ass"
	case ".vtt":
		fileType = "vtt"
	}

	return &SubtitleClip{
		filePath:   filePath,
		fileType:   fileType,
		fontFamily: "",
		fontSize:   0,
		fontColor:  "",
		marginV:    0,
		marginH:    0,
		charenc:    "UTF-8",
	}
}

func (sc *SubtitleClip) SetFont(family string, size int) *SubtitleClip {
	sc.fontFamily = family
	sc.fontSize = size
	return sc
}

func (sc *SubtitleClip) SetFontFamily(family string) *SubtitleClip {
	sc.fontFamily = family
	return sc
}

func (sc *SubtitleClip) SetFontSize(size int) *SubtitleClip {
	sc.fontSize = size
	return sc
}

func (sc *SubtitleClip) SetColor(color string) *SubtitleClip {
	sc.fontColor = color
	return sc
}

func (sc *SubtitleClip) SetMargins(vertical, horizontal int) *SubtitleClip {
	sc.marginV = vertical
	sc.marginH = horizontal
	return sc
}

func (sc *SubtitleClip) SetEncoding(encoding string) *SubtitleClip {
	sc.charenc = encoding
	return sc
}

func (sc *SubtitleClip) GetFilePath() string {
	return sc.filePath
}

func (sc *SubtitleClip) GetFileType() string {
	return sc.fileType
}

func buildTextFilterString(tc *TextClip, videoWidth, videoHeight uint64, videoDuration float64) string {
	var parts []string

	escapedText := escapeFFmpegText(tc.text)
	parts = append(parts, fmt.Sprintf("text='%s'", escapedText))

	if tc.positionAnim != nil {
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", tc.startTime)
		xExpr := linearExpr(tOffsetExpr, tc.positionAnim.Start, tc.positionAnim.Duration,
			tc.positionAnim.FromX, tc.positionAnim.ToX)
		yExpr := linearExpr(tOffsetExpr, tc.positionAnim.Start, tc.positionAnim.Duration,
			tc.positionAnim.FromY, tc.positionAnim.ToY)
		parts = append(parts, fmt.Sprintf("x='%s'", xExpr))
		parts = append(parts, fmt.Sprintf("y='%s'", yExpr))
	} else if tc.useAlign {
		x, y := alignmentToPosition(tc.alignment, videoWidth, videoHeight)
		parts = append(parts, fmt.Sprintf("x=%s", x))
		parts = append(parts, fmt.Sprintf("y=%s", y))
	} else {
		parts = append(parts, fmt.Sprintf("x=%d", tc.x))
		parts = append(parts, fmt.Sprintf("y=%d", tc.y))
	}

	if tc.fontFamily != "" {
		fontPath := resolveFontPath(tc.fontFamily)
		fontPath = strings.ReplaceAll(fontPath, ":", "\\\\:")
		parts = append(parts, fmt.Sprintf("fontfile=%s", fontPath))
	}
	parts = append(parts, fmt.Sprintf("fontsize=%d", tc.fontSize))
	parts = append(parts, fmt.Sprintf("fontcolor=%s", normalizeColor(tc.fontColor)))

	if tc.shadowColor != "" {
		parts = append(parts, fmt.Sprintf("shadowx=%d", tc.shadowX))
		parts = append(parts, fmt.Sprintf("shadowy=%d", tc.shadowY))
		parts = append(parts, fmt.Sprintf("shadowcolor=%s", normalizeColor(tc.shadowColor)))
	}

	if tc.borderWidth > 0 {
		parts = append(parts, fmt.Sprintf("borderw=%d", tc.borderWidth))
		if tc.borderColor != "" {
			parts = append(parts, fmt.Sprintf("bordercolor=%s", normalizeColor(tc.borderColor)))
		}
	}

	if tc.boxEnabled {
		parts = append(parts, "box=1")
		parts = append(parts, fmt.Sprintf("boxcolor=%s@%.2f", normalizeColor(tc.boxColor), tc.boxOpacity))
		parts = append(parts, "boxborderw=5")
	}

	if tc.startTime > 0 || tc.duration > 0 {
		endTime := tc.startTime + tc.duration
		if tc.duration == 0 {
			endTime = videoDuration
		}
		parts = append(parts, fmt.Sprintf("enable='between(t,%.3f,%.3f)'", tc.startTime, endTime))
	}

	if tc.fadeIn > 0 || tc.fadeOut > 0 {
		alphaExpr := buildAlphaExpression(tc.startTime, tc.duration, tc.fadeIn, tc.fadeOut, videoDuration)
		parts = append(parts, fmt.Sprintf("alpha='%s'", alphaExpr))
	}

	return "drawtext=" + strings.Join(parts, ":")
}

func textNeedsRotationOrScale(tc *TextClip) bool {
	return tc.rotationAnim != nil || tc.scaleAnim != nil
}


func buildRotatedScaledTextFilterString(tc *TextClip, videoWidth, videoHeight uint64, videoDuration float64, baseLabel, outputLabel string) string {
	// Determine text duration
	textDuration := tc.duration
	if textDuration == 0 {
		textDuration = videoDuration - tc.startTime
		if textDuration < 0 {
			textDuration = videoDuration
		}
	}
	if textDuration <= 0 {
		textDuration = videoDuration
	}

	// Create transparent base layer
	textBaseLabel := fmt.Sprintf("textbase_%p", tc) // Unique label
	baseFilter := fmt.Sprintf("color=c=black@0:s=%dx%d:r=%d:d=%.3f[%s]",
		videoWidth, videoHeight, 30, textDuration, textBaseLabel)

	// Build drawtext filter (without rotation/scale)
	var parts []string
	escapedText := escapeFFmpegText(tc.text)
	parts = append(parts, fmt.Sprintf("text='%s'", escapedText))
	
	// Position - always draw text centered on the transparent canvas
	// This ensures that when we rotate the canvas, it rotates around the text center
	parts = append(parts, "x=(w-text_w)/2")
	parts = append(parts, "y=(h-text_h)/2")

	// Font properties
	if tc.fontFamily != "" {
		fontPath := resolveFontPath(tc.fontFamily)
		fontPath = strings.ReplaceAll(fontPath, ":", "\\\\:")
		parts = append(parts, fmt.Sprintf("fontfile=%s", fontPath))
	}
	parts = append(parts, fmt.Sprintf("fontsize=%d", tc.fontSize))
	parts = append(parts, fmt.Sprintf("fontcolor=%s", normalizeColor(tc.fontColor)))

	// Shadow, border, box (same as normal text)
	if tc.shadowColor != "" {
		parts = append(parts, fmt.Sprintf("shadowx=%d", tc.shadowX))
		parts = append(parts, fmt.Sprintf("shadowy=%d", tc.shadowY))
		parts = append(parts, fmt.Sprintf("shadowcolor=%s", normalizeColor(tc.shadowColor)))
	}
	if tc.borderWidth > 0 {
		parts = append(parts, fmt.Sprintf("borderw=%d", tc.borderWidth))
		if tc.borderColor != "" {
			parts = append(parts, fmt.Sprintf("bordercolor=%s", normalizeColor(tc.borderColor)))
		}
	}
	if tc.boxEnabled {
		parts = append(parts, "box=1")
		parts = append(parts, fmt.Sprintf("boxcolor=%s@%.2f", normalizeColor(tc.boxColor), tc.boxOpacity))
		parts = append(parts, "boxborderw=5")
	}

	// Timing
	if tc.startTime > 0 || tc.duration > 0 {
		endTime := tc.startTime + tc.duration
		if tc.duration == 0 {
			endTime = videoDuration
		}
		parts = append(parts, fmt.Sprintf("enable='between(t,%.3f,%.3f)'", tc.startTime, endTime))
	}

	// Fade effects
	if tc.fadeIn > 0 || tc.fadeOut > 0 {
		alphaExpr := buildAlphaExpression(tc.startTime, tc.duration, tc.fadeIn, tc.fadeOut, videoDuration)
		parts = append(parts, fmt.Sprintf("alpha='%s'", alphaExpr))
	}

	drawTextFilter := fmt.Sprintf("drawtext=%s", strings.Join(parts, ":"))
	textDrawLabel := fmt.Sprintf("textdraw_%p", tc)

	// Build transform filters
	var transformFilters []string
	if tc.scaleAnim != nil {
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", tc.startTime)
		scaleExpr := linearExpr(tOffsetExpr, tc.scaleAnim.Start, tc.scaleAnim.Duration,
			tc.scaleAnim.From, tc.scaleAnim.To)
		transformFilters = append(transformFilters, buildScaleAnimationFilter(scaleExpr))
	}
	if tc.rotationAnim != nil {
		fromRad := tc.rotationAnim.FromDeg * math.Pi / 180.0
		toRad := tc.rotationAnim.ToDeg * math.Pi / 180.0
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", tc.startTime)
		angleExpr := linearExpr(tOffsetExpr, tc.rotationAnim.Start, tc.rotationAnim.Duration, fromRad, toRad)
		transformFilters = append(transformFilters, buildRotationAnimationFilter(angleExpr))
	}

	textTransformedLabel := fmt.Sprintf("textxf_%p", tc)
	transformPart := ""
	if len(transformFilters) > 0 {
		transformPart = fmt.Sprintf("[%s]%s[%s];", textDrawLabel, strings.Join(transformFilters, ","), textTransformedLabel)
	} else {
		textTransformedLabel = textDrawLabel
	}

	formatLabel := fmt.Sprintf("textfmt_%p", tc)
	if transformPart != "" {
		transformPart = fmt.Sprintf("%s[%s]format=rgba[%s];", transformPart, textTransformedLabel, formatLabel)
	} else {
		transformPart = fmt.Sprintf("[%s]format=rgba[%s];", textDrawLabel, formatLabel)
	}
	textTransformedLabel = formatLabel

	// Build overlay with animated or static position
	// Since we drew the text centered on its canvas, the center of the text is at (w/2, h/2) 
	// of the overlay canvas. To put the text center at (targetX, targetY), 
	// we overlay at (targetX - w/2, targetY - h/2).
	var overlayFilter string
	if tc.positionAnim != nil {
		tOffsetExpr := fmt.Sprintf("(t-%.3f)", tc.startTime)
		xExpr := linearExpr(tOffsetExpr, tc.positionAnim.Start, tc.positionAnim.Duration,
			tc.positionAnim.FromX, tc.positionAnim.ToX)
		yExpr := linearExpr(tOffsetExpr, tc.positionAnim.Start, tc.positionAnim.Duration,
			tc.positionAnim.FromY, tc.positionAnim.ToY)
		// For rotated text, we treat the position as the CENTER of the text
		overlayFilter = fmt.Sprintf("[%s][%s]overlay=x='%s-w/2':y='%s-h/2'[%s]", baseLabel, textTransformedLabel, xExpr, yExpr, outputLabel)
	} else if !tc.useAlign {
		// Static position - treat as center
		overlayFilter = fmt.Sprintf("[%s][%s]overlay=x=%d-w/2:y=%d-h/2[%s]", baseLabel, textTransformedLabel, tc.x, tc.y, outputLabel)
	} else {
		// Aligned position - calculate target center based on alignment
		targetCX, targetCY := getAlignmentCenter(tc.alignment, videoWidth, videoHeight)
		overlayFilter = fmt.Sprintf("[%s][%s]overlay=x=%s-w/2:y=%s-h/2[%s]", baseLabel, textTransformedLabel, targetCX, targetCY, outputLabel)
	}

	if transformPart != "" {
		return fmt.Sprintf("%s;[%s]%s[%s];%s%s", baseFilter, textBaseLabel, drawTextFilter, textDrawLabel, transformPart, overlayFilter)
	}
	return fmt.Sprintf("%s;[%s]%s[%s];%s", baseFilter, textBaseLabel, drawTextFilter, textDrawLabel, overlayFilter)
}

// getAlignmentCenter calculates the center position for a given alignment
func getAlignmentCenter(alignment TextAlignment, videoWidth, videoHeight uint64) (string, string) {
	switch alignment {
	case AlignTopLeft:
		return "w/4", "h/8" // Rough estimates since we don't know text_w
	case AlignTopCenter:
		return "w/2", "h/8"
	case AlignTopRight:
		return "3*w/4", "h/8"
	case AlignCenterLeft:
		return "w/4", "h/2"
	case AlignCenter:
		return "w/2", "h/2"
	case AlignCenterRight:
		return "3*w/4", "h/2"
	case AlignBottomLeft:
		return "w/4", "7*h/8"
	case AlignBottomCenter:
		return "w/2", "7*h/8"
	case AlignBottomRight:
		return "3*w/4", "7*h/8"
	default:
		return "w/2", "h/2"
	}
}


// buildAlphaExpression creates the FFmpeg alpha expression for fade effects
func buildAlphaExpression(startTime, duration, fadeIn, fadeOut, videoDuration float64) string {
	endTime := startTime + duration
	if duration == 0 {
		endTime = videoDuration
	}

	fadeInEnd := startTime + fadeIn
	fadeOutStart := endTime - fadeOut

	// Complex alpha expression:
	// - Before startTime: 0
	// - During fade-in: linear ramp from 0 to 1
	// - Full visibility: 1
	// - During fade-out: linear ramp from 1 to 0
	// - After endTime: 0

	if fadeIn > 0 && fadeOut > 0 {
		return fmt.Sprintf("if(lt(t,%.3f),0,if(lt(t,%.3f),(t-%.3f)/%.3f,if(lt(t,%.3f),1,if(lt(t,%.3f),(%.3f-t)/%.3f,0))))",
			startTime, fadeInEnd, startTime, fadeIn, fadeOutStart, endTime, endTime, fadeOut)
	} else if fadeIn > 0 {
		return fmt.Sprintf("if(lt(t,%.3f),0,if(lt(t,%.3f),(t-%.3f)/%.3f,1))",
			startTime, fadeInEnd, startTime, fadeIn)
	} else if fadeOut > 0 {
		return fmt.Sprintf("if(lt(t,%.3f),1,if(lt(t,%.3f),(%.3f-t)/%.3f,0))",
			fadeOutStart, endTime, endTime, fadeOut)
	}

	return "1"
}

// alignmentToPosition converts TextAlignment to FFmpeg position expression
func alignmentToPosition(alignment TextAlignment, videoWidth, videoHeight uint64) (string, string) {
	switch alignment {
	case AlignTopLeft:
		return "10", "10"
	case AlignTopCenter:
		return "(w-text_w)/2", "10"
	case AlignTopRight:
		return "(w-text_w-10)", "10"
	case AlignCenterLeft:
		return "10", "(h-text_h)/2"
	case AlignCenter:
		return "(w-text_w)/2", "(h-text_h)/2"
	case AlignCenterRight:
		return "(w-text_w-10)", "(h-text_h)/2"
	case AlignBottomLeft:
		return "10", "(h-text_h-10)"
	case AlignBottomCenter:
		return "(w-text_w)/2", "(h-text_h-10)"
	case AlignBottomRight:
		return "(w-text_w-10)", "(h-text_h-10)"
	default:
		return "(w-text_w)/2", "(h-text_h-10)"
	}
}

// escapeFFmpegText escapes text for use in FFmpeg drawtext filter
func escapeFFmpegText(text string) string {
	// Escape special characters for FFmpeg
	text = strings.ReplaceAll(text, "\\", "\\\\")
	text = strings.ReplaceAll(text, "'", "\\'")
	text = strings.ReplaceAll(text, ":", "\\:")
	text = strings.ReplaceAll(text, "%", "\\%")
	return text
}

// normalizeColor converts color names or hex codes to FFmpeg format
func normalizeColor(color string) string {
	// If it starts with #, it's a hex color - FFmpeg expects it without #
	if strings.HasPrefix(color, "#") {
		return "0x" + strings.TrimPrefix(color, "#")
	}
	// Otherwise, assume it's a named color (FFmpeg supports many)
	return color
}

// resolveFontPath attempts to find a font file path
// This is a simplified version - production code would need platform-specific logic
func resolveFontPath(fontFamily string) string {
	// Common font paths on different platforms
	fontPaths := []string{
		// Windows
		"C:/Windows/Fonts/" + fontFamily + ".ttf",
		"C:/Windows/Fonts/" + strings.ToLower(fontFamily) + ".ttf",
		// Linux
		"/usr/share/fonts/truetype/" + fontFamily + ".ttf",
		"/usr/share/fonts/TTF/" + fontFamily + ".ttf",
		// macOS
		"/Library/Fonts/" + fontFamily + ".ttf",
		"/System/Library/Fonts/" + fontFamily + ".ttf",
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fallback - just use the font family name and hope FFmpeg can find it
	return fontFamily
}

// ============================================================================
// FFmpeg Filter Generation - Subtitles
// ============================================================================

// buildSubtitleFilterString generates FFmpeg subtitle filter for a SubtitleClip
func buildSubtitleFilterString(sc *SubtitleClip) (string, error) {
	// Check if file exists
	if _, err := os.Stat(sc.filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("subtitle file not found: %s", sc.filePath)
	}

	// Convert Windows paths to FFmpeg format (forward slashes, escape colons)
	filePath := filepath.ToSlash(sc.filePath)
	filePath = strings.ReplaceAll(filePath, ":", "\\:")

	if sc.fileType == "ass" {
		// Use ass filter for .ass files
		return fmt.Sprintf("ass=%s", filePath), nil
	}

	// Use subtitles filter for .srt and .vtt files
	var parts []string
	parts = append(parts, fmt.Sprintf("filename=%s", filePath))

	// Add style overrides if provided
	if sc.fontFamily != "" {
		parts = append(parts, fmt.Sprintf("force_style='FontName=%s'", sc.fontFamily))
	}
	if sc.fontSize > 0 {
		parts = append(parts, fmt.Sprintf("force_style='FontSize=%d'", sc.fontSize))
	}
	if sc.fontColor != "" {
		parts = append(parts, fmt.Sprintf("force_style='PrimaryColour=%s'", normalizeColor(sc.fontColor)))
	}

	if sc.charenc != "" {
		parts = append(parts, fmt.Sprintf("charenc=%s", sc.charenc))
	}

	return "subtitles=" + strings.Join(parts, ":"), nil
}
