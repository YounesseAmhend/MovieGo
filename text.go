package moviego

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// TextExpansion controls how text is expanded in the drawtext filter.
type TextExpansion string

const (
	ExpansionNone   TextExpansion = "none"
	ExpansionNormal TextExpansion = "normal"
)

// Shadow controls the drop shadow behind text.
type Shadow struct {
	X     int    // horizontal offset (default: 0)
	Y     int    // vertical offset (default: 0)
	Color string // shadow color (default: "black")
}

// Stroke controls the outline/border around each glyph.
type Stroke struct {
	Width int    // outline thickness in pixels (default: 0)
	Color string // outline color (default: "black")
}

// Background controls the box drawn behind text.
type Background struct {
	Enabled bool   // show background box
	Color   string // box fill color, e.g. "black@0.5" (default: "white")
	Padding string // padding inside box: "10" or per-side "10|20|30|40"
	Width   int    // explicit box width (0 = auto-fit text)
	Height  int    // explicit box height (0 = auto-fit text)
}

// Layout controls multi-line text arrangement.
type Layout struct {
	LineSpacing int    // extra spacing between lines in pixels
	TabSize     int    // tab character width in spaces (default: 4)
	Align       string // multi-line alignment: "L"/"C"/"R" + "T"/"M"/"B"
	YAlign      string // what y refers to: "text", "baseline", "font"
	FixBounds   bool   // prevent text from going outside the video frame
}

// TypewriterParams configures the typewriter (character-by-character) effect.
type TypewriterParams struct {
	CharDelay float64 // seconds between each character appearing
	StartTime float64 // when typing starts
	Cursor    string  // cursor character (e.g. "|"), empty = no cursor
}

// TextClip represents a text overlay to be drawn on a video.
type TextClip struct {
	// Content
	Text     string // text string to draw (required if TextFile is empty)
	TextFile string // path to text file (alternative to Text)

	// Font -- FontFamily is auto-detected:
	//   ends with .ttf/.otf/.woff or contains path separator -> fontfile
	//   otherwise -> font family name via fontconfig
	FontFamily string // e.g. "Arial", "Sans", or "/path/to/font.ttf"
	FontSize   int    // font size in pixels (default: 24)
	FontColor  string // "white", "0xFF0000", "black@0.5" (default: "white")

	// Position & Timing
	Position  Position // X, Y as FFmpeg drawtext expressions
	StartTime float64  // when text appears in seconds (0 = from start)
	EndTime   float64  // when text disappears in seconds (0 = until end)

	// Opacity
	Opacity float64 // 0.0 (transparent) to 1.0 (opaque), default 1.0

	// Appearance sub-structs
	Background Background
	Stroke     Stroke
	Shadow     Shadow
	Layout     Layout

	// Advanced
	TextShaping bool          // enable RTL/Arabic text shaping (default: true)
	Expansion   TextExpansion // text expansion mode (default: ExpansionNormal)

	// Animations (if set, override static Position/Opacity)
	AnimatePosition *AnimatedPosition // animated X,Y position
	AnimateOpacity  *Animation        // animated alpha (0-1)
	Typewriter      *TypewriterParams // character-by-character reveal
}

// drawtext position expression constants
const (
	posCenterX    = "(w-tw)/2"
	posCenterY    = "(h-th)/2"
	posTopY       = "10"
	posBottomY    = "h-th-10"
	posLeftX      = "10"
	posRightX     = "w-tw-10"
)

// TextCenter returns a position that centers the text on the video.
func TextCenter() Position {
	return Position{X: posCenterX, Y: posCenterY}
}

// TextTopLeft returns a position at the top-left corner with 10px margin.
func TextTopLeft() Position {
	return Position{X: posLeftX, Y: posTopY}
}

// TextTopCenter returns a position at the top center.
func TextTopCenter() Position {
	return Position{X: posCenterX, Y: posTopY}
}

// TextTopRight returns a position at the top-right corner.
func TextTopRight() Position {
	return Position{X: posRightX, Y: posTopY}
}

// TextBottomLeft returns a position at the bottom-left corner.
func TextBottomLeft() Position {
	return Position{X: posLeftX, Y: posBottomY}
}

// TextBottomCenter returns a position at the bottom center.
func TextBottomCenter() Position {
	return Position{X: posCenterX, Y: posBottomY}
}

// TextBottomRight returns a position at the bottom-right corner.
func TextBottomRight() Position {
	return Position{X: posRightX, Y: posBottomY}
}

// isFontFile returns true if FontFamily looks like a font file path.
func isFontFile(s string) bool {
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	return strings.HasSuffix(lower, ".ttf") ||
		strings.HasSuffix(lower, ".otf") ||
		strings.HasSuffix(lower, ".woff") ||
		strings.HasSuffix(lower, ".woff2") ||
		strings.Contains(s, string(filepath.Separator)) ||
		strings.Contains(s, "/")
}

