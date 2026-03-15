package moviego

import "fmt"

func Concatenate(videos []Video) (*Video, error) {

	filenames := []string{}
	videoFilterComplex := []FilterComplex{}
	audioFilterComplex := []FilterComplex{}
	seen := make(map[string]struct{})

	filterElement := ""
	var duration float64
	for _, video := range videos {
		for _, filename := range video.filenames {
			if _, exists := seen[filename]; exists {
				continue
			}
			seen[filename] = struct{}{}
			filenames = append(filenames, filename)
		}
		videoFilterComplex = append(videoFilterComplex, video.filterComplex...)
		audioFilterComplex = append(audioFilterComplex, video.audio.filterComplex...)

		filterElement += fmt.Sprintf("[%s][%s]", video.lastVideoLabel(), video.audio.lastAudioLabel())
		duration += video.duration
	}
	label := fmt.Sprintf("concat_%d", incrementGlobalCounter())

	filterElement += fmt.Sprintf("concat=n=%d:a=1:v=1[%s_v][%s_a]", len(videos), label, label)

	order := incrementOrderCounter()

	audioFilterComplex = append(audioFilterComplex, FilterComplex{
		Order:         order,
		Label:         label + "_a",
		FilterElement: "",
	})
	videoFilterComplex = append(videoFilterComplex, FilterComplex{
		Order:         order,
		Label:         label + "_v",
		FilterElement: filterElement,
	})

	newAudio := videos[0].audio
	newAudio.filterComplex = audioFilterComplex
	newAudio.duration = duration

	return &Video{
		filenames:          filenames,
		startTime:          0,
		endTime:            duration,
		filterComplex: videoFilterComplex,
		duration:           duration,
		codec:              videos[0].codec,
		width:              videos[0].width,
		height:             videos[0].height,
		fps:                videos[0].fps,
		frames:             uint64(float64(videos[0].fps) * duration),
		ffmpegArgs:         videos[0].ffmpegArgs,
		isTemp:             false,
		audio:              newAudio,
		bitRate:            videos[0].bitRate,
		preset:             videos[0].preset,
		withMask:           videos[0].withMask,
		pixelFormat:        videos[0].pixelFormat,
		position:           videos[0].position,
	}, nil
}
