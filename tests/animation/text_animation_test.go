package animation_test

import (
	"math"
	"path/filepath"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestAnimatedTextPosition(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	pos := moviego.AnimatedPosition{
		Start:     moviego.Position{X: "-100", Y: "200"},
		End:       moviego.Position{X: "400", Y: "200"},
		StartTime: 0,
		EndTime:   4,
		Curve:     moviego.EaseInOut,
	}
	clip := moviego.TextClip{
		Text:            "Sliding Text",
		FontSize:        36,
		FontColor:        "white",
		Position:        moviego.Position{Y: "h/2-18"},
		AnimatePosition: &pos,
		Stroke:          moviego.Stroke{Width: 2, Color: "black"},
	}
	withText, err := cut.AddText(clip)
	if err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}
	outputPath := filepath.Join("output", "animated_text_position.mp4")
	if err := withText.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestAnimatedTextOpacity(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	opacity := moviego.Animation{Start: 0, End: 1, StartTime: 0, EndTime: 2, Curve: moviego.EaseInOut}
	clip := moviego.TextClip{
		Text:           "Fade In",
		FontSize:       48,
		FontColor:      "white",
		Position:       moviego.TextCenter(),
		AnimateOpacity: &opacity,
		Stroke:         moviego.Stroke{Width: 2, Color: "black"},
	}
	withText, err := cut.AddText(clip)
	if err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}
	outputPath := filepath.Join("output", "animated_text_opacity.mp4")
	if err := withText.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestAnimatedTextPositionAndOpacity(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	pos := moviego.AnimatedPosition{
		Start:     moviego.Position{X: "50", Y: "50"},
		End:       moviego.Position{X: "300", Y: "200"},
		StartTime: 0,
		EndTime:   3,
		Curve:     moviego.Linear,
	}
	opacity := moviego.Animation{Start: 0.3, End: 1, StartTime: 0, EndTime: 3, Curve: moviego.Linear}
	clip := moviego.TextClip{
		Text:            "Animated",
		FontSize:        32,
		FontColor:       "white",
		AnimatePosition: &pos,
		AnimateOpacity:  &opacity,
		Stroke:          moviego.Stroke{Width: 2, Color: "black"},
	}
	withText, err := cut.AddText(clip)
	if err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}
	outputPath := filepath.Join("output", "animated_text_position_opacity.mp4")
	if err := withText.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestTypewriter(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 5)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	clip := moviego.TextClip{
		Text:     "Hello",
		FontSize: 36,
		FontColor: "white",
		Position: moviego.TextTopLeft(),
		Typewriter: &moviego.TypewriterParams{
			CharDelay: 0.3,
			StartTime: 0.5,
		},
		Stroke: moviego.Stroke{Width: 2, Color: "black"},
	}
	withText, err := cut.AddText(clip)
	if err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}
	outputPath := filepath.Join("output", "typewriter.mp4")
	if err := withText.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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

func TestTypewriterWithCursor(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to load video: %v", err)
	}
	cut, err := video.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	clip := moviego.TextClip{
		Text:     "Hi",
		FontSize: 32,
		FontColor: "white",
		Position: moviego.Position{X: "20", Y: "20"},
		Typewriter: &moviego.TypewriterParams{
			CharDelay: 0.4,
			StartTime: 0,
			Cursor:    "|",
		},
		Stroke: moviego.Stroke{Width: 2, Color: "black"},
	}
	withText, err := cut.AddText(clip)
	if err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}
	outputPath := filepath.Join("output", "typewriter_cursor.mp4")
	if err := withText.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
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