// escapeDrawText escapes special characters in drawtext filter values.
func escapeDrawText(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, ":", `\:`)
	return s
}

// buildDrawTextFilter constructs the FFmpeg drawtext filter string from the TextClip.
func (tc TextClip) buildDrawTextFilter(videoDuration float64) string {
	parts := tc.appendContentParts(nil)
	parts = tc.appendFontParts(parts)
	parts = tc.appendPositionParts(parts)
	parts = tc.appendTimingParts(parts, videoDuration)
	parts = tc.appendAppearanceParts(parts)
	parts = tc.appendAdvancedParts(parts)
	return "drawtext=" + strings.Join(parts, ":")
}

func (tc TextClip) appendContentParts(parts []string) []string {
	if tc.TextFile != "" {
		return append(parts, "textfile='"+escapeDrawText(tc.TextFile)+"'")
	}
	if tc.Text != "" {
		return append(parts, "text='"+escapeDrawText(tc.Text)+"'")
	}
	return parts
}

func (tc TextClip) appendFontParts(parts []string) []string {
	if tc.FontFamily != "" {
		if isFontFile(tc.FontFamily) {
			parts = append(parts, "fontfile='"+escapeDrawText(tc.FontFamily)+"'")
		} else {
			parts = append(parts, "font='"+escapeDrawText(tc.FontFamily)+"'")
		}
	}
	if tc.FontSize > 0 {
		parts = append(parts, fmt.Sprintf("fontsize=%d", tc.FontSize))
	}
	if tc.FontColor != "" {
		parts = append(parts, "fontcolor="+tc.FontColor)
	}
	return parts
}

func (tc TextClip) appendPositionParts(parts []string) []string {
	if tc.AnimatePosition != nil {
		xExpr := tc.AnimatePosition.toExprX("t")
		yExpr := tc.AnimatePosition.toExprY("t")
		parts = append(parts, "x='"+xExpr+"'", "y='"+yExpr+"'")
	} else if tc.Position.X != "" || tc.Position.Y != "" {
		if tc.Position.X != "" {
			parts = append(parts, "x="+tc.Position.X)
		}
		if tc.Position.Y != "" {
			parts = append(parts, "y="+tc.Position.Y)
		}
	}
	return parts
}

func (tc TextClip) appendTimingParts(parts []string, videoDuration float64) []string {
	endTime := tc.EndTime
	if endTime <= 0 || endTime > videoDuration {
		endTime = videoDuration
	}
	if tc.StartTime > 0 || tc.EndTime > 0 {
		parts = append(parts, fmt.Sprintf("enable='between(t,%.4f,%.4f)'", tc.StartTime, endTime))
	}
	return parts
}

func (tc TextClip) appendAppearanceParts(parts []string) []string {
	parts = tc.appendOpacityPart(parts)
	parts = tc.appendBackgroundParts(parts)
	parts = tc.appendStrokeParts(parts)
	parts = tc.appendShadowParts(parts)
	parts = tc.appendLayoutParts(parts)
	return parts
}

func (tc TextClip) appendOpacityPart(parts []string) []string {
	if tc.AnimateOpacity != nil {
		expr := tc.AnimateOpacity.toExprDefault()
		return append(parts, "alpha='"+expr+"'")
	}
	if tc.Opacity > 0 && tc.Opacity < 1 {
		return append(parts, fmt.Sprintf("alpha=%.4f", tc.Opacity))
	}
	return parts
}

func (tc TextClip) appendBackgroundParts(parts []string) []string {
	if !tc.Background.Enabled {
		return parts
	}
	parts = append(parts, "box=1")
	if tc.Background.Color != "" {
		parts = append(parts, "boxcolor="+tc.Background.Color)
	}
	if tc.Background.Padding != "" {
		parts = append(parts, "boxborderw="+tc.Background.Padding)
	}
	if tc.Background.Width > 0 {
		parts = append(parts, fmt.Sprintf("boxw=%d", tc.Background.Width))
	}
	if tc.Background.Height > 0 {
		parts = append(parts, fmt.Sprintf("boxh=%d", tc.Background.Height))
	}
	return parts
}

func (tc TextClip) appendStrokeParts(parts []string) []string {
	if tc.Stroke.Width <= 0 {
		return parts
	}
	parts = append(parts, fmt.Sprintf("borderw=%d", tc.Stroke.Width))
	if tc.Stroke.Color != "" {
		parts = append(parts, "bordercolor="+tc.Stroke.Color)
	}
	return parts
}

