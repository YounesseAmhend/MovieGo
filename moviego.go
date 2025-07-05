package moviego

import (
	"encoding/json"
	"os/exec"
	"strconv"
)

type Video struct {
	filename   string
	codec      string
	width      int64
	height     int64
	duration   float64
	ffmpegArgs map[string][]string
	isTemp     bool
	audio      Audio
	bitRate    int64
}

type Audio struct {
	codec      string
	sampleRate int64
	channels   int8
	bps        int
	bitRate    int64
	duration   float64
}

func NewVideoFile(filename string) *Video {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_format", "-show_streams", filename, "-of", "json")
	output, err := cmd.Output()
	if err != nil {
		return &Video{}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return &Video{}
	}

	video := &Video{
		filename:   filename,
		ffmpegArgs: make(map[string][]string),
	}

	if format, ok := result["format"].(map[string]interface{}); ok {
		if filename, ok := format["filename"].(string); ok {
			video.SetFilename(filename)
		}
	}

	if streams, ok := result["streams"].([]interface{}); ok {
		for _, stream := range streams {
			if streamMap, ok := stream.(map[string]interface{}); ok {
				if codecType, ok := streamMap["codec_type"].(string); ok && codecType == "video" {
					if codec, ok := streamMap["codec_name"].(string); ok {
						video.Codec(codec)
					}
					if width, ok := streamMap["width"].(float64); ok {
						video.Width(int64(width))
					}
					if height, ok := streamMap["height"].(float64); ok {
						video.Height(int64(height))
					}
					if duration, ok := streamMap["duration"].(string); ok {
						if dur, err := strconv.ParseFloat(duration, 64); err == nil {
							video.Duration(dur)
						}
					}
					if bitRate, ok := streamMap["bit_rate"].(string); ok {
						if br, err := strconv.ParseInt(bitRate, 10, 64); err == nil {
							video.bitRate = br
						}
					}
				} else if codecType == "audio" {
					audio := Audio{}
					if codec, ok := streamMap["codec_name"].(string); ok {
						audio.Codec(codec)
					}
					if sampleRate, ok := streamMap["sample_rate"].(string); ok {
						if sr, err := strconv.ParseInt(sampleRate, 10, 64); err == nil {
							audio.SampleRate(sr)
						}
					}
					if channels, ok := streamMap["channels"].(float64); ok {
						audio.Channels(int8(channels))
					}
					if bitRate, ok := streamMap["bit_rate"].(string); ok {
						if br, err := strconv.ParseInt(bitRate, 10, 64); err == nil {
							audio.BitRate(br)
						}
					}
					if duration, ok := streamMap["duration"].(string); ok {
						if dur, err := strconv.ParseFloat(duration, 64); err == nil {
							audio.Duration(dur)
						}
					}
					video.SetAudio(audio)
				}
			}
		}
	}
	return video
}

func (v *Video) SetFilename(filePath string) *Video {
	v.filename = filePath
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

func (v *Video) GetAudio() Audio {
	return v.audio
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
