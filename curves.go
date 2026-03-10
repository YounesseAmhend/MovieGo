package moviego

import (
	"fmt"
	"strconv"
)

// Curve defines the interpolation curve between start and end values.
type Curve int

const (
	Linear Curve = iota

	// Quadratic
	EaseIn
	EaseOut
	EaseInOut

	// Cubic
	CubicIn
	CubicOut
	CubicInOut

	// Quartic
	QuartIn
	QuartOut
	QuartInOut

	// Quintic
	QuintIn
	QuintOut
	QuintInOut

	// Sine
	SineIn
	SineOut
	SineInOut

	// Exponential
	ExpoIn
	ExpoOut
	ExpoInOut

	// Circular
	CircIn
	CircOut
	CircInOut

	// Back (overshoots then settles)
	BackIn
	BackOut
	BackInOut

	// Elastic (spring-like oscillation)
	ElasticIn
	ElasticOut
	ElasticInOut

	// Bounce (ball drop)
	BounceIn
	BounceOut
	BounceInOut

	// Spring (damped spring)
	Spring

	// Steps (discrete jumps)
	Steps
)

// Animation interpolates a single float from Start to End over a time range.
type Animation struct {
	Start     float64
	End       float64
	StartTime float64
	EndTime   float64
	Curve     Curve
}

// AnimatedPosition interpolates a position from Start to End.
type AnimatedPosition struct {
	Start     Position
	End       Position
	StartTime float64
	EndTime   float64
	Curve     Curve
}

// Oscillation creates periodic motion (shake, wiggle, pulse).
type Oscillation struct {
	Amplitude float64 // max displacement from center
	Frequency float64 // cycles per second
	StartTime float64
	EndTime   float64
	Decay     float64 // damping factor (0 = no decay, higher = fades faster)
}

// AnimatedColor animates color properties over time. Nil fields are not animated.
type AnimatedColor struct {
	Brightness *Animation // -1 to 1
	Contrast   *Animation // -1000 to 1000
	Saturation *Animation // 0 to 3
	Hue        *Animation // degrees
}

// toExpr compiles Animation to an FFmpeg expression using the given time variable.
func (a Animation) toExpr(timeVar string) string {
	if a.Start == a.End {
		return fmt.Sprintf("%.4f", a.Start)
	}
	duration := a.EndTime - a.StartTime
	if duration <= 0 {
		return fmt.Sprintf("%.4f", a.End)
	}
	// progress: 0 at StartTime, 1 at EndTime
	progress := fmt.Sprintf("(min(max((%s-%.4f)/%.4f,0),1))", timeVar, a.StartTime, duration)
	eased := applyCurve(progress, a.Curve)
	// lerp: Start + (End - Start) * eased
	return fmt.Sprintf("%.4f+(%.4f)*%s", a.Start, a.End-a.Start, eased)
}

// toExprDefault returns the expression using "t" as the time variable.
func (a Animation) toExprDefault() string {
	return a.toExpr("t")
}

// ToExprForTest returns the FFmpeg expression for testing.
func (a Animation) ToExprForTest(timeVar string) string {
	return a.toExpr(timeVar)
}

// toExpr compiles Oscillation to an FFmpeg expression.
func (o Oscillation) toExpr(timeVar string) string {
	// amp * sin(2*PI*freq*t) * exp(-decay*t) with enable guard
	sinPart := fmt.Sprintf("sin(2*PI*%.4f*%s)", o.Frequency, timeVar)
	osc := fmt.Sprintf("%.4f*%s", o.Amplitude, sinPart)
	if o.Decay > 0 {
		osc = fmt.Sprintf("%s*exp(-%.4f*%s)", osc, o.Decay, timeVar)
	}
	if o.StartTime > 0 || o.EndTime > 0 {
		enable := fmt.Sprintf("between(%s,%.4f,%.4f)", timeVar, o.StartTime, o.EndTime)
		return fmt.Sprintf("if(%s,%s,0)", enable, osc)
	}
	return osc
}

// toExprDefault returns the oscillation expression using "t".
func (o Oscillation) toExprDefault() string {
	return o.toExpr("t")
}

