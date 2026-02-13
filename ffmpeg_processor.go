package moviego

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/YounesseAmhend/MovieGo/utils"
	"github.com/fatih/color"
)

// TempFiles holds paths to temporary files used during video processing
type TempFiles struct {
	VideoPath string
	AudioPath string
}

// createTempFiles creates temporary files for video and audio processing
func createTempFiles() (*TempFiles, error) {
	tempVideoFile, err := os.CreateTemp("", "moviego_video_*.mp4")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary video file: %w", err)
	}
	tempVideoPath := tempVideoFile.Name()
	tempVideoFile.Close()

	tempAudioFile, err := os.CreateTemp("", "moviego_audio_*.aac")
	if err != nil {
		os.Remove(tempVideoPath)
		return nil, fmt.Errorf("failed to create temporary audio file: %w", err)
	}
	tempAudioPath := tempAudioFile.Name()
	tempAudioFile.Close()

	return &TempFiles{
		VideoPath: tempVideoPath,
		AudioPath: tempAudioPath,
	}, nil
}

// Cleanup removes temporary files
func (tf *TempFiles) Cleanup() {
	if tf.VideoPath != "" {
		os.Remove(tf.VideoPath)
	}
	if tf.AudioPath != "" {
		os.Remove(tf.AudioPath)
	}
}

// extractAudio extracts audio from the source video
func extractAudio(sourceVideo string, outputPath string, threads int16) error {
	audioCmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-i", sourceVideo,
		"-vn", // No video
		"-threads", fmt.Sprintf("%d", threads),
		"-acodec", "copy", // Copy audio codec
		"-y",
		outputPath,
	)

	if err := audioCmd.Run(); err != nil {
		return fmt.Errorf("could not extract audio: %w", err)
	}
	return nil
}

// buildInputCommand creates the FFmpeg command to read frames from video
func buildInputCommand(video *Video, threads uint16) (*exec.Cmd, error) {
	// Validate video properties
	if video.GetFilename() == "" {
		return nil, fmt.Errorf("video filename is empty")
	}
	if video.GetFps() <= 0 {
		return nil, fmt.Errorf("video FPS is invalid (%d), must be greater than 0", video.GetFps())
	}
	if video.GetWidth() <= 0 || video.GetHeight() <= 0 {
		return nil, fmt.Errorf("video dimensions are invalid (%dx%d)", video.GetWidth(), video.GetHeight())
	}

	args := []string{"-loglevel", "error"}

	// Add start time if specified (subclip)
	if video.GetStartTime() > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", video.GetStartTime()))
	}

	// Add input file
	args = append(args, "-i", video.GetFilename())

	// Add duration if end time is specified (subclip)
	if video.GetEndTime() > 0 {
		duration := video.GetEndTime() - video.GetStartTime()
		args = append(args, "-t", fmt.Sprintf("%.3f", duration))
	}

	// Add remaining parameters
	args = append(args,
		"-f", "rawvideo",
		"-threads", fmt.Sprintf("%d", threads),
		"-pix_fmt", "rgba",
		"-r", fmt.Sprintf("%d", video.GetFps()),
		"-",
	)

	return exec.Command("ffmpeg", args...), nil
}

