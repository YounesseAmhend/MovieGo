package moviego

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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
	videoFilenames := v.GetFilenames()
	for _, filename := range videoFilenames {
		ffmpegArgs = append(ffmpegArgs, "-i", filename)
	}

	// Collect audio-only filenames (from v.audio.filenames) that aren't already
	// in the video input list. These come from standalone Audio files mixed in
	// via Composite() + SetAudio().
	videoFilenameSet := make(map[string]struct{}, len(videoFilenames))
	for _, fn := range videoFilenames {
		videoFilenameSet[fn] = struct{}{}
	}
	var audioOnlyFilenames []string
	for _, fn := range v.audio.filenames {
		if _, exists := videoFilenameSet[fn]; !exists {
			audioOnlyFilenames = append(audioOnlyFilenames, fn)
			videoFilenameSet[fn] = struct{}{} // avoid duplicates
		}
	}
	for _, filename := range audioOnlyFilenames {
		ffmpegArgs = append(ffmpegArgs, "-i", filename)
	}

	filterComplex := ""

	// split part – video+audio inputs
	for i, filename := range videoFilenames {
		videoLabels := []string{}
		for _, filter := range v.filterComplex {
			currentFilename := filter.FileCopy.Filename
			if currentFilename == filename {
				videoLabels = append(videoLabels, filter.FileCopy.Label)
			}
		}
		filterComplex += fmt.Sprintf("[%d:v]format=yuva420p,split=%d[%s];", i, len(videoLabels), strings.Join(videoLabels, "]["))

		audioLabels := []string{}
		for _, filter := range v.audio.filterComplex {
			currentFilename := filter.FileCopy.Filename
			if currentFilename == filename {
				audioLabels = append(audioLabels, filter.FileCopy.Label)
			}
		}
		filterComplex += fmt.Sprintf("[%d:a]asplit=%d[%s];", i, len(audioLabels), strings.Join(audioLabels, "]["))

	}

	// split part – audio-only inputs (no video stream)
	for j, filename := range audioOnlyFilenames {
		inputIndex := len(videoFilenames) + j
		audioLabels := []string{}
		for _, filter := range v.audio.filterComplex {
			if filter.FileCopy.Filename == filename {
				audioLabels = append(audioLabels, filter.FileCopy.Label)
			}
		}
		if len(audioLabels) > 1 {
			filterComplex += fmt.Sprintf("[%d:a]asplit=%d[%s];", inputIndex, len(audioLabels), strings.Join(audioLabels, "]["))
		} else if len(audioLabels) == 1 {
			filterComplex += fmt.Sprintf("[%d:a]anull[%s];", inputIndex, audioLabels[0])
		}
	}

	videoIndex, audioIndex := 0, 0
	videoLen, audioLen := len(v.filterComplex), len(v.audio.filterComplex)

	for videoIndex < videoLen || audioIndex < audioLen {
		nextOrder := uint64(math.MaxUint64)
		if videoIndex < videoLen {
			nextOrder = v.filterComplex[videoIndex].Order
		}
		if audioIndex < audioLen && v.audio.filterComplex[audioIndex].Order < nextOrder {
			nextOrder = v.audio.filterComplex[audioIndex].Order
		}

		if videoIndex < videoLen && v.filterComplex[videoIndex].Order == nextOrder {
			filter := v.filterComplex[videoIndex]
			if filter.FilterElement != "" {
				filterComplex += filter.FilterElement
				if !strings.HasSuffix(filter.FilterElement, "]") {
					filterComplex += fmt.Sprintf("[%s]", filter.Label)
				}
				filterComplex += ";"
			}
			videoIndex++
		}
		if audioIndex < audioLen && v.audio.filterComplex[audioIndex].Order == nextOrder {
			filter := v.audio.filterComplex[audioIndex]
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

	videoLabel := v.lastVideoLabel()
	if videoLabel == "" {
		return fmt.Errorf("no video output label generated")
	}

	audioLabel := v.audio.lastAudioLabel()
	if audioLabel == "" {
		return fmt.Errorf("no audio output label generated")
	}

	mapVideo := fmt.Sprintf("[%s]", videoLabel)
	encoder := resolveVideoEncoder(parms.Codec, v.GetCodec())

	mapAudio := fmt.Sprintf("[%s]", audioLabel)
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

	// Pixel format (default to yuv420p to strip alpha from internal YUVA pipeline)
	pf := resolvePixelFormat(parms.PixelFormat, v.GetPixelFormat())
	if pf == "" {
		pf = PixelFormatYUV420P
	}
	ffmpegArgs = append(ffmpegArgs, "-pix_fmt", string(pf))

	// Audio codec for MP4 output
	outputExt := strings.ToLower(filepath.Ext(parms.OutputPath))
	if outputExt == ".mp4" || outputExt == ".m4a" {
		ffmpegArgs = append(ffmpegArgs, "-c:a", "aac")
	}

	progressEnabled := parms.OnProgress != nil || !parms.SilentProgress
	if progressEnabled {
		ffmpegArgs = append(ffmpegArgs, "-progress", "pipe:1", "-nostats")
	}

	ffmpegArgs = append(ffmpegArgs, "-metadata:s:v:0", "rotate=0", "-y", parms.OutputPath)

	cmd := exec.Command(ffmpegPath, ffmpegArgs...)
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	displayProgram := filepath.Base(ffmpegPath)
	displayProgram = strings.TrimSuffix(displayProgram, filepath.Ext(displayProgram))
	displayCmd := &exec.Cmd{Args: append([]string{displayProgram}, ffmpegArgs...)}

	fmt.Println(formatCmd(displayCmd))

	if progressEnabled {
		handler := parms.OnProgress
		if handler == nil {
			handler = defaultProgressHandler(parms.OutputPath)
		}
		return v.runWithProgress(cmd, &stderrBuf, handler)
	}

	if err := cmd.Run(); err != nil {
		stderr := strings.TrimSpace(stderrBuf.String())
		if stderr != "" {
			return fmt.Errorf("failed to execute ffmpeg: %w\nffmpeg stderr: %s", err, stderr)
		}
		return fmt.Errorf("failed to execute ffmpeg: %w", err)
	}

	return nil
}

func (v *Video) runWithProgress(cmd *exec.Cmd, stderrBuf *bytes.Buffer, onProgress func(Progress)) error {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	totalDuration := v.GetDuration()
	scanner := bufio.NewScanner(stdoutPipe)
	cur := Progress{TotalDuration: totalDuration}

	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		switch key {
		case "frame":
			cur.Frame, _ = strconv.ParseInt(value, 10, 64)
		case "fps":
			cur.FPS, _ = strconv.ParseFloat(value, 64)
		case "bitrate":
			cur.Bitrate = value
		case "out_time_us":
			us, _ := strconv.ParseInt(value, 10, 64)
			cur.OutTime = float64(us) / 1_000_000
			if totalDuration > 0 {
				cur.Percentage = math.Min((cur.OutTime/totalDuration)*100, 100)
			}
		case "speed":
			s := strings.TrimSuffix(strings.TrimSpace(value), "x")
			cur.Speed, _ = strconv.ParseFloat(s, 64)
		case "progress":
			cur.Done = value == "end"
			if cur.Done {
				cur.Percentage = 100
			}
			onProgress(cur)
		}
	}

	if err := cmd.Wait(); err != nil {
		stderr := strings.TrimSpace(stderrBuf.String())
		if stderr != "" {
			return fmt.Errorf("failed to execute ffmpeg: %w\nffmpeg stderr: %s", err, stderr)
		}
		return fmt.Errorf("failed to execute ffmpeg: %w", err)
	}

	return nil
}

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorWhite  = "\033[37m"
	clearLine   = "\033[K"
)

