# MovieGo Test Suite

This directory contains the organized test suite for the MovieGo video editing library.

## Directory Structure

```
tests/
├── testdata/           # Input test files (test videos)
│   └── test.mp4
├── output/             # Test output files (gitignored)
├── moviego_test.go     # Main test suite with all test functions
├── test_helpers.go     # Shared test utilities and helpers
└── README.md           # This file
```

## Test data

Video tests expect `testdata/test.mp4` to exist. If this file is not in the repo (e.g. binary/large), place a small valid MP4 at `tests/testdata/test.mp4` to run the tests.

## Running Tests

### Run all tests
```bash
make test
# or
cd tests && go test -v
```

### Run specific test functions
```bash
make test-subclip        # Run subclip tests only
make test-concat         # Run concatenation tests only
make test-composite      # Run composite tests only
```

### Run with verbose output
```bash
make test-verbose
```

### Run tests and inspect output files
```bash
make test-inspect
```

### Clean test output files
```bash
make test-clean
```

## Test Coverage

The test suite includes the following test functions:

### TestSubclip
Tests basic subclip functionality with multiple test cases:
- Basic 2-5s subclip
- Beginning clip (0-2s)
- Mid clip (3-6s)

### TestMultipleSubclips
Tests creating multiple subclips from the same source video.

### TestConcatenation
Tests video concatenation with filters applied.

### TestCombinedOperations
Tests complex workflows combining:
- Subclip creation
- Filter application
- Video concatenation

### TestNestedConcatenation
Tests concatenating concatenated videos (nested operations).

### TestCompositeClip
Tests video composition features:
- Fluent API for building composites
- Configuration struct API
- Positioning and sizing
- Opacity control
- Layer management

### TestCompositeLayering
Tests advanced composite layouts:
- Picture-in-Picture (PiP) effects
- Split-screen (2x2 grid) layouts

### TestVideoProperties
Tests video property getters (duration, dimensions, frames).

## Test Helpers

The `test_helpers.go` file provides utility functions:

- `setupTestFFmpeg(t)` - Configure FFmpeg (runs once)
- `loadTestVideo(t, path)` - Load a test video file
- `cleanupOutputDir(t)` - Clean output directory before tests
- `verifyVideoFile(t, path)` - Verify video file exists and is valid
- `assertDuration(t, video, expected, tolerance)` - Assert video duration
- `assertDimensions(t, video, width, height)` - Assert video dimensions
- `writeTestVideo(t, video, path)` - Write video with standard parameters
- `getOutputPath(filename)` - Get path in output directory

## Writing New Tests

To add a new test:

1. Add a test function in `moviego_test.go`:
```go
func TestYourFeature(t *testing.T) {
    setupTestFFmpeg(t)
    cleanupOutputDir(t)
    
    video := loadTestVideo(t, "testdata/test.mp4")
    
    // Your test logic here
    
    // Use assertions
    assertDuration(t, video, expectedDuration, 0.1)
    
    // Write and verify output
    outputPath := getOutputPath("your_output.mp4")
    writeTestVideo(t, video, outputPath)
}
```

2. For table-driven tests:
```go
func TestYourFeature(t *testing.T) {
    setupTestFFmpeg(t)
    cleanupOutputDir(t)
    
    tests := []struct {
        name     string
        param1   float64
        expected float64
    }{
        {"Case1", 1.0, 2.0},
        {"Case2", 2.0, 4.0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## Best Practices

1. **Always call `setupTestFFmpeg(t)` at the start of each test**
2. **Call `cleanupOutputDir(t)` to start with a clean slate**
3. **Use descriptive test names** - they appear in test output
4. **Use table-driven tests** for testing multiple similar cases
5. **Use assertion helpers** instead of manual checks
6. **Write output to `output/` directory** using `getOutputPath()`
7. **Keep test files small and focused** on specific functionality

## Output Files

All test output files are written to `tests/output/` which is:
- Automatically created if it doesn't exist
- Cleaned by `make test-clean`
- Gitignored (won't be committed)
- Named descriptively by test case

## CI/CD Integration

To run tests in CI:
```bash
cd tests && go test -v -count=1
```

The `-count=1` flag disables test caching, ensuring tests always run fresh.