// buildOutputCommand creates the FFmpeg command to encode raw frames
func buildOutputCommand(video *Video, outputPath string, threads uint16) (*exec.Cmd, error) {
	// Validate video properties
	if video.GetFps() <= 0 {
		return nil, fmt.Errorf("video FPS is invalid (%d), must be greater than 0", video.GetFps())
	}
	if video.GetWidth() <= 0 || video.GetHeight() <= 0 {
		return nil, fmt.Errorf("video dimensions are invalid (%dx%d)", video.GetWidth(), video.GetHeight())
	}
	if outputPath == "" {
		return nil, fmt.Errorf("output path is empty")
	}

	codec := resolveCodec(video.GetCodec(), "")

	presetValue := video.GetPreset()
	if presetValue == "" {
		presetValue = Medium
	}

	// Map preset to codec-appropriate value
	presetStr := mapPresetForCodec(codec, string(presetValue))

	withMask := video.GetWithMask()
	// Internal processing in MovieGo currently always uses 4 bytes per pixel (RGBA)
	inputPixelFormat := "rgba"

	args := []string{
		"-loglevel", "warning",
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", video.width, video.height),
		"-pix_fmt", inputPixelFormat,
		"-r", fmt.Sprintf("%d", video.GetFps()),
		"-an",
		"-i", "-",
	}

	// Codec (use -c:v for consistency)
	args = append(args, "-c:v", codec)

	// Preset (only add if not empty - some encoders don't use preset)
	if presetStr != "" {
		args = append(args, "-preset", presetStr)
	}

	// Custom FFmpeg arguments
	if video.ffmpegArgs != nil {
		for key, values := range video.ffmpegArgs {
			for _, value := range values {
				args = append(args, key, value)
			}
		}
	}

	// Bitrate
	if video.GetBitRate() != "" {
		args = append(args, "-b:v", video.GetBitRate())
	}

	// Threads
	if threads != 0 {
		args = append(args, "-threads", fmt.Sprintf("%d", threads))
	}

	// Pixel format and special flags
	isH264 := codec == "libx264" || codec == "h264_nvenc" || codec == "h264"
	isH265 := codec == "libx265" || codec == "hevc"
	isVPX := codec == "libvpx" || codec == "libvpx-vp9" || codec == "vp8" || codec == "vp9"

	if isVPX && withMask {
		args = append(args, "-pix_fmt", "yuva420p", "-auto-alt-ref", "0")
	} else if isH264 || isH265 || isVPX {
		if video.width%2 == 0 && video.height%2 == 0 {
			if withMask && (isH264 || isH265) {
				args = append(args, "-pix_fmt", "yuva420p")
			} else {
				args = append(args, "-pix_fmt", "yuv420p")
			}
		} else {
			// Odd dimensions: yuv420p/yuva420p won't work for many encoders
			args = append(args, "-pix_fmt", "yuv444p")
		}
	} else {
		pixFmt := video.GetPixelFormat()
		if pixFmt == "" {
			if withMask {
				pixFmt = "rgba"
			} else {
				// Choose pixel format based on codec for optimal compatibility
				switch codec {
				case "h264_amf", "hevc_amf":
					pixFmt = "nv12"
				case "h264_nvenc", "hevc_nvenc":
					pixFmt = "nv12"
				case "h264_qsv", "hevc_qsv":
					pixFmt = "nv12"
				default:
					pixFmt = "rgb24"
				}
			}
		}
		args = append(args, "-pix_fmt", string(pixFmt))
	}

	// Alpha mode for webm
	if isVPX && withMask {
		args = append(args, "-metadata:s:v:0", "alpha_mode=1")
	}

	// Overwrite and output path
	args = append(args, "-y", outputPath)

	return exec.Command("ffmpeg", args...), nil
}

// PipeSetup holds the pipes for input and output FFmpeg processes
type PipeSetup struct {
	InputStdout  io.ReadCloser
	OutputStdin  io.WriteCloser
	InputStderr  io.ReadCloser
	OutputStderr io.ReadCloser
}

// setupPipes sets up pipes between FFmpeg processes
func setupPipes(inputCmd, outputCmd *exec.Cmd) (*PipeSetup, error) {
	inputStdout, err := inputCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create input stdout pipe: %w", err)
	}

	outputStdin, err := outputCmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create output stdin pipe: %w", err)
	}

	inputStderr, err := inputCmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create input stderr pipe: %w", err)
	}

	outputStderr, err := outputCmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create output stderr pipe: %w", err)
	}

	// Forward stderr to console
	go io.Copy(os.Stderr, inputStderr)
	go io.Copy(os.Stderr, outputStderr)

	return &PipeSetup{
		InputStdout:  inputStdout,
		OutputStdin:  outputStdin,
		InputStderr:  inputStderr,
		OutputStderr: outputStderr,
	}, nil
}

// combineAudioVideo combines the processed video with original audio
func combineAudioVideo(video *Video, processedVideoPath, outputPath string) error {
	args := []string{
		"-loglevel", "error",
		"-i", processedVideoPath,
	}

	hasAudio := video.HasAudio()
	var duration float64
	if hasAudio {
		// Add audio input with subclip timing if specified
		if video.GetStartTime() > 0 {
			args = append(args, "-ss", fmt.Sprintf("%.3f", video.GetStartTime()))
		}
		args = append(args, "-i", video.GetFilename())
		if video.GetEndTime() > 0 {
			duration = video.GetEndTime() - video.GetStartTime()
			args = append(args, "-t", fmt.Sprintf("%.3f", duration))
		} else {
			duration = video.GetDuration()
		}
	} else {
		duration = video.GetDuration()
	}

	args = append(args, "-c:v", "copy")

	if hasAudio {
		args = append(args, "-c:a", "copy", "-map", "0:v:0", "-map", "1:a:0", "-shortest")
	} else {
		args = append(args, "-map", "0:v:0")
	}

	args = append(args, "-y", outputPath)

	// Use progress parser for combining audio
	config := FFmpegProgressConfig{
		Args:          args,
		TotalDuration: duration,
		OperationName: "Combining audio",
		OutputPath:    outputPath,
		Bitrate:       video.GetBitRate(),
	}

	if err := runFFmpegWithProgress(config); err != nil {
		// If combining fails, try to just copy the processed video
		fmt.Printf("%s Could not combine audio: %v. Copying video without audio.\n", color.YellowString("⚠ Warning:"), err)
		if err := utils.CopyFile(processedVideoPath, outputPath); err != nil {
			return fmt.Errorf("failed to copy temporary video to output: %w", err)
		}
	}

	return nil
}
