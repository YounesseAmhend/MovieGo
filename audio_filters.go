package moviego

import (
	"fmt"
	"strings"
)

// audioFilter applies an audio FFmpeg filter.
func (a *Audio) audioFilter(filter string) (*Audio, error) {
	filterComplex, _ := deepCopySlice(a.filterComplex)
    order := incrementOrderCounter()
	
	if len(a.filterComplex) == 0 {
		filename := a.filenames[0]
		fileLabel := a.nextLabel()
		fileCopyAudio := FileCopy{
			Filename: filename,
			Label:    fmt.Sprintf("%s_a", fileLabel),
		}
		label := a.nextLabel()
		audioLabel := fmt.Sprintf("%s_a", label)
		
		filterComplex = append(filterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]%s", fileCopyAudio.Label, filter),
			FileCopy:      fileCopyAudio,
			Label:         audioLabel,
		})
	} else {
		label := a.nextLabel()
		audioLabel := fmt.Sprintf("%s_a", label)
		filterComplex = append(filterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]%s", a.lastAudioLabel(), filter),
			Label:         audioLabel,
		})
	}

	return &Audio{
		filenames:       a.filenames,
		codec:           a.codec,
		sampleRate:      a.sampleRate,
		channels:        a.channels,
		bps:             a.bps,
		bitRate:         a.bitRate,
		duration:        a.duration,
		labelCounter:    a.labelCounter,
		filterComplex:   filterComplex,
	}, nil
}

// Volume adjusts the audio volume (0.0 = silent, 1.0 = normal, >1.0 = louder).
func (a *Audio) Volume(volume float64) (*Audio, error) {
	if volume < 0 {
		return nil, fmt.Errorf("Volume: must be non-negative (got=%f, file=%s, label=%s)", volume, safeFirstFilename(a.filenames), a.safeLabel())
	}
	return a.audioFilter(fmt.Sprintf("volume=%.4f", volume))
}

// FadeIn applies a fade-in effect at the start.
func (a *Audio) FadeIn(duration float64) (*Audio, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("FadeIn: duration must be positive (got=%f, file=%s, label=%s)", duration, safeFirstFilename(a.filenames), a.safeLabel())
	}
	return a.audioFilter(fmt.Sprintf("afade=t=in:st=0:d=%.4f", duration))
}

// FadeOut applies a fade-out effect ending at the audio's duration.
func (a *Audio) FadeOut(duration float64) (*Audio, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("FadeOut: duration must be positive (got=%f, file=%s, label=%s)", duration, safeFirstFilename(a.filenames), a.safeLabel())
	}
	if duration > a.duration {
		return nil, fmt.Errorf("FadeOut: duration %.4f exceeds audio duration %.4f (file=%s, label=%s)", duration, a.duration, safeFirstFilename(a.filenames), a.safeLabel())
	}
	startTime := a.duration - duration
	return a.audioFilter(fmt.Sprintf("afade=t=out:st=%.4f:d=%.4f", startTime, duration))
}

// LowPass applies a low-pass filter with the given frequency.
func (a *Audio) LowPass(freq float64) (*Audio, error) {
	if freq <= 0 {
		return nil, fmt.Errorf("LowPass: frequency must be positive (got=%f, file=%s, label=%s)", freq, safeFirstFilename(a.filenames), a.safeLabel())
	}
	return a.audioFilter(fmt.Sprintf("lowpass=f=%.4f", freq))
}

// HighPass applies a high-pass filter with the given frequency.
func (a *Audio) HighPass(freq float64) (*Audio, error) {
	if freq <= 0 {
		return nil, fmt.Errorf("HighPass: frequency must be positive (got=%f, file=%s, label=%s)", freq, safeFirstFilename(a.filenames), a.safeLabel())
	}
	return a.audioFilter(fmt.Sprintf("highpass=f=%.4f", freq))
}

// Tempo changes the audio tempo without changing pitch.
// Supported range: 0.5 to 100.0.
func (a *Audio) Tempo(tempo float64) (*Audio, error) {
	if tempo < 0.5 || tempo > 100.0 {
		return nil, fmt.Errorf("Tempo: must be between 0.5 and 100.0 (got=%f, file=%s, label=%s)", tempo, safeFirstFilename(a.filenames), a.safeLabel())
	}
	// ffmpeg's atempo only supports 0.5-2.0 per instance.
	var filters []string
	curr := tempo
	for curr > 2.0 {
		filters = append(filters, "atempo=2.0")
		curr /= 2.0
	}
	for curr < 0.5 {
		filters = append(filters, "atempo=0.5")
		curr /= 0.5
	}
	filters = append(filters, fmt.Sprintf("atempo=%.4f", curr))
	
	newDuration := a.duration / tempo
	newAudio, err := a.audioFilter(strings.Join(filters, ","))
	if err != nil {
		return nil, fmt.Errorf("Tempo[file=%s, label=%s]: %w", safeFirstFilename(a.filenames), a.safeLabel(), err)
	}
	newAudio.duration = newDuration
	return newAudio, nil
}

// BassBoost boosts the bass frequencies.
func (a *Audio) BassBoost(gain float64) (*Audio, error) {
	return a.audioFilter(fmt.Sprintf("bass=g=%.4f", gain))
}

// TrebleBoost boosts the treble frequencies.
func (a *Audio) TrebleBoost(gain float64) (*Audio, error) {
	return a.audioFilter(fmt.Sprintf("treble=g=%.4f", gain))
}

// Normalize normalizes the audio levels.
func (a *Audio) Normalize() (*Audio, error) {
	return a.audioFilter("loudnorm")
}

// Echo applies an echo effect.
// delay: delay in milliseconds, decay: decay amount (0-1).
func (a *Audio) Echo(delay, decay float64) (*Audio, error) {
	if delay <= 0 || decay <= 0 || decay >= 1 {
		return nil, fmt.Errorf("Echo: invalid parameters (delay=%.4f, decay=%.4f, file=%s, label=%s) -- delay must be > 0 and decay must be between 0 and 1", delay, decay, safeFirstFilename(a.filenames), a.safeLabel())
	}
	return a.audioFilter(fmt.Sprintf("aecho=0.8:0.9:%.4f:%.4f", delay, decay))
}
