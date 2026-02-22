package moviego

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Concatenate(videos []Video, outputPath string) (*Video, error) {
	path, err := getFFmpegPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get ffmpeg path: %w", err)
	}

	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", dir, err)
	}

	args := []string{path}
	filenamesCount := map[string]int{}
	filenamesOrder := map[string]int{}
	filenamesHashs := map[string]string{}

	counter := 0

	for _, video := range videos {
		filename := video.filenames[0]
		_, ok := filenamesCount[filename]
		if ok {
			filenamesCount[filename]++
			continue
		}
		filenamesOrder[filename] = counter
		filenamesHashs[filename] = video.fileHashCode(filename)
		counter++
		args = append(args, "-i", filename)
		filenamesCount[filename] = 0
	}

	var parts []string

	for filename, index := range filenamesOrder {
		filenameCount := filenamesCount[filename] + 1
		seg := fmt.Sprintf("[%d:v]split=%d", index, filenameCount)
		for i := range filenameCount {
			seg += fmt.Sprintf(" [v_%s_%d]", filenamesHashs[filename], i)
		}
		parts = append(parts, seg)
	}
	for filename, index := range filenamesOrder {
		filenameCount := filenamesCount[filename] + 1
		seg := fmt.Sprintf("[%d:a]asplit=%d", index, filenameCount)
		for i := range filenameCount {
			seg += fmt.Sprintf(" [a_%s_%d]", filenamesHashs[filename], i)
		}
		parts = append(parts, seg)
	}

	for _, video := range videos {
		filename := video.filenames[0]
		index, _ := filenamesCount[filename]
		filenamesCount[filename]--
		parts = append(parts, fmt.Sprintf(
			"[v_%s_%d]trim=start=%f:end=%f,setpts=PTS-STARTPTS[%s]",
			filenamesHashs[filename], index, video.startTime, video.endTime, video.HashCode(filename)+"_v"))
		parts = append(parts, fmt.Sprintf(
			"[a_%s_%d]atrim=start=%f:end=%f,asetpts=PTS-STARTPTS[%s]",
			filenamesHashs[filename], index, video.startTime, video.endTime, video.HashCode(filename)+"_a"))
	}

	concatInputs := ""
	for _, video := range videos {
		filename := video.filenames[0]
		concatInputs += fmt.Sprintf("[%s] [%s] ", video.HashCode(filename)+"_v", video.HashCode(filename)+"_a")
	}
	parts = append(parts, strings.TrimSpace(concatInputs)+fmt.Sprintf(" concat=n=%d:v=1:a=1[outv][outa]", len(videos)))

	filterComplex := strings.Join(parts, "; ")

	args = append(args, "-filter_complex", filterComplex, "-map", "[outv]", "-map", "[outa]", "-y", outputPath)

	cmd := exec.Command(path, args[1:]...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	logger.Info("executing ffmpeg concatenate", "command", cmd.String())

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to concatenate videos: %w (ffmpeg: %s)", err, stderr.String())
	}

	filenames := make([]string, len(filenamesOrder))

	for filename, index := range filenamesOrder {
		filenames[index] = filename
	}

	duration := 0.0
	for _, video := range videos {
		duration += video.endTime - video.startTime
	}

	return &Video{
		filenames:     filenames,
		// filterComplex: filterComplex,
		startTime:     0,
		endTime:       duration,
		duration:      duration,
		codec:         videos[0].codec,
		width:         videos[0].width,
		height:        videos[0].height,
		fps:           videos[0].fps,
		frames:        uint64(float64(videos[0].fps) * duration),
		ffmpegArgs:    videos[0].ffmpegArgs,
		filters:       videos[0].filters,
		isTemp:        false,
		audio:         videos[0].audio,
		bitRate:       videos[0].bitRate,
		preset:        videos[0].preset,
		withMask:      videos[0].withMask,
		pixelFormat:   videos[0].pixelFormat,
		textClips:     videos[0].textClips,
		subtitleClips: videos[0].subtitleClips,
	}, nil
}
