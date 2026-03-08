package speed_test

import (
	"math"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestSpeedBasic(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	originalDuration := video.GetDuration()

	spedUp, err := video.Speed(2.0)
	if err != nil {
		t.Fatalf("Failed to speed up video: %v", err)
	}

	const outputPath = "output/speed_2x.mp4"
	err = spedUp.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	expectedDuration := originalDuration / 2
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

func TestSpeedSlowDown(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	originalDuration := video.GetDuration()

	slowedDown, err := video.Speed(0.5)
	if err != nil {
		t.Fatalf("Failed to slow down video: %v", err)
	}

	const outputPath = "output/speed_half.mp4"
	err = slowedDown.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	expectedDuration := originalDuration * 2
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

func TestSpeedSlowDownWithCut(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}

	cut, err := video.Cut(0, 4)
	if err != nil {
		t.Fatalf("Failed to cut video: %v", err)
	}
	slowedDown, err := cut.Speed(0.5)
	if err != nil {
		t.Fatalf("Failed to slow down cut: %v", err)
	}

	const outputPath = "output/speed_slowdown_with_cut.mp4"
	err = slowedDown.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	expectedDuration := 4.0 / 0.5 // 4s cut at 0.5x speed = 8s
	if math.Abs(exportedVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, exportedVideo.GetDuration())
	}
}

func TestSpeedInvalid(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}

	badSpeed, err := video.Speed(0)
	if err == nil {
		t.Fatalf("Expected error for Speed(0), got nil")
	}
	if badSpeed != nil {
		t.Fatalf("Expected nil video for Speed(0), got %v", badSpeed)
	}

	badSpeed, err = video.Speed(-1)
	if err == nil {
		t.Fatalf("Expected error for Speed(-1), got nil")
	}
	if badSpeed != nil {
		t.Fatalf("Expected nil video for Speed(-1), got %v", badSpeed)
	}

	badSpeed, err = video.Speed(2, -0.5)
	if err == nil {
		t.Fatalf("Expected error for Speed(2, -0.5), got nil")
	}
	if badSpeed != nil {
		t.Fatalf("Expected nil video for Speed(2, -0.5), got %v", badSpeed)
	}
}

func TestSpeedWithCut(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}

	cut, err := video.Cut(0, 5)
	if err != nil {
		t.Fatalf("Failed to cut video: %v", err)
	}
	spedUp, err := cut.Speed(2.0)
	if err != nil {
		t.Fatalf("Failed to speed up cut: %v", err)
	}

	const outputPath = "output/speed_with_cut.mp4"
	err = spedUp.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	expectedDuration := 5.0 / 2.0 // 5s cut at 2x speed = 2.5s
	if math.Abs(exportedVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, exportedVideo.GetDuration())
	}
}

func TestSpeedNested(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	originalDuration := video.GetDuration()

	spedUp, err := video.Speed(2.0)
	if err != nil {
		t.Fatalf("Failed to speed up: %v", err)
	}
	netOneX, err := spedUp.Speed(0.5)
	if err != nil {
		t.Fatalf("Failed to slow down: %v", err)
	}

	const outputPath = "output/speed_nested.mp4"
	err = netOneX.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	if math.Abs(exportedVideo.GetDuration()-originalDuration) > 0.1 {
		t.Fatalf("Expected duration %f (net 1x), got %f", originalDuration, exportedVideo.GetDuration())
	}
}

func TestSpeedWithPitch(t *testing.T) {
	video, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
	originalDuration := video.GetDuration()

	spedUpWithPitch, err := video.Speed(2.0, 1.5)
	if err != nil {
		t.Fatalf("Failed to speed up with pitch: %v", err)
	}

	const outputPath = "output/speed_2x_pitch_up.mp4"
	err = spedUpWithPitch.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write video: %v", err)
	}

	exportedVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load output: %v", err)
	}

	expectedDuration := originalDuration / 2
	if math.Abs(exportedVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, exportedVideo.GetDuration())
	}
}
