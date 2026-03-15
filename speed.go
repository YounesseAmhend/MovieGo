package moviego

import (
	"fmt"
	"strings"
)

// buildAtempoChain builds an atempo filter chain for the given speed.
// atempo supports 0.5-2.0 per instance; chain multiple for other values.
func buildAtempoChain(speed float64) string {
	var filters []string
	for speed > 2.0 {
		filters = append(filters, "atempo=2")
		speed /= 2
	}
	for speed < 0.5 && speed > 0 {
		filters = append(filters, "atempo=0.5")
		speed /= 0.5
	}
	if speed > 0 && speed != 1.0 {
		filters = append(filters, fmt.Sprintf("atempo=%.4f", speed))
	}
	if len(filters) == 0 {
		return "atempo=1"
	}
	return strings.Join(filters, ",")
}

// buildAudioSpeedFilter builds the audio filter string for speed and optional pitch change.
// If pitch is 0 or 1, pitch is preserved (atempo only). Otherwise asetrate/aresample are used.
func buildAudioSpeedFilter(speed, pitch float64, sampleRate uint64) string {
	sr := sampleRate
	if sr == 0 {
		sr = 44100
	}
	atempoChain := buildAtempoChain(speed)
	if pitch == 0 || pitch == 1 {
		return atempoChain
	}
	// asetrate=sr*pitch,atempo=1/pitch,aresample=sr for pitch change, then speed
	pitchPart := fmt.Sprintf("asetrate=%.0f,atempo=%.4f,aresample=%d", float64(sr)*pitch, 1/pitch, sr)
	return pitchPart + "," + atempoChain
}

// Speed changes playback speed for both video and audio.
// Parameters:
//   - speed: Playback speed multiplier (e.g. 2.0 = 2x faster, 0.5 = half speed)
//   - pitch: Optional. If omitted or 0 or 1, pitch is preserved. Otherwise a multiplier (e.g. 1.5 = higher, 0.5 = lower)
//
// Returns a new Video object with updated metadata (no file is created until WriteVideo is called)
func (v *Video) Speed(speed float64, pitch ...float64) (*Video, error) {
	if speed <= 0 {
		return nil, fmt.Errorf("speed must be positive, got %f", speed)
	}
	pitchVal := 0.0
	if len(pitch) > 0 {
		pitchVal = pitch[0]
		if pitchVal < 0 {
			return nil, fmt.Errorf("pitch must be non-negative, got %f", pitchVal)
		}
	}

	audioFilterComplex, _ := deepCopySlice(v.audio.filterComplex)
	videoFilterComplex, _ := deepCopySlice(v.filterComplex)
	order := incrementOrderCounter()

	newDuration := v.duration / speed
	sampleRate := v.GetAudio().GetSampleRate()
	videoFilter := fmt.Sprintf("setpts=PTS/%.4f", speed)
	audioFilter := buildAudioSpeedFilter(speed, pitchVal, sampleRate)

	if len(v.filterComplex) == 0 {
		filename := v.filenames[0]
		fileLabel := v.nextLabel(filename)
		fileCopyVideo := &FileCopy{
			Filename: filename,
			Label:    fmt.Sprintf("%s_v", fileLabel),
		}
		fileCopyAudio := &FileCopy{
			Filename: filename,
			Label:    fmt.Sprintf("%s_a", fileLabel),
		}
		label := v.nextLabel(filename)
		videoLabel := fmt.Sprintf("%s_v", label)
		audioLabel := fmt.Sprintf("%s_a", label)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]%s", fileCopyVideo.Label, videoFilter),
			FileCopy:      *fileCopyVideo,
			Label:         videoLabel,
		})
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]%s", fileCopyAudio.Label, audioFilter),
			FileCopy:      *fileCopyAudio,
			Label:         audioLabel,
		})
	} else {
		label := v.nextLabel(v.lastFilename())
		videoLabel := fmt.Sprintf("%s_v", label)
		audioLabel := fmt.Sprintf("%s_a", label)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]%s", v.lastVideoLabel(), videoFilter),
			Label:         videoLabel,
		})
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]%s", v.audio.lastAudioLabel(), audioFilter),
			Label:         audioLabel,
		})
	}

	newAudio := v.audio
	newAudio.filterComplex = audioFilterComplex

	newVideo := &Video{
		filenames:          v.filenames,
		codec:              v.codec,
		width:              v.width,
		height:             v.height,
		fps:                v.fps,
		duration:           newDuration,
		frames:             uint64(float64(v.fps) * newDuration),
		ffmpegArgs:         v.ffmpegArgs,
		filterComplex: videoFilterComplex,
		isTemp:             v.isTemp,
		audio:              newAudio,
		bitRate:            v.bitRate,
		preset:             v.preset,
		withMask:           v.withMask,
		pixelFormat:        v.pixelFormat,
		startTime:          0,
		endTime:            newDuration,
		position:           v.position,
		animatedPosition:   v.animatedPosition,
		animatedOpacity:    v.animatedOpacity,
	}

	return newVideo, nil
}
