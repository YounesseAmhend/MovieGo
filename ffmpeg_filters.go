package moviego

import (
	"fmt"
	"runtime"
)

// ============================================================================
// FFmpeg Filter Chain Utilities
// ============================================================================
//
// This file provides centralized, cross-platform utilities for building
// FFmpeg filter chains. It handles proper escaping of filter options that
// works consistently on Windows, Linux, and macOS.
//
// Key Design Points:
// - FFmpeg filter syntax is identical across all platforms
// - The backslash escape (\:) is interpreted by FFmpeg's filter parser
// - We use exec.Command() which passes args directly to FFmpeg (no shell)
// - This eliminates OS-specific shell escaping differences

// ============================================================================
// Linear Interpolation Expression Builder
// ============================================================================

// linearExpr builds a linear interpolation expression for FFmpeg filters
// This creates time-based animations that smoothly transition from one value to another
//
// Parameters:
//   - tExpr: time expression (usually "t" or "(t-offset)")
//   - start: animation start time in seconds
//   - duration: animation duration in seconds
//   - from: starting value
//   - to: ending value
//
// Returns: FFmpeg expression string with clamping at boundaries
//
// Example: linearExpr("t", 0, 2, 100, 200) creates an expression that:
//   - Returns 100 when t < 0
//   - Returns 200 when t >= 2
//   - Interpolates linearly between 100 and 200 when 0 <= t < 2
func linearExpr(tExpr string, start, duration, from, to float64) string {
	if duration <= 0 {
		return fmt.Sprintf("%.3f", from)
	}

	// Clamp expression: if(t < start, from, if(t >= start+duration, to, interpolated))
	// Interpolation: from + (to - from) * ((t - start) / duration)
	endTime := start + duration

	interpolated := fmt.Sprintf("%.3f+(%.3f-%.3f)*((%s-%.3f)/%.3f)", from, to, from, tExpr, start, duration)

	return fmt.Sprintf("if(lt(%s,%.3f),%.3f,if(gte(%s,%.3f),%.3f,%s))",
		tExpr, start, from, tExpr, endTime, to, interpolated)
}

// ============================================================================
// OS-Aware Filter Escaping
// ============================================================================

// getFilterEscapeChar returns the escape sequence for FFmpeg filter options
// This is used to escape colons in filter parameters (e.g., "rotate=angle\:c=none")
//
// Current implementation: All platforms (Windows, Linux, macOS) use "\\" which
// becomes "\:" in the final string. This is the FFmpeg standard.
//
// The function uses runtime.GOOS to detect the platform, allowing for future
// platform-specific adjustments if needed (though unlikely given FFmpeg's
// consistent cross-platform behavior).
func getFilterEscapeChar() string {
	// FFmpeg uses the same filter syntax on all platforms
	// The backslash escape is interpreted by FFmpeg's filter parser,
	// not by the OS shell (since we use exec.Command)
	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		return "\\:"
	default:
		// Default to FFmpeg standard for any other Unix-like OS
		return "\\:"
	}
}

// ============================================================================
// Animation Filter Builders
// ============================================================================

// buildScaleAnimationFilter creates a scale animation filter with proper escaping
// This filter scales content dynamically over time using a scale expression
//
// Parameters:
//   - scaleExpr: FFmpeg expression for the scale factor (e.g., from linearExpr)
//
// Returns: Properly escaped FFmpeg scale filter string
//
// The filter format:
//   - scale=w='trunc(iw*expr)':h='trunc(ih*expr)':eval=frame
//   - trunc() ensures integer dimensions
//   - eval=frame enables per-frame evaluation of expressions with time variables
//   - Single quotes protect complex expressions from being misinterpreted
//
// Example: buildScaleAnimationFilter("1.5") → "scale=w='trunc(iw*1.5)':h='trunc(ih*1.5)':eval=frame"
func buildScaleAnimationFilter(scaleExpr string) string {
	return fmt.Sprintf("scale=w='trunc(iw*%s)':h='trunc(ih*%s)':eval=frame", scaleExpr, scaleExpr)
}

// buildRotationAnimationFilter creates a rotation animation filter with proper escaping
// This filter rotates content dynamically over time using an angle expression
//
// Parameters:
//   - angleExpr: FFmpeg expression for the rotation angle in radians (e.g., from linearExpr)
//
// Returns: Properly escaped FFmpeg rotate filter string
//
// The filter format:
//   - rotate='expr':fillcolor=none:ow='hypot(iw,ih)':oh='ow'
//   - fillcolor=none makes the background transparent (preserves alpha)
//   - ow/oh uses hypot(iw,ih) to expand output dimensions to fit rotated content at any angle
//   - Single quotes protect complex expressions from being misinterpreted
//
// Example: buildRotationAnimationFilter("PI/4") → "rotate='PI/4':fillcolor=none:ow='hypot(iw,ih)':oh='ow'"
func buildRotationAnimationFilter(angleExpr string) string {
	return fmt.Sprintf("rotate='%s':fillcolor=none:ow='hypot(iw,ih)':oh='ow'", angleExpr)
}
