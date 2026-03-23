package complex_test

import (
	"math"
	"os"
	"os/exec"
	"testing"

	moviego "github.com/YounesseAmhend/MovieGo"
	"github.com/YounesseAmhend/MovieGo/tests/common"
)

func TestComplexIntegration(t *testing.T) {
	// 1. Multiple Video Clips
	v1, err := moviego.NewVideoFile(common.TestVideoPath)
	if err != nil {
		t.Fatalf("v1 error: %v", err)
	}
	v2, err := moviego.NewVideoFile(common.TestVideo2Path)
	if err != nil {
		t.Fatalf("v2 error: %v", err)
	}
	v3, err := moviego.NewVideoFile(common.TestVideo3Path)
	if err != nil {
		t.Fatalf("v3 error: %v", err)
	}

	// 2. Audio Background
	bgAudio, err := moviego.AudioFile(common.TestAudioPath)
	if err != nil {
		t.Fatalf("bgAudio error: %v", err)
	}

	// --- Process Clip 1 ---
	clip1, _ := v1.Cut(1, 4) // 3 seconds
	clip1, _ = clip1.Scale(moviego.ScaleParams{Width: 1280, Height: 720})
	clip1, _ = clip1.Rotate(math.Pi / 10)
	clip1, _ = clip1.Eq(moviego.EqParams{Brightness: moviego.F(0.1), Contrast: moviego.F(1.2), Saturation: moviego.F(1.5), Gamma: moviego.F(1.0)})
	clip1, _ = clip1.FadeIn(1.0)
	// Audio
	clip1Audio := clip1.GetAudio()
	clip1Audio, _ = clip1Audio.BassBoost(10)
	clip1Audio, _ = clip1Audio.Echo(500, 0.4)
	clip1.SetAudio(*clip1Audio)

	// --- Process Clip 2 ---
	clip2, _ := v2.Cut(0, 3) // 3 seconds
	clip2, _ = clip2.Scale(moviego.ScaleParams{Width: 1280, Height: 720})
	clip2, _ = clip2.Speed(1.5)           // now 2 seconds
	clip2, _ = clip2.HorizontalFlip()
	clip2, _ = clip2.Sepia()

	// --- Process Clip 3 (to overlay) ---
	clip3, _ := v3.Cut(0, 2)
	clip3, _ = clip3.Scale(moviego.ScaleParams{Width: 400, Height: 300})
	clip3, _ = clip3.Vignette(math.Pi / 4)
	clip3 = clip3.SetPosition(moviego.Position{X: "100", Y: "100"})

	// --- Additional processing for Clip 1 (since text isn't working on this environment) ---
	clip1, _ = clip1.Pad(moviego.PadParams{
		Width:  1400,
		Height: 800,
		X:      60,
		Y:      40,
	})
	clip1, _ = clip1.Crop(moviego.CropParams{
		Width:  1280,
		Height: 720,
		X:      60,
		Y:      40,
	})
	clip1, _ = clip1.Blur(1.5)
	clip1, _ = clip1.Sharpen(1.2)

	// --- Concatenate clip1 and clip2 with a transition (e.g. WipeLeft) ---
	// clip1 is ~3s, clip2 is ~2s. overlap = 1s -> total = 4s
	combined, err := moviego.ConcatenateWithTransition(clip1, clip2, moviego.TransitionParams{
		Transition: moviego.TransitionWipeLeft,
		Duration:   1.0,
	})
	if err != nil {
		t.Fatalf("Transition error: %v", err)
	}

	// --- Composite clip3 on top of the joined video ---
	// combined is 4s, clip3 is 2s.
	composited, err := moviego.CompositeClip([]moviego.Video{*combined, *clip3})
	if err != nil {
		t.Fatalf("Composite error: %v", err)
	}

	// --- Audio Mastering ---
	bgAudio, _ = bgAudio.Volume(0.2)
	bgAudio, _ = bgAudio.TrebleBoost(5)
	
	// Mix audio from composited with background audio
	mixedAudio, err := moviego.Composite([]moviego.Audio{*composited.GetAudio(), *bgAudio})
	if err != nil {
		t.Fatalf("Audio mix error: %v", err)
	}
	mixedAudio, _ = mixedAudio.Normalize() // Final master polish
	composited.SetAudio(*mixedAudio)

	// Add final fade out
	composited, _ = composited.FadeOut(0.5)

	// Render
	outputPath := "output/extremely_complex_test.mp4"
	err = composited.WriteVideo(moviego.VideoParameters{
		OutputPath:     outputPath,
		Codec:          moviego.CodecH264,
		Fps:            30,
		Bitrate:        "2M",
		Preset:         moviego.Fast,
		SilentProgress: true,
	})

	if err != nil {
		t.Fatalf("Failed to write complex video: %v", err)
	}

	// Verify Output
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist")
	}

	out, err := moviego.NewVideoFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to probe output: %v", err)
	}
	if out.GetWidth() != 1280 || out.GetHeight() != 720 {
		t.Errorf("Expected 1280x720 output, got %dx%d", out.GetWidth(), out.GetHeight())
	}
}

func TestMain(m *testing.M) {
	_ = os.MkdirAll("output", 0755)

	// Ensure test data exists
	if _, err := os.Stat(common.TestAudioPath); os.IsNotExist(err) {
		cmd := exec.Command("ffmpeg", "-i", common.TestVideoPath, "-vn", "-acodec", "libmp3lame", "-y", common.TestAudioPath)
		_ = cmd.Run()
	}

	os.Exit(m.Run())
}
