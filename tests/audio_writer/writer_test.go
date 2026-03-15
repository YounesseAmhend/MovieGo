package audio_writer_test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestWriteBasic(t *testing.T) {
	audio, err := moviego.AudioFile(common.TestAudioPath)
	if err != nil {
		t.Fatalf("Failed to load audio: %v", err)
	}

	outputPath := "output/test_write_basic.mp3"
	err = audio.Write(moviego.AudioParameters{
		OutputPath: outputPath,
		Codec:      moviego.AudioCodecMP3,
		Bitrate:    128,
	})
	if err != nil {
		t.Fatalf("Failed to write audio: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist")
	}
}

func TestWriteWithProgress(t *testing.T) {
	audio, err := moviego.AudioFile(common.TestAudioPath)
	if err != nil {
		t.Fatalf("Failed to load audio: %v", err)
	}

	progressCalled := false
	outputPath := "output/test_write_progress.mp3"
	err = audio.Write(moviego.AudioParameters{
		OutputPath: outputPath,
		OnProgress: func(p moviego.Progress) {
			progressCalled = true
			fmt.Printf("Progress: %.1f%%\n", p.Percentage)
		},
	})
	if err != nil {
		t.Fatalf("Failed to write audio: %v", err)
	}

	if !progressCalled {
		t.Errorf("Progress callback was not called")
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)
	
	// Prepare test audio if it doesn't exist
	if _, err := os.Stat(common.TestAudioPath); os.IsNotExist(err) {
		fmt.Println("Extracting test audio from video...")
		cmd := exec.Command("ffmpeg", "-i", common.TestVideoPath, "-vn", "-acodec", "libmp3lame", "-y", common.TestAudioPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to prepare test audio: %v\n", err)
		}
	}

	code := m.Run()
	
	// Clean up
	// os.Remove(common.TestAudioPath)
	
	os.Exit(code)
}
