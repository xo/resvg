package resvg

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

func testRender(t *testing.T, name string) {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	var opts []Option
	if name == "testdata/folder.svg" {
		opts = append(opts, WithScaleMode(ScaleBestFit), WithWidth(200))
	}
	img, err := Render(data, opts...)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	size := img.Bounds().Size()
	t.Logf("size: %d / %d", size.X, size.Y)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	b := buf.Bytes()
	out := name + ".png"
	t.Logf("writing to: %s", out)
	if err := os.WriteFile(out, b, 0o644); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	orig := name + ".orig.png"
	exp, err := os.ReadFile(orig)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	switch equal := bytes.Equal(b, exp); {
	case equal:
		t.Logf("%s and %s match!", orig, out)
	case os.Getenv("CI") != "":
		expEncoded := base64.StdEncoding.EncodeToString(exp)
		bEncoded := base64.StdEncoding.EncodeToString(b)
		t.Logf("WARNING: expected %s and %s to be equal!", orig, out)
		t.Logf("%s (expected):\n%s", orig, expEncoded)
		t.Logf("%s:\n%s", out, bEncoded)
	default:
		t.Errorf("expected %s and %s to be equal!", orig, out)
	}
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

func cleanString(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "v")
}

// versionTxt is the embedded resvg version.
//
//go:embed version.txt
var versionTxt []byte
