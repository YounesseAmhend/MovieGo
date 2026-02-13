package moviego

import (
	"crypto/sha256"
	"fmt"
)

// Video represents a video file with its properties and processing options
type Video struct {
	filename      string
	codec         Codec
	width         uint64
	height        uint64
	fps           uint64
	duration      float64
	hash          string
	frames        uint64
	ffmpegArgs    map[string][]string
	filters       []Filter
	customFilters []func([]byte, int)
	isTemp        bool
	audio         Audio
	bitRate       string
	preset        preset
	withMask      bool
	pixelFormat   PixelFormat
	startTime     float64
	endTime       float64
	textClips     []*TextClip
	subtitleClips []*SubtitleClip
}

// VideoParameters holds configuration for video processing

type preset string

const (
	// Software encoder presets (libx264, libx265, etc.)
	UltraFast preset = "ultrafast"
	SuperFast preset = "superfast"
	VeryFast  preset = "veryfast"
	Fast      preset = "fast"
	Medium    preset = "medium"
	Slow      preset = "slow"
	VerySlow  preset = "veryslow"
	Placebo   preset = "placebo"

	// NVIDIA NVENC presets (internal use only)
	presetNvencFast   preset = "fast"
	presetNvencMedium preset = "medium"
	presetNvencSlow   preset = "slow"
	presetNvencHQ     preset = "hq"

	// AMD AMF presets (internal use only)
	presetAmfSpeed    preset = "speed"
	presetAmfBalanced preset = "balanced"
	presetAmfQuality  preset = "quality"

	// Intel QSV presets (internal use only)
	presetQsvVeryFast preset = "veryfast"
	presetQsvFast     preset = "fast"
	presetQsvMedium   preset = "medium"
	presetQsvSlow     preset = "slow"
	presetQsvVerySlow preset = "veryslow"
)

type VideoParameters struct {
	OutputPath  string
	Threads     uint16
	Codec       Codec
	Fps         uint64
	Preset      preset
	WithMask    bool
	Bitrate     string
	PixelFormat PixelFormat
}

type PixelFormat string

const (
	PixelFormatRGBA    PixelFormat = "rgba"
	PixelFormatRGB     PixelFormat = "rgb"
	PixelFormatYUV420P PixelFormat = "yuv420p"
	PixelFormatYUV422P PixelFormat = "yuv422p"
	PixelFormatYUV444P PixelFormat = "yuv444p"
)

type Codec string

const (
	// H.264/AVC codecs
	CodecH264      Codec = "h264"
	CodecLibx264   Codec = "libx264"
	CodecH264Auto  Codec = "h264_auto"
	CodecH264Nvenc Codec = "h264_nvenc"
	CodecH264Qsv   Codec = "h264_qsv"
	CodecH264Amf   Codec = "h264_amf"
	CodecH264Vt    Codec = "h264_videotoolbox"

	// H.265/HEVC codecs
	CodecH265      Codec = "h265"
	CodecHevc      Codec = "hevc"
	CodecLibx265   Codec = "libx265"
	CodecHevcNvenc Codec = "hevc_nvenc"
	CodecHevcQsv   Codec = "hevc_qsv"
	CodecHevcAmf   Codec = "hevc_amf"
	CodecHevcVt    Codec = "hevc_videotoolbox"

	// VP8/VP9 codecs
	CodecVP8       Codec = "vp8"
	CodecVP9       Codec = "vp9"
	CodecLibvpx    Codec = "libvpx"
	CodecLibvpxVP9 Codec = "libvpx-vp9"

	// AV1 codecs
	CodecAV1       Codec = "av1"
	CodecLibaomAV1 Codec = "libaom-av1"
	CodecLibsvtav1 Codec = "libsvtav1"
	CodecAV1Nvenc  Codec = "av1_nvenc"
	CodecAV1Qsv    Codec = "av1_qsv"

	// MPEG codecs
	CodecMpeg2video Codec = "mpeg2video"
	CodecMpeg4      Codec = "mpeg4"
	CodecMpeg1video Codec = "mpeg1video"

	// Other codecs
	CodecTheora   Codec = "theora"
	CodecWmv1     Codec = "wmv1"
	CodecWmv2     Codec = "wmv2"
	CodecWmv3     Codec = "wmv3"
	CodecVc1      Codec = "vc1"
	CodecProres   Codec = "prores"
	CodecProresKS Codec = "prores_ks"
	CodecDNxHD    Codec = "dnxhd"
	CodecDNxHR    Codec = "dnxhr"
	CodecHuffYUV  Codec = "huffyuv"
	CodecFFV1     Codec = "ffv1"
	CodecUtvideo  Codec = "utvideo"
	CodecMjpeg    Codec = "mjpeg"
	CodecLibxvid  Codec = "libxvid"
)

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
func (v *Video) SetFilename(filename string) *Video {
	v.filename = filename
	return v
}

