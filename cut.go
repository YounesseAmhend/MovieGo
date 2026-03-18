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
		return nil, fmt.Errorf("Cut: start must be less than end (start=%.4f, end=%.4f, duration=%.4f, file=%s, label=%s)",
			start, end, v.duration, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}

	audioFilterComplex, _ := deepCopySlice(v.audio.filterComplex)
	videoFilterComplex, _ := deepCopySlice(v.filterComplex)
	order := incrementOrderCounter()

	if len(v.filterComplex) == 0 { // No need to check for audio\
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
			FilterElement: fmt.Sprintf("[%s]atrim=start=%.2f:end=%.2f,asetpts=PTS-STARTPTS", v.audio.lastAudioLabel(), start, end),
			Label:         audioLabel,
		})
	}

	// Create a new video with copied properties
	newAudio := v.audio
	newAudio.filterComplex = audioFilterComplex

	newVideo := &Video{
		filenames:        v.filenames,
		codec:            v.codec,
		width:            v.width,
		height:           v.height,
		fps:              v.fps,
		duration:         end - start,
		frames:           uint64(float64(v.fps) * (end - start)),
		ffmpegArgs:       v.ffmpegArgs,
		filterComplex:    videoFilterComplex,
		isTemp:           v.isTemp,
		audio:            newAudio,
		bitRate:          v.bitRate,
		preset:           v.preset,
		withMask:         v.withMask,
		pixelFormat:      v.pixelFormat,
		startTime:        0,
		endTime:          end - start,
		position:         v.position,
		animatedPosition: v.animatedPosition,
		animatedOpacity:  v.animatedOpacity,
	}

	return newVideo, nil
}
