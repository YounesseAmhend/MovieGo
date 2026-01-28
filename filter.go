package moviego

type Filter int

const (
	// Inverse represents a filter that inverts the colors of an image
	Inverse Filter = iota

	// BlackWhite represents a filter that converts an image to grayscale
	BlackWhite

	// Sepia represents a filter that applies a sepia tone effect to an image
	Sepia

	// Edge represents a filter that detects edges in an image
	Edge

	// SepiaTone represents a filter that applies a sepia tone effect to an image
	SepiaTone
)

func inverseColors(data []byte, i int) {
	data[i] = 255 - data[i]     // Red
	data[i+1] = 255 - data[i+1] // Green
	data[i+2] = 255 - data[i+2] // Blue
	// data[i+3] is alpha - preserve it unchanged
}

// blackAndWhite converts to grayscale using fixed-point integer math
// Coefficients scaled by 256: 0.299*256≈77, 0.587*256≈150, 0.114*256≈29
func blackAndWhite(data []byte, i int) {
	r := uint32(data[i])
	g := uint32(data[i+1])
	b := uint32(data[i+2])

	// Fixed-point arithmetic: (77*r + 150*g + 29*b) >> 8
	gray := uint8((77*r + 150*g + 29*b) >> 8)

	data[i] = gray
	data[i+1] = gray
	data[i+2] = gray
	// data[i+3] is alpha - preserve it unchanged
}

// sepiaTone applies sepia effect using fixed-point integer math
// Coefficients scaled by 1024 for better precision
// Red:   0.393*1024≈402, 0.769*1024≈787, 0.189*1024≈194
// Green: 0.349*1024≈357, 0.686*1024≈702, 0.168*1024≈172
// Blue:  0.272*1024≈278, 0.534*1024≈547, 0.131*1024≈134
func sepiaTone(data []byte, i int) {
	r := uint32(data[i])
	g := uint32(data[i+1])
	b := uint32(data[i+2])

	// Fixed-point arithmetic with clamping
	newRed := (402*r + 787*g + 194*b) >> 10
	if newRed > 255 {
		newRed = 255
	}

	newGreen := (357*r + 702*g + 172*b) >> 10
	if newGreen > 255 {
		newGreen = 255
	}

	newBlue := (278*r + 547*g + 134*b) >> 10
	if newBlue > 255 {
		newBlue = 255
	}

	data[i] = uint8(newRed)
	data[i+1] = uint8(newGreen)
	data[i+2] = uint8(newBlue)
	// data[i+3] is alpha - preserve it unchanged
}
func edgeDetection(data []byte, i int) {
	// Calculate gradients using only current pixel data
	// Simplified edge detection using only the current pixel's color values
	// We'll use the average color intensity as a simple edge indicator
	avg := (int(data[i]) + int(data[i+1]) + int(data[i+2])) / 3

	// If the average is in the middle range, consider it an edge
	if avg > 85 && avg < 170 {
		data[i] = 255 // White edge
		data[i+1] = 255
		data[i+2] = 255
	} else {
		data[i] = 0 // Black background
		data[i+1] = 0
		data[i+2] = 0
	}
	// data[i+3] is alpha - preserve it unchanged
}

// composeFilters combines all custom filters and built-in filters into a single function
// This optimization eliminates nested loops in the hot path
func composeFilters(customFilters []func([]byte, int), filters []Filter) func([]byte, int) {
	// If no filters, return a no-op function
	if len(customFilters) == 0 && len(filters) == 0 {
		return nil
	}

	// Create a single composed function that applies all filters
	return func(data []byte, i int) {
		// Apply custom filters
		for _, fun := range customFilters {
			fun(data, i)
		}

		// Apply built-in filters
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
}
