package tests

import (
	"math"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
)

func TestConcatenate(t *testing.T) {
	video1, err := moviego.NewVideoFile(TEST_VIDEO_PATH)
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
	result, err := moviego.Concatenate([]moviego.Video{*cut1, *cut2},)

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
	video, err := moviego.NewVideoFile(TEST_VIDEO_PATH)
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
