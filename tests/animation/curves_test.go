package animation_test

import (
	"strings"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
)

func TestLinearExpr(t *testing.T) {
	a := moviego.Animation{
		Start:     0,
		End:       100,
		StartTime: 0,
		EndTime:   2,
		Curve:     moviego.Linear,
	}
	expr := a.ToExprForTest("t")
	if expr == "" {
		t.Fatal("expected non-empty expression for linear animation")
	}
	if !strings.Contains(expr, "100") {
		t.Errorf("expected expression to contain end value 100, got %q", expr)
	}
	if !strings.Contains(expr, "0") {
		t.Errorf("expected expression to contain start value 0, got %q", expr)
	}
}

func TestEaseInExpr(t *testing.T) {
	a := moviego.Animation{
		Start:     0,
		End:       1,
		StartTime: 0,
		EndTime:   1,
		Curve:     moviego.EaseIn,
	}
	expr := a.ToExprForTest("t")
	if expr == "" {
		t.Fatal("expected non-empty expression for ease-in")
	}
	if !strings.Contains(expr, "pow") {
		t.Errorf("ease-in should use pow, got %q", expr)
	}
}

func TestEaseOutExpr(t *testing.T) {
	a := moviego.Animation{
		Start:     0,
		End:       1,
		StartTime: 0,
		EndTime:   1,
		Curve:     moviego.EaseOut,
	}
	expr := a.ToExprForTest("t")
	if expr == "" {
		t.Fatal("expected non-empty expression for ease-out")
	}
}

func TestAllCurvesProduceValidExpr(t *testing.T) {
	curves := []moviego.Curve{
		moviego.Linear, moviego.EaseIn, moviego.EaseOut, moviego.EaseInOut,
		moviego.CubicIn, moviego.CubicOut, moviego.CubicInOut,
		moviego.QuartIn, moviego.QuartOut, moviego.QuartInOut,
		moviego.QuintIn, moviego.QuintOut, moviego.QuintInOut,
		moviego.SineIn, moviego.SineOut, moviego.SineInOut,
		moviego.ExpoIn, moviego.ExpoOut, moviego.ExpoInOut,
		moviego.CircIn, moviego.CircOut, moviego.CircInOut,
		moviego.BackIn, moviego.BackOut, moviego.BackInOut,
		moviego.ElasticIn, moviego.ElasticOut, moviego.ElasticInOut,
		moviego.BounceIn, moviego.BounceOut, moviego.BounceInOut,
		moviego.Spring, moviego.Steps,
	}
	a := moviego.Animation{Start: 0, End: 100, StartTime: 0, EndTime: 2, Curve: moviego.Linear}
	for _, c := range curves {
		a.Curve = c
		expr := a.ToExprForTest("t")
		if expr == "" {
			t.Errorf("curve %v produced empty expression", c)
		}
	}
}

func TestOscillationExpr(t *testing.T) {
	o := moviego.Oscillation{
		Amplitude: 10,
		Frequency: 2,
		StartTime: 0,
		EndTime:   5,
		Decay:     0.1,
	}
	expr := o.ToExprForTest("t")
	if expr == "" {
		t.Fatal("expected non-empty oscillation expression")
	}
	if !strings.Contains(expr, "sin") {
		t.Errorf("oscillation should use sin, got %q", expr)
	}
}

func TestAnimatedPositionExpr(t *testing.T) {
	ap := moviego.AnimatedPosition{
		Start:     moviego.Position{X: "0", Y: "0"},
		End:       moviego.Position{X: "100", Y: "200"},
		StartTime: 0,
		EndTime:   2,
		Curve:     moviego.Linear,
	}
	xExpr, yExpr := ap.ToExprXForTest("t"), ap.ToExprYForTest("t")
	if xExpr == "" || yExpr == "" {
		t.Fatalf("expected both X and Y expressions, got x=%q y=%q", xExpr, yExpr)
	}
}

func TestSingleValueAnimation(t *testing.T) {
	a := moviego.Animation{
		Start:     42,
		End:       42,
		StartTime: 0,
		EndTime:   2,
		Curve:     moviego.Linear,
	}
	expr := a.ToExprForTest("t")
	if expr != "42.0000" {
		t.Errorf("single value animation should return constant, got %q", expr)
	}
}
