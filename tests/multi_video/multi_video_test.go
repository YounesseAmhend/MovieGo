package multivideo_test

import (
	"fmt"
	"math"
	"os"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func requireTestVideo(t *testing.T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("Test video %s not found. Run: make download-extra-test-videos", path)
	}
}

var extraTestVideoPaths = []string{common.TestVideo2Path, common.TestVideo3Path, common.TestVideo4Path}

func TestLoadAllTestVideos(t *testing.T) {
	allPaths := append([]string{common.TestVideoPath}, extraTestVideoPaths...)
	for _, path := range allPaths {
		requireTestVideo(t, path)
		video, err := moviego.NewVideoFile(path)
		if err != nil {
			t.Fatalf("Failed to load %s: %v", path, err)
		}
		if video.GetDuration() <= 0 {
			t.Fatalf("Video %s has zero or negative duration", path)
		}
	}
}

func TestCutMultipleVideos(t *testing.T) {
	allPaths := append([]string{common.TestVideoPath}, extraTestVideoPaths...)
	const cutDuration = 2.0

	for i, path := range allPaths {
		requireTestVideo(t, path)
		video, err := moviego.NewVideoFile(path)
		if err != nil {
			t.Fatalf("Failed to load %s: %v", path, err)
		}

		end := cutDuration
		if video.GetDuration() < cutDuration {
			end = video.GetDuration()
		}
		cut, err := video.Cut(0, end)
		if err != nil {
			t.Fatalf("Failed to cut %s: %v", path, err)
		}

		exportPath := fmt.Sprintf("output/cut_test%d.mp4", i+1)
		err = cut.WriteVideo(moviego.VideoParameters{OutputPath: exportPath})
		if err != nil {
			t.Fatalf("Failed to write cut %s: %v", path, err)
		}

		exportedVideo, err := moviego.NewVideoFile(exportPath)
		if err != nil {
			t.Fatalf("Failed to load cut output %s: %v", exportPath, err)
		}

		expectedDuration := end
		if math.Abs(exportedVideo.GetDuration()-expectedDuration) > 0.1 {
			t.Fatalf("Video %s: expected duration %f, got %f", path, expectedDuration, exportedVideo.GetDuration())
		}
	}
}

func TestConcatenateMultipleVideos(t *testing.T) {
	for _, path := range extraTestVideoPaths {
		requireTestVideo(t, path)
	}

	video2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("Failed to load test2: %v", err)
	}
	video3, err := moviego.NewVideoFile(common.TestVideo3Path)
	if err != nil {
		t.Fatalf("Failed to load test3: %v", err)
	}
	video4, err := moviego.NewVideoFile(common.TestVideo4Path)
	if err != nil {
		t.Fatalf("Failed to load test4: %v", err)
	}

	const segmentDuration = 1.0
	cut2, err := video2.Cut(0, segmentDuration)
	if err != nil {
		t.Fatalf("Failed to cut test2: %v", err)
	}
	cut3, err := video3.Cut(0, segmentDuration)
	if err != nil {
		t.Fatalf("Failed to cut test3: %v", err)
	}
	cut4, err := video4.Cut(0, segmentDuration)
	if err != nil {
		t.Fatalf("Failed to cut test4: %v", err)
	}

	result, err := moviego.Concatenate([]moviego.Video{*cut2, *cut3, *cut4})
	if err != nil {
		t.Fatalf("Failed to concatenate: %v", err)
	}

	const outputPath = "output/multi_concatenated.mp4"
	err = result.WriteVideo(moviego.VideoParameters{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Failed to write concatenated video: %v", err)
	}

	resultVideo, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to load concatenated output: %v", err)
	}

	expectedDuration := 3 * segmentDuration
	if math.Abs(resultVideo.GetDuration()-expectedDuration) > 0.1 {
		t.Fatalf("Expected duration %f, got %f", expectedDuration, resultVideo.GetDuration())
	}
}
