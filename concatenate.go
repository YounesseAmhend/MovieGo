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
		_, ok := filenamesCount[video.filename]
		if ok {
			filenamesCount[video.filename]++
			continue
		}
		filenamesOrder[video.filename] = counter
		filenamesHashs[video.filename] = video.fileHashCode()
		counter++
		args = append(args, "-i", video.filename)
		filenamesCount[video.filename] = 0
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
		index, _ := filenamesCount[video.filename]
		filenamesCount[video.filename]--
		parts = append(parts, fmt.Sprintf(
			"[v_%s_%d]trim=start=%f:end=%f,setpts=PTS-STARTPTS[%s]",
			filenamesHashs[video.filename], index, video.startTime, video.endTime, video.HashCode()+"_v"))
		parts = append(parts, fmt.Sprintf(
			"[a_%s_%d]atrim=start=%f:end=%f,asetpts=PTS-STARTPTS[%s]",
			filenamesHashs[video.filename], index, video.startTime, video.endTime, video.HashCode()+"_a"))
	}

	concatInputs := ""
	for _, video := range videos {
		concatInputs += fmt.Sprintf("[%s] [%s] ", video.HashCode()+"_v", video.HashCode()+"_a")
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

	return NewVideoFile(outputPath)
}