func defaultProgressHandler(outputPath string) func(Progress) {
	filename := filepath.Base(outputPath)
	return func(p Progress) {
		if p.Done {
			bar := strings.Repeat("█", 30)
			dur := formatDuration(p.TotalDuration)
			speed := formatSpeed(p.Speed)
			line := "\r" + colorGreen + colorBold + fmt.Sprintf("%-18s", filename) + colorReset +
				" " + colorGreen + bar + colorReset + " " +
				colorCyan + "100.0%" + colorReset + "  " +
				colorCyan + dur + colorReset + "  " +
				colorYellow + speed + colorReset + clearLine + "\n"
			fmt.Fprint(os.Stderr, line)
			return
		}

		pct := p.Percentage
		filled := int(pct / 100 * 30)
		if filled > 30 {
			filled = 30
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", 30-filled)

		elapsed := formatDuration(p.OutTime)
		total := formatDuration(p.TotalDuration)

		line := "\r" +
			colorCyan + colorBold + fmt.Sprintf("%-18s", filename) + colorReset +
			" " + colorGreen + bar + colorReset + " " +
			colorWhite + fmt.Sprintf("%5.1f%%", pct) + colorReset + "  " +
			colorDim + elapsed + " / " + total + colorReset + "  " +
			colorYellow + formatSpeed(p.Speed) + colorReset + clearLine
		fmt.Fprint(os.Stderr, line)
	}
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "00:00"
	}
	m := int(seconds) / 60
	s := int(seconds) % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func formatSpeed(speed float64) string {
	if speed <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1fx", speed)
}

