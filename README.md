# MovieGo

Go library for video editing using FFmpeg. Build video pipelines programmatically.

## Requirements

- [FFmpeg](https://ffmpeg.org/) installed and available in `PATH`

## Installation

```bash
go get github.com/YounesseAmhend/MovieGo
```

## Quick Start

```go
package main

import (
    "github.com/YounesseAmhend/MovieGo"
)

func main() {
    v, err := moviego.NewVideoFile("input.mp4")
    if err != nil {
        panic(err)
    }

    cut, _ := v.Cut(0, 10)
    cut, _ = cut.Scale(moviego.ScaleParams{Width: 1280, Height: 720})
    cut, _ = cut.Eq(moviego.EqParams{Gamma: moviego.F(1.2)})
    cut, _ = cut.FadeIn(1.0)

    cut.WriteVideo(moviego.VideoParameters{
        OutputPath: "output.mp4",
        Codec:      moviego.CodecH264,
        Fps:        30,
    })
}
```

## Features

- **Video I/O** – Load, cut, trim, concatenate
- **Filters** – Brightness, contrast, saturation, gamma, blur, sharpen, hue, vignette
- **Effects** – Fade in/out, grayscale, sepia, negate
- **Transitions** – Wipe, dissolve, fade between clips
- **Compositing** – Overlay clips, mix audio
- **Animations** – Scale, rotate, blur, color over time
- **Text** – Burn-in subtitles
- **Speed** – Change playback speed with optional pitch correction

## Optional Parameters

For filters like `Eq` and `Hue`, use `moviego.F(value)` for optional params. Nil means “use FFmpeg default”:

```go
// Only gamma, rest use defaults
video.Eq(moviego.EqParams{Gamma: moviego.F(1.2)})

// Full control
video.Eq(moviego.EqParams{
    Brightness: moviego.F(0.1),
    Contrast:   moviego.F(1.2),
    Saturation: moviego.F(1.5),
    Gamma:      moviego.F(1.0),
})
```

## License

See LICENSE file.
