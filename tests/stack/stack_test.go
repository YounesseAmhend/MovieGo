package stack_test

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

const (
	stackDuration        = 4.0
	durationErrFmt       = "Expected duration ~%.1f, got %f"
	dimensionErrFmt      = "Expected exported dimensions %dx%d, got %dx%d"
)

type tileSpec struct {
	path        string
	label       string
	fontPath    string
	fontColor   string
	textShaping bool
	background  moviego.Background
	stroke      moviego.Stroke
	shadow      moviego.Shadow
	filter      func(*moviego.Video) (*moviego.Video, error)
}

func mustLoadVideo(t *testing.T, path string) *moviego.Video {
	t.Helper()
	video, err := moviego.NewVideoFile(path)
	if err != nil {
		t.Fatalf("Failed to load video %s: %v", path, err)
	}
	return video
}

func mustCutVideo(t *testing.T, video *moviego.Video, start, end float64) *moviego.Video {
	t.Helper()
	cut, err := video.Cut(start, end)
	if err != nil {
		t.Fatalf("Failed to cut video: %v", err)
	}
	return cut
}

func mustBuildTile(t *testing.T, spec tileSpec) *moviego.Video {
	t.Helper()
	video := mustCutVideo(t, mustLoadVideo(t, spec.path), 0, stackDuration)
	if spec.filter != nil {
		var err error
		video, err = spec.filter(video)
		if err != nil {
			t.Fatalf("Failed to apply tile filter %q: %v", spec.label, err)
		}
	}

	withText, err := video.AddText(moviego.TextClip{
		Text:        spec.label,
		FontFamily:  spec.fontPath,
		FontSize:    24,
		FontColor:   spec.fontColor,
		Position:    moviego.TextTopCenter(),
		TextShaping: spec.textShaping,
		Background:  spec.background,
		Stroke:      spec.stroke,
		Shadow:      spec.shadow,
	})
	if err != nil {
		t.Fatalf("Failed to add tile text %q: %v", spec.label, err)
	}
	return withText
}

func mustWriteAndLoad(t *testing.T, video *moviego.Video, outputPath string) *moviego.Video {
	t.Helper()
	if err := video.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write video %s: %v", outputPath, err)
	}
	return mustLoadVideo(t, outputPath)
}

func mustFontPath(t *testing.T, path string) string {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Missing required font %s: %v", path, err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("Failed to resolve font path %s: %v", path, err)
	}
	return absPath
}

func TestHStackCombinesFiltersAndLabels(t *testing.T) {
	left := mustBuildTile(t, tileSpec{
		path:      common.TestVideoPath,
		label:     "Brightness",
		fontColor: "white",
		filter: func(video *moviego.Video) (*moviego.Video, error) {
			return video.Brightness(0.12)
		},
	})
	right := mustBuildTile(t, tileSpec{
		path:      common.TestVideoPath,
		label:     "Saturation",
		fontColor: "yellow",
		filter: func(video *moviego.Video) (*moviego.Video, error) {
			return video.Saturation(1.6)
		},
	})

	stacked, err := moviego.HStack([]moviego.Video{*left, *right})
	if err != nil {
		t.Fatalf("Failed to hstack videos: %v", err)
	}

	expectedW := left.GetWidth() + right.GetWidth()
	expectedH := left.GetHeight()
	if stacked.GetWidth() != expectedW || stacked.GetHeight() != expectedH {
		t.Fatalf("Expected stacked dimensions %dx%d (sum of widths, common height), got %dx%d", expectedW, expectedH, stacked.GetWidth(), stacked.GetHeight())
	}

	out := mustWriteAndLoad(t, stacked, filepath.Join("output", "hstack_combined.mp4"))
	if out.GetWidth() != expectedW || out.GetHeight() != expectedH {
		t.Fatalf(dimensionErrFmt, expectedW, expectedH, out.GetWidth(), out.GetHeight())
	}
	if math.Abs(out.GetDuration()-stackDuration) > 0.2 {
		t.Fatalf(durationErrFmt, stackDuration, out.GetDuration())
	}
}

func TestVStackCombinesLayoutAndStyledText(t *testing.T) {
	top := mustBuildTile(t, tileSpec{
		path:      common.TestVideoPath,
		label:     "Contrast+Stroke",
		fontColor: "white",
		stroke:    moviego.Stroke{Width: 2, Color: "black"},
		filter: func(video *moviego.Video) (*moviego.Video, error) {
			return video.Contrast(1.2)
		},
	})
	bottom := mustBuildTile(t, tileSpec{
		path:       common.TestVideoPath,
		label:      "Rotate+Shadow",
		fontColor:  "white",
		background: moviego.Background{Enabled: true, Color: "black@0.5", Padding: "8"},
		shadow:     moviego.Shadow{X: 2, Y: 2, Color: "black"},
		filter: func(video *moviego.Video) (*moviego.Video, error) {
			return video.Rotate(0.08)
		},
	})

	stacked, err := moviego.VStack([]moviego.Video{*top, *bottom})
	if err != nil {
		t.Fatalf("Failed to vstack videos: %v", err)
	}

	expectedW := top.GetWidth()
	expectedH := top.GetHeight() + bottom.GetHeight()
	if stacked.GetWidth() != expectedW || stacked.GetHeight() != expectedH {
		t.Fatalf("Expected stacked dimensions %dx%d (common width, sum of heights), got %dx%d", expectedW, expectedH, stacked.GetWidth(), stacked.GetHeight())
	}

	out := mustWriteAndLoad(t, stacked, filepath.Join("output", "vstack_combined.mp4"))
	if out.GetWidth() != expectedW || out.GetHeight() != expectedH {
		t.Fatalf(dimensionErrFmt, expectedW, expectedH, out.GetWidth(), out.GetHeight())
	}
	if math.Abs(out.GetDuration()-stackDuration) > 0.2 {
		t.Fatalf(durationErrFmt, stackDuration, out.GetDuration())
	}
}

