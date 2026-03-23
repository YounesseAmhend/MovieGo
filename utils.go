package moviego

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
)

// safeLastVideoLabel returns the last video filter label, or "<none>" if the filter chain is empty.
func safeLastVideoLabel(v *Video) string {
	if len(v.filterComplex) == 0 {
		return "<none>"
	}
	return v.lastVideoLabel()
}

// safeLastAudioLabel returns the last audio filter label, or "<none>" if the filter chain is empty.
func safeLastAudioLabel(v *Video) string {
	if len(v.audio.filterComplex) == 0 {
		return "<none>"
	}
	return v.audio.lastAudioLabel()
}

// safeFirstFilename returns the first filename, or "<none>" if the slice is empty.
func safeFirstFilename(filenames []string) string {
	if len(filenames) == 0 {
		return "<none>"
	}
	return filenames[0]
}

func deepCopySlice[T any](src []T) ([]T, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return nil, fmt.Errorf("deepCopySlice encode: %w", err)
	}

	var dst []T
	if err := gob.NewDecoder(&buf).Decode(&dst); err != nil {
		return nil, fmt.Errorf("deepCopySlice decode: %w", err)
	}

	return dst, nil
}

var safeLabelRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
func sanitize(filename string) string {
	base := filepath.Base(filename)
	safeName := safeLabelRegex.ReplaceAllString(base, "_")
	safeName = strings.Trim(safeName, "_")
	
	if safeName == "" {
		return "vid"
	}

	return safeName
}

func incrementGlobalCounter() uint64 {
	return atomic.AddUint64(&globalLabelCounter, 1);
}

func incrementOrderCounter() uint64 {
	return atomic.AddUint64(&globalOrderCounter, 1);
}