package moviego

import (
	"fmt"
	"math"
	"os/exec"
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
	v.applyParameters(parms)


	ffmpegPath, err := getFFmpegPath()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg path: %w", err)
	}

	args := []string{ffmpegPath}
	for _, filename := range v.GetFilenames() {
		args = append(args, "-i", filename)
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
			if len(filter.FilterElements) > 0 {
				filterComplex += strings.Join(filter.FilterElements, ",")
				if !strings.HasSuffix(filter.FilterElements[len(filter.FilterElements)-1], "]") {
					filterComplex += fmt.Sprintf("[%s]", filter.Label)
				}
				filterComplex += ";"
			}
			videoIndex++
		}
		if audioIndex < audioLen && v.audioFilterComplex[audioIndex].Order == nextOrder {
			filter := v.audioFilterComplex[audioIndex]
			if len(filter.FilterElements) > 0 {
				filterComplex += strings.Join(filter.FilterElements, ",")
				if !strings.HasSuffix(filter.FilterElements[len(filter.FilterElements)-1], "]") {
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

	args = append(args, "-filter_complex", filterComplex, "-map", mapVideo, "-map", mapAudio, "-y", parms.OutputPath)

	cmd := exec.Command(ffmpegPath, args[1:]...)

	logger.Info("ffmpeg command", "cmd", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute ffmpeg: %w", err)
	}

	return nil
}

