package moviego

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

func NewVideoFile(filename string) (*Video, error) {
	if filename == "" {
		return nil, fmt.Errorf("NewVideoFile: filename cannot be empty")
	}

	ffprobePath, err := getFFprobePath()
	if err != nil {
		return nil, fmt.Errorf("NewVideoFile: ffprobe not found for '%s': %w", filename, err)
	}
	cmd := exec.Command(ffprobePath, "-v", "error", "-show_format", "-show_streams", filename, "-of", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("NewVideoFile: failed to probe video file '%s': %w", filename, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("NewVideoFile: failed to parse metadata for '%s': %w", filename, err)
	}

	video := &Video{
		filenames:   []string{filename},
		ffmpegArgs: make(map[string][]string),
	}

	if format, ok := result["format"].(map[string]interface{}); ok {
		if filename, ok := format["filename"].(string); ok {
			video.SetFilename([]string{filename})
		}
	}

	if streams, ok := result["streams"].([]interface{}); ok {
		for _, stream := range streams {
			if streamMap, ok := stream.(map[string]interface{}); ok {
				if codecType, ok := streamMap["codec_type"].(string); ok && codecType == "video" {
					if codec, ok := streamMap["codec_name"].(string); ok {
						video.Codec(Codec(codec))
					}
					if width, ok := streamMap["width"].(float64); ok {
						video.Width(uint64(width))
					}
					if height, ok := streamMap["height"].(float64); ok {
						video.Height(uint64(height))
					}
					if duration, ok := streamMap["duration"].(string); ok {
						if dur, err := strconv.ParseFloat(duration, 64); err == nil {
							video.Duration(dur)
						}
					}
					if frames, ok := streamMap["nb_frames"].(string); ok {
						if fra, err := strconv.ParseUint(frames, 10, 64); err == nil {
							video.SetFrames(fra)
						}
					}
					if bitRate, ok := streamMap["bit_rate"].(string); ok {
						video.BitRate(bitRate)
					}
					if frameRate, ok := streamMap["avg_frame_rate"].(string); ok {
						parts := strings.Split(frameRate, "/")
						if len(parts) == 2 {
							numerator, err1 := strconv.ParseFloat(parts[0], 64)
							denominator, err2 := strconv.ParseFloat(parts[1], 64)
							if err1 == nil && err2 == nil && denominator != 0 {
								fps := math.Round(numerator / denominator)
								if fps > 0 {
									video.fps = uint64(fps)
								}
							}
						}
					}
					// Set default FPS if parsing failed or resulted in 0
					if video.fps == 0 {
						video.fps = 30 // Default to 30 fps
					}
				} else if codecType, ok := streamMap["codec_type"].(string); ok && codecType == "audio" {
					audio := Audio{}
					if codec, ok := streamMap["codec_name"].(string); ok {
						audio.Codec(codec)
					}
					if sampleRate, ok := streamMap["sample_rate"].(string); ok {
						if sr, err := strconv.ParseUint(sampleRate, 10, 64); err == nil {
							audio.SampleRate(sr)
						}
					}
					if channels, ok := streamMap["channels"].(float64); ok {
						audio.Channels(uint8(channels))
					}
					if bitRate, ok := streamMap["bit_rate"].(string); ok {
						if br, err := strconv.ParseUint(bitRate, 10, 64); err == nil {
							audio.BitRate(br)
						}
					}
					if duration, ok := streamMap["duration"].(string); ok {
						if dur, err := strconv.ParseFloat(duration, 64); err == nil {
							audio.Duration(dur)
						}
					}
					audio.SetFilename([]string{filename})
					video.SetAudio(audio)
				}
			}
		}
	}

	// Validate essential video properties
	if video.GetWidth() <= 0 || video.GetHeight() <= 0 {
		return nil, fmt.Errorf("NewVideoFile: video file '%s' has invalid dimensions (%dx%d)", filename, video.GetWidth(), video.GetHeight())
	}
	if video.GetDuration() <= 0 {
		return nil, fmt.Errorf("NewVideoFile: video file '%s' has invalid duration (%.2f)", filename, video.GetDuration())
	}
	// FPS is already set to default 30 if parsing failed, so it should be valid

	return video, nil
}

// AudioFile probes an audio file using ffprobe and returns an Audio with
// its metadata populated (codec, sample rate, channels, bit rate, duration).
func AudioFile(filename string) (*Audio, error) {
	if filename == "" {
		return nil, fmt.Errorf("AudioFile: filename cannot be empty")
	}

	ffprobePath, err := getFFprobePath()
	if err != nil {
		return nil, fmt.Errorf("AudioFile: ffprobe not found for '%s': %w", filename, err)
	}
	cmd := exec.Command(ffprobePath, "-v", "error", "-show_format", "-show_streams", filename, "-of", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("AudioFile: failed to probe audio file '%s': %w", filename, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("AudioFile: failed to parse metadata for '%s': %w", filename, err)
	}

	audio := &Audio{
		filenames: []string{filename},
	}

	if streams, ok := result["streams"].([]interface{}); ok {
		for _, stream := range streams {
			streamMap, ok := stream.(map[string]interface{})
			if !ok {
				continue
			}
			codecType, _ := streamMap["codec_type"].(string)
			if codecType != "audio" {
				continue
			}
			if codec, ok := streamMap["codec_name"].(string); ok {
				audio.Codec(codec)
			}
			if sampleRate, ok := streamMap["sample_rate"].(string); ok {
				if sr, err := strconv.ParseUint(sampleRate, 10, 64); err == nil {
					audio.SampleRate(sr)
				}
			}
			if channels, ok := streamMap["channels"].(float64); ok {
				audio.Channels(uint8(channels))
			}
			if bitRate, ok := streamMap["bit_rate"].(string); ok {
				if br, err := strconv.ParseUint(bitRate, 10, 64); err == nil {
					audio.BitRate(br)
				}
			}
			if duration, ok := streamMap["duration"].(string); ok {
				if dur, err := strconv.ParseFloat(duration, 64); err == nil {
					audio.Duration(dur)
				}
			}
			break // use the first audio stream
		}
	}

	if audio.GetCodec() == "" {
		return nil, fmt.Errorf("AudioFile: no audio stream found in '%s'", filename)
	}

	return audio, nil
}
