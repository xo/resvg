package resvg

import (
	"bytes"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	for _, nn := range files {
		name := nn
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
	img, err := Render(data)
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
	if !bytes.Equal(exp, b) {
		t.Errorf("expected %s and %s to have same content", orig, out)
	}
}