// GetFilename returns the video filename
func (v *Video) GetFilename() string {
	return v.filename
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

func (v *Video) GetHash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v.filename)))
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
// Filter Configuration
// ============================================================================

// AddFilter adds a built-in filter to the video processing pipeline
func (v *Video) AddFilter(filter Filter) *Video {
	v.filters = append(v.filters, filter)
	return v
}

// AddCustomFilter adds a custom filter function to the video processing pipeline
func (v *Video) AddCustomFilter(filterFunc func([]byte, int)) *Video {
	v.customFilters = append(v.customFilters, filterFunc)
	return v
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

// Subclip creates a new video segment with specified start and end times (lazy operation)
// Parameters:
//   - start: Start time in seconds (must be >= 0)
//   - end: End time in seconds (must be > start and <= video duration)
//
// Returns a new Video object with updated metadata (no file is created until WriteVideo is called)
func (v *Video) Subclip(start, end float64) *Video {
	// Validate inputs
	if start < 0 {
		start = 0
	}
	if end > v.duration {
		end = v.duration
	}
	if start >= end {
		// Return empty video if invalid range
		return &Video{}
	}

	// Create a new video with copied properties
	newVideo := &Video{
		filename:      v.filename,
		codec:         v.codec,
		width:         v.width,
		height:        v.height,
		fps:           v.fps,
		duration:      end - start,
		frames:        uint64(float64(v.fps) * (end - start)),
		ffmpegArgs:    v.ffmpegArgs,
		filters:       v.filters,
		customFilters: v.customFilters,
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

	return newVideo
}

// AddText adds a text overlay to the video
func (v *Video) AddText(textClip *TextClip) *Video {
	v.textClips = append(v.textClips, textClip)
	return v
}

// AddSubtitle adds a subtitle file to the video and returns the SubtitleClip for further configuration
func (v *Video) AddSubtitle(subtitlePath string) *SubtitleClip {
	subtitleClip := NewSubtitleClip(subtitlePath)
	v.subtitleClips = append(v.subtitleClips, subtitleClip)
	return subtitleClip
}

// AddSubtitleClip adds a pre-configured SubtitleClip to the video
func (v *Video) AddSubtitleClip(subtitleClip *SubtitleClip) *Video {
	v.subtitleClips = append(v.subtitleClips, subtitleClip)
	return v
}

// RemoveText removes a text overlay at the specified index
func (v *Video) RemoveText(index int) *Video {
	if index >= 0 && index < len(v.textClips) {
		v.textClips = append(v.textClips[:index], v.textClips[index+1:]...)
	}
	return v
}

// RemoveSubtitle removes a subtitle clip at the specified index
func (v *Video) RemoveSubtitle(index int) *Video {
	if index >= 0 && index < len(v.subtitleClips) {
		v.subtitleClips = append(v.subtitleClips[:index], v.subtitleClips[index+1:]...)
	}
	return v
}

// ClearText removes all text overlays
func (v *Video) ClearText() *Video {
	v.textClips = []*TextClip{}
	return v
}

// ClearSubtitles removes all subtitle clips
func (v *Video) ClearSubtitles() *Video {
	v.subtitleClips = []*SubtitleClip{}
	return v
}

// GetTextClips returns all text overlays
func (v *Video) GetTextClips() []*TextClip {
	return v.textClips
}

// GetSubtitleClips returns all subtitle clips
func (v *Video) GetSubtitleClips() []*SubtitleClip {
	return v.subtitleClips
}

// HasText returns whether the video has any text overlays
func (v *Video) HasText() bool {
	return len(v.textClips) > 0
}

// HasSubtitles returns whether the video has any subtitle clips
func (v *Video) HasSubtitles() bool {
	return len(v.subtitleClips) > 0
}
