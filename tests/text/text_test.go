package text_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func mustLoadVideo(t *testing.T, path string) *moviego.Video {
	t.Helper()
	video, err := moviego.NewVideoFile(path)
	if err != nil {
		t.Fatalf("Failed to load video %s: %v", path, err)
	}
	return video
}

func mustWriteVideo(t *testing.T, video *moviego.Video, outputPath string) *moviego.Video {
	t.Helper()
	if err := video.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write video %s: %v", outputPath, err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output %s: %v", outputPath, err)
	}
	return out
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

func TestAddTextBasic(t *testing.T) {
	video := mustLoadVideo(t, common.TestVideoPath)

	withText, err := video.AddText(moviego.TextClip{
		Text:      "Hello World",
		FontSize:  48,
		FontColor: "white",
		Position:  moviego.TextCenter(),
	})
	if err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}

	out := mustWriteVideo(t, withText, filepath.Join("output", "text_basic.mp4"))
	if math.Abs(out.GetDuration()-video.GetDuration()) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", video.GetDuration(), out.GetDuration())
	}
	if out.GetWidth() != video.GetWidth() || out.GetHeight() != video.GetHeight() {
		t.Fatalf("Expected dimensions %dx%d, got %dx%d",
			video.GetWidth(), video.GetHeight(), out.GetWidth(), out.GetHeight())
	}
}

func TestAddTextStyledTimed(t *testing.T) {
	video := mustLoadVideo(t, common.TestVideoPath)

	withText, err := video.AddText(moviego.TextClip{
		Text:       "Styled Chapter",
		FontFamily: "Sans",
		FontSize:   36,
		FontColor:  "white",
		Position:   moviego.TextBottomCenter(),
		StartTime:  0.5,
		EndTime:    3.5,
		Background: moviego.Background{Enabled: true, Color: "black@0.6", Padding: "10"},
		Stroke:     moviego.Stroke{Width: 2, Color: "black"},
		Shadow:     moviego.Shadow{X: 2, Y: 2, Color: "black"},
	})
	if err != nil {
		t.Fatalf("Failed to add styled text: %v", err)
	}

	out := mustWriteVideo(t, withText, filepath.Join("output", "text_styled_timed.mp4"))
	if math.Abs(out.GetDuration()-video.GetDuration()) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", video.GetDuration(), out.GetDuration())
	}
}

func TestArabicTextShapingWithFont(t *testing.T) {
	video := mustLoadVideo(t, common.TestVideoPath)
	arabicFont := mustFontPath(t, common.ArabicFontPath)

	withText, err := video.AddText(moviego.TextClip{
		Text:        "مرحبا بالعالم",
		FontFamily:  arabicFont,
		FontSize:    48,
		FontColor:   "white",
		Position:    moviego.TextCenter(),
		TextShaping: true,
	})
	if err != nil {
		t.Fatalf("Failed to add Arabic text: %v", err)
	}

	_ = mustWriteVideo(t, withText, filepath.Join("output", "text_arabic.mp4"))
}

func TestChineseTextRenderingWithFont(t *testing.T) {
	video := mustLoadVideo(t, common.TestVideoPath)
	chineseFont := mustFontPath(t, common.ChineseFontPath)

	withText, err := video.AddText(moviego.TextClip{
		Text:       "你好世界",
		FontFamily: chineseFont,
		FontSize:   48,
		FontColor:  "white",
		Position:   moviego.TextCenter(),
	})
	if err != nil {
		t.Fatalf("Failed to add Chinese text: %v", err)
	}

	_ = mustWriteVideo(t, withText, filepath.Join("output", "text_chinese.mp4"))
}

func TestAddTextRequiresText(t *testing.T) {
	video := mustLoadVideo(t, common.TestVideoPath)

	_, err := video.AddText(moviego.TextClip{
		FontSize:  48,
		FontColor: "white",
		Position:  moviego.TextCenter(),
	})
	if err == nil {
		t.Fatal("Expected error when Text and TextFile are both empty")
	}
}

func TestAddTextsNilClip(t *testing.T) {
	video := mustLoadVideo(t, common.TestVideoPath)

	_, err := video.AddTexts([]*moviego.TextClip{nil})
	if err == nil {
		t.Fatal("Expected error when TextClip is nil")
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	os.Exit(m.Run())
}
