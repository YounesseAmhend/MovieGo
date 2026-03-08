package concatenate_test

import (
	"math"
	"os"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestConcatenate(t *testing.T) {
	video1, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}

	const startTimeCut1 = 0
	const endTimeCut1 = 2

	const startTimeCut2 = 2
	const endTimeCut2 = 5

	const expectedDuration float64 = endTimeCut1 - startTimeCut1 + endTimeCut2 - startTimeCut2

	cut1, err := video1.Cut(startTimeCut1, endTimeCut1)
	if err != nil {
		t.Fatalf("Failed to cut video (0-3): %v", err)
	}
	cut2, err := video1.Cut(startTimeCut2, endTimeCut2)
	if err != nil {
		t.Fatalf("Failed to cut video (2-3): %v", err)
	}

	const outputPath = "output/concatenated_video.mp4"
	result, err := moviego.Concatenate([]moviego.Video{*cut1, *cut2})

	result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})

	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	resultVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	if resultVideo.GetDuration() != expectedDuration {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, result.GetDuration())
	}

}

func TestNestedConcatenate(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}

	cut1, err := video.Cut(0, 1)
	if err != nil {
		t.Fatalf("Failed to cut video (0-1): %v", err)
	}
	cut2, err := video.Cut(1, 2)
	if err != nil {
		t.Fatalf("Failed to cut video (1-2): %v", err)
	}
	cut3, err := video.Cut(2, 3)
	if err != nil {
		t.Fatalf("Failed to cut video (2-3): %v", err)
	}
	cut4, err := video.Cut(3, 4)
	if err != nil {
		t.Fatalf("Failed to cut video (3-4): %v", err)
	}

	concatAB, err := moviego.Concatenate([]moviego.Video{*cut1, *cut2})
	if err != nil {
		t.Fatalf("Failed to concatenate AB: %v", err)
	}
	concatCD, err := moviego.Concatenate([]moviego.Video{*cut3, *cut4})
	if err != nil {
		t.Fatalf("Failed to concatenate CD: %v", err)
	}

	result, err := moviego.Concatenate([]moviego.Video{*concatAB, *concatCD})
	if err != nil {
		t.Fatalf("Failed nested concatenation: %v", err)
	}

	const outputPath = "output/nested_concatenated.mp4"
	err = result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write nested concatenation: %v", err)
	}

	resultVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load nested concatenation output: %v", err)
	}

	const expectedDuration float64 = 4.0
	if math.Abs(resultVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, resultVideo.GetDuration())
	}
	if resultVideo.GetWidth() != video.GetWidth() {
		t.Fatalf("Expected width %d, got %d", video.GetWidth(), resultVideo.GetWidth())
	}
	if resultVideo.GetHeight() != video.GetHeight() {
		t.Fatalf("Expected height %d, got %d", video.GetHeight(), resultVideo.GetHeight())
	}
}

// TestInsaneConcatenate: 4 cuts from test2/test3/test4, nested concat in shuffled order (4,1,3,2)
func TestInsaneConcatenate(t *testing.T) {
	for _, path := range []string{common.TestVideo2Path, common.TestVideo3Path, common.TestVideo4Path} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Skipf("Test video %s not found. Run: make download-extra-test-videos", path)
		}
	}

	v2, _ := moviego.NewVideoFile(common.TestVideo2Path)
	v3, _ := moviego.NewVideoFile(common.TestVideo3Path)
	v4, _ := moviego.NewVideoFile(common.TestVideo4Path)

	const seg float64 = 8.0
	// 4 cuts: v4[0:0.6], v2[0:0.6], v3[0:0.6], v2[0.6:1.2] — test2 sliced twice
	c1, _ := v2.Cut(0, seg)
	c2, _ := v2.Cut(seg, seg*2)
	c3, _ := v3.Cut(0, seg)
	c4, _ := v4.Cut(0, seg)

	// Nested concat: (c4 + c1) + (c3 + c2) → order: 4, 1, 3, 2 (shuffled)
	left, _ := moviego.Concatenate([]moviego.Video{*c4, *c1})
	right, _ := moviego.Concatenate([]moviego.Video{*c3, *c2})
	result, err := moviego.Concatenate([]moviego.Video{*left, *right})
	if err != nil {
		t.Fatalf("Failed insane concat: %v", err)
	}

	const outputPath = "output/insane_concatenated.mp4"
	if err := result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath}); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}
	expected := 4 * seg
	if math.Abs(out.GetDuration()-expected) > 0.15 {
		t.Fatalf("Expected duration %f, got %f", expected, out.GetDuration())
	}
}
