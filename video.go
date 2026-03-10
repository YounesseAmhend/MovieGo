package moviego

import (
	"fmt"
	"runtime"
)

var globalLabelCounter uint64
var globalOrderCounter uint64

type FileCopy struct {
	Filename string
	Label    string
}

type FilterComplex struct {
	FilterElement string
	FileCopy       FileCopy
	Label          string
	Order          uint64
}

// Position defines where a video is placed when used in a CompositeClip.
// X and Y are FFmpeg expressions (e.g. "(W-w)/2", "0", "100").
type Position struct {
	X string
	Y string
}

// CenterPosition returns a position that centers the overlay on the background.
func CenterPosition() Position {
	return Position{X: "(W-w)/2", Y: "(H-h)/2"}
}

// TopLeftPosition returns a position at the top-left corner.
func TopLeftPosition() Position {
	return Position{X: "0", Y: "0"}
}

// Video represents a video file with its properties and processing options
type Video struct {
	filenames          []string
	codec              Codec
	width              uint64
	labelCounter       uint64
	height             uint64
	fps                uint64
	duration           float64
	frames             uint64
	ffmpegArgs         map[string][]string
	videoFilterComplex []FilterComplex
	audioFilterComplex []FilterComplex
	isTemp             bool
	audio              Audio
	bitRate            string
	preset             preset
	withMask           bool
	pixelFormat        PixelFormat
	startTime          float64
	endTime            float64
	position           Position
	animatedPosition   *AnimatedPosition // nil = use static position
	animatedOpacity    *Animation         // nil = fully opaque
}

// ============================================================================
// Video Property Getters and Setters
// ============================================================================

// SetFrames sets the total number of frames in the video
func (v *Video) SetFrames(frames uint64) *Video {
	v.frames = frames
	return v
}

// GetFrames returns the total number of frames in the video
func (v *Video) GetFrames() int64 {
	return int64(v.frames)
}

// SetFilename sets the video filename
func (v *Video) SetFilename(filenames []string) *Video {
	v.filenames = filenames
	return v
}

// GetFilenames returns the video filename
func (v *Video) GetFilenames() []string {
	return v.filenames
}

// Codec sets the video codec
func (v *Video) Codec(codec Codec) *Video {
	v.codec = codec
	return v
}

// GetCodec returns the video codec
func (v *Video) GetCodec() string {
	return string(v.codec)
}

// Width sets the video width
func (v *Video) Width(width uint64) *Video {
	v.width = width
	return v
}

// GetWidth returns the video width
func (v *Video) GetWidth() uint64 {
	return v.width
}

// Height sets the video height
func (v *Video) Height(height uint64) *Video {
	v.height = height
	return v
}

// GetHeight returns the video height
func (v *Video) GetHeight() uint64 {
	return v.height
}

// Duration sets the video duration
func (v *Video) Duration(duration float64) *Video {
	v.duration = duration
	return v
}

// GetDuration returns the video duration
func (v *Video) GetDuration() float64 {
	return v.duration
}

// SetFps sets the video frames per second
func (v *Video) SetFps(fps uint64) *Video {
	v.fps = fps
	return v
}

// GetFps returns the video frames per second
func (v *Video) GetFps() uint64 {
	return v.fps
}

// BitRate sets the video bitrate
func (v *Video) BitRate(bitRate string) *Video {
	v.bitRate = bitRate
	return v
}

// GetBitRate returns the video bitrate
func (v *Video) GetBitRate() string {
	return v.bitRate
}

// Preset sets the video preset
func (v *Video) Preset(p preset) *Video {
	v.preset = p
	return v
}

// GetPreset returns the video preset
func (v *Video) GetPreset() preset {
	return v.preset
}

// WithMask sets whether the video has a mask
func (v *Video) WithMask(withMask bool) *Video {
	v.withMask = withMask
	return v
}

// GetWithMask returns whether the video has a mask
func (v *Video) GetWithMask() bool {
	return v.withMask
}

// PixelFormat sets the video pixel format
func (v *Video) PixelFormat(pixelFormat PixelFormat) *Video {
	v.pixelFormat = pixelFormat
	return v
}

// GetPixelFormat returns the video pixel format
func (v *Video) GetPixelFormat() PixelFormat {
	return v.pixelFormat
}

