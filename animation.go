package moviego

import "fmt"

// AnimatedRotate applies a time-based rotation. Angle is in radians.
func (v *Video) AnimatedRotate(a Animation) (*Video, error) {
	expr := a.toExprDefault()
	filter := fmt.Sprintf("rotate='%s'", expr)
	return v.videoFilter(filter)
}

// AnimatedScale applies a time-based scale. Start/End are scale ratios (1.0 = 100%).
func (v *Video) AnimatedScale(a Animation) (*Video, error) {
	if a.Start <= 0 || a.End <= 0 {
		return nil, fmt.Errorf("AnimatedScale: start and end must be positive")
	}
	expr := a.toExprDefault()
	filter := fmt.Sprintf("scale=w='trunc(iw*(%s)/2)*2':h='trunc(ih*(%s)/2)*2':eval=frame", expr, expr)
	return v.videoFilter(filter)
}

// AnimatedBlur applies a time-based blur. Start/End are blur radius values (0 = sharp).
// Uses boxblur for expression support; radius is in pixels (uses floor for integer radius).
func (v *Video) AnimatedBlur(a Animation) (*Video, error) {
	if a.Start < 0 || a.End < 0 {
		return nil, fmt.Errorf("AnimatedBlur: start and end must be non-negative")
	}
	expr := a.toExprDefault()
	filter := fmt.Sprintf("boxblur=lr='floor(%s+0.5)':lp=1", expr)
	return v.videoFilter(filter)
}

// AnimatedColor applies time-based color adjustments. Nil fields are not animated.
func (v *Video) AnimatedColor(ac AnimatedColor) (*Video, error) {
	var err error
	if ac.Brightness != nil {
		v, err = v.animatedEqPart("brightness", *ac.Brightness, -1, 1)
		if err != nil {
			return nil, err
		}
	}
	if ac.Contrast != nil {
		v, err = v.animatedEqPart("contrast", *ac.Contrast, -1000, 1000)
		if err != nil {
			return nil, err
		}
	}
	if ac.Saturation != nil {
		v, err = v.animatedEqPart("saturation", *ac.Saturation, 0, 3)
		if err != nil {
			return nil, err
		}
	}
	if ac.Hue != nil {
		expr := ac.Hue.toExprDefault()
		filter := fmt.Sprintf("hue=h='%s'", expr)
		v, err = v.videoFilter(filter)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (v *Video) animatedEqPart(name string, a Animation, minVal, maxVal float64) (*Video, error) {
	if a.Start < minVal || a.Start > maxVal || a.End < minVal || a.End > maxVal {
		return nil, fmt.Errorf("AnimatedColor %s: values must be in [%v, %v]", name, minVal, maxVal)
	}
	expr := a.toExprDefault()
	filter := fmt.Sprintf("eq=%s='%s'", name, expr)
	return v.videoFilter(filter)
}

// Shake applies a position oscillation (camera shake effect).
func (v *Video) Shake(o Oscillation) (*Video, error) {
	if o.Amplitude <= 0 {
		return nil, fmt.Errorf("Shake: amplitude must be positive")
	}
	amp := int(o.Amplitude)
	if amp < 1 {
		amp = 1
	}
	oscExpr := o.toExprDefault()
	padFilter := fmt.Sprintf("pad=width=iw+%d:height=ih+%d:x=%d:y=%d", amp*2, amp*2, amp, amp)
	cropFilter := fmt.Sprintf("crop=w='trunc((iw-%d)/2)*2':h='trunc((ih-%d)/2)*2':x='%d+%s':y='%d+%s'", amp*2, amp*2, amp, oscExpr, amp, oscExpr)
	combined := padFilter + "," + cropFilter
	return v.videoFilter(combined)
}

// Wiggle applies a rotation oscillation.
func (v *Video) Wiggle(o Oscillation) (*Video, error) {
	expr := o.toExprDefault()
	filter := fmt.Sprintf("rotate='%s'", expr)
	return v.videoFilter(filter)
}

// Pulse applies a scale oscillation (zoom in/out effect).
func (v *Video) Pulse(o Oscillation) (*Video, error) {
	expr := o.toExprDefault()
	// scale factor = 1 + oscillation (osc can be negative)
	filter := fmt.Sprintf("scale=w='trunc(iw*(1+%s)/2)*2':h='trunc(ih*(1+%s)/2)*2':eval=frame", expr, expr)
	return v.videoFilter(filter)
}

// ZoomPan applies a Ken Burns style zoom and pan effect.
func (v *Video) ZoomPan(params ZoomPanParams) (*Video, error) {
	if params.Duration <= 0 {
		return nil, fmt.Errorf("ZoomPan: duration must be positive")
	}
	fps := int(v.fps)
	if params.FPS > 0 {
		fps = params.FPS
	}
	frames := int(params.Duration * float64(fps))
	if frames < 1 {
		frames = 1
	}
	zoomExpr := params.Zoom.toExpr("on/" + fmt.Sprintf("%d", fps))
	xExpr := params.Pan.toExprX("on/" + fmt.Sprintf("%d", fps))
	yExpr := params.Pan.toExprY("on/" + fmt.Sprintf("%d", fps))
	filter := fmt.Sprintf("zoompan=z='%s':x='%s':y='%s':d=%d:s=%dx%d:fps=%d",
		zoomExpr, xExpr, yExpr, frames, v.width, v.height, fps)
	zoomed, err := v.videoFilter(filter)
	if err != nil {
		return nil, err
	}
	zoomed.duration = params.Duration
	zoomed.frames = uint64(frames)
	return zoomed, nil
}

// ZoomPanParams holds parameters for the ZoomPan (Ken Burns) effect.
type ZoomPanParams struct {
	Zoom     Animation
	Pan      AnimatedPosition
	Duration float64
	FPS      int
}
