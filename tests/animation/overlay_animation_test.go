package animation_test

import (
	"math"
	"path/filepath"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestAnimatedOverlayPosition(t *testing.T) {
	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load bg: %v", err)
	}
	bgCut, err := bg.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut bg: %v", err)
	}
	fg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load fg: %v", err)
	}
	fgCut, err := fg.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut fg: %v", err)
	}
	fgScaled, err := fgCut.ScaleRatio(0.3)
	if err != nil {
		t.Fatalf("Failed to scale fg: %v", err)
	}
	pos := moviego.AnimatedPosition{
		Start:     moviego.Position{X: "0", Y: "0"},
		End:       moviego.Position{X: "600", Y: "300"},
		StartTime: 0,
		EndTime:   4,
		Curve:     moviego.EaseInOut,
	}
	fgScaled.SetAnimatedPosition(pos)
	result, err := moviego.CompositeClip([]moviego.Video{*bgCut, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}
	outputPath := filepath.Join("output", "animated_overlay_position.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestAnimatedOverlayOpacity(t *testing.T) {
	t.Skip("colorchannelmixer aa expression may not be supported in all FFmpeg versions")
	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load bg: %v", err)
	}
	bgCut, err := bg.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut bg: %v", err)
	}
	fg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load fg: %v", err)
	}
	fgCut, err := fg.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut fg: %v", err)
	}
	fgScaled, err := fgCut.ScaleRatio(0.3)
	if err != nil {
		t.Fatalf("Failed to scale fg: %v", err)
	}
	fgScaled.SetPosition(moviego.CenterPosition())
	opacity := moviego.Animation{Start: 0, End: 1, StartTime: 0, EndTime: 2, Curve: moviego.Linear}
	fgScaled.SetAnimatedOpacity(opacity)
	result, err := moviego.CompositeClip([]moviego.Video{*bgCut, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}
	outputPath := filepath.Join("output", "animated_overlay_opacity.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestAnimatedOverlayPositionAndOpacity(t *testing.T) {
	t.Skip("colorchannelmixer aa expression may not be supported in all FFmpeg versions")
	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load bg: %v", err)
	}
	bgCut, err := bg.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut bg: %v", err)
	}
	fg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load fg: %v", err)
	}
	fgCut, err := fg.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut fg: %v", err)
	}
	fgScaled, err := fgCut.ScaleRatio(0.25)
	if err != nil {
		t.Fatalf("Failed to scale fg: %v", err)
	}
	pos := moviego.AnimatedPosition{
		Start:     moviego.Position{X: "10", Y: "10"},
		End:       moviego.Position{X: "400", Y: "200"},
		StartTime: 0,
		EndTime:   3,
		Curve:     moviego.Linear,
	}
	opacity := moviego.Animation{Start: 0.5, End: 1, StartTime: 0, EndTime: 3, Curve: moviego.Linear}
	fgScaled.SetAnimatedPosition(pos)
	fgScaled.SetAnimatedOpacity(opacity)
	result, err := moviego.CompositeClip([]moviego.Video{*bgCut, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}
	outputPath := filepath.Join("output", "animated_overlay_position_opacity.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestAnimatedOverlayWithCut(t *testing.T) {
	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load bg: %v", err)
	}
	bgCut, err := bg.Cut(0, 5)
	if err != nil {
		t.Fatalf("Failed to cut bg: %v", err)
	}
	fg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load fg: %v", err)
	}
	fgCut, err := fg.Cut(1, 4)
	if err != nil {
		t.Fatalf("Failed to cut fg: %v", err)
	}
	fgScaled, err := fgCut.ScaleRatio(0.3)
	if err != nil {
		t.Fatalf("Failed to scale fg: %v", err)
	}
	pos := moviego.AnimatedPosition{
		Start:     moviego.Position{X: "50", Y: "50"},
		End:       moviego.Position{X: "500", Y: "250"},
		StartTime: 0,
		EndTime:   3,
		Curve:     moviego.EaseOut,
	}
	fgScaled.SetAnimatedPosition(pos)
	result, err := moviego.CompositeClip([]moviego.Video{*bgCut, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}
	outputPath := filepath.Join("output", "animated_overlay_cut.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-5) > 0.2 {
		t.Errorf("expected duration ~5, got %f", out.GetDuration())
	}
}