// ToExprForTest returns the oscillation expression for testing.
func (o Oscillation) ToExprForTest(timeVar string) string {
	return o.toExpr(timeVar)
}

// toExprX compiles AnimatedPosition to an FFmpeg expression for the X coordinate.
func (ap AnimatedPosition) toExprX(timeVar string) string {
	startX := parsePositionValue(ap.Start.X)
	endX := parsePositionValue(ap.End.X)
	a := Animation{
		Start:     startX,
		End:       endX,
		StartTime: ap.StartTime,
		EndTime:   ap.EndTime,
		Curve:     ap.Curve,
	}
	return a.toExpr(timeVar)
}

// toExprY compiles AnimatedPosition to an FFmpeg expression for the Y coordinate.
func (ap AnimatedPosition) toExprY(timeVar string) string {
	startY := parsePositionValue(ap.Start.Y)
	endY := parsePositionValue(ap.End.Y)
	a := Animation{
		Start:     startY,
		End:       endY,
		StartTime: ap.StartTime,
		EndTime:   ap.EndTime,
		Curve:     ap.Curve,
	}
	return a.toExpr(timeVar)
}

// toExprDefault returns both X and Y expressions using "t".
func (ap AnimatedPosition) toExprDefault() (xExpr, yExpr string) {
	return ap.toExprX("t"), ap.toExprY("t")
}

// ToExprXForTest returns the X expression for testing.
func (ap AnimatedPosition) ToExprXForTest(timeVar string) string {
	return ap.toExprX(timeVar)
}

// ToExprYForTest returns the Y expression for testing.
func (ap AnimatedPosition) ToExprYForTest(timeVar string) string {
	return ap.toExprY(timeVar)
}

