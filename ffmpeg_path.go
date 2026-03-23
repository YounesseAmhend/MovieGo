package moviego

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const cacheDirName = "MovieGo"

var (
	ffprobePath   string
	ffprobeErr    error
	ffprobeOnce   sync.Once
	ffmpegPath    string
	ffmpegErr     error
	ffmpegOnce    sync.Once
)

// getCacheDir returns the directory for caching ffprobe/ffmpeg paths (e.g. UserCacheDir/MovieGo).
func getCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheDirName), nil
}

// readPathFromCache reads the cached path for the given executable name. Returns the path if the cache file exists and the path points to an existing file; otherwise returns empty string and false.
func readPathFromCache(name string) (string, bool) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", false
	}
	cacheFile := filepath.Join(cacheDir, name+".path")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", false
	}
	path := strings.TrimSpace(string(data))
	if path == "" {
		return "", false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "", false
	}
	return path, true
}

// writePathToCache writes the resolved path to the cache file for the given executable name. Ignores errors so the app still works if the filesystem is read-only.
func writePathToCache(name, path string) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return
	}
	_ = os.MkdirAll(cacheDir, 0755)
	cacheFile := filepath.Join(cacheDir, name+".path")
	_ = os.WriteFile(cacheFile, []byte(path), 0644)
}

// searchDirs returns OS-specific directories to search for ffprobe/ffmpeg when not in PATH.
func searchDirs() []string {
	if runtime.GOOS == "windows" {
		dirs := []string{
			`C:\ffmpeg\bin`,
			`C:\Program Files\ffmpeg\bin`,
			`C:\Program Files (x86)\ffmpeg\bin`,
		}
		if p := os.Getenv("ProgramFiles"); p != "" {
			dirs = append(dirs, filepath.Join(p, "ffmpeg", "bin"))
		}
		if p := os.Getenv("ProgramFiles(x86)"); p != "" {
			dirs = append(dirs, filepath.Join(p, "ffmpeg", "bin"))
		}
		if p := os.Getenv("LocalAppData"); p != "" {
			dirs = append(dirs, filepath.Join(p, "Programs", "ffmpeg", "bin"))
		}
		return dirs
	}
	return []string{"/usr/local/bin", "/usr/bin"}
}

// executableName returns the name to use when looking for the binary on disk (e.g. "ffprobe.exe" on Windows).
func executableName(name string) string {
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		return name + ".exe"
	}
	return name
}

// findExecutable resolves the path to the named executable. Order: 1) on-disk cache, 2) PATH (LookPath), 3) warn and search system dirs. Writes resolved path to disk cache when found.
func findExecutable(name string) (string, error) {
	// 1. Try on-disk cache first
	if path, ok := readPathFromCache(name); ok {
		return path, nil
	}

	// 2. Look in PATH
	path, err := exec.LookPath(name)
	if err == nil {
		writePathToCache(name, path)
		return path, nil
	}

	// 3. Not in PATH: warn and search system
	slog.Warn(name+" not found in PATH, searching system...", "name", name)
	for _, dir := range searchDirs() {
		candidate := filepath.Join(dir, executableName(name))
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		writePathToCache(name, abs)
		return abs, nil
	}

	return "", fmt.Errorf("%s: executable file not found in PATH or common install locations", name)
}

// getFFprobePath returns the path to the ffprobe executable. Cached in memory for the process lifetime.
func getFFprobePath() (string, error) {
	ffprobeOnce.Do(func() {
		ffprobePath, ffprobeErr = findExecutable("ffprobe")
	})
	return ffprobePath, ffprobeErr
}

// getFFmpegPath returns the path to the ffmpeg executable. Cached in memory for the process lifetime.
func getFFmpegPath() (string, error) {
	ffmpegOnce.Do(func() {
		ffmpegPath, ffmpegErr = findExecutable("ffmpeg")
	})
	return ffmpegPath, ffmpegErr
}