// SetStartTime sets the start time for subclip
func (v *Video) SetStartTime(startTime float64) *Video {
	v.startTime = startTime
	return v
}

// GetStartTime returns the start time for subclip
func (v *Video) GetStartTime() float64 {
	return v.startTime
}

// SetEndTime sets the end time for subclip
func (v *Video) SetEndTime(endTime float64) *Video {
	v.endTime = endTime
	return v
}

// GetEndTime returns the end time for subclip
func (v *Video) GetEndTime() float64 {
	return v.endTime
}

// SetPosition sets the overlay position used by CompositeClip.
func (v *Video) SetPosition(position Position) *Video {
	v.position = position
	return v
}

// GetPosition returns the overlay position.
// Returns center position if none was explicitly set.
func (v *Video) GetPosition() Position {
	if v.position.X == "" && v.position.Y == "" {
		return CenterPosition()
	}
	return v.position
}

// SetAnimatedPosition sets the overlay position animation for CompositeClip.
func (v *Video) SetAnimatedPosition(ap AnimatedPosition) *Video {
	v.animatedPosition = &ap
	return v
}

// SetAnimatedOpacity sets the overlay opacity animation for CompositeClip.
func (v *Video) SetAnimatedOpacity(a Animation) *Video {
	v.animatedOpacity = &a
	return v
}

// ============================================================================
// FFmpeg Configuration
// ============================================================================

// FfmpegArgs sets custom FFmpeg arguments
func (v *Video) FfmpegArgs(ffmpegArgs map[string][]string) *Video {
	v.ffmpegArgs = ffmpegArgs
	return v
}

// GetFfmpegArgs returns the custom FFmpeg arguments
func (v *Video) GetFfmpegArgs() map[string][]string {
	return v.ffmpegArgs
}

// ============================================================================
// Audio Properties
// ============================================================================

// SetAudio sets the audio configuration
func (v *Video) SetAudio(audio Audio) *Video {
	v.audio = audio
	return v
}

// GetAudio returns the audio configuration
func (v *Video) GetAudio() *Audio {
	return &v.audio
}

// ============================================================================
// Temporary File Management
// ============================================================================

// SetIsTemp marks whether this video is a temporary file
func (v *Video) SetIsTemp(isTemp bool) *Video {
	v.isTemp = isTemp
	return v
}

// GetIsTemp returns whether this video is a temporary file
func (v *Video) GetIsTemp() bool {
	return v.isTemp
}

// HasAudio returns whether the video has an audio stream
func (v *Video) HasAudio() bool {
	return v.audio.codec != ""
}

// ============================================================================
// Internal Helpers
// ============================================================================

func (v *Video) lastAudioLabel() string {
	return v.audioFilterComplex[len(v.audioFilterComplex)-1].Label
}

func (v *Video) lastVideoLabel() string {
	return v.videoFilterComplex[len(v.videoFilterComplex)-1].Label
}

func (v *Video) applyParameters(parms VideoParameters) *Video {
	if parms.Codec != "" {
		v.Codec(parms.Codec)
	}
	if parms.Fps != 0 {
		v.SetFps(parms.Fps)
	}
	if parms.Preset != "" {
		v.Preset(parms.Preset)
	}
	if parms.WithMask {
		v.WithMask(parms.WithMask)
	}
	if parms.PixelFormat != "" {
		v.PixelFormat(parms.PixelFormat)
	}
	if parms.Bitrate != "" {
		v.BitRate(parms.Bitrate)
	}
	if parms.Threads == 0 {
		// Optimize thread allocation: reserve CPUs for Go workers
		// FFmpeg gets 60% of CPUs, Go workers use 40%
		totalCPUs := runtime.GOMAXPROCS(0)
		if totalCPUs <= 2 {
			parms.Threads = uint16(totalCPUs)
		} else {
			parms.Threads = uint16((totalCPUs * 6) / 10)
			if parms.Threads < 2 {
				parms.Threads = 2
			}
		}
	}

	return v
}

func (v *Video) nextLabel(filename string) string {
	id := incrementGlobalCounter()
	safeName := sanitize(filename)
	// Result: [1_hello_world_v] and [2_hello_world_v]
	return fmt.Sprintf("%d_%s", id, safeName)
}



func (v *Video) lastFilename() string {
	return v.filenames[len(v.filenames)-1]
}
