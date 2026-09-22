package resvg

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestVersion(t *testing.T) {
	ver := Version()
	if v, exp := cleanString(ver), cleanString(string(versionTxt)); v != exp {
		t.Fatalf("expected %s, got: %s", exp, v)
	}
	t.Logf("resvg: %s", ver)
}

func TestRender(t *testing.T) {
	var files []string
	err := filepath.Walk("testdata", func(name string, info fs.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case info.IsDir() || strings.ToLower(filepath.Ext(name)) != ".svg":
			return nil
		}
		files = append(files, name)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	for _, name := range files {
		t.Run(strings.TrimSuffix(filepath.Base(name), ".svg"), func(t *testing.T) {
			testRender(t, name)
		})
	}
}

// testFont is a fixed font, loaded instead of system fonts so that
// TestRender's output doesn't depend on which fonts happen to be installed
// on the machine running the test. Rendering itself (tiny-skia, rustybuzz)
// has no OS dependency; font *selection* is the only reason the same SVG
// paints differently across CI runners, and pinning the font file removes
// that variable the same way upstream resvg's own test suite does (a
// private fontdb loaded from a committed font directory, never system
// fonts). See testdata/fonts/LICENSE.
//
//go:embed testdata/fonts/DejaVuSans.ttf
var testFont []byte

// testRender renders name and checks that the result is structurally sound:
// it decodes, has positive, self-consistent dimensions, and isn't entirely
// transparent. It does not compare pixels against a golden image. resvg's
// rasterizer and text shaper are pure, OS-independent software (tiny-skia,
// rustybuzz), but anti-aliased pixel values can still differ by a bit or two
// across CPU architectures (amd64 vs arm64 float rounding) even with an
// identical font pinned, so a byte- or pixel-exact comparison would still be
// flaky across the platform matrix this package builds for. Verifying
// pixel-perfect rendering is resvg's own job upstream, not this wrapper's;
// this package's job is to prove the cgo boundary -- data in, dimensions and
// paint out -- works.
func testRender(t *testing.T, name string) {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	opts := []Option{
		WithLoadSystemFonts(false),
		WithFonts(testFont),
		WithSansSerifFamily("DejaVu Sans"),
	}
	if name == "testdata/folder.svg" {
		opts = append(opts, WithScaleMode(ScaleBestFit), WithWidth(200))
	}
	img, err := Render(data, opts...)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	size := img.Bounds().Size()
	t.Logf("size: %d / %d", size.X, size.Y)
	if size.X <= 0 || size.Y <= 0 {
		t.Fatalf("expected positive dimensions, got: %dx%d", size.X, size.Y)
	}
	if !nonBlank(img) {
		t.Errorf("expected %s to render something, got a fully transparent image", name)
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// a corrupted cgo buffer round-trips as a decode error or the wrong size,
	// not necessarily as an encode error above
	decoded, err := png.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("expected no error decoding the rendered png, got: %v", err)
	}
	if b := decoded.Bounds().Size(); b != size {
		t.Fatalf("expected decoded png to be %dx%d, got: %dx%d", size.X, size.Y, b.X, b.Y)
	}
	out := name + ".png"
	t.Logf("writing to: %s", out)
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// nonBlank reports whether img has at least one non-transparent pixel, as a
// cheap proxy for "something was actually painted".
func nonBlank(img *image.RGBA) bool {
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 {
			return true
		}
	}
	return false
}

func TestScale(t *testing.T) {
	tests := []struct {
		mode   ScaleMode
		width  uint
		height uint
		w      uint
		h      uint
		expw   int
		exph   int
		expx   float32
		expy   float32
	}{
		{ScaleNone, 100, 100, 0, 0, 100, 100, 1.0, 1.0},
		{ScaleNone, 100, 100, 200, 50, 200, 50, 2.0, 0.5},
		{ScaleNone, 100, 100, 50, 0, 50, 100, 0.5, 1.0},
		{ScaleNone, 100, 100, 0, 200, 100, 200, 1.0, 2.0},
		{ScaleMinWidth, 100, 100, 200, 0, 200, 200, 2.0, 2.0},
		{ScaleMinWidth, 1000, 1000, 200, 0, 1000, 1000, 1.0, 1.0},
		{ScaleMaxWidth, 100, 100, 200, 0, 100, 100, 1.0, 1.0},
		{ScaleMaxWidth, 1000, 1000, 500, 0, 500, 500, 0.5, 0.5},
		{ScaleMinHeight, 100, 100, 0, 200, 200, 200, 2.0, 2.0},
		{ScaleMinHeight, 1000, 1000, 0, 200, 1000, 1000, 1.0, 1.0},
		{ScaleMaxHeight, 100, 100, 0, 200, 100, 100, 1.0, 1.0},
		{ScaleMaxHeight, 1000, 1000, 0, 500, 500, 500, 0.5, 0.5},
		{ScaleBestFit, 100, 100, 960, 1000, 960, 960, 9.6, 9.6},
		{ScaleBestFit, 100, 100, 1000, 960, 960, 960, 9.6, 9.6},
		{ScaleBestFit, 1000, 1000, 200, 300, 200, 200, 0.2, 0.2},
		{ScaleBestFit, 1000, 5000, 100, 200, 40, 200, 0.04, 0.04},
		{ScaleBestFit, 16, 16, 200, 200, 200, 200, 12.5, 12.5},
		{ScaleBestFit, 200, 200, 16, 16, 16, 16, 0.08, 0.08},
		{ScaleBestFit, 250, 200, 100, 90, 100, 80, 0.4, 0.4},
		{ScaleBestFit, 16, 16, 200, 0, 200, 200, 12.5, 12.5},
		{ScaleBestFit, 200, 200, 0, 16, 16, 16, 0.08, 0.08},
		{ScaleBestFit, 250, 200, 0, 90, 113, 90, 0.45, 0.45},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%d_%d_%d_%d", test.width, test.height, test.w, test.h), func(t *testing.T) {
			w, h, x, y := test.mode.Scale(test.width, test.height, test.w, test.h)
			if w != test.expw {
				t.Errorf("expected w %d, got: %d", test.expw, w)
			}
			if h != test.exph {
				t.Errorf("expected h %d, got: %d", test.exph, h)
			}
			if x != test.expx {
				t.Errorf("expected x %f, got: %f", test.expx, x)
			}
			if y != test.expy {
				t.Errorf("expected y %f, got: %f", test.expy, y)
			}
		})
	}
}

