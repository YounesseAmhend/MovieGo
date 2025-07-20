package moviego

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

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

						if frames, ok := streamMap["nb_frames"].(string); ok {
							if fra, err := strconv.ParseInt(frames, 10, 64); err == nil {
								video.SetFrames(fra)
							}
						}
						if bitRate, ok := streamMap["bit_rate"].(string); ok {
							if br, err := strconv.ParseInt(bitRate, 10, 64); err == nil {
								video.bitRate = br
							}
						}
						if frameRate, ok := streamMap["avg_frame_rate"].(string); ok {
							parts := strings.Split(frameRate, "/")
							if len(parts) == 2 {
								numerator, err1 := strconv.ParseInt(parts[0], 10, 16)
								denominator, err2 := strconv.ParseInt(parts[1], 10, 16)
								if err1 == nil && err2 == nil && denominator != 0 {
									video.fps = int16(numerator / denominator)
								}
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
	}
	return video
}
