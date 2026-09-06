// Command shadow turns a raw window screenshot into the picture the README
// shows.
//
// Two steps. The first repairs what the capture got wrong, the second is the
// drop shadow on transparent margins that makes a window look like a window.
//
// The repair is for captures taken through Wine, which is how docs/window-shot.sh
// photographs a Win32 window on a machine that is not running Windows. The strip
// beside the last tab is left unpainted there, and what shows through is whatever
// the frame buffer held, which is black. Windows paints the button face colour.
// Any run of pure black wider than a glyph gets that colour, so text is left
// alone and the bands are not.
//
// The shadow is taken from the capture's own alpha channel, so a window with
// rounded corners, which is what the Windows 11 snipping tool hands back, casts a
// rounded shadow, and a plain rectangle casts a rectangular one.
//
// Usage: shadow <input.png> <output.png>
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

const (
	// The margin has to hold the blur and the offset, or the shadow meets the
	// edge of the picture and stops looking like light.
	margin  = 72
	offsetY = 18
	blur    = 22
	opacity = 140

	// What Windows paints where a control does not, and how long a run of
	// black has to be before it is one of those rather than a letter. Ten
	// times the width of a glyph at the size the launcher draws them.
	unpaintedRunMin = 120
)

var face = color.NRGBA{R: 240, G: 240, B: 240, A: 255}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: shadow <input.png> <output.png>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "shadow:", err)
		os.Exit(1)
	}
}

func run(sourcePath, targetPath string) error {
	source, err := readPNG(sourcePath)
	if err != nil {
		return err
	}
	repaintUnpainted(source)
	result := shadowed(source)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(out, result); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	fmt.Printf("wrote %s (%dx%d)\n", targetPath, result.Rect.Dx(), result.Rect.Dy())
	return nil
}

func readPNG(path string) (*image.NRGBA, error) {
	file, err := os.Open(path) //nolint:gosec // the path is an argument
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	decoded, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	source := image.NewNRGBA(decoded.Bounds())
	draw.Draw(source, source.Rect, decoded, decoded.Bounds().Min, draw.Src)
	return source, nil
}

// repaintUnpainted fills the black bands a Wine capture leaves with the button
// face colour.
func repaintUnpainted(img *image.NRGBA) {
	width, height := img.Rect.Dx(), img.Rect.Dy()
	for y := range height {
		run := 0
		for x := 0; x <= width; x++ {
			if x < width {
				p := img.NRGBAAt(x, y)
				if p.R == 0 && p.G == 0 && p.B == 0 {
					run++
					continue
				}
			}
			if run >= unpaintedRunMin {
				for back := x - run; back < x; back++ {
					img.SetNRGBA(back, y, face)
				}
			}
			run = 0
		}
	}
}

// shadowed returns the capture on transparent margins, over its own drop shadow.
func shadowed(source *image.NRGBA) *image.NRGBA {
	width, height := source.Rect.Dx(), source.Rect.Dy()
	canvas := image.NewNRGBA(image.Rect(0, 0, width+margin*2, height+margin*2))

	// The shape of the shadow is the shape of whatever is opaque in the
	// capture, and the blur runs over the shadow's alpha alone: black stays
	// black, only how much of it shows changes.
	alpha := make([]float64, canvas.Rect.Dx()*canvas.Rect.Dy())
	stride := canvas.Rect.Dx()
	for y := range height {
		for x := range width {
			if source.NRGBAAt(x, y).A > 0 {
				alpha[(y+margin+offsetY)*stride+x+margin] = opacity
			}
		}
	}
	gaussianBlur(alpha, stride, canvas.Rect.Dy(), blur)
	for i, a := range alpha {
		canvas.Pix[i*4+3] = uint8(math.Round(a))
	}

	draw.Draw(canvas, source.Rect.Add(image.Pt(margin, margin)), source, source.Rect.Min, draw.Over)
	return canvas
}

// gaussianBlur is a separable blur with sigma equal to radius, which is what
// the picture had before: soft enough that the shadow reads as light.
func gaussianBlur(values []float64, width, height int, sigma float64) {
	reach := int(math.Ceil(sigma * 3))
	kernel := make([]float64, 2*reach+1)
	sum := 0.0
	for i := range kernel {
		d := float64(i - reach)
		kernel[i] = math.Exp(-d * d / (2 * sigma * sigma))
		sum += kernel[i]
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	pass := make([]float64, len(values))
	for y := range height {
		for x := range width {
			acc := 0.0
			for k, w := range kernel {
				if sx := x + k - reach; sx >= 0 && sx < width {
					acc += values[y*width+sx] * w
				}
			}
			pass[y*width+x] = acc
		}
	}
	for x := range width {
		for y := range height {
			acc := 0.0
			for k, w := range kernel {
				if sy := y + k - reach; sy >= 0 && sy < height {
					acc += pass[sy*width+x] * w
				}
			}
			values[y*width+x] = acc
		}
	}
}
