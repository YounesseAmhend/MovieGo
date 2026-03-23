package moviego

import "fmt"

// Transition defines the type of crossfade between two clips.
type Transition string

const (
	TransitionFade       Transition = "fade"
	TransitionFadeBlack  Transition = "fadeblack"
	TransitionFadeWhite  Transition = "fadewhite"
	TransitionWipeLeft   Transition = "wipeleft"
	TransitionWipeRight  Transition = "wiperight"
	TransitionWipeUp     Transition = "wipeup"
	TransitionWipeDown   Transition = "wipedown"
	TransitionSlideLeft  Transition = "slideleft"
	TransitionSlideRight Transition = "slideright"
	TransitionSlideUp    Transition = "slideup"
	TransitionSlideDown  Transition = "slidedown"
	TransitionCircleCrop Transition = "circlecrop"
	TransitionDissolve   Transition = "dissolve"
	TransitionPixelize   Transition = "pixelize"
	TransitionRadial     Transition = "radial"
	TransitionZoomIn     Transition = "zoomin"
	TransitionDiagTL     Transition = "diagtl"
	TransitionDiagTR     Transition = "diagtr"
	TransitionDiagBL     Transition = "diagbl"
	TransitionDiagBR     Transition = "diagbr"
)

// TransitionParams holds parameters for clip-to-clip transitions.
type TransitionParams struct {
	Transition Transition
	Duration   float64 // overlap duration in seconds
}

// ConcatenateWithTransition joins two clips with a transition effect.
// The transition overlaps the end of clip1 with the start of clip2.
func ConcatenateWithTransition(clip1, clip2 *Video, params TransitionParams) (*Video, error) {
	if clip1 == nil || clip2 == nil {
		return nil, fmt.Errorf("ConcatenateWithTransition: both clips must be non-nil")
	}
	if params.Duration <= 0 {
		return nil, fmt.Errorf("ConcatenateWithTransition: duration must be positive (duration=%.4f)", params.Duration)
	}
	if params.Duration >= clip1.GetDuration() {
		return nil, fmt.Errorf("ConcatenateWithTransition: duration %f must be less than clip1 duration %f", params.Duration, clip1.GetDuration())
	}
	if params.Duration >= clip2.GetDuration() {
		return nil, fmt.Errorf("ConcatenateWithTransition: duration %f must be less than clip2 duration %f", params.Duration, clip2.GetDuration())
	}
	if params.Transition == "" {
		params.Transition = TransitionFade
	}

	initRawVideo(clip1)
	initRawVideo(clip2)

	// Ensure matching dimensions
	if clip1.GetWidth() != clip2.GetWidth() || clip1.GetHeight() != clip2.GetHeight() {
		var err error
		clip2, err = clip2.Scale(ScaleParams{Width: int(clip1.GetWidth()), Height: int(clip1.GetHeight())})
		if err != nil {
			return nil, fmt.Errorf("ConcatenateWithTransition: clip2 scale failed: %w", err)
		}
	}

	offset := clip1.GetDuration() - params.Duration

	filenames := []string{}
	seen := make(map[string]struct{})
	for _, f := range clip1.GetFilenames() {
		if _, exists := seen[f]; !exists {
			seen[f] = struct{}{}
			filenames = append(filenames, f)
		}
	}
	for _, f := range clip2.GetFilenames() {
		if _, exists := seen[f]; !exists {
			seen[f] = struct{}{}
			filenames = append(filenames, f)
		}
	}

	videoFilterComplex := append([]FilterComplex{}, clip1.filterComplex...)
	videoFilterComplex = append(videoFilterComplex, clip2.filterComplex...)
	audioFilterComplex := append([]FilterComplex{}, clip1.audio.filterComplex...)
	audioFilterComplex = append(audioFilterComplex, clip2.audio.filterComplex...)

	order := incrementOrderCounter()
	label := fmt.Sprintf("xfade_%d", incrementGlobalCounter())

	xfadeFilter := fmt.Sprintf("[%s][%s]xfade=transition=%s:duration=%.4f:offset=%.4f[%s]",
		clip1.lastVideoLabel(), clip2.lastVideoLabel(),
		string(params.Transition), params.Duration, offset,
		label+"_v")
	acrossfadeFilter := fmt.Sprintf("[%s][%s]acrossfade=d=%.4f:c1=tri:c2=tri[%s]",
		clip1.audio.lastAudioLabel(), clip2.audio.lastAudioLabel(),
		params.Duration, label+"_a")

	videoFilterComplex = append(videoFilterComplex, FilterComplex{
		Order:         order,
		Label:         label + "_v",
		FilterElement: xfadeFilter,
	})
	audioFilterComplex = append(audioFilterComplex, FilterComplex{
		Order:         order,
		Label:         label + "_a",
		FilterElement: acrossfadeFilter,
	})

	newDuration := clip1.GetDuration() + clip2.GetDuration() - params.Duration

	newAudio := clip1.audio
	newAudio.filterComplex = audioFilterComplex
	newAudio.duration = newDuration

	return &Video{
		filenames:          filenames,
		startTime:          0,
		endTime:            newDuration,
		filterComplex: videoFilterComplex,
		duration:           newDuration,
		codec:              clip1.codec,
		width:              clip1.width,
		height:             clip1.height,
		fps:                clip1.fps,
		frames:             uint64(float64(clip1.fps) * newDuration),
		ffmpegArgs:         clip1.ffmpegArgs,
		isTemp:             false,
		audio:              newAudio,
		bitRate:            clip1.bitRate,
		preset:             clip1.preset,
		withMask:           clip1.withMask,
		pixelFormat:        clip1.pixelFormat,
	}, nil
}