func TestXStackCombinesFiltersStylingAndLanguages(t *testing.T) {
	arabicFont := mustFontPath(t, common.ArabicFontPath)
	chineseFont := mustFontPath(t, common.ChineseFontPath)

	videos := []moviego.Video{
		*mustBuildTile(t, tileSpec{
			path:      common.TestVideoPath,
			label:     "Hello Brightness",
			fontColor: "white",
			filter: func(video *moviego.Video) (*moviego.Video, error) {
				return video.Brightness(0.1)
			},
		}),
		*mustBuildTile(t, tileSpec{
			path:        common.TestVideo2Path,
			label:       "مرحبا بالعالم",
			fontPath:    arabicFont,
			fontColor:   "white",
			textShaping: true,
			background:  moviego.Background{Enabled: true, Color: "black@0.6", Padding: "8"},
			filter: func(video *moviego.Video) (*moviego.Video, error) {
				return video.Saturation(1.4)
			},
		}),
		*mustBuildTile(t, tileSpec{
			path:      common.TestVideo3Path,
			label:     "你好世界",
			fontPath:  chineseFont,
			fontColor: "yellow",
			shadow:    moviego.Shadow{X: 2, Y: 2, Color: "black"},
			filter: func(video *moviego.Video) (*moviego.Video, error) {
				return video.Contrast(1.15)
			},
		}),
		*mustBuildTile(t, tileSpec{
			path:      common.TestVideo4Path,
			label:     "Crop+Box",
			fontColor: "cyan",
			background: moviego.Background{
				Enabled: true,
				Color:   "black@0.5",
				Padding: "6",
			},
			filter: func(video *moviego.Video) (*moviego.Video, error) {
				w, h := video.GetWidth(), video.GetHeight()
				return video.Crop(moviego.CropParams{X: 0, Y: 0, Width: int(w) / 2, Height: int(h) / 2})
			},
		}),
	}

	stacked, err := moviego.XStack(videos, "0_0|w0_0|0_h0|w0_h0")
	if err != nil {
		t.Fatalf("Failed to xstack videos: %v", err)
	}

	// For 2x2 layout "0_0|w0_0|0_h0|w0_h0": width = w0+w1, height = h0+h2
	expectedW := videos[0].GetWidth() + videos[1].GetWidth()
	expectedH := videos[0].GetHeight() + videos[2].GetHeight()
	if stacked.GetWidth() != expectedW || stacked.GetHeight() != expectedH {
		t.Fatalf("Expected stacked dimensions %dx%d (layout-derived), got %dx%d", expectedW, expectedH, stacked.GetWidth(), stacked.GetHeight())
	}

	out := mustWriteAndLoad(t, stacked, filepath.Join("output", "xstack_combined.mp4"))
	if out.GetWidth() != expectedW || out.GetHeight() != expectedH {
		t.Fatalf(dimensionErrFmt, expectedW, expectedH, out.GetWidth(), out.GetHeight())
	}
	if math.Abs(out.GetDuration()-stackDuration) > 0.2 {
		t.Fatalf(durationErrFmt, stackDuration, out.GetDuration())
	}
}

func TestXStackRejectsInvalidLayout(t *testing.T) {
	left := mustBuildTile(t, tileSpec{path: common.TestVideoPath, label: "Left", fontColor: "white"})
	right := mustBuildTile(t, tileSpec{path: common.TestVideo2Path, label: "Right", fontColor: "white"})

	_, err := moviego.XStack([]moviego.Video{*left, *right}, "0_0")
	if err == nil {
		t.Fatal("Expected invalid xstack layout to fail")
	}
}

func mustBuildCroppedTile(t *testing.T, path string, width, height int) *moviego.Video {
	t.Helper()
	video := mustCutVideo(t, mustLoadVideo(t, path), 0, stackDuration)
	cropped, err := video.Crop(moviego.CropParams{X: 0, Y: 0, Width: width, Height: height})
	if err != nil {
		t.Fatalf("Failed to crop video: %v", err)
	}
	return cropped
}

func TestHStackRejectsMismatchedHeights(t *testing.T) {
	// Same width, different heights
	left := mustBuildCroppedTile(t, common.TestVideoPath, 320, 180)
	right := mustBuildCroppedTile(t, common.TestVideo2Path, 320, 200)

	_, err := moviego.HStack([]moviego.Video{*left, *right})
	if err == nil {
		t.Fatal("Expected HStack with mismatched heights to fail")
	}
	if !strings.Contains(err.Error(), "height") {
		t.Fatalf("Expected error to mention height mismatch, got: %v", err)
	}
}

func TestVStackRejectsMismatchedWidths(t *testing.T) {
	// Same height, different widths
	top := mustBuildCroppedTile(t, common.TestVideoPath, 320, 180)
	bottom := mustBuildCroppedTile(t, common.TestVideo2Path, 400, 180)

	_, err := moviego.VStack([]moviego.Video{*top, *bottom})
	if err == nil {
		t.Fatal("Expected VStack with mismatched widths to fail")
	}
	if !strings.Contains(err.Error(), "width") {
		t.Fatalf("Expected error to mention width mismatch, got: %v", err)
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	os.Exit(m.Run())
}
