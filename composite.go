package moviego

import "fmt"

// initRawVideo ensures a video has at least one filter complex entry by adding
// identity filters (null/anull) for raw videos loaded directly from files.
func initRawVideo(v *Video) {
	if len(v.videoFilterComplex) > 0 {
		return
	}
	filename := v.filenames[0]
	fileLabel := v.nextLabel(filename)
	fileCopyVideo := FileCopy{Filename: filename, Label: fmt.Sprintf("%s_v", fileLabel)}
	fileCopyAudio := FileCopy{Filename: filename, Label: fmt.Sprintf("%s_a", fileLabel)}

	label := v.nextLabel(filename)
	order := incrementOrderCounter()

	v.videoFilterComplex = append(v.videoFilterComplex, FilterComplex{
		Order:         order,
		FilterElement: fmt.Sprintf("[%s]null", fileCopyVideo.Label),
		FileCopy:      fileCopyVideo,
		Label:         fmt.Sprintf("%s_v", label),
	})
	v.audioFilterComplex = append(v.audioFilterComplex, FilterComplex{
		Order:         order,
		FilterElement: fmt.Sprintf("[%s]anull", fileCopyAudio.Label),
		FileCopy:      fileCopyAudio,
		Label:         fmt.Sprintf("%s_a", label),
	})
}

// CompositeClip overlays multiple videos on top of each other, similar to
// MoviePy's CompositeVideoClip. The first video is the background; each
// subsequent video is overlaid using its Position (defaults to center).
// Audio from all layers is mixed together with amix.
func CompositeClip(videos []Video) (*Video, error) {
	if len(videos) == 0 {
		return nil, fmt.Errorf("CompositeClip: no videos provided")
	}
	if len(videos) == 1 {
		v := videos[0]
		return &v, nil
	}

	for i := range videos {
		initRawVideo(&videos[i])
	}

	filenames := []string{}
	videoFilterComplex := []FilterComplex{}
	audioFilterComplex := []FilterComplex{}
	seen := make(map[string]struct{})

	var maxDuration float64
	for _, video := range videos {
		for _, filename := range video.filenames {
			if _, exists := seen[filename]; exists {
				continue
			}
			seen[filename] = struct{}{}
			filenames = append(filenames, filename)
		}
		videoFilterComplex = append(videoFilterComplex, video.videoFilterComplex...)
		audioFilterComplex = append(audioFilterComplex, video.audioFilterComplex...)

		if video.duration > maxDuration {
			maxDuration = video.duration
		}
	}

	order := incrementOrderCounter()
	compositeLabel := fmt.Sprintf("composite_%d", incrementGlobalCounter())

	// Build the overlay chain:
	//   [bg][fg1]overlay=x=...:y=...[ov1];
	//   [ov1][fg2]overlay=x=...:y=...[ov2];
	//   ...last one gets the final label
	currentLabel := videos[0].lastVideoLabel()
	filterElement := ""

	for i := 1; i < len(videos); i++ {
		pos := videos[i].GetPosition()
		fgLabel := videos[i].lastVideoLabel()

		overlayExpr := fmt.Sprintf("[%s][%s]overlay=x=%s:y=%s", currentLabel, fgLabel, pos.X, pos.Y)

		if i < len(videos)-1 {
			intermediateLabel := fmt.Sprintf("%s_ov%d", compositeLabel, i)
			filterElement += fmt.Sprintf("%s[%s];", overlayExpr, intermediateLabel)
			currentLabel = intermediateLabel
		} else {
			filterElement += overlayExpr
		}
	}

	videoFilterComplex = append(videoFilterComplex, FilterComplex{
		Order:         order,
		Label:         compositeLabel + "_v",
		FilterElement: filterElement,
	})

	// Build audio mix: [a0][a1][a2]...amix=inputs=N:duration=longest
	audioMixElement := ""
	for _, video := range videos {
		audioMixElement += fmt.Sprintf("[%s]", video.lastAudioLabel())
	}
	audioMixElement += fmt.Sprintf("amix=inputs=%d:duration=longest", len(videos))

	audioFilterComplex = append(audioFilterComplex, FilterComplex{
		Order:         order,
		Label:         compositeLabel + "_a",
		FilterElement: audioMixElement,
	})

	bg := videos[0]
	return &Video{
		filenames:          filenames,
		startTime:          0,
		endTime:            maxDuration,
		audioFilterComplex: audioFilterComplex,
		videoFilterComplex: videoFilterComplex,
		duration:           maxDuration,
		codec:              bg.codec,
		width:              bg.width,
		height:             bg.height,
		fps:                bg.fps,
		frames:             uint64(float64(bg.fps) * maxDuration),
		ffmpegArgs:         bg.ffmpegArgs,
		isTemp:             false,
		audio:              bg.audio,
		bitRate:            bg.bitRate,
		preset:             bg.preset,
		withMask:           bg.withMask,
		pixelFormat:        bg.pixelFormat,
	}, nil
}
