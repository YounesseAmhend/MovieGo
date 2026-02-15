package tests


import (
	moviego "github.com/YounesseAmhend/MovieGo"
	"testing"
)


func LoadVideo(t *testing.T) {
	_, err := moviego.NewVideoFile(TEST_VIDEO_PATH)
	if err != nil {
		t.Fatalf("Failed to create video file: %v", err)
	}
}