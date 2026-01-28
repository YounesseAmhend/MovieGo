package moviego

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
)

// frameBuffer wraps a byte slice for pooling
type frameBuffer struct {
	data []byte
}

// bufferPool manages frame buffer reuse
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &frameBuffer{}
	},
}

// getBuffer retrieves a buffer from the pool or creates a new one
func getBuffer(size int) *frameBuffer {
	buf := bufferPool.Get().(*frameBuffer)
	if cap(buf.data) < size {
		buf.data = make([]byte, size)
	} else {
		buf.data = buf.data[:size]
	}
	return buf
}

// putBuffer returns a buffer to the pool
func putBuffer(buf *frameBuffer) {
	bufferPool.Put(buf)
}

// pixelTask represents a chunk of pixels to process
type pixelTask struct {
	data   []byte
	start  int
	end    int
	filter func([]byte, int)
	done   chan struct{}
}

// workerPool manages a pool of workers for pixel processing
type workerPool struct {
	tasks   chan pixelTask
	workers int
	wg      sync.WaitGroup
}

// newWorkerPool creates a new worker pool
func newWorkerPool(workers int) *workerPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	wp := &workerPool{
		tasks:   make(chan pixelTask, workers*2),
		workers: workers,
	}

	// Start workers
	for i := 0; i < workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}

	return wp
}

// worker processes pixel tasks from the queue
func (wp *workerPool) worker() {
	defer wp.wg.Done()
	for task := range wp.tasks {
		for i := task.start; i < task.end; i += 4 {
			task.filter(task.data, i)
		}
		task.done <- struct{}{}
	}
}

// processFrame applies filters to a frame using the worker pool
func (wp *workerPool) processFrame(data []byte, composedFilter func([]byte, int)) {
	if composedFilter == nil {
		return
	}

	chunkSize := len(data) / wp.workers
	if chunkSize%4 != 0 {
		chunkSize -= chunkSize % 4
	}

	doneChan := make(chan struct{}, wp.workers)

	for worker := 0; worker < wp.workers; worker++ {
		start := worker * chunkSize
		end := start + chunkSize
		if worker == wp.workers-1 {
			end = len(data)
		}

		wp.tasks <- pixelTask{
			data:   data,
			start:  start,
			end:    end,
			filter: composedFilter,
			done:   doneChan,
		}
	}

	// Wait for all workers to finish
	for i := 0; i < wp.workers; i++ {
		<-doneChan
	}
}

// close shuts down the worker pool
func (wp *workerPool) close() {
	close(wp.tasks)
	wp.wg.Wait()
}

// frameData holds a frame's buffer and metadata
type frameData struct {
	buffer     *frameBuffer
	err        error
	frameIndex int
}

// FrameProcessorConfig holds configuration for frame processing
type FrameProcessorConfig struct {
	Video          *Video
	InputReader    io.Reader
	OutputWriter   io.WriteCloser
	TotalFrames    int64
	ComposedFilter func([]byte, int)
}

// processFrameLoop reads, processes, and writes video frames using a pipelined approach
func processFrameLoop(config FrameProcessorConfig) error {
	frameSize := int(config.Video.width * config.Video.height * 4)

	fmt.Printf("%s | %s %d | %s %s | %s %d\n",
		color.CyanString("Processing video frames"),
		color.New(color.FgWhite, color.Bold).Sprint("Total Frames:"),
		config.TotalFrames,
		color.New(color.FgWhite, color.Bold).Sprint("Bitrate:"),
		config.Video.GetBitRate(),
		color.New(color.FgWhite, color.Bold).Sprint("FPS:"),
		config.Video.GetFps())

	// Create progress bar with reduced update frequency
	bar := pb.New64(config.TotalFrames)
	bar.SetTemplateString(`{{string . "prefix"}}{{counters . }} {{cyan (bar . "[" "|" "}>" " " "]")}} {{percent . }} | {{speed . }} | {{rtime . "ETA %s"}}{{string . "suffix"}}`)
	bar.Set("prefix", "\033[32mProcessing frames:\033[0m ")
	bar.Set("suffix", fmt.Sprintf(" | %s", config.Video.GetBitRate()))
	bar.SetRefreshRate(100 * time.Millisecond)
	bar.Set(pb.Terminal, true)
	bar.Set(pb.Color, true)
	bar.Start()
	defer bar.Finish()

	// Create worker pool for pixel processing
	workerPool := newWorkerPool(runtime.NumCPU())
	defer workerPool.close()

	// Create buffered writer for output
	bufferedWriter := bufio.NewWriterSize(config.OutputWriter, frameSize*4)

	// Pipeline channels
	const pipelineDepth = 3
	readChan := make(chan frameData, pipelineDepth)
	processChan := make(chan frameData, pipelineDepth)

	var pipelineWg sync.WaitGroup
	var pipelineErr error
	var errMutex sync.Mutex

	setError := func(err error) {
		errMutex.Lock()
		defer errMutex.Unlock()
		if pipelineErr == nil {
			pipelineErr = err
		}
	}

	// Stage 1: Reader goroutine
	pipelineWg.Add(1)
	go func() {
		defer pipelineWg.Done()
		defer close(readChan)

		frameIndex := 0
		for {
			buf := getBuffer(frameSize)
			_, err := io.ReadFull(config.InputReader, buf.data)

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				putBuffer(buf)
				break
			}

			if err != nil {
				setError(fmt.Errorf("failed to read frame %d: %w", frameIndex, err))
				putBuffer(buf)
				break
			}

			readChan <- frameData{buffer: buf, err: nil, frameIndex: frameIndex}
			frameIndex++
		}
	}()

	// Stage 2: Processor goroutine
	pipelineWg.Add(1)
	go func() {
		defer pipelineWg.Done()
		defer close(processChan)

		for frame := range readChan {
			if frame.err != nil {
				processChan <- frame
				continue
			}

			// Process the frame with composed filter using worker pool
			workerPool.processFrame(frame.buffer.data, config.ComposedFilter)

			processChan <- frame
		}
	}()

	// Stage 3: Writer (main goroutine)
	frameCount := 0
	progressUpdateInterval := 10 // Update progress every N frames

	for frame := range processChan {
		if frame.err != nil {
			setError(frame.err)
			putBuffer(frame.buffer)
			break
		}

		// Write the frame to FFmpeg encoder
		_, err := bufferedWriter.Write(frame.buffer.data)
		if err != nil {
			setError(fmt.Errorf("failed to write frame %d: %w", frameCount, err))
			putBuffer(frame.buffer)
			break
		}

		// Return buffer to pool
		putBuffer(frame.buffer)

		frameCount++

		// Update progress bar less frequently
		if frameCount%progressUpdateInterval == 0 || frameCount == int(config.TotalFrames) {
			bar.SetCurrent(int64(frameCount))
		}
	}

	// Flush buffered writer
	if err := bufferedWriter.Flush(); err != nil {
		setError(fmt.Errorf("failed to flush output: %w", err))
	}

	// Wait for all pipeline stages to complete
	pipelineWg.Wait()

	// Check for pipeline errors
	errMutex.Lock()
	defer errMutex.Unlock()
	return pipelineErr
}
