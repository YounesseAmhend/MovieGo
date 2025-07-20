package moviego

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// Video struct and its methods moved from moviego.go

type Video struct {
	filename   string
	codec      string
	width      int64
	height     int64
	fps        int16
	duration   float64
	frames     int64
	ffmpegArgs map[string][]string
	filters    []Filter
	customFilters []func([]byte, int)
	isTemp     bool
	audio      Audio
	bitRate    int64
}

func (video *Video) writeVideo(outputFile string) {
	const tempVideoFile = "temp_video.mp4"

	// First, extract audio from the original video
	audioCmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-i", video.GetFilename(),
		"-vn",             // No video
		"-acodec", "copy", // Copy audio codec
		"-y",
		"temp_audio.aac",
	)

	if err := audioCmd.Run(); err != nil {
		fmt.Printf("Warning: Could not extract audio: %v\n", err)
	}

	// FFmpeg command to read frames from test.mp4
	inputCmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-i", video.GetFilename(),
		"-f", "rawvideo",
		"-threads", "16",
		"-pix_fmt", "rgb24",
		"-r", fmt.Sprintf("%d", video.GetFps()),
		"-",
	)

	// FFmpeg command to encode raw frames into temp video file
	outputCmd := exec.Command("ffmpeg",
		"-loglevel", "warning",
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-s", fmt.Sprintf("%dx%d", video.width, video.height),
		"-r", fmt.Sprintf("%d", video.GetFps()),
		"-i", "-",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-threads", "16",
		"-b:v", fmt.Sprintf("%d", video.GetBitRate()),
		"-y",
		tempVideoFile,
	)

	// Set up pipes
	inputStdout, err := inputCmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	outputStdin, err := outputCmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	inputStderr, _ := inputCmd.StderrPipe()
	outputStderr, _ := outputCmd.StderrPipe()

	go io.Copy(os.Stderr, inputStderr)
	go io.Copy(os.Stderr, outputStderr)

	// Start both processes
	if err := inputCmd.Start(); err != nil {
		panic(err)
	}
	if err := outputCmd.Start(); err != nil {
		panic(err)
	}
	frameSize := video.width * video.height * 3

	buf := make([]byte, frameSize)
	frameCount := 0
	totalFrames := video.GetFrames()

	fmt.Printf("Processing video frames | Total Frames: %d | Bitrate: %d kbps | FPS: %d\n",
		totalFrames, video.GetBitRate()/1000, video.GetFps())

	// Create progress bar with pb/v3
	bar := pb.New64(totalFrames)
	bar.SetTemplateString(`{{string . "prefix"}}{{counters . }} {{cyan (bar . "[" "|" "}>" " " "]")}} {{percent . }} {{speed . }} {{rtime . "ETA %s"}}{{string . "suffix"}}`)
	bar.Set("prefix", "\033[32mProcessing frames:\033[0m ")
	bar.Set("suffix", fmt.Sprintf(" | %d kbps", video.GetBitRate()/1000))
	bar.SetRefreshRate(50 * time.Millisecond)
	bar.Set(pb.Terminal, true)
	bar.Set(pb.Color, true)
	bar.Start()

	for {
		// Read one frame
		_, err := io.ReadFull(inputStdout, buf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			panic(err)
		}

		// Optional: Process the frame (edit pixels, filter, etc.)
		// processFrame(buf, []func([]byte, int){})
		processFrame(buf, video.customFilters, video.filters)
		// / inverseColors(data, i)
		// blackAndWhite(data, i)
		// sepiaTone(data, i)
		// edgeDetection(data, i)

		// Write the frame to FFmpeg encoder
		_, err = outputStdin.Write(buf)
		if err != nil {
			panic(err)
		}

		frameCount++

		// Update progress bar
		bar.Increment()
	}

	bar.Finish()

	// Close stdin to signal EOF to FFmpeg encoder
	outputStdin.Close()

	// Wait for processes to finish
	if err := inputCmd.Wait(); err != nil {
		panic(err)
	}
	if err := outputCmd.Wait(); err != nil {
		panic(err)
	}

	fmt.Println("\nVideo processing complete. Now combining with audio...")

	// Combine processed video with original audio
	combineCmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-i", tempVideoFile,
		"-i", video.GetFilename(),
		"-c:v", "copy", // Copy video codec
		"-c:a", "copy", // Copy audio codec
		"-map", "0:v:0", // Use video from first input (processed video)
		"-map", "1:a:0", // Use audio from second input (original video)
		"-shortest", // End when shortest stream ends
		"-y",
		outputFile,
	)

	if err := combineCmd.Run(); err != nil {
		fmt.Printf("Warning: Could not combine audio: %v\n", err)
		// If combining fails, just copy the temp video as output
		copyCmd := exec.Command("cp", tempVideoFile, outputFile)
		if err := copyCmd.Run(); err != nil {
			panic(err)
		}
	}

	// Clean up temporary files
	os.Remove(tempVideoFile)
	os.Remove("temp_audio.aac")

	fmt.Printf("Video processing complete: %s (with audio)", outputFile)

	// Print FFmpeg errors
}

