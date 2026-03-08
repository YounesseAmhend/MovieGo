package moviego

import "fmt"

// Cut creates a new video segment with specified start and end times (lazy operation)
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
	order := incrementOrderCounter()


	if len(v.videoFilterComplex) == 0 { // No need to check for audio\
		filename := v.filenames[0]
		fileLabel := v.nextLabel(filename)
		fileCopyVideo := &FileCopy{
			Filename: filename,
			Label: fmt.Sprintf("%s_v", fileLabel),
		}
		fileCopyAudio := &FileCopy{
			Filename: filename,
			Label: fmt.Sprintf("%s_a", fileLabel),
		}
		label := v.nextLabel(filename)
		videoLabel := fmt.Sprintf("%s_v", label)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS", fileCopyVideo.Label, start, end),
			FileCopy:      *fileCopyVideo,
			Label:         videoLabel,
		})
		audioLabel := fmt.Sprintf("%s_a", label)
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]atrim=start=%.2f:end=%.2f,asetpts=PTS-STARTPTS", fileCopyAudio.Label, start, end),
			FileCopy:      *fileCopyAudio,
			Label:         audioLabel,
		})
	} else {
		label := v.nextLabel(v.lastFilename())
		videoLabel := fmt.Sprintf("%s_v", label)
		audioLabel := fmt.Sprintf("%s_a", label)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]trim=start=%.2f:end=%.2f,setpts=PTS-STARTPTS", v.lastVideoLabel(), start, end),
			Label:         videoLabel,
		})
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf("[%s]atrim=start=%.2f:end=%.2f,asetpts=PTS-STARTPTS", v.lastAudioLabel(), start, end),
			Label:         audioLabel,
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
		videoFilterComplex: videoFilterComplex,
		audioFilterComplex: audioFilterComplex,
		isTemp:        v.isTemp,
		audio:         v.audio,
		bitRate:       v.bitRate,
		preset:        v.preset,
		withMask:      v.withMask,
		pixelFormat:   v.pixelFormat,
		startTime:     0,
		endTime:       end - start,
		position:      v.position,
	}

	return newVideo, nil
}
