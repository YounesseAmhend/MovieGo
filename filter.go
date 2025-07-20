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
	data[i+2] = 255 - data[i+2]
}
func blackAndWhite(data []byte, i int) {
	gray := uint8(0.299*float64(data[i]) + 0.587*float64(data[i+1]) + 0.114*float64(data[i+2]))
	data[i] = gray
	data[i+1] = gray
	data[i+2] = gray
}

func sepiaTone(data []byte, i int) {
	// Apply sepia tone formula
	red := float64(data[i])
	green := float64(data[i+1])
	blue := float64(data[i+2])

	newRed := (red * 0.393) + (green * 0.769) + (blue * 0.189)
	newGreen := (red * 0.349) + (green * 0.686) + (blue * 0.168)
	newBlue := (red * 0.272) + (green * 0.534) + (blue * 0.131)

	// Clamp values to 255
	data[i] = min(255, uint8(newRed))
	data[i+1] = min(255, uint8(newGreen))
	data[i+2] = min(255, uint8(newBlue))
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
}
