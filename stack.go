package moviego

import (
	"fmt"
	"strconv"
	"strings"
)

// HStack stacks videos horizontally from left to right.
// All input videos must have the same height before calling HStack.
func HStack(videos []Video) (*Video, error) {
	if len(videos) == 0 {
		return nil, fmt.Errorf("HStack: no videos provided")
	}
	if len(videos) == 1 {
		v := videos[0]
		return &v, nil
	}

	prepared := append([]Video(nil), videos...)
	for i := range prepared {
		initRawVideo(&prepared[i])
	}

	height := prepared[0].height
	var width uint64
	for i, video := range prepared {
		if video.height != height {
			return nil, fmt.Errorf("HStack: input %d height %d does not match expected height %d", i, video.height, height)
		}
		width += video.width
	}

	return stackVideos(prepared, buildStackFilter(prepared, "hstack=inputs=%d"), width, height, "hstack")
}

// VStack stacks videos vertically from top to bottom.
// All input videos must have the same width before calling VStack.
func VStack(videos []Video) (*Video, error) {
	if len(videos) == 0 {
		return nil, fmt.Errorf("VStack: no videos provided")
	}
	if len(videos) == 1 {
		v := videos[0]
		return &v, nil
	}

	prepared := append([]Video(nil), videos...)
	for i := range prepared {
		initRawVideo(&prepared[i])
	}

	width := prepared[0].width
	var height uint64
	for i, video := range prepared {
		if video.width != width {
			return nil, fmt.Errorf("VStack: input %d width %d does not match expected width %d", i, video.width, width)
		}
		height += video.height
	}

	return stackVideos(prepared, buildStackFilter(prepared, "vstack=inputs=%d"), width, height, "vstack")
}

// XStack stacks videos according to the provided FFmpeg xstack layout.
// Example layout: "0_0|w0_0|0_h0|w0_h0"
func XStack(videos []Video, layout string) (*Video, error) {
	if len(videos) == 0 {
		return nil, fmt.Errorf("XStack: no videos provided")
	}
	if len(videos) == 1 {
		v := videos[0]
		return &v, nil
	}
	if strings.TrimSpace(layout) == "" {
		return nil, fmt.Errorf("XStack: layout is required")
	}

	prepared := append([]Video(nil), videos...)
	for i := range prepared {
		initRawVideo(&prepared[i])
	}

	width, height, err := inferXStackDimensions(layout, prepared)
	if err != nil {
		return nil, fmt.Errorf("XStack: %w", err)
	}

	filter := buildStackInputs(prepared) + fmt.Sprintf("xstack=inputs=%d:layout=%s", len(prepared), layout)
	return stackVideos(prepared, filter, width, height, "xstack")
}

func stackVideos(videos []Video, videoFilter string, width, height uint64, labelPrefix string) (*Video, error) {
	prepared := append([]Video(nil), videos...)
	for i := range prepared {
		initRawVideo(&prepared[i])
	}

	filenames := []string{}
	videoFilterComplex := []FilterComplex{}
	audioFilterComplex := []FilterComplex{}
	seen := make(map[string]struct{})

	var maxDuration float64
	for _, video := range prepared {
		for _, filename := range video.filenames {
			if _, exists := seen[filename]; exists {
				continue
			}
			seen[filename] = struct{}{}
			filenames = append(filenames, filename)
		}
		videoFilterComplex = append(videoFilterComplex, video.filterComplex...)
		audioFilterComplex = append(audioFilterComplex, video.audio.filterComplex...)
		if video.duration > maxDuration {
			maxDuration = video.duration
		}
	}

	order := incrementOrderCounter()
	label := fmt.Sprintf("%s_%d", labelPrefix, incrementGlobalCounter())

	videoFilterComplex = append(videoFilterComplex, FilterComplex{
		Order:         order,
		Label:         label + "_v",
		FilterElement: videoFilter,
	})
	audioFilterComplex = append(audioFilterComplex, FilterComplex{
		Order:         order,
		Label:         label + "_a",
		FilterElement: buildAudioMix(prepared),
	})

	base := prepared[0]
	newAudio := base.audio
	newAudio.filterComplex = audioFilterComplex
	newAudio.duration = maxDuration

	return &Video{
		filenames:          filenames,
		startTime:          0,
		endTime:            maxDuration,
		filterComplex: videoFilterComplex,
		duration:           maxDuration,
		codec:              base.codec,
		width:              width,
		height:             height,
		fps:                base.fps,
		frames:             uint64(float64(base.fps) * maxDuration),
		ffmpegArgs:         base.ffmpegArgs,
		isTemp:             false,
		audio:              newAudio,
		bitRate:            base.bitRate,
		preset:             base.preset,
		withMask:           base.withMask,
		pixelFormat:        base.pixelFormat,
	}, nil
}

