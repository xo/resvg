package resvg_test

import (
	"bytes"
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/xo/resvg"
)

func Example() {
	img, err := resvg.Render(svgData)
	if err != nil {
		log.Fatal(err)
	}
	b := img.Bounds()
	fmt.Printf("width: %d height: %d", b.Max.X, b.Max.Y)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("rect.png", buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	// Output:
	// width: 400 height: 180
}

var svgData = []byte(`<?xml version="1.0" encoding="iso-8859-1"?>
<svg width="400" height="180" xmlns="http://www.w3.org/2000/svg" version="1.1">
  <rect x="50" y="20" width="150" height="150" style="fill:blue;stroke:pink;stroke-width:5;fill-opacity:0.1;stroke-opacity:0.9" />
  Sorry, your browser does not support inline SVG.
</svg>`)
