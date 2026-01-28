package moviego

import (
	"fmt"
	"os/exec"
	"strings"
)

// ConcatenateVideos creates a lazy Video object representing the concatenation of multiple videos
// No file is written to disk - the concatenation happens later when WriteVideo is called
// Parameters:
//   - videos: Slice of Video objects to concatenate (must have at least 1 video)
//
// Returns a new Video object representing the concatenated video (lazy) and an error if any
func ConcatenateVideos(videos []*Video) (*Video, error) {
	// Validate input
	if len(videos) == 0 {
		return nil, fmt.Errorf("no videos provided for concatenation")
	}

	// Flatten nested concatenations and filter out invalid videos
	// If a video is itself a concatenation (has sourceVideos), expand it
	validVideos := make([]*Video, 0, len(videos))
	for _, video := range videos {
		if video == nil {
			continue
		}

		// If this video is a concatenation (lazy), expand its source videos
		if video.sourceVideos != nil {
			validVideos = append(validVideos, video.sourceVideos...)
		} else if video.GetFilename() != "" && video.GetDuration() > 0 {
			// Regular video with a filename
			validVideos = append(validVideos, video)
		}
	}

	if len(validVideos) == 0 {
		return nil, fmt.Errorf("no valid videos to concatenate")
	}

	// Calculate combined properties from source videos
	var totalDuration float64
	firstVideo := validVideos[0]
	hasAudio := false

	for _, video := range validVideos {
		totalDuration += video.GetDuration()
		if video.HasAudio() {
			hasAudio = true
		}
	}

	// Create a lazy Video object representing the concatenation
	concatenated := &Video{
		filename:      "", // No filename yet (lazy)
		codec:         Codec(firstVideo.GetCodec()),
		width:         firstVideo.GetWidth(),
		height:        firstVideo.GetHeight(),
		fps:           firstVideo.GetFps(),
		duration:      totalDuration,
		frames:        uint64(float64(firstVideo.GetFps()) * totalDuration),
		ffmpegArgs:    make(map[string][]string),
		filters:       []Filter{},
		customFilters: []func([]byte, int){},
		isTemp:        false,
		bitRate:       firstVideo.GetBitRate(),
		preset:        firstVideo.GetPreset(),
		withMask:      firstVideo.GetWithMask(),
		pixelFormat:   firstVideo.GetPixelFormat(),
		startTime:     0,
		endTime:       0,
		sourceVideos:  validVideos, // Store source videos for lazy concatenation
	}

	// Copy audio info if any video has audio
	if hasAudio {
		concatenated.audio = *firstVideo.GetAudio()
	}

	return concatenated, nil
}

// concatenateSingleVideo handles the case of a single video
func concatenateSingleVideo(video *Video, outputPath string, params VideoParameters) (*Video, error) {
	// If it's a subclip or has filters, process it using filter_complex
	if video.GetStartTime() > 0 || video.GetEndTime() > 0 || len(video.filters) > 0 || len(video.customFilters) > 0 {
		// Use filter_complex even for single video to handle subclips and filters natively
		if err := concatenateWithFilterComplex([]*Video{video}, outputPath, params); err != nil {
			return nil, fmt.Errorf("failed to process single video: %w", err)
		}
	} else {
		// Just copy the file
		cmd := exec.Command("ffmpeg",
			"-loglevel", "error",
			"-i", video.GetFilename(),
			"-c", "copy",
			"-y",
			outputPath,
		)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to copy single video: %w", err)
		}
	}

	return NewVideoFile(outputPath)
}

