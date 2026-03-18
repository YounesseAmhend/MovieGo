package moviego

import (
	"fmt"
	"strings"
)

const filterInputFormat = "[%s]%s"

// EqParams holds parameters for the eq (equalizer) filter: brightness, contrast, saturation, gamma.
type EqParams struct {
	Brightness float64 // -1 to 1, default 0
	Contrast   float64 // -1000 to 1000, default 1
	Saturation float64 // 0 to 3, default 1
	Gamma      float64 // 0.1 to 10, default 1
}

// HueParams holds parameters for the hue filter.
type HueParams struct {
	Degrees    float64 // hue angle in degrees
	Saturation float64 // -10 to 10, default 1
}

// ScaleParams holds parameters for the scale filter.
type ScaleParams struct {
	Width  int // pixels, use -1 to preserve aspect when Height set
	Height int // pixels, use -1 to preserve aspect when Width set
}

// CropParams holds parameters for the crop filter.
type CropParams struct {
	X      int
	Y      int
	Width  int
	Height int
}

// PadParams holds parameters for the pad filter.
type PadParams struct {
	Width  int
	Height int
	X      int
	Y      int
}

func evenDimension(value int) int {
	if value%2 != 0 {
		return value - 1
	}
	return value
}

func resolveScaledDimensions(srcW, srcH uint64, params ScaleParams) (uint64, uint64, error) {
	width := params.Width
	height := params.Height

	switch {
	case width == 0:
		width = int(srcW)
	case width == -1 && height > 0:
		width = int(float64(srcW) * (float64(height) / float64(srcH)))
	case width == -1:
		return 0, 0, fmt.Errorf("Scale: width can only be -1 when height is specified (width=%d, height=%d)", params.Width, params.Height)
	}

	switch {
	case height == 0:
		height = int(srcH)
	case height == -1 && width > 0:
		height = int(float64(srcH) * (float64(width) / float64(srcW)))
	case height == -1:
		return 0, 0, fmt.Errorf("Scale: height can only be -1 when width is specified (width=%d, height=%d)", params.Width, params.Height)
	}

	width = evenDimension(width)
	height = evenDimension(height)
	if width < 2 || height < 2 {
		return 0, 0, fmt.Errorf("Scale: produces invalid dimensions %dx%d", width, height)
	}

	return uint64(width), uint64(height), nil
}

// videoFilter applies a video-only FFmpeg filter and passes audio through unchanged.
// All public filter methods delegate to this.
func (v *Video) videoFilter(filter string) (*Video, error) {
	audioFilterComplex, _ := deepCopySlice(v.audio.filterComplex)
	videoFilterComplex, _ := deepCopySlice(v.filterComplex)
	order := incrementOrderCounter()

	videoFilter := filter
	audioFilter := "anull"

	if len(v.filterComplex) == 0 {
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
		audioLabel := fmt.Sprintf("%s_a", label)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf(filterInputFormat, fileCopyVideo.Label, videoFilter),
			FileCopy:      *fileCopyVideo,
			Label:         videoLabel,
		})
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf(filterInputFormat, fileCopyAudio.Label, audioFilter),
			FileCopy:      *fileCopyAudio,
			Label:         audioLabel,
		})
	} else {
		label := v.nextLabel(v.lastFilename())
		videoLabel := fmt.Sprintf("%s_v", label)
		audioLabel := fmt.Sprintf("%s_a", label)
		videoFilterComplex = append(videoFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf(filterInputFormat, v.lastVideoLabel(), videoFilter),
			Label:         videoLabel,
		})
		audioFilterComplex = append(audioFilterComplex, FilterComplex{
			Order:         order,
			FilterElement: fmt.Sprintf(filterInputFormat, v.audio.lastAudioLabel(), audioFilter),
			Label:         audioLabel,
		})
	}

	newAudio := v.audio
	newAudio.filterComplex = audioFilterComplex

	return &Video{
		filenames:          v.filenames,
		codec:              v.codec,
		width:              v.width,
		height:             v.height,
		fps:                v.fps,
		duration:           v.duration,
		frames:             v.frames,
		ffmpegArgs:         v.ffmpegArgs,
		filterComplex: videoFilterComplex,
		isTemp:             v.isTemp,
		audio:              newAudio,
		bitRate:            v.bitRate,
		preset:             v.preset,
		withMask:           v.withMask,
		pixelFormat:        v.pixelFormat,
		startTime:          v.startTime,
		endTime:            v.endTime,
		position:           v.position,
		animatedPosition:   v.animatedPosition,
		animatedOpacity:    v.animatedOpacity,
	}, nil
}

