package transition_test

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestTransitionFade(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	clip1Cut, err := clip1.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut clip1: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	clip2Cut, err := clip2.Cut(1, 5)
	if err != nil {
		t.Fatalf("Failed to cut clip2: %v", err)
	}
	result, err := moviego.ConcatenateWithTransition(clip1Cut, clip2Cut, moviego.TransitionParams{
		Transition: moviego.TransitionFade,
		Duration:   1.5,
	})
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	outputPath := filepath.Join("output", "transition_fade.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	expectedDuration := 4.0 + 4.0 - 1.5 // clip1 + clip2 - overlap
	if math.Abs(out.GetDuration()-expectedDuration) > 0.3 {
		t.Errorf("expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestTransitionWipeLeft(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	clip1Cut, err := clip1.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut clip1: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	clip2Cut, err := clip2.Cut(1, 5)
	if err != nil {
		t.Fatalf("Failed to cut clip2: %v", err)
	}
	result, err := moviego.ConcatenateWithTransition(clip1Cut, clip2Cut, moviego.TransitionParams{
		Transition: moviego.TransitionWipeLeft,
		Duration:   1,
	})
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	outputPath := filepath.Join("output", "transition_wipe_left.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	expectedDuration := 4.0 + 4.0 - 1.0
	if math.Abs(out.GetDuration()-expectedDuration) > 0.3 {
		t.Errorf("expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestTransitionSlideRight(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	clip1Cut, err := clip1.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut clip1: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	clip2Cut, err := clip2.Cut(1, 5)
	if err != nil {
		t.Fatalf("Failed to cut clip2: %v", err)
	}
	result, err := moviego.ConcatenateWithTransition(clip1Cut, clip2Cut, moviego.TransitionParams{
		Transition: moviego.TransitionSlideRight,
		Duration:   1,
	})
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	outputPath := filepath.Join("output", "transition_slide_right.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	expectedDuration := 4.0 + 4.0 - 1.0
	if math.Abs(out.GetDuration()-expectedDuration) > 0.3 {
		t.Errorf("expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestTransitionDissolve(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	clip1Cut, err := clip1.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut clip1: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	clip2Cut, err := clip2.Cut(1, 5)
	if err != nil {
		t.Fatalf("Failed to cut clip2: %v", err)
	}
	result, err := moviego.ConcatenateWithTransition(clip1Cut, clip2Cut, moviego.TransitionParams{
		Transition: moviego.TransitionDissolve,
		Duration:   1.2,
	})
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	outputPath := filepath.Join("output", "transition_dissolve.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	expectedDuration := 4.0 + 4.0 - 1.2
	if math.Abs(out.GetDuration()-expectedDuration) > 0.3 {
		t.Errorf("expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestTransitionZoomIn(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	clip1Cut, err := clip1.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut clip1: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	clip2Cut, err := clip2.Cut(1, 5)
	if err != nil {
		t.Fatalf("Failed to cut clip2: %v", err)
	}
	result, err := moviego.ConcatenateWithTransition(clip1Cut, clip2Cut, moviego.TransitionParams{
		Transition: moviego.TransitionZoomIn,
		Duration:   1,
	})
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	outputPath := filepath.Join("output", "transition_zoom_in.mp4")
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	expectedDuration := 4.0 + 4.0 - 1.0
	if math.Abs(out.GetDuration()-expectedDuration) > 0.3 {
		t.Errorf("expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestTransitionInvalidDuration(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	clip1Cut, err := clip1.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut clip1: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	clip2Cut, err := clip2.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut clip2: %v", err)
	}
	_, err = moviego.ConcatenateWithTransition(clip1Cut, clip2Cut, moviego.TransitionParams{
		Transition: moviego.TransitionFade,
		Duration:   5, // longer than clip1 (3s)
	})
	if err == nil {
		t.Fatal("expected error for duration > clip length")
	}
}

func TestTransitionChained(t *testing.T) {
	clip1, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip1: %v", err)
	}
	c1, err := clip1.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	clip2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip2: %v", err)
	}
	c2, err := clip2.Cut(1, 4)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	clip3, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load clip3: %v", err)
	}
	c3, err := clip3.Cut(2, 5)
	if err != nil {
		t.Fatalf("Failed to cut: %v", err)
	}
	// A -> B with transition
	ab, err := moviego.ConcatenateWithTransition(c1, c2, moviego.TransitionParams{
		Transition: moviego.TransitionFade,
		Duration:   0.8,
	})
	if err != nil {
		t.Fatalf("Failed A->B transition: %v", err)
	}
	// (A+B) -> C with transition
	abc, err := moviego.ConcatenateWithTransition(ab, c3, moviego.TransitionParams{
		Transition: moviego.TransitionDissolve,
		Duration:   0.8,
	})
	if err != nil {
		t.Fatalf("Failed (A+B)->C transition: %v", err)
	}
	outputPath := filepath.Join("output", "transition_chained.mp4")
	if err := abc.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	// 3+3-0.8 + 3-0.8 = 5.2 + 2.2 = 7.4
	expectedDuration := (3.0 + 3.0 - 0.8) + 3.0 - 0.8
	if math.Abs(out.GetDuration()-expectedDuration) > 0.5 {
		t.Errorf("expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	os.Exit(m.Run())
}