// concatenateWithFilterComplex uses FFmpeg's filter_complex to concatenate videos natively
// This handles subclips, filters, and concatenation in a single FFmpeg command
func concatenateWithFilterComplex(videos []*Video, outputPath string, params VideoParameters) error {
	// Filter out invalid/empty videos
	validVideos := make([]*Video, 0, len(videos))
	for _, video := range videos {
		if video.GetFilename() != "" && video.GetDuration() > 0 {
			validVideos = append(validVideos, video)
		}
	}

	if len(validVideos) == 0 {
		return fmt.Errorf("no valid videos to concatenate")
	}

	// Group videos by source file and track their input indices
	sourceMap := make(map[string]int)
	sources := []string{}

	for _, video := range validVideos {
		filename := video.GetFilename()
		if _, exists := sourceMap[filename]; !exists {
			sourceMap[filename] = len(sources)
			sources = append(sources, filename)
		}
	}

	// Update videos to use only valid ones
	videos = validVideos

	// Calculate total duration for progress tracking
	var totalDuration float64
	for _, video := range videos {
		// For subclips, use the actual duration
		if video.GetEndTime() > 0 && video.GetStartTime() > 0 {
			totalDuration += video.GetEndTime() - video.GetStartTime()
		} else if video.GetEndTime() > 0 {
			totalDuration += video.GetEndTime()
		} else {
			totalDuration += video.GetDuration()
		}
	}

	// Build FFmpeg arguments
	args := []string{"-loglevel", "error"}

	// Add all source files as inputs (keep original paths for sourceMap consistency)
	for _, source := range sources {
		args = append(args, "-i", source)
	}

	// Build filter_complex string
	filterComplex, hasAudio := buildFilterString(videos, sourceMap)

	if filterComplex != "" {
		args = append(args, "-filter_complex", filterComplex)
		args = append(args, "-map", "[outv]")
		if hasAudio {
			args = append(args, "-map", "[outa]")
		}
	} else {
		// No filtering needed, just concat
		args = append(args, "-map", "0")
	}

	// Apply encoding parameters
	selectedCodec := resolveCodec(string(params.Codec), "")
	args = append(args, "-c:v", selectedCodec)

	selectedPreset := mapPresetForCodec(selectedCodec, resolvePreset(params.Preset, ""))
	if selectedPreset != "" {
		args = append(args, "-preset", selectedPreset)
	}

	if bitrate := resolveBitrate(params.Bitrate, ""); bitrate != "" {
		args = append(args, "-b:v", bitrate)
	}

	if fps := resolveFps(params.Fps, 0); fps > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", fps))
	}

	if params.PixelFormat != "" {
		args = append(args, "-pix_fmt", params.PixelFormat)
	}

	if params.Threads != 0 {
		args = append(args, "-threads", fmt.Sprintf("%d", params.Threads))
	}

	// Audio encoding
	if hasAudio {
		args = append(args, "-c:a", "aac")
	}

	// Output
	args = append(args, "-y", outputPath)

	// Use progress parser for concatenation
	config := FFmpegProgressConfig{
		Args:          args,
		TotalDuration: totalDuration,
		OperationName: "Concatenating videos",
		OutputPath:    outputPath,
		Bitrate:       resolveBitrate(params.Bitrate, ""),
	}

	if err := runFFmpegWithProgress(config); err != nil {
		return fmt.Errorf("ffmpeg filter_complex concatenation failed: %w", err)
	}

	return nil
}

