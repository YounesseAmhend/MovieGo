package tests

import (
	"math"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
)

func TestCut(t *testing.T) {
	video, err := moviego.NewVideoFile(TEST_VIDEO_PATH)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	_, err = video.Cut(0, 10)
	if err != nil {
		t.Fatalf("Failed to cut video: %v", err)
	}

	badCut, err := video.Cut(10, 0)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	if badCut != nil {
		t.Fatalf("Expected nil, got %v", badCut)
	}

	const start = 0;
	end := video.GetDuration() / 2;
	goodCut, err := video.Cut(start, end)

	const exportPath = "output/good_cut.mp4"
	err = goodCut.WriteVideo(moviego.VideoParameters{
		OutputPath: exportPath,
	})

	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(exportPath)

	exportedDuration := exportedVideo.GetDuration()
	if math.Abs(exportedDuration - (end - start)) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", end - start, exportedDuration)
	}

	if exportedVideo.GetWidth() != video.GetWidth() {
		t.Fatalf("Expected width %d, got %d", video.GetWidth(), exportedVideo.GetWidth())
	}
	if exportedVideo.GetHeight() != video.GetHeight() {
		t.Fatalf("Expected height %d, got %d", video.GetHeight(), exportedVideo.GetHeight())
	}
}

func TestNestedCut(t *testing.T) {
	video, err := moviego.NewVideoFile(TEST_VIDEO_PATH)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}

	halfDuration := video.GetDuration() / 2
	firstCut, err := video.Cut(0, halfDuration)
	if err != nil {
		t.Fatalf("Failed first cut: %v", err)
	}

	quarterDuration := firstCut.GetDuration() / 2
	nestedCut, err := firstCut.Cut(0, quarterDuration)
	if err != nil {
		t.Fatalf("Failed nested cut: %v", err)
	}

	const exportPath = "output/nested_cut.mp4"
	err = nestedCut.WriteVideo(moviego.VideoParameters{
		OutputPath: exportPath,
	})
	if err != nil {
		t.Fatalf("Failed to write nested cut: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to load nested cut output: %v", err)
	}

	expectedDuration := video.GetDuration() / 4
	if math.Abs(exportedVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, exportedVideo.GetDuration())
	}
	if exportedVideo.GetWidth() != video.GetWidth() {
		t.Fatalf("Expected width %d, got %d", video.GetWidth(), exportedVideo.GetWidth())
	}
	if exportedVideo.GetHeight() != video.GetHeight() {
		t.Fatalf("Expected height %d, got %d", video.GetHeight(), exportedVideo.GetHeight())
	}
}
