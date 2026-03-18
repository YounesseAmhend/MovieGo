package filter_test

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

type valueFilterSpec struct {
	name        string
	values      []float64
	apply       func(*moviego.Video, float64) (*moviego.Video, error)
	formatLabel func(float64) string
	scale       bool // whether output may differ from original dimensions
}

func TestAllFiltersStacked(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	origW := int(video.GetWidth())
	origH := int(video.GetHeight())
	if origW%2 != 0 {
		origW--
	}
	if origH%2 != 0 {
		origH--
	}
	cutDuration := 2.0
	cut, err := video.Cut(0, cutDuration)
	if err != nil {
		t.Fatalf("Failed to cut video: %v", err)
	}

	specs := []valueFilterSpec{
		{
			name:   "saturation",
			values: []float64{0.5, 1.0, 1.5, 2.0},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Saturation(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Saturation %.1f", x) },
		},
		{
			name:   "brightness",
			values: []float64{-0.1, 0, 0.05, 0.1},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Brightness(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Brightness %.2f", x) },
		},
		{
			name:   "contrast",
			values: []float64{0.8, 1.0, 1.3, 1.6},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Contrast(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Contrast %.1f", x) },
		},
		{
			name:   "hue",
			values: []float64{0, 90, 180, 270},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Hue(moviego.HueParams{Degrees: moviego.F(x), Saturation: moviego.F(1)}) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Hue %.0f°", x) },
		},
		{
			name:   "scale_ratio",
			values: []float64{0.25, 0.5, 0.75, 1.0},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.ScaleRatio(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("ScaleRatio %.2f", x) },
			scale:  true,
		},
		{
			name:   "rotate",
			values: []float64{0, 0.05, 0.1, 0.15},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Rotate(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Rotate %.2f", x) },
		},
		{
			name:   "fade_in",
			values: []float64{0.3, 0.6, 1.0, 1.5},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.FadeIn(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("FadeIn %.1fs", x) },
		},
		{
			name:   "fade_out",
			values: []float64{0.3, 0.6, 1.0, 1.5},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.FadeOut(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("FadeOut %.1fs", x) },
		},
		{
			name:   "eq_saturation",
			values: []float64{0.5, 1.0, 1.5, 2.0},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Eq(moviego.EqParams{Brightness: moviego.F(0), Contrast: moviego.F(1), Saturation: moviego.F(x), Gamma: moviego.F(1)}) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Eq sat %.1f", x) },
		},
		{
			name:   "horizontal_flip",
			values: []float64{1, 1, 1, 1},
			apply:  func(v *moviego.Video, _ float64) (*moviego.Video, error) { return v.HorizontalFlip() },
			formatLabel: func(_ float64) string { return "HorizontalFlip" },
		},
		{
			name:   "vertical_flip",
			values: []float64{1, 1, 1, 1},
			apply:  func(v *moviego.Video, _ float64) (*moviego.Video, error) { return v.VerticalFlip() },
			formatLabel: func(_ float64) string { return "VerticalFlip" },
		},
		{
			name:   "blur",
			values: []float64{0.5, 1.0, 2.0, 4.0},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Blur(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Blur %.1f", x) },
		},
		{
			name:   "sharpen",
			values: []float64{0.5, 1.0, 1.5, 2.0},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Sharpen(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Sharpen %.1f", x) },
		},
		{
			name:   "grayscale",
			values: []float64{1, 1, 1, 1},
			apply:  func(v *moviego.Video, _ float64) (*moviego.Video, error) { return v.Grayscale() },
			formatLabel: func(_ float64) string { return "Grayscale" },
		},
		{
			name:   "sepia",
			values: []float64{1, 1, 1, 1},
			apply:  func(v *moviego.Video, _ float64) (*moviego.Video, error) { return v.Sepia() },
			formatLabel: func(_ float64) string { return "Sepia" },
		},
		{
			name:   "vignette",
			values: []float64{0.3, 0.5, 0.8, 1.0},
			apply:  func(v *moviego.Video, x float64) (*moviego.Video, error) { return v.Vignette(x) },
			formatLabel: func(x float64) string { return fmt.Sprintf("Vignette %.1f", x) },
		},
		{
			name:   "negate",
			values: []float64{1, 1, 1, 1},
			apply:  func(v *moviego.Video, _ float64) (*moviego.Video, error) { return v.Negate() },
			formatLabel: func(_ float64) string { return "Negate" },
		},
	}

	for _, spec := range specs {
		var tiles []moviego.Video
		for i, val := range spec.values {
			filtered, err := spec.apply(cut, val)
			if err != nil {
				t.Fatalf("Failed to apply filter %q value %v: %v", spec.name, val, err)
			}
			if spec.scale && (filtered.GetWidth() != uint64(origW) || filtered.GetHeight() != uint64(origH)) {
				filtered, err = filtered.Scale(moviego.ScaleParams{Width: origW, Height: origH})
				if err != nil {
					t.Fatalf("Failed to scale filter %q tile %d to original dimensions: %v", spec.name, i, err)
				}
			}
			label := spec.formatLabel(val)
			withText, err := filtered.AddText(moviego.TextClip{
				Text:     label,
				FontSize: 24,
				FontColor: "white",
				Position: moviego.TextTopCenter(),
				Stroke:   moviego.Stroke{Width: 2, Color: "black"},
			})
			if err != nil {
				t.Fatalf("Failed to add label %q: %v", label, err)
			}
			tiles = append(tiles, *withText)
		}

		stacked, err := moviego.XStack(tiles, "0_0|w0_0|0_h0|w0_h0")
		if err != nil {
			t.Fatalf("Failed to xstack filter %q: %v", spec.name, err)
		}

		outputPath := filepath.Join("output", spec.name+"_values.mp4")
		if err := stacked.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
			t.Fatalf("Failed to write %s: %v", outputPath, err)
		}

		exported, err := moviego.NewVideoFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to load output %s: %v", outputPath, err)
		}
		expectedW := origW * 2
		expectedH := origH * 2
		if exported.GetWidth() != uint64(expectedW) || exported.GetHeight() != uint64(expectedH) {
			t.Fatalf("%s: expected dimensions %dx%d, got %dx%d", spec.name, expectedW, expectedH, exported.GetWidth(), exported.GetHeight())
		}
		if math.Abs(exported.GetDuration()-cutDuration) > 0.2 {
			t.Fatalf("%s: expected duration ~%.1f, got %f", spec.name, cutDuration, exported.GetDuration())
		}
	}
}

func TestFilterChain(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	originalDuration := video.GetDuration()

	filtered, err := video.Saturation(1.2)
	if err != nil {
		t.Fatalf("Failed to apply saturation: %v", err)
	}
	filtered, err = filtered.Brightness(0.05)
	if err != nil {
		t.Fatalf("Failed to apply brightness: %v", err)
	}
	filtered, err = filtered.HorizontalFlip()
	if err != nil {
		t.Fatalf("Failed to apply horizontal flip: %v", err)
	}

	outputPath := filepath.Join("output", "filter_chain.mp4")
	err = filtered.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	if math.Abs(exportedVideo.GetDuration()-originalDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", originalDuration, exportedVideo.GetDuration())
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	os.Exit(m.Run())
}