// Saturation adjusts color saturation (0-3, 1 = no change).
func (v *Video) Saturation(saturation float64) (*Video, error) {
	if saturation < 0 || saturation > 3 {
		return nil, fmt.Errorf("Saturation: must be 0-3 (got=%f, file=%s, label=%s)", saturation, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("eq=saturation=%.4f", saturation))
}

// Brightness adjusts brightness (-1 to 1, 0 = no change).
func (v *Video) Brightness(brightness float64) (*Video, error) {
	if brightness < -1 || brightness > 1 {
		return nil, fmt.Errorf("Brightness: must be -1 to 1 (got=%f, file=%s, label=%s)", brightness, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("eq=brightness=%.4f", brightness))
}

// Contrast adjusts contrast (-1000 to 1000, 1 = no change).
func (v *Video) Contrast(contrast float64) (*Video, error) {
	if contrast < -1000 || contrast > 1000 {
		return nil, fmt.Errorf("Contrast: must be -1000 to 1000 (got=%f, file=%s, label=%s)", contrast, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("eq=contrast=%.4f", contrast))
}

// ScaleRatio scales video by a ratio multiplier (e.g., 0.5 = half, 2.0 = double).
// Uses explicit dimensions from video metadata; ensures dimensions are even for codec compatibility.
func (v *Video) ScaleRatio(ratio float64) (*Video, error) {
	if ratio <= 0 {
		return nil, fmt.Errorf("ScaleRatio: ratio must be positive (got=%f, file=%s, label=%s)", ratio, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	width := int(float64(v.width) * ratio)
	height := int(float64(v.height) * ratio)
	if width < 1 || height < 1 {
		return nil, fmt.Errorf("ScaleRatio: ratio %f produces invalid dimensions %dx%d (file=%s, label=%s)", ratio, width, height, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	// Ensure dimensions are divisible by 2 (required by most codecs)
	width = width & 0xFFFFFFFE
	height = height & 0xFFFFFFFE
	if width < 2 {
		width = 2
	}
	if height < 2 {
		height = 2
	}
	return v.Scale(ScaleParams{Width: width, Height: height})
}

// Rotate rotates the video by the given angle in radians.
func (v *Video) Rotate(angle float64) (*Video, error) {
	return v.videoFilter(fmt.Sprintf("rotate=%.4f:fillcolor=none", angle))
}

// HorizontalFlip flips the video horizontally.
func (v *Video) HorizontalFlip() (*Video, error) {
	return v.videoFilter("hflip")
}

// VerticalFlip flips the video vertically.
func (v *Video) VerticalFlip() (*Video, error) {
	return v.videoFilter("vflip")
}

// FadeIn applies a fade-in effect for the given duration (seconds) at the start.
func (v *Video) FadeIn(duration float64) (*Video, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("FadeIn: duration must be positive (got=%f, file=%s, label=%s)", duration, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("fade=t=in:st=0:d=%.4f", duration))
}

// FadeOut applies a fade-out effect for the given duration (seconds) at the end.
func (v *Video) FadeOut(duration float64) (*Video, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("FadeOut: duration must be positive (got=%f, file=%s, label=%s)", duration, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	if duration > v.duration {
		return nil, fmt.Errorf("FadeOut: duration %f exceeds video duration %f (file=%s, label=%s)", duration, v.duration, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	startTime := v.duration - duration
	return v.videoFilter(fmt.Sprintf("fade=t=out:st=%.4f:d=%.4f", startTime, duration))
}

// Eq applies brightness, contrast, saturation, and gamma adjustments.
func (v *Video) Eq(params EqParams) (*Video, error) {
	file, label := safeFirstFilename(v.filenames), safeLastVideoLabel(v)
	if params.Brightness < -1 || params.Brightness > 1 {
		return nil, fmt.Errorf("Eq: brightness must be -1 to 1 (got=%f, file=%s, label=%s)", params.Brightness, file, label)
	}
	if params.Contrast < -1000 || params.Contrast > 1000 {
		return nil, fmt.Errorf("Eq: contrast must be -1000 to 1000 (got=%f, file=%s, label=%s)", params.Contrast, file, label)
	}
	if params.Saturation < 0 || params.Saturation > 3 {
		return nil, fmt.Errorf("Eq: saturation must be 0-3 (got=%f, file=%s, label=%s)", params.Saturation, file, label)
	}
	if params.Gamma < 0.1 || params.Gamma > 10 {
		return nil, fmt.Errorf("Eq: gamma must be 0.1 to 10 (got=%f, file=%s, label=%s)", params.Gamma, file, label)
	}
	parts := []string{
		fmt.Sprintf("brightness=%.4f", params.Brightness),
		fmt.Sprintf("contrast=%.4f", params.Contrast),
		fmt.Sprintf("saturation=%.4f", params.Saturation),
		fmt.Sprintf("gamma=%.4f", params.Gamma),
	}
	return v.videoFilter("eq=" + strings.Join(parts, ":"))
}

// Hue adjusts hue and saturation.
func (v *Video) Hue(params HueParams) (*Video, error) {
	if params.Saturation < -10 || params.Saturation > 10 {
		return nil, fmt.Errorf("Hue: saturation must be -10 to 10 (got=%f, file=%s, label=%s)", params.Saturation, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	filter := fmt.Sprintf("hue=h=%.4f", params.Degrees)
	if params.Saturation != 1 {
		filter += fmt.Sprintf(":s=%.4f", params.Saturation)
	}
	return v.videoFilter(filter)
}

// Scale resizes the video. Use -1 for width or height to preserve aspect ratio.
func (v *Video) Scale(params ScaleParams) (*Video, error) {
	file, label := safeFirstFilename(v.filenames), safeLastVideoLabel(v)
	if params.Width < -1 || params.Height < -1 {
		return nil, fmt.Errorf("Scale: width and height must be >= -1 (width=%d, height=%d, file=%s, label=%s)", params.Width, params.Height, file, label)
	}
	if params.Width == 0 && params.Height == 0 {
		return nil, fmt.Errorf("Scale: at least one of width or height must be non-zero (width=%d, height=%d, file=%s, label=%s)", params.Width, params.Height, file, label)
	}
	widthStr := fmt.Sprintf("%d", params.Width)
	heightStr := fmt.Sprintf("%d", params.Height)
	if params.Width == 0 {
		widthStr = "iw"
	}
	if params.Height == 0 {
		heightStr = "ih"
	}
	targetWidth, targetHeight, err := resolveScaledDimensions(v.width, v.height, params)
	if err != nil {
		return nil, fmt.Errorf("Scale[file=%s, label=%s]: %w", file, label, err)
	}
	// -1 preserves aspect ratio based on the other dimension
	scaled, err := v.videoFilter(fmt.Sprintf("scale=w=%s:h=%s", widthStr, heightStr))
	if err != nil {
		return nil, fmt.Errorf("Scale[file=%s, label=%s]: %w", file, label, err)
	}
	scaled.width = targetWidth
	scaled.height = targetHeight
	return scaled, nil
}

// Crop crops the video to the specified region.
func (v *Video) Crop(params CropParams) (*Video, error) {
	file, label := safeFirstFilename(v.filenames), safeLastVideoLabel(v)
	if params.Width <= 0 || params.Height <= 0 {
		return nil, fmt.Errorf("Crop: width and height must be positive (width=%d, height=%d, file=%s, label=%s)", params.Width, params.Height, file, label)
	}
	if params.X < 0 || params.Y < 0 {
		return nil, fmt.Errorf("Crop: x and y must be non-negative (x=%d, y=%d, file=%s, label=%s)", params.X, params.Y, file, label)
	}
	cropped, err := v.videoFilter(fmt.Sprintf("crop=w=%d:h=%d:x=%d:y=%d",
		params.Width, params.Height, params.X, params.Y))
	if err != nil {
		return nil, fmt.Errorf("Crop[file=%s, label=%s]: %w", file, label, err)
	}
	cropped.width = uint64(evenDimension(params.Width))
	cropped.height = uint64(evenDimension(params.Height))
	return cropped, nil
}

// Pad adds padding to the video.
func (v *Video) Pad(params PadParams) (*Video, error) {
	file, label := safeFirstFilename(v.filenames), safeLastVideoLabel(v)
	if params.Width < 0 || params.Height < 0 {
		return nil, fmt.Errorf("Pad: width and height must be non-negative (width=%d, height=%d, file=%s, label=%s)", params.Width, params.Height, file, label)
	}
	if params.X < 0 || params.Y < 0 {
		return nil, fmt.Errorf("Pad: x and y must be non-negative (x=%d, y=%d, file=%s, label=%s)", params.X, params.Y, file, label)
	}
	padded, err := v.videoFilter(fmt.Sprintf("pad=width=%d:height=%d:x=%d:y=%d",
		params.Width, params.Height, params.X, params.Y))
	if err != nil {
		return nil, fmt.Errorf("Pad[file=%s, label=%s]: %w", file, label, err)
	}
	if params.Width > 0 {
		padded.width = uint64(evenDimension(params.Width))
	}
	if params.Height > 0 {
		padded.height = uint64(evenDimension(params.Height))
	}
	return padded, nil
}

// Blur applies a Gaussian blur. Sigma controls blur strength (higher = more blur).
func (v *Video) Blur(sigma float64) (*Video, error) {
	if sigma <= 0 {
		return nil, fmt.Errorf("Blur: sigma must be positive (got=%f, file=%s, label=%s)", sigma, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("gblur=sigma=%.4f", sigma))
}

// Sharpen applies an unsharp mask. Amount controls sharpening strength (1.0-5.0 typical).
func (v *Video) Sharpen(amount float64) (*Video, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("Sharpen: amount must be positive (got=%f, file=%s, label=%s)", amount, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("unsharp=5:5:%.4f:5:5:0", amount))
}

// Grayscale converts the video to grayscale.
func (v *Video) Grayscale() (*Video, error) {
	return v.videoFilter("hue=s=0")
}

// Sepia applies a sepia tone effect.
func (v *Video) Sepia() (*Video, error) {
	return v.videoFilter("colorchannelmixer=.393:.769:.189:0:.349:.686:.168:0:.272:.534:.131:0")
}

// Vignette applies a vignette (darkened corners) effect. Angle controls darkness (radians, PI/5 typical).
func (v *Video) Vignette(angle float64) (*Video, error) {
	if angle <= 0 {
		return nil, fmt.Errorf("Vignette: angle must be positive (got=%f, file=%s, label=%s)", angle, safeFirstFilename(v.filenames), safeLastVideoLabel(v))
	}
	return v.videoFilter(fmt.Sprintf("vignette=angle=%.4f", angle))
}

// Negate inverts the colors of the video.
func (v *Video) Negate() (*Video, error) {
	return v.videoFilter("negate")
}
