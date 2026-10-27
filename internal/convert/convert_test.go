package convert

import (
	"archive/zip"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func fixture(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{uint8(x * 255 / w), uint8(y * 255 / h), uint8((x + y) % 256), 255})
		}
	}
	return img
}

func encoded(t *testing.T, f Format, w, h int) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := encode(&b, fixture(w, h), f, 90); err != nil {
		t.Fatalf("encode %s: %v", f, err)
	}
	return b.Bytes()
}

func TestRoundTripAllFormats(t *testing.T) {
	for _, in := range Outputs {
		src := encoded(t, in, 64, 48)
		for _, out := range Outputs {
			got, err := Bytes(src, Options{Format: out, Quality: 80})
			if err != nil {
				t.Errorf("%s→%s: %v", in, out, err)
				continue
			}
			img, name, err := image.Decode(bytes.NewReader(got))
			if err != nil {
				t.Errorf("%s→%s decode: %v", in, out, err)
				continue
			}
			if want, _ := ParseFormat(name); want != out {
				t.Errorf("%s→%s: got format %s", in, out, name)
			}
			if b := img.Bounds(); b.Dx() != 64 || b.Dy() != 48 {
				t.Errorf("%s→%s: size %v", in, out, b)
			}
		}
	}
}

func TestQualityAffectsSize(t *testing.T) {
	src := encoded(t, PNG, 200, 200)
	for _, f := range []Format{JPEG, WebP} {
		lo, _ := Bytes(src, Options{Format: f, Quality: 10})
		hi, _ := Bytes(src, Options{Format: f, Quality: 95})
		if len(lo) >= len(hi) {
			t.Errorf("%s: q10 (%d) not smaller than q95 (%d)", f, len(lo), len(hi))
		}
	}
}

func TestFit(t *testing.T) {
	cases := []struct{ w, h, mw, mh, ew, eh int }{
		{400, 200, 100, 0, 100, 50},
		{400, 200, 0, 50, 100, 50},
		{400, 200, 100, 100, 100, 50},
		{400, 200, 1000, 1000, 400, 200}, // never upscale
		{400, 200, 0, 0, 400, 200},
	}
	for _, c := range cases {
		b := Fit(fixture(c.w, c.h), c.mw, c.mh).Bounds()
		if b.Dx() != c.ew || b.Dy() != c.eh {
			t.Errorf("%+v: got %dx%d", c, b.Dx(), b.Dy())
		}
	}
}

func TestOrient(t *testing.T) {
	if b := orient(fixture(40, 20), 6).Bounds(); b.Dx() != 20 || b.Dy() != 40 {
		t.Errorf("orientation 6 should swap dimensions, got %v", b)
	}
	if b := orient(fixture(40, 20), 3).Bounds(); b.Dx() != 40 || b.Dy() != 20 {
		t.Errorf("orientation 3 should keep dimensions, got %v", b)
	}
}

func TestOutputHasNoExif(t *testing.T) {
	var b bytes.Buffer
	jpeg.Encode(&b, fixture(32, 32), nil)
	// splice a fake APP1 Exif segment after SOI
	exif := append([]byte{0xFF, 0xE1, 0x00, 0x10}, []byte("Exif\x00\x00MM\x00*\x00\x00\x00\x08")...)
	src := append(append([]byte{0xFF, 0xD8}, exif...), b.Bytes()[2:]...)
	out, err := Bytes(src, Options{Format: JPEG, Quality: 80})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("Exif")) {
		t.Error("output still contains Exif")
	}
}

