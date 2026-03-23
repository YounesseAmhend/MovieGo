package composite_test

import (
	"math"
	"os"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestCompositeClip(t *testing.T) {
	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load background video: %v", err)
	}

	fg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load foreground video: %v", err)
	}

	fgScaled, err := fg.ScaleRatio(0.5)
	if err != nil {
		t.Fatalf("Failed to scale foreground: %v", err)
	}

	result, err := moviego.CompositeClip([]moviego.Video{*bg, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}

	const outputPath = "output/composite_center.mp4"
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write composite video: %v", err)
	}

	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-bg.GetDuration()) > 0.2 {
		t.Fatalf("Expected duration ~%f, got %f", bg.GetDuration(), out.GetDuration())
	}
	if out.GetWidth() != bg.GetWidth() || out.GetHeight() != bg.GetHeight() {
		t.Fatalf("Expected dimensions %dx%d, got %dx%d",
			bg.GetWidth(), bg.GetHeight(), out.GetWidth(), out.GetHeight())
	}
}

func TestCompositeClipWithPosition(t *testing.T) {
	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load background video: %v", err)
	}

	fg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load foreground video: %v", err)
	}

	fgScaled, err := fg.ScaleRatio(0.5)
	if err != nil {
		t.Fatalf("Failed to scale foreground: %v", err)
	}
	fgScaled.SetPosition(moviego.TopLeftPosition())

	result, err := moviego.CompositeClip([]moviego.Video{*bg, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}

	const outputPath = "output/composite_top_left.mp4"
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write composite video: %v", err)
	}

	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-bg.GetDuration()) > 0.2 {
		t.Fatalf("Expected duration ~%f, got %f", bg.GetDuration(), out.GetDuration())
	}
}

func TestCompositeClipNested(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}

	bgCut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut background: %v", err)
	}

	fgCut, err := video.Cut(1, 3)
	if err != nil {
		t.Fatalf("Failed to cut foreground: %v", err)
	}
	fgScaled, err := fgCut.ScaleRatio(0.3)
	if err != nil {
		t.Fatalf("Failed to scale foreground: %v", err)
	}
	fgScaled.SetPosition(moviego.Position{X: "10", Y: "10"})

	result, err := moviego.CompositeClip([]moviego.Video{*bgCut, *fgScaled})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}

	const outputPath = "output/composite_nested.mp4"
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write composite video: %v", err)
	}

	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if math.Abs(out.GetDuration()-3.0) > 0.2 {
		t.Fatalf("Expected duration ~3.0, got %f", out.GetDuration())
	}
}

func TestCompositeClipMultiLayer(t *testing.T) {
	for _, path := range []string{common.TestVideo2Path, common.TestVideo3Path} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Skipf("Test video %s not found", path)
		}
	}

	bg, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load bg: %v", err)
	}
	v2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load v2: %v", err)
	}
	v3, err := moviego.NewVideoFile(common.TestVideo3Path)
	if err != nil {
		t.Fatalf("Failed to load v3: %v", err)
	}

	v2Scaled, err := v2.ScaleRatio(0.3)
	if err != nil {
		t.Fatalf("Failed to scale v2: %v", err)
	}
	v2Scaled.SetPosition(moviego.TopLeftPosition())

	v3Scaled, err := v3.ScaleRatio(0.3)
	if err != nil {
		t.Fatalf("Failed to scale v3: %v", err)
	}
	v3Scaled.SetPosition(moviego.Position{X: "W-w", Y: "H-h"})

	result, err := moviego.CompositeClip([]moviego.Video{*bg, *v2Scaled, *v3Scaled})
	if err != nil {
		t.Fatalf("Failed to composite multi-layer: %v", err)
	}

	const outputPath = "output/composite_multi_layer.mp4"
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write multi-layer composite: %v", err)
	}

	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	if out.GetWidth() != bg.GetWidth() || out.GetHeight() != bg.GetHeight() {
		t.Fatalf("Expected dimensions %dx%d, got %dx%d",
			bg.GetWidth(), bg.GetHeight(), out.GetWidth(), out.GetHeight())
	}
}
