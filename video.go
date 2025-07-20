package moviego

// Video struct and its methods moved from moviego.go

type Video struct {
	filename   string
	codec      string
	width      int64
	height     int64
	fps        int16
	duration   float64
	frames     int64
	ffmpegArgs map[string][]string
	isTemp     bool
	audio      Audio
	bitRate    int64
}

func (v *Video) SetFrames(frames int64) *Video {
	v.frames = frames
	return v
}

func (v *Video) GetFrames() int64 {
	return v.frames
}

func (v *Video) SetFilename(filename string) *Video {
	v.filename = filename
	return v
}

func (v *Video) GetFilename() string {
	return v.filename
}

func (v *Video) Codec(codec string) *Video {
	v.codec = codec
	return v
}

func (v *Video) GetCodec() string {
	return v.codec
}

func (v *Video) Width(width int64) *Video {
	v.width = width
	return v
}

func (v *Video) GetWidth() int64 {
	return v.width
}

func (v *Video) Height(height int64) *Video {
	v.height = height
	return v
}

func (v *Video) GetHeight() int64 {
	return v.height
}

func (v *Video) Duration(duration float64) *Video {
	v.duration = duration
	return v
}

func (v *Video) GetDuration() float64 {
	return v.duration
}

func (v *Video) FfmpegArgs(ffmpegArgs map[string][]string) *Video {
	v.ffmpegArgs = ffmpegArgs
	return v
}

func (v *Video) GetFfmpegArgs() map[string][]string {
	return v.ffmpegArgs
}

func (v *Video) SetIsTemp(isTemp bool) *Video {
	v.isTemp = isTemp
	return v
}

func (v *Video) GetIsTemp() bool {
	return v.isTemp
}

func (v *Video) BitRate(bitRate int64) *Video {
	v.bitRate = bitRate
	return v
}

func (v *Video) GetBitRate() int64 {
	return v.bitRate
}

func (v *Video) SetAudio(audio Audio) *Video {
	v.audio = audio
	return v
}

func (v *Video) GetAudio() *Audio {
	return &v.audio
}

func (v *Video) SetFps(fps int16) *Video {
	v.fps = fps
	return v
}

func (v *Video) GetFps() int16 {
	return v.fps
} 