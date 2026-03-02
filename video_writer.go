package moviego

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

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
	if parms.Codec != "" {
		v.Codec(parms.Codec)
	}
	if parms.Fps != 0 {
		v.SetFps(parms.Fps)
	}
	if parms.Preset != "" {
		v.Preset(parms.Preset)
	}
	if parms.WithMask {
		v.WithMask(parms.WithMask)
	}
	if parms.PixelFormat != "" {
		v.PixelFormat(parms.PixelFormat)
	}
	if parms.Bitrate != "" {
		v.BitRate(parms.Bitrate)
	}
	if parms.Threads == 0 {
		// Optimize thread allocation: reserve CPUs for Go workers
		// FFmpeg gets 60% of CPUs, Go workers use 40%
		totalCPUs := runtime.GOMAXPROCS(0)
		if totalCPUs <= 2 {
			parms.Threads = uint16(totalCPUs)
		} else {
			parms.Threads = uint16((totalCPUs * 6) / 10)
			if parms.Threads < 2 {
				parms.Threads = 2
			}
		}
	}

	// Validate FPS after applying parameters
	if v.GetFps() <= 0 {
		return fmt.Errorf("video FPS is invalid (%d), cannot process video", v.GetFps())
	}

	ffmpegPath, err := getFFmpegPath()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg path: %w", err)
	}
	args := []string{ffmpegPath}
	for _, filename := range v.GetFilenames() {
		args = append(args, "-i", filename)
	}

	filterComplex := ""

	//split part
	for i, filename := range v.GetFilenames() {
		videoLabels := []string{}
		for _, filter := range v.videoFilterComplex {
			currentFilename := filter.fileCopy.filename
			if currentFilename == filename {
				videoLabels = append(videoLabels, filter.fileCopy.label)
			}
		}
		filterComplex += fmt.Sprintf("[%d:v]split=%d[%s];", i, len(videoLabels), strings.Join(videoLabels, "]["))

		audioLabels := []string{}
		for _, filter := range v.audioFilterComplex {
			currentFilename := filter.fileCopy.filename
			if currentFilename == filename {
				audioLabels = append(audioLabels, filter.fileCopy.label)
			}
		}
		filterComplex += fmt.Sprintf("[%d:a]asplit=%d[%s];", i, len(audioLabels), strings.Join(audioLabels, "]["))

	}

	audioIndex := 0
	videoIndex := 0

	videoLen := len(v.videoFilterComplex)
	audioLen := len(v.audioFilterComplex)

	for i := uint64(0); i <= globalOrderCounter; i++ {
		if videoIndex < (videoLen) && i == v.videoFilterComplex[videoIndex].order {
			filter := v.videoFilterComplex[videoIndex]
			if len(filter.filterElements) > 0 {
				filterComplex += strings.Join(filter.filterElements, ",")
				if !strings.HasSuffix(filter.filterElements[len(filter.filterElements)-1], "]") {
					filterComplex += fmt.Sprintf("[%s];", filter.label)
				}
			}
			videoIndex++
		}
		if audioIndex < (audioLen) && i == v.audioFilterComplex[audioIndex].order {
			filter := v.audioFilterComplex[audioIndex]
			if len(filter.filterElements) > 0 {
				filterComplex += strings.Join(filter.filterElements, ",")
				if !strings.HasSuffix(filter.filterElements[len(filter.filterElements)-1], "]") {
					filterComplex += fmt.Sprintf("[%s];", filter.label)
				}
			}
			audioIndex++
		}
	}

	lastAudioLabel := v.lastAudioLabel()
	lastVideoLabel := v.lastVideoLabel()

	mapVideo := fmt.Sprintf("[%s]", lastVideoLabel)
	mapAudio := fmt.Sprintf("[%s]", lastAudioLabel)

	args = append(args, "-filter_complex", filterComplex, "-map", mapVideo, "-map", mapAudio, "-y", parms.OutputPath)

	cmd := exec.Command(ffmpegPath, args[1:]...)

	logger.Info("ffmpeg command", "cmd", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute ffmpeg: %w", err)
	}

	return nil
}

// ProcessVideoFramesParams contains parameters for processing video frames
type ProcessVideoFramesParams struct {
	OutputPath string
	Threads    uint16
}

// validateFilterPart checks if a filter part contains valid FFmpeg filter syntax
// It ensures the filter contains actual filter operations, not just labels
func validateFilterPart(filterPart string) bool {
	if strings.TrimSpace(filterPart) == "" {
		return false
	}

	// Remove all label references (e.g., [label], [0:v], [outv]) to check for actual filter operations
	// Labels are in the format [label_name] or [number:stream_type]
	labelPattern := regexp.MustCompile(`\[[^\]]+\]`)
	filterWithoutLabels := labelPattern.ReplaceAllString(filterPart, "")

	// After removing labels, there should be filter operations remaining
	// Filter operations typically contain '=' (e.g., scale=100:100, overlay=x=10:y=10)
	// or filter names (e.g., trim, setpts, drawtext, overlay)
	filterWithoutLabels = strings.TrimSpace(filterWithoutLabels)

	// Check if there's actual filter content (not just separators)
	if filterWithoutLabels == "" || filterWithoutLabels == ";" {
		return false
	}

	// Check for common filter operations (contains '=' or known filter names)
	hasFilterOps := strings.Contains(filterWithoutLabels, "=") ||
		strings.Contains(filterWithoutLabels, "trim") ||
		strings.Contains(filterWithoutLabels, "setpts") ||
		strings.Contains(filterWithoutLabels, "overlay") ||
		strings.Contains(filterWithoutLabels, "drawtext") ||
		strings.Contains(filterWithoutLabels, "scale") ||
		strings.Contains(filterWithoutLabels, "subtitles") ||
		strings.Contains(filterWithoutLabels, "ass") ||
		strings.Contains(filterWithoutLabels, "colorchannelmixer") ||
		strings.Contains(filterWithoutLabels, "format") ||
		strings.Contains(filterWithoutLabels, "color")

	return hasFilterOps
}
