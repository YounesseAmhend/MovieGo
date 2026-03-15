package moviego

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// Write processes the audio with applied filters and writes to output file
func (a *Audio) Write(parms AudioParameters) error {
	if parms.OutputPath == "" {
		return fmt.Errorf("output path is empty, cannot write audio")
	}

	if len(a.filenames) == 0 {
		return fmt.Errorf("audio filename is empty, cannot process audio")
	}

	ffmpegPath, err := getFFmpegPath()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg path: %w", err)
	}

	ffmpegArgs := []string{}
	for _, filename := range a.filenames {
		ffmpegArgs = append(ffmpegArgs, "-i", filename)
	}

	filterComplex := ""
	// split part
	for i, filename := range a.filenames {
		audioLabels := []string{}
		for _, filter := range a.filterComplex {
			if filter.FileCopy.Filename == filename {
				audioLabels = append(audioLabels, filter.FileCopy.Label)
			}
		}
		if len(audioLabels) > 1 {
			filterComplex += fmt.Sprintf("[%d:a]asplit=%d[%s];", i, len(audioLabels), strings.Join(audioLabels, "]["))
		} else if len(audioLabels) == 1 {
			filterComplex += fmt.Sprintf("[%d:a]anull[%s];", i, audioLabels[0])
		}
	}

	for _, filter := range a.filterComplex {
		if filter.FilterElement != "" {
			filterComplex += filter.FilterElement
			if !strings.HasSuffix(filter.FilterElement, "]") {
				filterComplex += fmt.Sprintf("[%s]", filter.Label)
			}
			filterComplex += ";"
		}
	}

	audioLabel := a.lastAudioLabel()
	if audioLabel == "" && len(a.filenames) > 0 {
		// No filters applied, map first input stream directly
		ffmpegArgs = append(ffmpegArgs, "-map", "0:a")
	} else if audioLabel != "" {
		ffmpegArgs = append(ffmpegArgs, "-filter_complex", filterComplex, "-map", fmt.Sprintf("[%s]", audioLabel))
	} else {
		return fmt.Errorf("no audio stream to map")
	}

	// Threads
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

	// Audio parameters
	if parms.Codec != "" {
		ffmpegArgs = append(ffmpegArgs, "-c:a", string(parms.Codec))
	} else {
		// Default to aac for common extensions
		ext := strings.ToLower(filepath.Ext(parms.OutputPath))
		if ext == ".mp4" || ext == ".m4a" || ext == ".mov" {
			ffmpegArgs = append(ffmpegArgs, "-c:a", "aac")
		}
	}

	if parms.SampleRate > 0 {
		ffmpegArgs = append(ffmpegArgs, "-ar", fmt.Sprintf("%d", parms.SampleRate))
	}
	if parms.Channels > 0 {
		ffmpegArgs = append(ffmpegArgs, "-ac", fmt.Sprintf("%d", parms.Channels))
	}
	if parms.Bitrate > 0 {
		ffmpegArgs = append(ffmpegArgs, "-b:a", fmt.Sprintf("%dk", parms.Bitrate))
	}

	progressEnabled := parms.OnProgress != nil || !parms.SilentProgress
	if progressEnabled {
		ffmpegArgs = append(ffmpegArgs, "-progress", "pipe:1", "-nostats")
	}

	ffmpegArgs = append(ffmpegArgs, "-vn", "-y", parms.OutputPath)

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
		return a.runAudioWithProgress(cmd, &stderrBuf, handler)
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

func (a *Audio) runAudioWithProgress(cmd *exec.Cmd, stderrBuf *bytes.Buffer, onProgress func(Progress)) error {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	totalDuration := a.duration
	scanner := bufio.NewScanner(stdoutPipe)
	cur := Progress{TotalDuration: totalDuration}

	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		switch key {
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
