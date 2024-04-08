package resvg

import (
	"bytes"
	_ "embed"
	"encoding/base64"
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
		opts = append(opts, WithBestFit(true), WithWidth(200))
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

func cleanString(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(s), "v")
}

// versionTxt is the embedded resvg version.
//
//go:embed version.txt
var versionTxt []byte
