package tests

import (
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