func TestBatchAndZip(t *testing.T) {
	dir := t.TempDir()
	var jobs []Job
	for _, n := range []string{"a", "b"} {
		p := filepath.Join(dir, n+".png")
		os.WriteFile(p, encoded(t, PNG, 30, 30), 0o644)
		jobs = append(jobs, Job{ID: n, Path: p})
	}
	jobs = append(jobs, Job{ID: "bad", Path: filepath.Join(dir, "missing.png")})
	var res []Result
	Batch(context.Background(), jobs, Options{Format: WebP, Quality: 70}, "", func(r Result) { res = append(res, r) })
	if len(res) != 3 {
		t.Fatalf("got %d results", len(res))
	}
	var ok []Result
	for _, r := range res {
		if r.Err == nil {
			ok = append(ok, r)
		}
	}
	if len(ok) != 2 {
		t.Fatalf("want 2 successes, got %d", len(ok))
	}
	var zb bytes.Buffer
	if err := WriteZip(&zb, append(append([]Result{}, res...), ok[0])); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zb.Bytes()), int64(zb.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) != 3 { // errors skipped, duplicate name de-duplicated
		t.Errorf("zip has %d files, want 3", len(zr.File))
	}
}

func TestUniqueName(t *testing.T) {
	taken := map[string]bool{"a.jpg": true, "a (1).jpg": true}
	if got := UniqueName("a.jpg", func(s string) bool { return taken[s] }); got != "a (2).jpg" {
		t.Errorf("got %s", got)
	}
}

const testSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="12" viewBox="0 0 24 12"><rect width="24" height="12" fill="#5266eb"/></svg>`

func TestSVGInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "i.svg")
	os.WriteFile(path, []byte(testSVG), 0o644)
	info, err := Probe(path)
	if err != nil || info.Format != SVG || info.Width != 24 || info.Height != 12 {
		t.Fatalf("probe: %+v %v", info, err)
	}
	cases := []struct {
		o      Options
		ew, eh int
	}{
		{Options{Format: PNG}, 512, 256}, // no limit → long edge 512
		{Options{Format: PNG, MaxWidth: 200}, 200, 100},
		{Options{Format: WebP, Quality: 80, MaxHeight: 60}, 120, 60},
	}
	for _, c := range cases {
		out, err := File(path, c.o)
		if err != nil {
			t.Fatalf("%+v: %v", c.o, err)
		}
		img, _, err := image.Decode(bytes.NewReader(out))
		if err != nil {
			t.Fatal(err)
		}
		if b := img.Bounds(); b.Dx() != c.ew || b.Dy() != c.eh {
			t.Errorf("%+v: got %dx%d want %dx%d", c.o, b.Dx(), b.Dy(), c.ew, c.eh)
		}
	}
}

func TestSVGText(t *testing.T) {
	p := filepath.Join(t.TempDir(), "t.svg")
	os.WriteFile(p, []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="200" height="60"><rect width="200" height="60" fill="white"/><text x="5" y="40" font-size="30" fill="black">Hola</text></svg>`), 0o644)
	out, err := File(p, Options{Format: PNG})
	if err != nil {
		t.Fatal(err)
	}
	img, _, _ := image.Decode(bytes.NewReader(out))
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if r, _, _, _ := img.At(x, y).RGBA(); r < 0x8000 {
				return
			}
		}
	}
	t.Skip("no system font found; text not rendered")
}

func TestBatchPerFileFormat(t *testing.T) {
	dir := t.TempDir()
	var jobs []Job
	for _, n := range []string{"a", "b"} {
		p := filepath.Join(dir, n+".png")
		os.WriteFile(p, encoded(t, PNG, 20, 20), 0o644)
		jobs = append(jobs, Job{ID: n, Path: p})
	}
	jobs[1].Format = JPEG // override; "a" follows Options.Format
	got := map[string]Result{}
	Batch(context.Background(), jobs, Options{Format: WebP, Quality: 70}, "", func(r Result) { got[r.ID] = r })
	for id, want := range map[string]struct {
		name string
		fmt  string
	}{"a": {"a.webp", "webp"}, "b": {"b.jpg", "jpeg"}} {
		r := got[id]
		if r.Err != nil || r.Name != want.name {
			t.Fatalf("%s: name=%q err=%v", id, r.Name, r.Err)
		}
		if _, f, _ := image.Decode(bytes.NewReader(r.Data)); f != want.fmt {
			t.Errorf("%s: encoded as %s, want %s", id, f, want.fmt)
		}
	}
}