func (tc TextClip) appendShadowParts(parts []string) []string {
	if tc.Shadow.X == 0 && tc.Shadow.Y == 0 {
		return parts
	}
	parts = append(parts, fmt.Sprintf("shadowx=%d", tc.Shadow.X))
	parts = append(parts, fmt.Sprintf("shadowy=%d", tc.Shadow.Y))
	if tc.Shadow.Color != "" {
		parts = append(parts, "shadowcolor="+tc.Shadow.Color)
	}
	return parts
}

func (tc TextClip) appendLayoutParts(parts []string) []string {
	if tc.Layout.LineSpacing != 0 {
		parts = append(parts, fmt.Sprintf("line_spacing=%d", tc.Layout.LineSpacing))
	}
	if tc.Layout.TabSize > 0 {
		parts = append(parts, fmt.Sprintf("tabsize=%d", tc.Layout.TabSize))
	}
	if tc.Layout.Align != "" {
		parts = append(parts, "text_align="+tc.Layout.Align)
	}
	if tc.Layout.YAlign != "" {
		parts = append(parts, "y_align="+tc.Layout.YAlign)
	}
	if tc.Layout.FixBounds {
		parts = append(parts, "fix_bounds=1")
	}
	return parts
}

func (tc TextClip) appendAdvancedParts(parts []string) []string {
	if !tc.TextShaping {
		parts = append(parts, "text_shaping=0")
	}
	if tc.Expansion != "" && tc.Expansion != ExpansionNormal {
		parts = append(parts, "expansion="+string(tc.Expansion))
	}
	return parts
}

// AddText adds a single text overlay to the video.
func (v *Video) AddText(clip TextClip) (*Video, error) {
	if clip.Text == "" && clip.TextFile == "" {
		return nil, fmt.Errorf("AddText: Text or TextFile is required")
	}
	if clip.Typewriter != nil {
		return v.addTextTypewriter(clip)
	}
	filter := clip.buildDrawTextFilter(v.duration)
	return v.videoFilter(filter)
}

// addTextTypewriter adds text with character-by-character typewriter effect.
func (v *Video) addTextTypewriter(clip TextClip) (*Video, error) {
	tw := clip.Typewriter
	if tw.CharDelay <= 0 {
		tw = &TypewriterParams{CharDelay: 0.1, StartTime: clip.StartTime, Cursor: tw.Cursor}
	}
	text := clip.Text
	if text == "" {
		return nil, fmt.Errorf("AddText typewriter: Text is required (TextFile not supported for typewriter)")
	}
	// Approximate character width as 0.6 * font size for positioning
	charWidth := clip.FontSize
	if charWidth <= 0 {
		charWidth = 24
	}
	charWidth = int(float64(charWidth) * 0.6)
	if charWidth < 8 {
		charWidth = 8
	}
	baseX := 10
	baseY := 10
	if clip.Position.X != "" {
		if x, err := strconv.ParseFloat(clip.Position.X, 64); err == nil {
			baseX = int(x)
		}
	}
	if clip.Position.Y != "" {
		if y, err := strconv.ParseFloat(clip.Position.Y, 64); err == nil {
			baseY = int(y)
		}
	}
	var err error
	for i, r := range text {
		charClip := clip
		charClip.Text = string(r)
		charClip.Typewriter = nil
		charClip.AnimatePosition = nil
		charClip.AnimateOpacity = nil
		charClip.Position = Position{X: fmt.Sprintf("%d", baseX+i*charWidth), Y: fmt.Sprintf("%d", baseY)}
		charClip.StartTime = tw.StartTime + float64(i)*tw.CharDelay
		charClip.EndTime = 0

		filter := charClip.buildDrawTextFilter(v.duration)
		v, err = v.videoFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	if tw.Cursor != "" {
		cursorClip := clip
		cursorClip.Text = tw.Cursor
		cursorClip.Typewriter = nil
		cursorClip.AnimatePosition = nil
		cursorClip.AnimateOpacity = nil
		cursorClip.Position = Position{X: fmt.Sprintf("%d", baseX+len(text)*charWidth), Y: fmt.Sprintf("%d", baseY)}
		cursorClip.StartTime = tw.StartTime
		cursorClip.EndTime = 0
		// Cursor blinks - show for 0.5s, hide for 0.5s (simplified: always visible)
		filter := cursorClip.buildDrawTextFilter(v.duration)
		v, err = v.videoFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

// AddTexts adds multiple text overlays to the video in one call.
func (v *Video) AddTexts(clips []*TextClip) (*Video, error) {
	var err error
	for _, clip := range clips {
		if clip == nil {
			return nil, fmt.Errorf("AddTexts: nil TextClip in slice")
		}
		v, err = v.AddText(*clip)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}
