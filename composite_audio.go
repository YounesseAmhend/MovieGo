package moviego

import "fmt"

// initRawAudio ensures an audio object has at least one filter complex entry by adding
// an anull filter for raw audio loaded directly from files.
func initRawAudio(a *Audio) {
	if len(a.filterComplex) > 0 {
		return
	}
	filename := a.filenames[0]
	fileLabel := a.nextLabel()
	fileCopyAudio := FileCopy{Filename: filename, Label: fmt.Sprintf("%s_a", fileLabel)}

	label := a.nextLabel()
	order := incrementOrderCounter()

	a.filterComplex = append(a.filterComplex, FilterComplex{
		Order:         order,
		FilterElement: fmt.Sprintf("[%s]anull", fileCopyAudio.Label),
		FileCopy:      fileCopyAudio,
		Label:         fmt.Sprintf("%s_a", label),
	})
}

// Composite mixes multiple audio tracks together.
// Audio from all sources is mixed using amix.
func Composite(audios []Audio) (*Audio, error) {
	if len(audios) == 0 {
		return nil, fmt.Errorf("Composite: no audios provided")
	}
	if len(audios) == 1 {
		a := audios[0]
		return &a, nil
	}

	for i := range audios {
		initRawAudio(&audios[i])
	}

	filenames := []string{}
	audioFilterComplex := []FilterComplex{}
	seen := make(map[string]struct{})

	var maxDuration float64
	for _, audio := range audios {
		for _, filename := range audio.filenames {
			if _, exists := seen[filename]; exists {
				continue
			}
			seen[filename] = struct{}{}
			filenames = append(filenames, filename)
		}
		audioFilterComplex = append(audioFilterComplex, audio.filterComplex...)

		if audio.duration > maxDuration {
			maxDuration = audio.duration
		}
	}

	order := incrementOrderCounter()
	compositeLabel := fmt.Sprintf("composite_a_%d", incrementGlobalCounter())

	// Build audio mix: [a0][a1][a2]...amix=inputs=N:duration=longest
	audioMixElement := ""
	for _, audio := range audios {
		audioMixElement += fmt.Sprintf("[%s]", audio.lastAudioLabel())
	}
	audioMixElement += fmt.Sprintf("amix=inputs=%d:duration=longest", len(audios))

	audioFilterComplex = append(audioFilterComplex, FilterComplex{
		Order:         order,
		Label:         compositeLabel,
		FilterElement: audioMixElement,
	})

	first := audios[0]
	return &Audio{
		filenames:       filenames,
		codec:           first.codec,
		sampleRate:      first.sampleRate,
		channels:        first.channels,
		bps:             first.bps,
		bitRate:         first.bitRate,
		duration:        maxDuration,
		labelCounter:    first.labelCounter,
		filterComplex:   audioFilterComplex,
		replacementPath: "",
		removed:         false,
	}, nil
}
