package animation_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestAnimatedRotate(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	anim := moviego.Animation{
		Start:     0,
		End:       2 * math.Pi,
		StartTime: 0,
		EndTime:   3,
		Curve:     moviego.Linear,
	}
	rotated, err := cut.AnimatedRotate(anim)
	if err != nil {
		t.Fatalf("Failed to animate rotate: %v", err)
	}
	outputPath := filepath.Join("output", "animated_rotate.mp4")
	if err := rotated.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestAnimatedScale(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	anim := moviego.Animation{
		Start:     1.0,
		End:       1.5,
		StartTime: 0,
		EndTime:   3,
		Curve:     moviego.EaseInOut,
	}
	scaled, err := cut.AnimatedScale(anim)
	if err != nil {
		t.Fatalf("Failed to animate scale: %v", err)
	}
	outputPath := filepath.Join("output", "animated_scale.mp4")
	if err := scaled.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestAnimatedBlur(t *testing.T) {
	t.Skip("boxblur lr expression may not be supported in all FFmpeg versions")
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	anim := moviego.Animation{
		Start:     0,
		End:       5,
		StartTime: 0,
		EndTime:   3,
		Curve:     moviego.Linear,
	}
	blurred, err := cut.AnimatedBlur(anim)
	if err != nil {
		t.Fatalf("Failed to animate blur: %v", err)
	}
	outputPath := filepath.Join("output", "animated_blur.mp4")
	if err := blurred.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestAnimatedColor(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	brightAnim := moviego.Animation{Start: 0, End: 0.3, StartTime: 0, EndTime: 3, Curve: moviego.Linear}
	ac := moviego.AnimatedColor{Brightness: &brightAnim}
	colored, err := cut.AnimatedColor(ac)
	if err != nil {
		t.Fatalf("Failed to animate color: %v", err)
	}
	outputPath := filepath.Join("output", "animated_color.mp4")
	if err := colored.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestShake(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	osc := moviego.Oscillation{
		Amplitude: 5,
		Frequency: 8,
		StartTime: 0,
		EndTime:   3,
		Decay:     0,
	}
	shaken, err := cut.Shake(osc)
	if err != nil {
		t.Fatalf("Failed to shake: %v", err)
	}
	outputPath := filepath.Join("output", "shake.mp4")
	if err := shaken.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestWiggle(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	osc := moviego.Oscillation{
		Amplitude: 0.1,
		Frequency: 2,
		StartTime: 0,
		EndTime:   3,
		Decay:     0,
	}
	wiggled, err := cut.Wiggle(osc)
	if err != nil {
		t.Fatalf("Failed to wiggle: %v", err)
	}
	outputPath := filepath.Join("output", "wiggle.mp4")
	if err := wiggled.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestPulse(t *testing.T) {
	t.Skip("scale with oscillation expression may not be supported in all FFmpeg versions")
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	osc := moviego.Oscillation{
		Amplitude: 0.1,
		Frequency: 1,
		StartTime: 0,
		EndTime:   3,
		Decay:     0,
	}
	pulsed, err := cut.Pulse(osc)
	if err != nil {
		t.Fatalf("Failed to pulse: %v", err)
	}
	outputPath := filepath.Join("output", "pulse.mp4")
	if err := pulsed.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestAnimatedRotateWithCut(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	anim := moviego.Animation{Start: 0, End: math.Pi, StartTime: 0, EndTime: 4, Curve: moviego.EaseOut}
	rotated, err := cut.AnimatedRotate(anim)
	if err != nil {
		t.Fatalf("Failed to animate rotate: %v", err)
	}
	outputPath := filepath.Join("output", "animated_rotate_cut.mp4")
	if err := rotated.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-4) > 0.2 {
		t.Errorf("expected duration ~4, got %f", out.GetDuration())
	}
}

func TestAnimatedScaleWithFilters(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	saturated, err := cut.Saturation(1.3)
	if err != nil {
		t.Fatalf("Failed to saturate: %v", err)
	}
	anim := moviego.Animation{Start: 1.0, End: 1.2, StartTime: 0, EndTime: 3, Curve: moviego.Linear}
	scaled, err := saturated.AnimatedScale(anim)
	if err != nil {
		t.Fatalf("Failed to animate scale: %v", err)
	}
	outputPath := filepath.Join("output", "animated_scale_filters.mp4")
	if err := scaled.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3) > 0.2 {
		t.Errorf("expected duration ~3, got %f", out.GetDuration())
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	os.Exit(m.Run())
}
