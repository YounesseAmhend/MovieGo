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
	const endTimeCut2 = 3

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
	result, err := moviego.Concatenate([]moviego.Video{*cut1, *cut2}, outputPath)
	if err != nil {
		t.Fatalf("Failed to concatenate videos: %v", err)
	}
	if result == nil {
		t.Fatalf("Concatenate returned nil video")
	}

	exportedVideo, err := moviego.NewVideoFile(result.GetFilenames()[0])
	if err != nil {
		t.Fatalf("Failed to load concatenated video file: %v", err)
	}

	if math.Abs(exportedVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, exportedVideo.GetDuration())
	}
	if exportedVideo.GetWidth() != video1.GetWidth() {
		t.Fatalf("Expected width %d, got %d", video1.GetWidth(), exportedVideo.GetWidth())
	}
	if exportedVideo.GetHeight() != video1.GetHeight() {
		t.Fatalf("Expected height %d, got %d", video1.GetHeight(), exportedVideo.GetHeight())
	}
}
