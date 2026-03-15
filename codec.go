package moviego

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

type AudioCodec string

const (
	AudioCodecAAC    AudioCodec = "aac"
	AudioCodecMP3    AudioCodec = "libmp3lame"
	AudioCodecFLAC   AudioCodec = "flac"
	AudioCodecOpus   AudioCodec = "libopus"
	AudioCodecVorbis AudioCodec = "libvorbis"
	AudioCodecPCM    AudioCodec = "pcm_s16le"
)

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

type PixelFormat string

const (
	PixelFormatRGBA    PixelFormat = "rgba"
	PixelFormatRGB     PixelFormat = "rgb"
	PixelFormatYUV420P PixelFormat = "yuv420p"
	PixelFormatYUVA420P PixelFormat = "yuva420p"
	PixelFormatYUV422P PixelFormat = "yuv422p"
	PixelFormatYUV444P PixelFormat = "yuv444p"
)

// Progress holds real-time encoding progress reported by FFmpeg.
type Progress struct {
	// Percentage of encoding completed (0.0 – 100.0).
	Percentage float64
	// Current output timestamp in seconds.
	OutTime float64
	// Total expected duration in seconds.
	TotalDuration float64
	// Encoding speed relative to real-time (e.g. 1.5 means 1.5× real-time).
	Speed float64
	// Current encoding bitrate string (e.g. "1024.5kbits/s").
	Bitrate string
	// Current frame number being encoded.
	Frame int64
	// Frames per second the encoder is running at.
	FPS float64
	// Whether encoding has finished.
	Done bool
}

// VideoParameters holds configuration for video processing.
//
// By default every encode shows a colored progress bar on stderr.
// Set SilentProgress to true to suppress it, or set OnProgress to
// replace the built-in output with your own handler.
type VideoParameters struct {
	OutputPath  string
	Threads     uint16
	Codec       Codec
	Fps         uint64
	Preset      preset
	WithMask    bool
	Bitrate     string
	PixelFormat PixelFormat
	// SilentProgress disables the default colored progress bar.
	// Has no effect when OnProgress is set.
	SilentProgress bool
	// OnProgress, when set, replaces the default colored progress bar.
	// Called periodically with encoding progress.
	OnProgress func(Progress)
}

// AudioParameters holds configuration for audio processing.
type AudioParameters struct {
	OutputPath string
	Threads    uint16
	Codec      AudioCodec
	SampleRate uint64
	Channels   uint8
	Bitrate    uint64
	// SilentProgress disables the default colored progress bar.
	SilentProgress bool
	// OnProgress, when set, replaces the default colored progress bar.
	OnProgress func(Progress)
}