// buildFilterString constructs the FFmpeg filter_complex string for concatenating videos
// Returns the filter string and a boolean indicating if audio streams exist
func buildFilterString(videos []*Video, sourceMap map[string]int) (string, bool) {
	var videoFilters []string
	var audioFilters []string
	hasAudio := false

	// Check if any video has audio
	for _, video := range videos {
		if video.HasAudio() {
			hasAudio = true
			break
		}
	}

	// Build filter for each video
	for i, video := range videos {
		sourceIndex := sourceMap[video.GetFilename()]

		// Build video filter chain
		videoFilterChain := fmt.Sprintf("[%d:v]", sourceIndex)

		// Add trim if this is a subclip
		startTime := video.GetStartTime()
		endTime := video.GetEndTime()

		if startTime > 0 || endTime > 0 {
			if endTime > 0 {
				videoFilterChain += fmt.Sprintf("trim=start=%.3f:end=%.3f,setpts=PTS-STARTPTS", startTime, endTime)
			} else {
				videoFilterChain += fmt.Sprintf("trim=start=%.3f,setpts=PTS-STARTPTS", startTime)
			}
		} else {
			// For full video, just copy and reset PTS
			videoFilterChain += "copy,setpts=PTS-STARTPTS"
		}

		// Add filters (Inverse, BlackWhite, etc.)
		filterString := translateFiltersToFFmpeg(video.filters)
		if filterString != "" {
			videoFilterChain += "," + filterString
		}

		videoFilterChain += fmt.Sprintf("[v%d]", i)
		videoFilters = append(videoFilters, videoFilterChain)

		// Build audio filter chain if audio exists
		if hasAudio && video.HasAudio() {
			audioFilterChain := fmt.Sprintf("[%d:a]", sourceIndex)

			if startTime > 0 || endTime > 0 {
				if endTime > 0 {
					audioFilterChain += fmt.Sprintf("atrim=start=%.3f:end=%.3f,asetpts=PTS-STARTPTS", startTime, endTime)
				} else {
					audioFilterChain += fmt.Sprintf("atrim=start=%.3f,asetpts=PTS-STARTPTS", startTime)
				}
			} else {
				audioFilterChain += "acopy,asetpts=PTS-STARTPTS"
			}

			audioFilterChain += fmt.Sprintf("[a%d]", i)
			audioFilters = append(audioFilters, audioFilterChain)
		} else if hasAudio {
			// Create silent audio for videos without audio
			audioFilterChain := "anullsrc=channel_layout=stereo:sample_rate=44100"
			duration := endTime - startTime
			if duration <= 0 {
				duration = video.GetDuration()
			}
			audioFilterChain = fmt.Sprintf("aevalsrc=0:d=%.3f", duration) + "[a" + fmt.Sprintf("%d]", i)
			audioFilters = append(audioFilters, audioFilterChain)
		}
	}

	// Build concat filter
	var concatInputs []string
	for i := 0; i < len(videos); i++ {
		concatInputs = append(concatInputs, fmt.Sprintf("[v%d]", i))
		if hasAudio {
			concatInputs = append(concatInputs, fmt.Sprintf("[a%d]", i))
		}
	}

	var filterComplex string

	// Combine all filters
	allFilters := append(videoFilters, audioFilters...)
	filterComplex = strings.Join(allFilters, ";")

	// Add concat filter
	if hasAudio {
		filterComplex += ";" + strings.Join(concatInputs, "") +
			fmt.Sprintf("concat=n=%d:v=1:a=1[outv][outa]", len(videos))
	} else {
		filterComplex += ";" + strings.Join(concatInputs, "") +
			fmt.Sprintf("concat=n=%d:v=1:a=0[outv]", len(videos))
	}

	return filterComplex, hasAudio
}

// translateFiltersToFFmpeg converts MovieGo Filter enums to FFmpeg filter names
func translateFiltersToFFmpeg(filters []Filter) string {
	if len(filters) == 0 {
		return ""
	}

	var ffmpegFilters []string

	for _, filter := range filters {
		switch filter {
		case Inverse:
			ffmpegFilters = append(ffmpegFilters, "negate")
		case BlackWhite:
			// Using colorchannelmixer for grayscale (standard ITU-R BT.601 coefficients)
			ffmpegFilters = append(ffmpegFilters, "colorchannelmixer=.3:.4:.3:0:.3:.4:.3:0:.3:.4:.3")
		case Sepia, SepiaTone:
			// Sepia tone using colorchannelmixer
			ffmpegFilters = append(ffmpegFilters, "colorchannelmixer=.393:.769:.189:0:.349:.686:.168:0:.272:.534:.131")
		case Edge:
			// Edge detection
			ffmpegFilters = append(ffmpegFilters, "edgedetect")
		}
	}

	return strings.Join(ffmpegFilters, ",")
}
