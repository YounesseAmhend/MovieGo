package moviego

import "fmt"

type Audio struct {
	filenames       []string
	codec           string
	sampleRate      uint64
	channels        uint8
	bps             uint
	bitRate         uint64
	duration        float64
	labelCounter    uint64
	filterComplex   []FilterComplex
}

func (a *Audio) SetFilename(filenames []string) *Audio {
	a.filenames = filenames
	return a
}

func (a *Audio) GetFilenames() []string {
	return a.filenames
}


func (a *Audio) Codec(codec string) *Audio {
	a.codec = codec
	return a
}

func (a *Audio) GetCodec() string {
	return a.codec
}

func (a *Audio) SampleRate(sampleRate uint64) *Audio {
	a.sampleRate = sampleRate
	return a
}

func (a *Audio) GetSampleRate() uint64 {
	return a.sampleRate
}

func (a *Audio) Channels(channels uint8) *Audio {
	a.channels = channels
	return a
}

func (a *Audio) GetChannels() uint8 {
	return a.channels
}

func (a *Audio) Bps(bps uint) *Audio {
	a.bps = bps
	return a
}

func (a *Audio) GetBps() uint {
	return a.bps
}

func (a *Audio) BitRate(bitRate uint64) *Audio {
	a.bitRate = bitRate
	return a
}

func (a *Audio) GetBitRate() uint64 {
	return a.bitRate
}

func (a *Audio) Duration(duration float64) *Audio {
	a.duration = duration
	return a
}

func (a *Audio) GetDuration() float64 {
	return a.duration
}

func (a *Audio) lastAudioLabel() string {
	if len(a.filterComplex) == 0 {
		return ""
	}
	return a.filterComplex[len(a.filterComplex)-1].Label
}

func (a *Audio) lastFilename() string {
	if len(a.filenames) == 0 {
		return ""
	}
	return a.filenames[len(a.filenames)-1]
}

func (a *Audio) nextLabel() string {
	id := incrementGlobalCounter()
	safeName := sanitize(a.lastFilename())
	return fmt.Sprintf("%d_%s", id, safeName)
}