func parsePositionValue(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// applyCurve maps progress (0-1) through the curve to eased progress.
func applyCurve(progress string, c Curve) string {
	p := progress
	switch c {
	case Linear:
		return p

	case EaseIn:
		return fmt.Sprintf("pow(%s,2)", p)
	case EaseOut:
		return fmt.Sprintf("(1-pow(1-%s,2))", p)
	case EaseInOut:
		return fmt.Sprintf("if(lt(%s,0.5),2*pow(%s,2),(2-pow(-2*%s+2,2))/2)", p, p, p)

	case CubicIn:
		return fmt.Sprintf("pow(%s,3)", p)
	case CubicOut:
		return fmt.Sprintf("(1-pow(1-%s,3))", p)
	case CubicInOut:
		return fmt.Sprintf("if(lt(%s,0.5),4*pow(%s,3),(1-pow(-2*%s+2,3)/2))", p, p, p)

	case QuartIn:
		return fmt.Sprintf("pow(%s,4)", p)
	case QuartOut:
		return fmt.Sprintf("(1-pow(1-%s,4))", p)
	case QuartInOut:
		return fmt.Sprintf("if(lt(%s,0.5),8*pow(%s,4),(1-pow(-2*%s+2,4)/2))", p, p, p)

	case QuintIn:
		return fmt.Sprintf("pow(%s,5)", p)
	case QuintOut:
		return fmt.Sprintf("(1-pow(1-%s,5))", p)
	case QuintInOut:
		return fmt.Sprintf("if(lt(%s,0.5),16*pow(%s,5),(1-pow(-2*%s+2,5)/2))", p, p, p)

	case SineIn:
		return fmt.Sprintf("(1-cos(PI*%s/2))", p)
	case SineOut:
		return fmt.Sprintf("sin(PI*%s/2)", p)
	case SineInOut:
		return fmt.Sprintf("(1-cos(PI*%s))/2", p)

	case ExpoIn:
		return fmt.Sprintf("if(lt(%s,0.0001),0,pow(2,10*%s-10))", p, p)
	case ExpoOut:
		return fmt.Sprintf("if(gt(%s,0.9999),1,1-pow(2,-10*%s))", p, p)
	case ExpoInOut:
		return fmt.Sprintf("if(lt(%s,0.0001),0,if(gt(%s,0.9999),1,if(lt(%s,0.5),pow(2,20*%s-10)/2,(2-pow(2,-20*%s+10))/2)))", p, p, p, p, p)

	case CircIn:
		return fmt.Sprintf("(1-sqrt(1-pow(%s,2)))", p)
	case CircOut:
		return fmt.Sprintf("sqrt(1-pow(%s-1,2))", p)
	case CircInOut:
		return fmt.Sprintf("if(lt(%s,0.5),(1-sqrt(1-pow(2*%s,2)))/2,(sqrt(1-pow(-2*%s+2,2))+1)/2)", p, p, p)

	case BackIn:
		return fmt.Sprintf("(2.70158*pow(%s,3)-1.70158*pow(%s,2))", p, p)
	case BackOut:
		return fmt.Sprintf("(1+2.70158*pow(%s-1,3)+1.70158*pow(%s-1,2))", p, p)
	case BackInOut:
		return fmt.Sprintf("if(lt(%[1]s,0.5),(pow(2*%[1]s,2)*((2.59491*%[1]s+0)*2*%[1]s-2.59491))/2,(pow(2*%[1]s-2,2)*((2.59491*(%[1]s-1)+1.59491)*2*(%[1]s-1)+2.59491)+2)/2)", p)

	case ElasticIn:
		return fmt.Sprintf("if(lt(%s,0.0001),0,if(gt(%s,0.9999),1,-pow(2,10*%s-10)*sin((%s*10-10.75)*(2*PI/3))))", p, p, p, p)
	case ElasticOut:
		return fmt.Sprintf("if(lt(%s,0.0001),0,if(gt(%s,0.9999),1,pow(2,-10*%s)*sin((%s*10-0.75)*(2*PI/3))+1))", p, p, p, p)
	case ElasticInOut:
		return fmt.Sprintf("if(lt(%s,0.0001),0,if(gt(%s,0.9999),1,if(lt(%s,0.5),-(pow(2,20*%s-10)*sin((20*%s-11.125)*(2*PI/4.5)))/2,(pow(2,-20*%s+10)*sin((20*%s-11.125)*(2*PI/4.5)))/2+1)))", p, p, p, p, p, p, p)

	case BounceIn:
		return fmt.Sprintf("(1-(if(lt(%[1]s,0.36363636),7.5625*pow(%[1]s,2),if(lt(%[1]s,0.72727272),7.5625*pow(%[1]s-0.54545454,2)+0.75,if(lt(%[1]s,0.90909090),7.5625*pow(%[1]s-0.81818181,2)+0.9375,7.5625*pow(%[1]s-0.95454545,2)+0.984375)))))", p)
	case BounceOut:
		return fmt.Sprintf("if(lt(%[1]s,0.36363636),7.5625*pow(%[1]s,2),if(lt(%[1]s,0.72727272),7.5625*pow(%[1]s-0.54545454,2)+0.75,if(lt(%[1]s,0.90909090),7.5625*pow(%[1]s-0.81818181,2)+0.9375,7.5625*pow(%[1]s-0.95454545,2)+0.984375)))", p)
	case BounceInOut:
		return fmt.Sprintf("if(lt(%[1]s,0.5),(1-(if(lt(1-2*%[1]s,0.36363636),7.5625*pow(1-2*%[1]s,2),if(lt(1-2*%[1]s,0.72727272),7.5625*pow(1-2*%[1]s-0.54545454,2)+0.75,if(lt(1-2*%[1]s,0.90909090),7.5625*pow(1-2*%[1]s-0.81818181,2)+0.9375,7.5625*pow(1-2*%[1]s-0.95454545,2)+0.984375)))))/2,(if(lt(2*%[1]s-1,0.36363636),7.5625*pow(2*%[1]s-1,2),if(lt(2*%[1]s-1,0.72727272),7.5625*pow(2*%[1]s-1-0.54545454,2)+0.75,if(lt(2*%[1]s-1,0.90909090),7.5625*pow(2*%[1]s-1-0.81818181,2)+0.9375,7.5625*pow(2*%[1]s-1-0.95454545,2)+0.984375)))+1)/2)", p)

	case Spring:
		// Damped spring: 1 - exp(-t*5) * cos(t*2*PI)
		return fmt.Sprintf("(1-exp(-%s*5)*cos(%s*2*PI))", p, p)

	case Steps:
		// 5 steps
		return fmt.Sprintf("floor(%s*5)/5", p)

	default:
		return p
	}
}
