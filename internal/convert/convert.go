// Package convert decodes, transforms and re-encodes images in pure Go.
package convert

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/rwcarlsen/goexif/exif"
	_ "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
)

// Options controls a single conversion.
type Options struct {
	Format    Format `json:"format"`
	Quality   int    `json:"quality"`   // 1–100, lossy formats only
	MaxWidth  int    `json:"maxWidth"`  // 0 = unlimited
	MaxHeight int    `json:"maxHeight"` // 0 = unlimited
}

// Info describes a source image without fully keeping it in memory.
type Info struct {
	Format        Format
	Width, Height int
}

// Probe reads just the header of the file.
func Probe(path string) (Info, error) {
	f, err := os.Open(path)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()
	cfg, name, err := image.DecodeConfig(f)
	if err != nil {
		f.Close()
		if info, isSvg, svgErr := probeSVG(path); isSvg {
			return info, svgErr
		}
		return Info{}, err
	}
	format, err := ParseFormat(name)
	if err != nil {
		return Info{}, err
	}
	return Info{Format: format, Width: cfg.Width, Height: cfg.Height}, nil
}

// File converts the image at path and returns the encoded bytes.
func File(path string, o Options) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Bytes(data, o)
}

// Bytes converts encoded image data. EXIF and other metadata are dropped
// because only pixels are re-encoded; orientation is applied first.
func Bytes(data []byte, o Options) ([]byte, error) {
	if o.Quality < 1 || o.Quality > 100 {
		o.Quality = 80
	}
	var img image.Image
	var err error
	if isSVG(data) {
		// Rendered straight at the target size, so Fit below is a no-op.
		img, err = rasterSVG(data, o.MaxWidth, o.MaxHeight)
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
		if err == nil {
			img = orient(img, readOrientation(data))
		}
	}
	if errors.Is(err, image.ErrFormat) {
		return nil, errors.New("formato de imagen no reconocido")
	}
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer la imagen: %w", err)
	}
	img = Fit(img, o.MaxWidth, o.MaxHeight)
	var buf bytes.Buffer
	if err := encode(&buf, img, o.Format, o.Quality); err != nil {
		return nil, fmt.Errorf("no se pudo codificar en %s: %w", o.Format, err)
	}
	return buf.Bytes(), nil
}

// Fit scales img down to fit within maxW×maxH, keeping aspect ratio.
// A zero bound is unlimited; images are never upscaled.
func Fit(img image.Image, maxW, maxH int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	nw, nh := w, h
	if maxW > 0 && nw > maxW {
		nh = nh * maxW / nw
		nw = maxW
	}
	if maxH > 0 && nh > maxH {
		nw = nw * maxH / nh
		nh = maxH
	}
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	if nw == w && nh == h {
		return img
	}
	dst := image.NewNRGBA(image.Rect(0, 0, nw, nh))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, b, xdraw.Over, nil)
	return dst
}

// flatten composites transparent pixels onto white for formats without alpha.
func flatten(img image.Image) image.Image {
	if o, ok := img.(interface{ Opaque() bool }); ok && o.Opaque() {
		return img
	}
	b := img.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, b, img, b.Min, draw.Over)
	return dst
}

func readOrientation(data []byte) int {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return 1
	}
	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}
	v, err := tag.Int(0)
	if err != nil {
		return 1
	}
	return v
}
