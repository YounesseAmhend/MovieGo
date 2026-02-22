package moviego

import "fmt"

// Subclip creates a new video segment with specified start and end times (lazy operation)
// Parameters:
//   - start: Start time in seconds (must be >= 0)
//   - end: End time in seconds (must be > start and <= video duration)
//
// Returns a new Video object with updated metadata (no file is created until WriteVideo is called)
func (v *Video) Cut(start, end float64) (*Video, error) {

	// Validate inputs
	if start < 0 {
		logger.Warn("Cut: Start time is less than 0 , setting to 0", "start", start)
		start = 0
	}
	if end > v.duration {
		logger.Warn("Cut: End time is greater than video duration, setting to video duration", "end", end, "duration", v.duration)
		end = v.duration
	}
	if start >= end {
		// Return empty video if invalid range
		return nil, fmt.Errorf("invalid range: start must be less than end")
	}

	audioFilterComplex, _ := deepCopySlice(v.audioFilterComplex)
	videoFilterComplex, _ := deepCopySlice(v.videoFilterComplex)

	if len(v.videoFilterComplex) == 0 { // No need to check for audio\
		filename := v.filenames[0]
		fileLabel := v.nextLabel(filename)
		fileCopyVideo := &FileCopy{
			filename: filename,
			label: fmt.Sprintf("%s_v", fileLabel),
		}
		fileCopyAudio := &FileCopy{
			filename: filename,
			label: fmt.Sprintf("%s_a", fileLabel),
		}
		label := v.nextLabel(filename)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			filterElements: []string{
				fmt.Sprintf("[%s]trim=start=%f:end=%f,setpts=PTS-STARTPTS", fileCopyVideo.label, start, end),
			},
			fileCopy: *fileCopyVideo,
			labelVideo: fmt.Sprintf("%s_v", label),
		})
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			filterElements: []string{
				fmt.Sprintf("[%s]atrim=start=%f:end=%f,asetpts=PTS-STARTPTS", fileCopyAudio.label, start, end),
			},
			fileCopy: *fileCopyAudio,
			labelAudio: fmt.Sprintf("%s_a", label),
		})
	}

	// Create a new video with copied properties
	newVideo := &Video{
		filenames:  v.filenames,
		codec:      v.codec,
		width:      v.width,
		height:     v.height,
		fps:        v.fps,
		duration:   end - start,
		frames:     uint64(float64(v.fps) * (end - start)),
		ffmpegArgs: v.ffmpegArgs,
		filters:    v.filters,
		videoFilterComplex: videoFilterComplex,
		audioFilterComplex: audioFilterComplex,
		isTemp:        v.isTemp,
		audio:         v.audio,
		bitRate:       v.bitRate,
		preset:        v.preset,
		withMask:      v.withMask,
		pixelFormat:   v.pixelFormat,
		startTime:     start,
		endTime:       end,
		textClips:     v.textClips,
		subtitleClips: v.subtitleClips,
	}

	return newVideo, nil
}
