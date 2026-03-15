package composite_audio_test

import (
	"math"
	"os"
	"os/exec"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestComposite(t *testing.T) {
	audio1, err := moviego.AudioFile(common.TestAudioPath)
	if err != nil {
		t.Fatalf("Failed to load audio1: %v", err)
	}

	audio2, err := moviego.AudioFile(common.TestAudioPath)
	if err != nil {
		t.Fatalf("Failed to load audio2: %v", err)
	}
	
	audio2, err = audio2.Volume(0.5)
	if err != nil {
		t.Fatalf("Failed to apply volume to audio2: %v", err)
	}

	result, err := moviego.Composite([]moviego.Audio{*audio1, *audio2})
	if err != nil {
		t.Fatalf("Failed to composite: %v", err)
	}

	outputPath := "output/test_composite.mp3"
	err = result.Write(moviego.AudioParameters{
		OutputPath:     outputPath,
		SilentProgress: true,
	})
	if err != nil {
		t.Fatalf("Failed to write composite: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist")
	}

	out, err := moviego.AudioFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to probe output: %v", err)
	}

	expectedDuration := audio1.GetDuration()
	if math.Abs(out.GetDuration()-expectedDuration) > 0.5 {
		t.Errorf("Expected duration ~%f, got %f", expectedDuration, out.GetDuration())
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	
	if _, err := os.Stat(common.TestAudioPath); os.IsNotExist(err) {
		cmd := exec.Command("ffmpeg", "-i", common.TestVideoPath, "-vn", "-acodec", "libmp3lame", "-y", common.TestAudioPath)
		_ = cmd.Run()
	}

	os.Exit(m.Run())
}
