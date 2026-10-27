package convert

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"github.com/gen2brain/webp"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

// Format is an output (or detected input) image format.
type Format string

const (
	JPEG Format = "jpeg"
	PNG  Format = "png"
	WebP Format = "webp"
	GIF  Format = "gif"
	BMP  Format = "bmp"
	TIFF Format = "tiff"

	// SVG is input-only: it is rasterized, never produced.
	SVG Format = "svg"
)

// Outputs lists the formats we can encode to.
var Outputs = []Format{JPEG, PNG, WebP, GIF, BMP, TIFF}

// Ext returns the file extension (with dot) for the format.
func (f Format) Ext() string {
	if f == JPEG {
		return ".jpg"
	}
	return "." + string(f)
}

// Lossy reports whether the quality setting applies.
func (f Format) Lossy() bool { return f == JPEG || f == WebP }

// ParseFormat maps a name or extension to a Format.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimPrefix(s, ".")) {
	case "jpeg", "jpg":
		return JPEG, nil
	case "png":
		return PNG, nil
	case "webp":
		return WebP, nil
	case "gif":
		return GIF, nil
	case "bmp":
		return BMP, nil
	case "tiff", "tif":
		return TIFF, nil
	case "svg":
		return SVG, nil
	}
	return "", fmt.Errorf("unsupported format %q", s)
}

// encode writes img in the given format. quality is 1–100 and only used for lossy formats.
func encode(w io.Writer, img image.Image, f Format, quality int) error {
	switch f {
	case JPEG:
		return jpeg.Encode(w, flatten(img), &jpeg.Options{Quality: quality})
	case PNG:
		return (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(w, img)
	case WebP:
		return webp.Encode(w, img, webp.Options{Quality: quality, Method: 4})
	case GIF:
		return gif.Encode(w, img, nil)
	case BMP:
		return bmp.Encode(w, img)
	case TIFF:
		return tiff.Encode(w, img, &tiff.Options{Compression: tiff.Deflate, Predictor: true})
	}
	return fmt.Errorf("unsupported output format %q", f)
}
