package moviego

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// GPU vendor types
type gpuVendor string

const (
	gpuNvidia  gpuVendor = "nvidia"
	gpuAMD     gpuVendor = "amd"
	gpuIntel   gpuVendor = "intel"
	gpuApple   gpuVendor = "apple"
	gpuUnknown gpuVendor = "unknown"
)

// Cached codec detection result
var (
	cachedCodec     string
	cachedCodecOnce sync.Once
)

// selectBestH264Codec detects the best available H.264 codec for the system
// Priority: h264_nvenc (NVIDIA) > h264_qsv (Intel) > h264_amf (AMD) > h264_videotoolbox (Apple) > libx264 (software)
func selectBestH264Codec() string {
	cachedCodecOnce.Do(func() {
		cachedCodec = detectBestH264Codec()
	})
	return cachedCodec
}

// detectBestH264Codec performs the actual codec detection
func detectBestH264Codec() string {
	// Get available FFmpeg encoders
	availableEncoders := getAvailableEncoders()
	if len(availableEncoders) == 0 {
		// If we can't get encoder list, fallback to software
		return "libx264"
	}

	// Detect GPU vendor
	vendor := detectGPUVendor()

	// Try to match hardware encoder based on GPU vendor and availability
	switch vendor {
	case gpuNvidia:
		if isEncoderAvailable("h264_nvenc", availableEncoders) {
			fmt.Println("✓ Using h264_nvenc (NVIDIA GPU detected)")
			return "h264_nvenc"
		}
	case gpuIntel:
		if isEncoderAvailable("h264_qsv", availableEncoders) {
			fmt.Println("✓ Using h264_qsv (Intel Quick Sync detected)")
			return "h264_qsv"
		}
	case gpuAMD:
		if isEncoderAvailable("h264_amf", availableEncoders) {
			fmt.Println("✓ Using h264_amf (AMD GPU detected)")
			return "h264_amf"
		}
	case gpuApple:
		if isEncoderAvailable("h264_videotoolbox", availableEncoders) {
			fmt.Println("✓ Using h264_videotoolbox (Apple VideoToolbox detected)")
			return "h264_videotoolbox"
		}
	}

	// If primary vendor encoder not available, try others in priority order
	priorityEncoders := []string{"h264_nvenc", "h264_qsv", "h264_amf", "h264_videotoolbox"}
	for _, encoder := range priorityEncoders {
		if isEncoderAvailable(encoder, availableEncoders) {
			fmt.Printf("✓ Using %s (hardware acceleration available)\n", encoder)
			return encoder
		}
	}

	// Fallback to software encoding
	fmt.Println("→ Using libx264 (software encoding - no hardware acceleration detected)")
	return "libx264"
}

// detectGPUVendor detects the GPU vendor on the system
func detectGPUVendor() gpuVendor {
	os := runtime.GOOS

	switch os {
	case "windows":
		return detectGPUWindows()
	case "linux":
		return detectGPULinux()
	case "darwin":
		return detectGPUMacOS()
	default:
		return gpuUnknown
	}
}

// detectGPUWindows detects GPU vendor on Windows using wmic
func detectGPUWindows() gpuVendor {
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "name")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return gpuUnknown
	}

	outputStr := strings.ToLower(string(output))

	// Check in priority order
	if strings.Contains(outputStr, "nvidia") {
		return gpuNvidia
	}
	if strings.Contains(outputStr, "intel") {
		return gpuIntel
	}
	if strings.Contains(outputStr, "amd") || strings.Contains(outputStr, "radeon") {
		return gpuAMD
	}

	return gpuUnknown
}

// detectGPULinux detects GPU vendor on Linux using multiple methods
func detectGPULinux() gpuVendor {
	// Method 1: Check NVIDIA driver
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return gpuNvidia
	}

	// Method 2: Check /proc/driver/nvidia
	cmd := exec.Command("cat", "/proc/driver/nvidia/version")
	if err := cmd.Run(); err == nil {
		return gpuNvidia
	}

	// Method 3: Use lspci
	cmd = exec.Command("lspci")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return gpuUnknown
	}

	outputStr := strings.ToLower(string(output))

	// Look for VGA/3D/Display controllers
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "vga") || strings.Contains(line, "3d") || strings.Contains(line, "display") {
			if strings.Contains(line, "nvidia") {
				return gpuNvidia
			}
			if strings.Contains(line, "intel") {
				return gpuIntel
			}
			if strings.Contains(line, "amd") || strings.Contains(line, "radeon") {
				return gpuAMD
			}
		}
	}

	// Method 4: Check /sys/class/drm
	cmd = exec.Command("ls", "/sys/class/drm")
	output, err = cmd.CombinedOutput()
	if err == nil {
		outputStr := strings.ToLower(string(output))
		if strings.Contains(outputStr, "nvidia") {
			return gpuNvidia
		}
	}

	return gpuUnknown
}