func buildStackFilter(videos []Video, format string) string {
	return buildStackInputs(videos) + fmt.Sprintf(format, len(videos))
}

func buildStackInputs(videos []Video) string {
	var b strings.Builder
	for _, video := range videos {
		b.WriteString("[")
		b.WriteString(video.lastVideoLabel())
		b.WriteString("]")
	}
	return b.String()
}

func buildAudioMix(videos []Video) string {
	var b strings.Builder
	for _, video := range videos {
		b.WriteString("[")
		b.WriteString(video.audio.lastAudioLabel())
		b.WriteString("]")
	}
	b.WriteString(fmt.Sprintf("amix=inputs=%d:duration=longest", len(videos)))
	return b.String()
}

func inferXStackDimensions(layout string, videos []Video) (uint64, uint64, error) {
	entries := strings.Split(layout, "|")
	if len(entries) != len(videos) {
		return 0, 0, fmt.Errorf("XStack: layout entries %d do not match input videos %d", len(entries), len(videos))
	}

	var maxWidth uint64
	var maxHeight uint64
	for i, entry := range entries {
		parts := strings.SplitN(strings.TrimSpace(entry), "_", 2)
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("XStack: invalid layout entry %q", entry)
		}

		x, err := evalXStackExpr(parts[0], videos)
		if err != nil {
			return 0, 0, fmt.Errorf("XStack: invalid x expression %q: %w", parts[0], err)
		}
		y, err := evalXStackExpr(parts[1], videos)
		if err != nil {
			return 0, 0, fmt.Errorf("XStack: invalid y expression %q: %w", parts[1], err)
		}

		right := uint64(x) + videos[i].width
		bottom := uint64(y) + videos[i].height
		if right > maxWidth {
			maxWidth = right
		}
		if bottom > maxHeight {
			maxHeight = bottom
		}
	}

	return maxWidth, maxHeight, nil
}

func evalXStackExpr(expr string, videos []Video) (int, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, nil
	}

	total := 0
	for _, token := range strings.Split(expr, "+") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		value, err := evalXStackToken(token, videos)
		if err != nil {
			return 0, fmt.Errorf("evalXStackExpr: %w", err)
		}
		total += value
	}

	return total, nil
}

func evalXStackToken(token string, videos []Video) (int, error) {
	if isNumericToken(token) {
		n, err := strconv.Atoi(token)
		if err != nil {
			return 0, fmt.Errorf("evalXStackToken: %w", err)
		}
		return n, nil
	}
	if len(token) <= 1 || (token[0] != 'w' && token[0] != 'h') {
		return 0, fmt.Errorf("evalXStackToken: unsupported token %q", token)
	}

	index, err := strconv.Atoi(token[1:])
	if err != nil {
		return 0, fmt.Errorf("evalXStackToken: %w", err)
	}
	if index < 0 || index >= len(videos) {
		return 0, fmt.Errorf("evalXStackToken: index %d out of range (len=%d)", index, len(videos))
	}

	if token[0] == 'w' {
		return int(videos[index].width), nil
	}
	return int(videos[index].height), nil
}

func isNumericToken(token string) bool {
	return token[0] >= '0' && token[0] <= '9'
}
