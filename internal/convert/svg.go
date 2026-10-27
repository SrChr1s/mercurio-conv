package convert

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	resvg "github.com/kanrichan/resvg-go"
)

// minSVGEdge is the long edge used when an SVG is converted without a size limit,
// so small icon-sized viewBoxes don't produce unusably tiny rasters.
const minSVGEdge = 512

// maxSVGEdge guards against absurd output sizes.
const maxSVGEdge = 16384

// The resvg WASM instance is single-threaded, so renders are serialized.
var svgRenderer struct {
	sync.Mutex
	once sync.Once
	r    *resvg.Renderer
	err  error
}

func renderer() (*resvg.Renderer, error) {
	svgRenderer.once.Do(func() {
		ctx, err := resvg.NewContext(context.Background())
		if err != nil {
			svgRenderer.err = err
			return
		}
		r, err := ctx.NewRenderer()
		if err != nil {
			svgRenderer.err = err
			return
		}
		loadDefaultFont(r)
		svgRenderer.r = r
	})
	return svgRenderer.r, svgRenderer.err
}

// loadDefaultFont gives <text> a fallback face. resvg-go's own LoadFontDir cannot open
// absolute paths (it validates them with fs.ValidPath), so we hand it the bytes instead.
func loadDefaultFont(r *resvg.Renderer) {
	want := []string{"LiberationSans-Regular.ttf", "DejaVuSans.ttf", "NotoSans-Regular.ttf", "NotoSans[wdth,wght].ttf", "Arial.ttf", "Helvetica.ttc"}
	roots := []string{"/usr/share/fonts", "/usr/local/share/fonts", "/Library/Fonts", "/System/Library/Fonts", `C:\Windows\Fonts`}
	for _, name := range want {
		for _, root := range roots {
			var found string
			_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil || found != "" {
					return filepath.SkipDir
				}
				if !d.IsDir() && d.Name() == name {
					found = p
					return filepath.SkipAll
				}
				return nil
			})
			if found == "" {
				continue
			}
			if data, err := os.ReadFile(found); err == nil && r.LoadFontData(data) == nil {
				return
			}
		}
	}
}

func isSVG(head []byte) bool {
	if len(head) > 4096 {
		head = head[:4096]
	}
	return bytes.Contains(head, []byte("<svg"))
}

// renderSVG renders at w×h (0,0 = the SVG's own size) and decodes the PNG resvg returns.
func renderSVG(data []byte, w, h uint32) (image.Image, error) {
	r, err := renderer()
	if err != nil {
		return nil, fmt.Errorf("motor SVG: %w", err)
	}
	svgRenderer.Lock()
	out, err := r.RenderWithSize(data, w, h)
	svgRenderer.Unlock()
	if err != nil {
		return nil, fmt.Errorf("SVG no válido: %w", err)
	}
	return png.Decode(bytes.NewReader(out))
}

// probeSVG reports whether path is an SVG and, if so, its intrinsic size.
// A file that looks like an SVG but can't be rendered returns the real error.
func probeSVG(path string) (info Info, isSvg bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil || !isSVG(data) {
		return Info{}, false, err
	}
	img, err := renderSVG(data, 0, 0)
	if err != nil {
		return Info{}, true, err
	}
	b := img.Bounds()
	return Info{Format: SVG, Width: b.Dx(), Height: b.Dy()}, true, nil
}

// rasterSVG renders the vector at its final size (transparent background).
// With a max bound it fills that bound (upscaling is fine for vectors);
// otherwise the long edge is at least minSVGEdge.
func rasterSVG(data []byte, maxW, maxH int) (image.Image, error) {
	img, err := renderSVG(data, 0, 0) // intrinsic size, also validates the file
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	vw, vh := float64(b.Dx()), float64(b.Dy())
	if vw < 1 || vh < 1 {
		return nil, fmt.Errorf("el SVG no tiene un tamaño utilizable")
	}
	scale := 1.0
	switch {
	case maxW > 0 || maxH > 0:
		scale = 1e9
		if maxW > 0 {
			scale = float64(maxW) / vw
		}
		if maxH > 0 {
			scale = min(scale, float64(maxH)/vh)
		}
	case max(vw, vh) < minSVGEdge:
		scale = minSVGEdge / max(vw, vh)
	}
	w, h := max(1, int(vw*scale+0.5)), max(1, int(vh*scale+0.5))
	if w > maxSVGEdge || h > maxSVGEdge {
		return nil, fmt.Errorf("el SVG saldría de %d×%d, demasiado grande", w, h)
	}
	if w == b.Dx() && h == b.Dy() {
		return img, nil
	}
	return renderSVG(data, uint32(w), uint32(h))
}
