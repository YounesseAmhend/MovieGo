package moviego

import (
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// WriteVideo processes the video with applied filters and writes to output file
func (video *Video) WriteVideo(parms VideoParameters) error {
	if parms.OutputPath == "" {
		return fmt.Errorf("output path is empty, cannot write video")
	}

	// Check if video has text overlays or subtitles - use FFmpeg filter approach
	if video.HasText() || video.HasSubtitles() {
		return video.writeVideoWithTextFilters(parms)
	}

	// Validate essential video properties before processing
	if video.GetFilename() == "" {
		return fmt.Errorf("video filename is empty, cannot process video")
	}
	if video.GetWidth() <= 0 || video.GetHeight() <= 0 {
		return fmt.Errorf("video dimensions are invalid (%dx%d), cannot process video", video.GetWidth(), video.GetHeight())
	}
	if video.GetDuration() <= 0 {
		return fmt.Errorf("video duration is invalid (%.2f), cannot process video", video.GetDuration())
	}

	// Apply parameters to video
	if parms.Codec != "" {
		video.Codec(parms.Codec)
	}
	if parms.Fps != 0 {
		video.SetFps(parms.Fps)
	}
	if parms.Preset != "" {
		video.Preset(parms.Preset)
	}
	if parms.WithMask {
		video.WithMask(parms.WithMask)
	}
	if parms.PixelFormat != "" {
		video.PixelFormat(parms.PixelFormat)
	}
	if parms.Bitrate != "" {
		video.BitRate(parms.Bitrate)
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
	if video.GetFps() <= 0 {
		return fmt.Errorf("video FPS is invalid (%d), cannot process video", video.GetFps())
	}

	// Create temporary files for processing
	tempFiles, err := createTempFiles()
	if err != nil {
		return err
	}
	defer tempFiles.Cleanup()

	// Process video frames
	if err := video.processVideoFrames(ProcessVideoFramesParams{
		OutputPath: tempFiles.VideoPath,
		Threads:    parms.Threads,
	}); err != nil {
		return err
	}

	fmt.Printf("\n%s Now combining with audio...\n", color.CyanString("Video processing complete."))

	// Combine processed video with original audio
	if err := combineAudioVideo(video, tempFiles.VideoPath, parms.OutputPath); err != nil {
		return err
	}

	fmt.Printf("%s %s %s\n",
		color.GreenString("Video processing complete:"),
		color.MagentaString(parms.OutputPath),
		color.GreenString("(with audio)"))
	return nil
}

// ProcessVideoFramesParams contains parameters for processing video frames
type ProcessVideoFramesParams struct {
	OutputPath string
	Threads    uint16
}

// processVideoFrames handles the frame reading, filtering, and encoding
func (video *Video) processVideoFrames(params ProcessVideoFramesParams) error {
	// Build FFmpeg commands
	inputCmd, err := buildInputCommand(video, params.Threads)
	if err != nil {
		return fmt.Errorf("failed to build input command: %w", err)
	}

	outputCmd, err := buildOutputCommand(video, params.OutputPath, params.Threads)
	if err != nil {
		return fmt.Errorf("failed to build output command: %w", err)
	}

	// Set up pipes
	pipes, err := setupPipes(inputCmd, outputCmd)
	if err != nil {
		return err
	}
	defer pipes.OutputStdin.Close()

	// Start both FFmpeg processes
	if err := inputCmd.Start(); err != nil {
		return fmt.Errorf("failed to start input FFmpeg process: %w", err)
	}
	if err := outputCmd.Start(); err != nil {
		return fmt.Errorf("failed to start output FFmpeg process: %w", err)
	}

	// Compose filters once before processing
	composedFilter := composeFilters(video.customFilters, video.filters)

	// Process all frames
	config := FrameProcessorConfig{
		Video:          video,
		InputReader:    pipes.InputStdout,
		OutputWriter:   pipes.OutputStdin,
		TotalFrames:    video.GetFrames(),
		ComposedFilter: composedFilter,
	}

	if err := processFrameLoop(config); err != nil {
		return err
	}

	// Close stdin to signal EOF to FFmpeg encoder
	pipes.OutputStdin.Close()

	// Wait for processes to finish
	if err := inputCmd.Wait(); err != nil {
		return fmt.Errorf("input FFmpeg process failed: %w", err)
	}
	if err := outputCmd.Wait(); err != nil {
		return fmt.Errorf("output FFmpeg process failed: %w", err)
	}

	return nil
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

// writeVideoWithTextFilters writes a video with text overlays and/or subtitles using FFmpeg filters
func (video *Video) writeVideoWithTextFilters(parms VideoParameters) error {
	// Build FFmpeg command with filter_complex for overlays
	args := []string{"-loglevel", "error"}

	// Input file
	args = append(args, "-i", video.GetFilename())

	// Calculate video duration
	videoDuration := video.GetDuration()
	if videoDuration <= 0 {
		videoDuration = 10.0
	}

	// Build filter chain
	var filterParts []string
	currentLabel := "0:v"

	// Apply video trim/subclip if needed
	if video.GetStartTime() > 0 || video.GetEndTime() > 0 {
		trimFilter := fmt.Sprintf("[%s]", currentLabel)
		if video.GetEndTime() > 0 {
			trimFilter += fmt.Sprintf("trim=start=%.3f:end=%.3f,setpts=PTS-STARTPTS", video.GetStartTime(), video.GetEndTime())
		} else {
			trimFilter += fmt.Sprintf("trim=start=%.3f,setpts=PTS-STARTPTS", video.GetStartTime())
		}
		trimFilter += "[trimmed]"
		if validateFilterPart(trimFilter) {
			filterParts = append(filterParts, trimFilter)
			currentLabel = "trimmed"
		}
	}

	// Collect all overlay clips and sort by layer
	type overlayItem struct {
		layer    int
		clipType string // "text", "subtitle"
		index    int
		clip     interface{}
	}
	var overlays []overlayItem

	// Add text clips
	if video.HasText() {
		for i, textClip := range video.GetTextClips() {
			overlays = append(overlays, overlayItem{
				layer:    textClip.GetLayer(),
				clipType: "text",
				index:    i,
				clip:     textClip,
			})
		}
	}

	// Add subtitle clips
	if video.HasSubtitles() {
		for i, subtitleClip := range video.GetSubtitleClips() {
			overlays = append(overlays, overlayItem{
				layer:    0, // Subtitles typically go on top
				clipType: "subtitle",
				index:    i,
				clip:     subtitleClip,
			})
		}
	}

	// Sort overlays by layer
	sort.Slice(overlays, func(i, j int) bool {
		return overlays[i].layer < overlays[j].layer
	})

	// Process overlays in layer order
	hasOverlayFilters := false
	for i, overlay := range overlays {
		isLast := i == len(overlays)-1
		var nextLabel string
		if isLast {
			nextLabel = "outv"
		} else {
			nextLabel = fmt.Sprintf("overlay%d", i)
		}

		switch overlay.clipType {
		case "text":
			textClip := overlay.clip.(*TextClip)
			// Check if text needs rotation/scale (requires special transparent layer approach)
			if textNeedsRotationOrScale(textClip) {
				textFilter := buildRotatedScaledTextFilterString(textClip, video.GetWidth(), video.GetHeight(), video.GetDuration(), currentLabel, nextLabel)
				if textFilter != "" {
					parts := strings.Split(textFilter, ";")
					allPartsValid := true
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if part != "" && !validateFilterPart(part) {
							allPartsValid = false
							break
						}
					}
					if allPartsValid {
						filterParts = append(filterParts, textFilter)
						currentLabel = nextLabel
						hasOverlayFilters = true
					}
				}
			} else {
				// Normal text overlay (position animation only, no rotation/scale)
				textFilter := buildTextFilterString(textClip, video.GetWidth(), video.GetHeight(), video.GetDuration())
				if textFilter != "" {
					filterPart := fmt.Sprintf("[%s]%s[%s]", currentLabel, textFilter, nextLabel)
					if validateFilterPart(filterPart) {
						filterParts = append(filterParts, filterPart)
						currentLabel = nextLabel
						hasOverlayFilters = true
					}
				}
			}

		case "subtitle":
			subtitleClip := overlay.clip.(*SubtitleClip)
			subtitleFilter, err := buildSubtitleFilterString(subtitleClip)
			if err != nil {
				// Skip subtitle on error
				continue
			}
			if subtitleFilter != "" {
				filterPart := fmt.Sprintf("[%s]%s[%s]", currentLabel, subtitleFilter, nextLabel)
				if validateFilterPart(filterPart) {
					filterParts = append(filterParts, filterPart)
					currentLabel = nextLabel
					hasOverlayFilters = true
				}
			}
		}
	}

	// Add filter_complex to args
	// Filter out any empty or invalid filter parts
	validFilterParts := make([]string, 0, len(filterParts))
	for _, part := range filterParts {
		if validateFilterPart(part) {
			validFilterParts = append(validFilterParts, part)
		}
	}

	if len(validFilterParts) > 0 {
		filterComplex := strings.Join(validFilterParts, ";")
		// Remove any empty segments that might have been created by joining
		filterComplex = strings.ReplaceAll(filterComplex, ";;", ";")
		filterComplex = strings.Trim(filterComplex, ";")
		filterComplex = strings.TrimSpace(filterComplex)

		// Final validation: ensure filter_complex contains actual filter operations
		if filterComplex != "" && filterComplex != ";" {
			// Split by semicolon and validate each part
			parts := strings.Split(filterComplex, ";")
			validParts := make([]string, 0, len(parts))
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part != "" && validateFilterPart(part) {
					validParts = append(validParts, part)
				}
			}

			// Rejoin only valid parts
			if len(validParts) > 0 {
				filterComplex = strings.Join(validParts, ";")
				if strings.TrimSpace(filterComplex) != "" {
					args = append(args, "-filter_complex", filterComplex)
					if hasOverlayFilters && currentLabel == "outv" {
						args = append(args, "-map", "[outv]")
					} else if currentLabel != "0:v" {
						args = append(args, "-map", "["+currentLabel+"]")
					} else {
						args = append(args, "-map", "0:v")
					}
				} else {
					args = append(args, "-map", "0:v")
				}
			} else {
				args = append(args, "-map", "0:v")
			}
		} else {
			args = append(args, "-map", "0:v")
		}
	} else {
		args = append(args, "-map", "0:v")
	}

	// Map audio stream if present
	if video.HasAudio() {
		args = append(args, "-map", "0:a")
	}

	// Apply encoding parameters
	selectedCodec := resolveCodec(string(parms.Codec), video.GetCodec())
	args = append(args, "-c:v", selectedCodec)

	// Map preset for the selected codec
	selectedPreset := mapPresetForCodec(selectedCodec, resolvePreset(parms.Preset, video.GetPreset()))
	if selectedPreset != "" {
		args = append(args, "-preset", selectedPreset)
	}

	if bitrate := resolveBitrate(parms.Bitrate, video.GetBitRate()); bitrate != "" {
		args = append(args, "-b:v", bitrate)
	}

	if fps := resolveFps(parms.Fps, video.GetFps()); fps > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", fps))
	}

	if parms.PixelFormat != "" {
		args = append(args, "-pix_fmt", string(parms.PixelFormat))
	} else if video.GetPixelFormat() != "" {
		args = append(args, "-pix_fmt", string(video.GetPixelFormat()))
	}

	if parms.Threads != 0 {
		args = append(args, "-threads", fmt.Sprintf("%d", parms.Threads))
	}

	// Audio encoding
	if video.HasAudio() {
		args = append(args, "-c:a", "aac", "-b:a", "192k")
	}

	// Output
	args = append(args, "-y", parms.OutputPath)

	// Final safety check: ensure we never pass an empty filter_complex
	for i, arg := range args {
		if arg == "-filter_complex" && i+1 < len(args) {
			filterValue := args[i+1]
			if strings.TrimSpace(filterValue) == "" || filterValue == ";" {
				args = append(args[:i], args[i+2:]...)
				hasMap := false
				for _, a := range args {
					if a == "-map" {
						hasMap = true
						break
					}
				}
				if !hasMap {
					args = append(args, "-map", "0:v")
				}
				break
			}
		}
	}

	// Use progress parser for video with text filters
	config := FFmpegProgressConfig{
		Args:          args,
		TotalDuration: videoDuration,
		OperationName: "Processing video with overlays",
		OutputPath:    parms.OutputPath,
		Bitrate:       resolveBitrate(parms.Bitrate, video.GetBitRate()),
	}

	if err := runFFmpegWithProgress(config); err != nil {
		return fmt.Errorf("ffmpeg text/subtitle processing failed: %w", err)
	}

	fmt.Printf("%s %s\n",
		color.GreenString("Video with overlays written successfully:"),
		color.MagentaString(parms.OutputPath))
	return nil
}
