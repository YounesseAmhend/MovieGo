package moviego

import (
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func writeFilterComplex(b *strings.Builder, raw string) {
	for _, segment := range strings.Split(raw, ";") {
		segment = strings.TrimSpace(segment)
		if segment != "" {
			b.WriteString("    ")
			b.WriteString(segment)
			b.WriteString(";\n")
		}
	}
}

// formatCmd formats an exec.Cmd into a readable multi-line string.
// The -filter_complex value is split by ";" so each filter gets its own line.
func formatCmd(cmd *exec.Cmd) string {
	args := cmd.Args
	if len(args) == 0 {
		return ""
	}

	var b strings.Builder
	exe := filepath.Base(args[0])
	if strings.HasSuffix(strings.ToLower(exe), ".exe") {
		exe = exe[:len(exe)-4]
	}
	b.WriteString(exe)
	b.WriteByte('\n')

	for i := 1; i < len(args); i++ {
		arg := args[i]

		if arg == "-filter_complex" && i+1 < len(args) {
			b.WriteString("  -filter_complex\n")
			i++
			writeFilterComplex(&b, args[i])
			continue
		}

		if strings.HasPrefix(arg, "-") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			b.WriteString("  ")
			b.WriteString(arg)
			b.WriteByte(' ')
			b.WriteString(args[i+1])
			b.WriteByte('\n')
			i++
		} else {
			b.WriteString("  ")
			b.WriteString(arg)
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// WriteVideo processes the video with applied filters and writes to output file
func (v *Video) WriteVideo(parms VideoParameters) error {
	if parms.OutputPath == "" {
		return fmt.Errorf("output path is empty, cannot write video")
	}

	// Validate essential video properties before processing
	if len(v.GetFilenames()) == 0 {
		return fmt.Errorf("video filename is empty, cannot process video")
	}
	if v.GetWidth() <= 0 || v.GetHeight() <= 0 {
		return fmt.Errorf("video dimensions are invalid (%dx%d), cannot process video", v.GetWidth(), v.GetHeight())
	}
	if v.GetDuration() <= 0 {
		return fmt.Errorf("video duration is invalid (%.2f), cannot process video", v.GetDuration())
	}

	// Apply parameters to video
	v.applyParameters(parms)


	ffmpegPath, err := getFFmpegPath()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg path: %w", err)
	}

	ffmpegArgs := []string{}
	for _, filename := range v.GetFilenames() {
		ffmpegArgs = append(ffmpegArgs, "-i", filename)
	}

	filterComplex := ""

	// split part
	for i, filename := range v.GetFilenames() {
		videoLabels := []string{}
		for _, filter := range v.videoFilterComplex {
			currentFilename := filter.FileCopy.Filename
			if currentFilename == filename {
				videoLabels = append(videoLabels, filter.FileCopy.Label)
			}
		}
		filterComplex += fmt.Sprintf("[%d:v]split=%d[%s];", i, len(videoLabels), strings.Join(videoLabels, "]["))

		audioLabels := []string{}
		for _, filter := range v.audioFilterComplex {
			currentFilename := filter.FileCopy.Filename
			if currentFilename == filename {
				audioLabels = append(audioLabels, filter.FileCopy.Label)
			}
		}
		filterComplex += fmt.Sprintf("[%d:a]asplit=%d[%s];", i, len(audioLabels), strings.Join(audioLabels, "]["))

	}

	videoIndex, audioIndex := 0, 0
	videoLen, audioLen := len(v.videoFilterComplex), len(v.audioFilterComplex)

	for videoIndex < videoLen || audioIndex < audioLen {
		nextOrder := uint64(math.MaxUint64)
		if videoIndex < videoLen {
			nextOrder = v.videoFilterComplex[videoIndex].Order
		}
		if audioIndex < audioLen && v.audioFilterComplex[audioIndex].Order < nextOrder {
			nextOrder = v.audioFilterComplex[audioIndex].Order
		}

		if videoIndex < videoLen && v.videoFilterComplex[videoIndex].Order == nextOrder {
			filter := v.videoFilterComplex[videoIndex]
			if filter.FilterElement != "" {
				filterComplex += filter.FilterElement
				if !strings.HasSuffix(filter.FilterElement, "]") {
					filterComplex += fmt.Sprintf("[%s]", filter.Label)
				}
				filterComplex += ";"
			}
			videoIndex++
		}
		if audioIndex < audioLen && v.audioFilterComplex[audioIndex].Order == nextOrder {
			filter := v.audioFilterComplex[audioIndex]
			if filter.FilterElement != "" {
				filterComplex += filter.FilterElement
				if !strings.HasSuffix(filter.FilterElement, "]") {
					filterComplex += fmt.Sprintf("[%s]", filter.Label)
				}
				filterComplex += ";"
			}
			audioIndex++
		}
	}

	audioLabel := v.lastAudioLabel()
	videoLabel := v.lastVideoLabel()

	if videoLabel == "" {
		return fmt.Errorf("no video output label generated")
	}
	if audioLabel == "" {
		return fmt.Errorf("no audio output label generated")
	}

	mapVideo := fmt.Sprintf("[%s]", videoLabel)
	mapAudio := fmt.Sprintf("[%s]", audioLabel)

	// Video codec
	encoder := resolveVideoEncoder(parms.Codec, v.GetCodec())
	ffmpegArgs = append(ffmpegArgs, "-filter_complex", filterComplex, "-map", mapVideo, "-map", mapAudio, "-c:v", encoder)

	// Threads (compute effective value - applyParameters modifies a copy)
	effectiveThreads := parms.Threads
	if effectiveThreads == 0 {
		totalCPUs := runtime.GOMAXPROCS(0)
		if totalCPUs <= 2 {
			effectiveThreads = uint16(totalCPUs)
		} else {
			effectiveThreads = uint16((totalCPUs * 6) / 10)
			if effectiveThreads < 2 {
				effectiveThreads = 2
			}
		}
	}
	ffmpegArgs = append(ffmpegArgs, "-threads", fmt.Sprintf("%d", effectiveThreads))

	// FPS (if set)
	if fps := resolveFps(parms.Fps, v.GetFps()); fps > 0 {
		ffmpegArgs = append(ffmpegArgs, "-r", fmt.Sprintf("%d", fps))
	}

	// Bitrate (if set)
	if br := resolveBitrate(parms.Bitrate, v.GetBitRate()); br != "" {
		ffmpegArgs = append(ffmpegArgs, "-b:v", br)
	}

	// Preset (codec-specific, only for encoders that support it)
	presetStr := resolvePreset(parms.Preset, v.GetPreset())
	mappedPreset := mapPresetForCodec(encoder, presetStr)
	if mappedPreset != "" {
		ffmpegArgs = append(ffmpegArgs, "-preset", mappedPreset)
	}

	// Pixel format (if set)
	if pf := resolvePixelFormat(parms.PixelFormat, v.GetPixelFormat()); pf != "" {
		ffmpegArgs = append(ffmpegArgs, "-pix_fmt", string(pf))
	}

	// Audio codec for MP4 output
	outputExt := strings.ToLower(filepath.Ext(parms.OutputPath))
	if outputExt == ".mp4" || outputExt == ".m4a" {
		ffmpegArgs = append(ffmpegArgs, "-c:a", "aac")
	}

	ffmpegArgs = append(ffmpegArgs, "-metadata:s:v:0", "rotate=0", "-y", parms.OutputPath)

	cmd := exec.Command(ffmpegPath, ffmpegArgs...)
	displayProgram := filepath.Base(ffmpegPath)
	displayProgram = strings.TrimSuffix(displayProgram, filepath.Ext(displayProgram))
	displayCmd := &exec.Cmd{Args: append([]string{displayProgram}, ffmpegArgs...)}

	fmt.Println(formatCmd(displayCmd))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute ffmpeg: %w", err)
	}

	return nil
}