func TestClose(t *testing.T) {
	data, err := os.ReadFile("testdata/rect.svg")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	r := New()
	if _, err := r.Render(data); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("expected second close to be a no-op, got: %v", err)
	}
	if _, err := r.Render(data); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected %v, got: %v", ErrClosed, err)
	}
	if _, err := r.ParseConfig(data); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected %v, got: %v", ErrClosed, err)
	}
	if _, err := r.Parse(data); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected %v, got: %v", ErrClosed, err)
	}
}

func TestTreeReuse(t *testing.T) {
	data, err := os.ReadFile("testdata/rect.svg")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	r := New()
	defer r.Close()
	tree, err := r.Parse(data)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	cfg, err := tree.Config()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	img1, err := tree.Render()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if b := img1.Bounds().Size(); b.X != cfg.Width || b.Y != cfg.Height {
		t.Fatalf("expected %dx%d, got: %dx%d", cfg.Width, cfg.Height, b.X, b.Y)
	}
	// render the same tree a second time, without reparsing
	img2, err := tree.Render()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !bytes.Equal(img1.Pix, img2.Pix) {
		t.Fatalf("expected repeated renders of the same tree to match")
	}
	if err := tree.Close(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if err := tree.Close(); err != nil {
		t.Fatalf("expected second close to be a no-op, got: %v", err)
	}
	if _, err := tree.Render(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected %v, got: %v", ErrClosed, err)
	}
	if _, err := tree.Config(); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected %v, got: %v", ErrClosed, err)
	}
}

func TestConcurrentRender(t *testing.T) {
	data, err := os.ReadFile("testdata/rect.svg")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	r := New()
	defer r.Close()
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := r.Render(data); err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		}()
	}
	wg.Wait()
}

func cleanString(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "v")
}

// versionTxt is the embedded resvg version.
//
//go:embed version.txt
var versionTxt []byte
