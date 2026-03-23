package audio_filters_test

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestFilters(t *testing.T) {
	audio, err := moviego.AudioFile(common.TestAudioPath)
	if err != nil {
		t.Fatalf("Failed to load audio: %v", err)
	}

	tests := []struct {
		name  string
		apply func(*moviego.Audio) (*moviego.Audio, error)
	}{
		{"volume", func(a *moviego.Audio) (*moviego.Audio, error) { return a.Volume(0.5) }},
		{"fade_in", func(a *moviego.Audio) (*moviego.Audio, error) { return a.FadeIn(1.0) }},
		{"fade_out", func(a *moviego.Audio) (*moviego.Audio, error) { return a.FadeOut(1.0) }},
		{"lowpass", func(a *moviego.Audio) (*moviego.Audio, error) { return a.LowPass(1000) }},
		{"highpass", func(a *moviego.Audio) (*moviego.Audio, error) { return a.HighPass(2000) }},
		{"tempo", func(a *moviego.Audio) (*moviego.Audio, error) { return a.Tempo(1.5) }},
		{"bass_boost", func(a *moviego.Audio) (*moviego.Audio, error) { return a.BassBoost(5) }},
		{"treble_boost", func(a *moviego.Audio) (*moviego.Audio, error) { return a.TrebleBoost(5) }},
		{"normalize", func(a *moviego.Audio) (*moviego.Audio, error) { return a.Normalize() }},
		{"echo", func(a *moviego.Audio) (*moviego.Audio, error) { return a.Echo(500, 0.5) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered, err := tt.apply(audio)
			if err != nil {
				t.Fatalf("Failed to apply filter %s: %v", tt.name, err)
			}

			outputPath := fmt.Sprintf("output/test_filter_%s.mp3", tt.name)
			err = filtered.Write(moviego.AudioParameters{
				OutputPath:     outputPath,
				SilentProgress: true,
			})
			if err != nil {
				t.Fatalf("Failed to write filtered audio %s: %v", tt.name, err)
			}

			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Fatalf("Output file %s does not exist", outputPath)
			}

			out, err := moviego.AudioFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to probe output %s: %v", outputPath, err)
			}

			if tt.name == "tempo" {
				expectedDuration := audio.GetDuration() / 1.5
				if math.Abs(out.GetDuration()-expectedDuration) > 0.5 {
					t.Errorf("Expected duration ~%f, got %f", expectedDuration, out.GetDuration())
				}
			}
		})
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
