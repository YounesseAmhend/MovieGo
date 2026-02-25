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
		videoFilterComplex = append(videoFilterComplex, video.videoFilterComplex...)
		audioFilterComplex = append(audioFilterComplex, video.audioFilterComplex...)

		filterElement += fmt.Sprintf("[%s][%s]", video.lastAudioLabel(), video.lastVideoLabel())
		duration += video.duration
	}
	label := fmt.Sprintf("concat_%d", incrementGlobalCounter())

	filterElement += fmt.Sprintf("concat=n=%d:a=1:v=1[%s_a]", len(videos), label)

	audioFilterComplex = append(audioFilterComplex, FilterComplex{
		label: label + "_a",
		filterElements: []string{},
	})
	videoFilterComplex = append(videoFilterComplex, FilterComplex{
		label:          label + "_v",
		filterElements: []string{filterElement},
	})

	return &Video{
		filenames:          filenames,
		startTime:          0,
		endTime:            duration,
		audioFilterComplex: audioFilterComplex,
		videoFilterComplex: videoFilterComplex,
		duration:           duration,
		codec:              videos[0].codec,
		width:              videos[0].width,
		height:             videos[0].height,
		fps:                videos[0].fps,
		frames:             uint64(float64(videos[0].fps) * duration),
		ffmpegArgs:         videos[0].ffmpegArgs,
		filters:            videos[0].filters,
		isTemp:             false,
		audio:              videos[0].audio,
		bitRate:            videos[0].bitRate,
		preset:             videos[0].preset,
		withMask:           videos[0].withMask,
		pixelFormat:        videos[0].pixelFormat,
		textClips:          videos[0].textClips,
		subtitleClips:      videos[0].subtitleClips,
	}, nil
}
