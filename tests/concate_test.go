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

	cut1, err := video1.Cut(0, 3)
	if err != nil {
		t.Fatalf("Failed to cut video (0-3): %v", err)
	}
	cut2, err := video1.Cut(2, 3)
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

	exportedVideo, err := moviego.NewVideoFile(result.GetFilename())
	if err != nil {
		t.Fatalf("Failed to load concatenated video file: %v", err)
	}

	expectedDuration := 4.0 // 3s (0-3) + 1s (2-3)
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
