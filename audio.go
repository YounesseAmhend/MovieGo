package moviego

type Audio struct {
	codec      string
	sampleRate int64
	channels   int8
	bps        int
	bitRate    int64
	duration   float64
}

func (a *Audio) Codec(codec string) *Audio {
	a.codec = codec
	return a
}

func (a *Audio) GetCodec() string {
	return a.codec
}

func (a *Audio) SampleRate(sampleRate int64) *Audio {
	a.sampleRate = sampleRate
	return a
}

func (a *Audio) GetSampleRate() int64 {
	return a.sampleRate
}

func (a *Audio) Channels(channels int8) *Audio {
	a.channels = channels
	return a
}

func (a *Audio) GetChannels() int8 {
	return a.channels
}

func (a *Audio) Bps(bps int) *Audio {
	a.bps = bps
	return a
}

func (a *Audio) GetBps() int {
	return a.bps
}

func (a *Audio) BitRate(bitRate int64) *Audio {
	a.bitRate = bitRate
	return a
}

func (a *Audio) GetBitRate() int64 {
	return a.bitRate
}

func (a *Audio) Duration(duration float64) *Audio {
	a.duration = duration
	return a
}

func (a *Audio) GetDuration() float64 {
	return a.duration
}