func processFrame(data []byte, funcs []func([]byte, int), filters []Filter) {
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	chunkSize := len(data) / numWorkers

	// Ensure chunkSize is divisible by 3 to maintain pixel boundaries
	if chunkSize%3 != 0 {
		chunkSize -= chunkSize % 3
	}

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		start := worker * chunkSize
		end := start + chunkSize
		if worker == numWorkers-1 {
			end = len(data) // Last worker gets remaining data
		}

		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i += 3 {

				for _, fun := range funcs {
					fun(data, i)
				}
				for _, f := range filters {
					switch f {
					case Inverse:
						inverseColors(data, i)
					case BlackWhite:
						blackAndWhite(data, i)
					case Sepia:
						sepiaTone(data, i)
					case Edge:
						edgeDetection(data, i)
					}
				}
			}
		}(start, end)
	}
	wg.Wait()
}

func (v *Video) SetFrames(frames int64) *Video {
	v.frames = frames
	return v
}

func (v *Video) GetFrames() int64 {
	return v.frames
}

func (v *Video) SetFilename(filename string) *Video {
	v.filename = filename
	return v
}

func (v *Video) GetFilename() string {
	return v.filename
}

func (v *Video) Codec(codec string) *Video {
	v.codec = codec
	return v
}

func (v *Video) GetCodec() string {
	return v.codec
}

func (v *Video) Width(width int64) *Video {
	v.width = width
	return v
}

func (v *Video) GetWidth() int64 {
	return v.width
}

func (v *Video) Height(height int64) *Video {
	v.height = height
	return v
}

func (v *Video) GetHeight() int64 {
	return v.height
}

func (v *Video) Duration(duration float64) *Video {
	v.duration = duration
	return v
}

func (v *Video) GetDuration() float64 {
	return v.duration
}

func (v *Video) FfmpegArgs(ffmpegArgs map[string][]string) *Video {
	v.ffmpegArgs = ffmpegArgs
	return v
}

func (v *Video) GetFfmpegArgs() map[string][]string {
	return v.ffmpegArgs
}

func (v *Video) SetIsTemp(isTemp bool) *Video {
	v.isTemp = isTemp
	return v
}

func (v *Video) GetIsTemp() bool {
	return v.isTemp
}

func (v *Video) BitRate(bitRate int64) *Video {
	v.bitRate = bitRate
	return v
}

func (v *Video) GetBitRate() int64 {
	return v.bitRate
}

func (v *Video) SetAudio(audio Audio) *Video {
	v.audio = audio
	return v
}

func (v *Video) GetAudio() *Audio {
	return &v.audio
}

func (v *Video) SetFps(fps int16) *Video {
	v.fps = fps
	return v
}

func (v *Video) GetFps() int16 {
	return v.fps
}

func (v *Video) AddFilter(filter Filter) *Video {
	v.filters = append(v.filters, filter)
	return v
}

func (v *Video) AddCustomFilter(filterFunc func([]byte, int)) *Video {
	v.customFilters = append(v.customFilters, filterFunc)
	return v
}
