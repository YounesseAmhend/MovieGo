package moviego

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// FFmpegProgressConfig holds configuration for FFmpeg progress tracking
type FFmpegProgressConfig struct {
	Args          []string
	TotalDuration float64 // Total duration in seconds for progress calculation
	TotalFrames   int64   // Optional: total frames if known (0 if unknown)
	OperationName string  // e.g., "Processing video", "Combining audio"
	OutputPath    string  // For display in progress bar suffix
	Bitrate       string  // Optional: bitrate for display
}

// parseTime parses FFmpeg time format (HH:MM:SS.mmm) to seconds
func parseTime(timeStr string) (float64, error) {
	// Format: HH:MM:SS.mmm or HH:MM:SS
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, err
	}

	return hours*3600 + minutes*60 + seconds, nil
}

// parseFFmpegProgressLine extracts progress information from FFmpeg stderr line
// Format: frame=  123 fps= 45 q=28.0 size=    1024kB time=00:00:05.12 bitrate=1638.4kbits/s speed=1.88x
func parseFFmpegProgressLine(line string) (frame int64, currentTime float64, speed float64, ok bool) {
	// Extract frame number
	frameRegex := regexp.MustCompile(`frame=\s*(\d+)`)
	if matches := frameRegex.FindStringSubmatch(line); len(matches) > 1 {
		if f, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
			frame = f
		}
	}

	// Extract time
	timeRegex := regexp.MustCompile(`time=(\d{2}:\d{2}:\d{2}\.?\d*)`)
	if matches := timeRegex.FindStringSubmatch(line); len(matches) > 1 {
		if t, err := parseTime(matches[1]); err == nil {
			currentTime = t
		}
	}

	// Extract speed (e.g., "speed=1.88x" or "speed=0.5x")
	speedRegex := regexp.MustCompile(`speed=\s*([\d.]+)x`)
	if matches := speedRegex.FindStringSubmatch(line); len(matches) > 1 {
		if s, err := strconv.ParseFloat(matches[1], 64); err == nil {
			speed = s
		}
	}

	// Return true if we found at least time or frame
	ok = currentTime > 0 || frame > 0
	return
}

// runFFmpegWithProgress runs an FFmpeg command with progress bar tracking
func runFFmpegWithProgress(config FFmpegProgressConfig) error {
	if len(config.Args) == 0 {
		return fmt.Errorf("no FFmpeg arguments provided")
	}

	// Modify args to allow progress output
	// Replace "-loglevel error" with "-loglevel info" to see progress lines
	args := make([]string, len(config.Args))
	copy(args, config.Args)
	for i, arg := range args {
		if arg == "-loglevel" && i+1 < len(args) && args[i+1] == "error" {
			args[i+1] = "info" // Change to info to show progress output
			break
		}
	}

	// Build command
	cmd := exec.Command("ffmpeg", args...)

	// Capture stderr for progress parsing and error output
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg command: %w", err)
	}

	// Create progress bar
	var bar *pb.ProgressBar
	if config.TotalFrames > 0 {
		// Use frames if available (more accurate)
		bar = pb.New64(config.TotalFrames)
	} else if config.TotalDuration > 0 {
		// Use duration converted to centiseconds for granularity
		bar = pb.New64(int64(config.TotalDuration * 100))
	} else {
		// Fallback: use a large number and update based on progress
		bar = pb.New64(1000)
	}

	// Set progress bar template (same as frame_processor.go)
	bar.SetTemplateString(`{{string . "prefix"}}{{counters . }} {{cyan (bar . "[" "|" "}>" " " "]")}} {{percent . }} | {{speed . }} | {{rtime . "ETA %s"}}{{string . "suffix"}}`)
	
	// Set prefix
	prefix := fmt.Sprintf("\033[32m%s:\033[0m ", config.OperationName)
	bar.Set("prefix", prefix)

	// Set suffix
	suffix := ""
	if config.Bitrate != "" {
		suffix = fmt.Sprintf(" | %s", config.Bitrate)
	} else if config.OutputPath != "" {
		// Show output filename if no bitrate
		outputName := config.OutputPath
		if len(outputName) > 30 {
			outputName = "..." + outputName[len(outputName)-27:]
		}
		suffix = fmt.Sprintf(" | %s", outputName)
	}
	bar.Set("suffix", suffix)

	bar.SetRefreshRate(100 * time.Millisecond)
	bar.Set(pb.Terminal, true)
	bar.Set(pb.Color, true)
	bar.Start()
	defer bar.Finish()

	// Parse stderr in real-time
	var stderrBuffer bytes.Buffer
	scanner := bufio.NewScanner(stderrPipe)
	
	// Track progress
	var lastFrame int64
	var lastTime float64

	// Use a channel to signal when stderr reading is complete
	stderrDone := make(chan struct{})

	// Read stderr in a goroutine
	go func() {
		defer close(stderrDone)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuffer.WriteString(line)
			stderrBuffer.WriteString("\n")

			// Parse progress line
			frame, currentTime, _, ok := parseFFmpegProgressLine(line)
			if ok {

				// Update progress bar based on what we have
				if config.TotalFrames > 0 && frame > 0 {
					// Use frame count if available
					if frame <= config.TotalFrames {
						bar.SetCurrent(frame)
					}
					lastFrame = frame
				} else if config.TotalDuration > 0 && currentTime > 0 {
					// Use time-based progress
					progress := (currentTime / config.TotalDuration) * 100
					if progress > 100 {
						progress = 100
					}
					// Convert to centiseconds for the bar
					currentCentiseconds := int64(currentTime * 100)
					totalCentiseconds := int64(config.TotalDuration * 100)
					if currentCentiseconds <= totalCentiseconds {
						bar.SetCurrent(currentCentiseconds)
					}
					lastTime = currentTime
				} else if frame > 0 {
					// Fallback: update based on frame count even if total unknown
					bar.SetCurrent(frame)
					lastFrame = frame
				}
			}
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	
	// Wait for stderr reading to complete
	<-stderrDone

	// Ensure progress bar shows completion (set to 100%)
	if config.TotalFrames > 0 {
		bar.SetCurrent(config.TotalFrames)
	} else if config.TotalDuration > 0 {
		bar.SetCurrent(int64(config.TotalDuration * 100))
	} else if lastFrame > 0 {
		bar.SetCurrent(lastFrame)
	} else if lastTime > 0 {
		bar.SetCurrent(int64(lastTime * 100))
	}

	// Check for errors
	if err != nil {
		stderrOutput := stderrBuffer.String()
		return fmt.Errorf("ffmpeg %s failed: %w\nOutput: %s", config.OperationName, err, stderrOutput)
	}

	return nil
}