// detectGPUMacOS detects GPU vendor on macOS
func detectGPUMacOS() gpuVendor {
	// macOS primarily uses Apple's VideoToolbox
	// But we can still detect discrete GPUs
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If command fails, assume Apple silicon or integrated
		return gpuApple
	}

	outputStr := strings.ToLower(string(output))

	// Check for discrete GPUs first
	if strings.Contains(outputStr, "nvidia") {
		return gpuNvidia
	}
	if strings.Contains(outputStr, "amd") || strings.Contains(outputStr, "radeon") {
		return gpuAMD
	}
	if strings.Contains(outputStr, "intel") {
		return gpuIntel
	}

	// Default to Apple (for VideoToolbox on Apple Silicon or Intel Macs)
	return gpuApple
}

// getAvailableEncoders queries FFmpeg for available encoders
func getAvailableEncoders() map[string]bool {
	cmd := exec.Command("ffmpeg", "-hide_banner", "-encoders")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return make(map[string]bool)
	}

	encoders := make(map[string]bool)
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Encoder lines start with "V" for video
		if strings.HasPrefix(line, "V") {
			// Format: "V..... encodername    Description"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				encoderName := parts[1]
				encoders[encoderName] = true
			}
		}
	}

	return encoders
}

// isEncoderAvailable checks if a specific encoder is available
func isEncoderAvailable(encoderName string, availableEncoders map[string]bool) bool {
	return availableEncoders[encoderName]
}


func resolveCodec(preferredCodec, fallbackCodec string) string {
	if preferredCodec != "" {
		return preferredCodec
	}
	if fallbackCodec != "" {
		return fallbackCodec
	}
	return selectBestH264Codec()
}

func mapPresetForCodec(codec string, presetValue string) string {
	// If no preset specified, return empty to use encoder defaults
	if presetValue == "" {
		return ""
	}

	switch codec {
	case "h264_nvenc", "hevc_nvenc", "av1_nvenc":
		// NVIDIA NVENC presets
		switch presetValue {
		case string(UltraFast), string(SuperFast), string(VeryFast):
			return string(presetNvencFast)
		case string(Fast):
			return string(presetNvencFast)
		case string(Medium):
			return string(presetNvencMedium)
		case string(Slow), string(VerySlow):
			return string(presetNvencSlow)
		case string(Placebo):
			return string(presetNvencHQ)
		default:
			return string(presetNvencMedium)
		}

	case "h264_amf", "hevc_amf":
		// AMD AMF presets
		switch presetValue {
		case string(UltraFast), string(SuperFast), string(VeryFast), string(Fast):
			return string(presetAmfSpeed)
		case string(Medium):
			return string(presetAmfBalanced)
		case string(Slow), string(VerySlow), string(Placebo):
			return string(presetAmfQuality)
		default:
			return string(presetAmfBalanced)
		}

	case "h264_qsv", "hevc_qsv", "av1_qsv":
		// Intel QSV presets (mostly same as software)
		switch presetValue {
		case string(UltraFast), string(SuperFast):
			return string(presetQsvVeryFast)
		case string(VeryFast):
			return string(presetQsvVeryFast)
		case string(Fast):
			return string(presetQsvFast)
		case string(Medium):
			return string(presetQsvMedium)
		case string(Slow):
			return string(presetQsvSlow)
		case string(VerySlow):
			return string(presetQsvVerySlow)
		case string(Placebo):
			return string(presetQsvVerySlow)
		default:
			return string(presetQsvMedium)
		}

	case "h264_videotoolbox", "hevc_videotoolbox":
		// VideoToolbox doesn't use preset parameter
		return ""

	default:
		// Software encoders (libx264, libx265, etc.) - return as-is
		return presetValue
	}
}

// resolvePreset resolves preset with fallback: preferredPreset → fallbackPreset → Medium
// Returns the preset string value, or empty string if no preset should be used.
func resolvePreset(preferredPreset, fallbackPreset preset) string {
	if preferredPreset != "" {
		return string(preferredPreset)
	}
	if fallbackPreset != "" {
		return string(fallbackPreset)
	}
	return string(Medium)
}

// resolveBitrate resolves bitrate with fallback: preferredBitrate → fallbackBitrate → empty
// Returns the bitrate string value, or empty string if no bitrate is specified.
func resolveBitrate(preferredBitrate, fallbackBitrate string) string {
	if preferredBitrate != "" {
		return preferredBitrate
	}
	return fallbackBitrate
}

// resolveFps resolves FPS with fallback: preferredFps → fallbackFps → 0
// Returns the FPS value, or 0 if no FPS is specified.
func resolveFps(preferredFps, fallbackFps uint64) uint64 {
	if preferredFps != 0 {
		return preferredFps
	}
	return fallbackFps
}
