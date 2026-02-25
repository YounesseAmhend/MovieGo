package moviego

import (
	"bytes"
	"encoding/gob"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
)

func deepCopySlice[T any](src []T) ([]T, error) {
    var buf bytes.Buffer
    if err := gob.NewEncoder(&buf).Encode(src); err != nil {
        return nil, err
    }

    var dst []T
    if err := gob.NewDecoder(&buf).Decode(&dst); err != nil {
        return nil, err
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