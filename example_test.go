package resvg_test

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/xo/resvg"
)

func Example() {
	img, _, err := image.Decode(bytes.NewReader(svgData))
	if err != nil {
		log.Fatal(err)
	}
	b := img.Bounds()
	fmt.Printf("width: %d height: %d\n", b.Max.X, b.Max.Y)
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

func Example_render() {
	img, err := resvg.Render(svgData)
	if err != nil {
		log.Fatal(err)
	}
	b := img.Bounds()
	fmt.Printf("width: %d height: %d\n", b.Max.X, b.Max.Y)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("rect_render.png", buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	// Output:
	// width: 400 height: 180
}

func Example_background() {
	img, err := resvg.Render(svgData, resvg.WithBackground(color.White))
	if err != nil {
		log.Fatal(err)
	}
	b := img.Bounds()
	fmt.Printf("width: %d height: %d\n", b.Max.X, b.Max.Y)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("rect_white.png", buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	// Output:
	// width: 400 height: 180
}

func Example_bestFit() {
	a, err := resvg.Render(svgData, resvg.WithBestFit(true), resvg.WithWidth(300))
	if err != nil {
		log.Fatal(err)
	}
	ab := a.Bounds()
	fmt.Printf("width: %d height: %d\n", ab.Max.X, ab.Max.Y)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, a); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("rect_bestfit_a.png", buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	b, err := resvg.Render(svgData, resvg.WithBestFit(true), resvg.WithHeight(135))
	if err != nil {
		log.Fatal(err)
	}
	bb := a.Bounds()
	fmt.Printf("width: %d height: %d\n", bb.Max.X, bb.Max.Y)
	buf.Reset()
	if err := png.Encode(buf, b); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("rect_bestfit_b.png", buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	// Output:
	// width: 300 height: 135
	// width: 300 height: 135
}

func Example_scale() {
	img, err := resvg.Render(svgData, resvg.WithWidth(200), resvg.WithHeight(700))
	if err != nil {
		log.Fatal(err)
	}
	b := img.Bounds()
	fmt.Printf("width: %d height: %d\n", b.Max.X, b.Max.Y)
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("rect_scale.png", buf.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
	// Output:
	// width: 200 height: 700
}

var svgData = []byte(`<?xml version="1.0" encoding="iso-8859-1"?>
<svg width="400" height="180" xmlns="http://www.w3.org/2000/svg" version="1.1">
  <rect x="50" y="20" width="150" height="150" style="fill:blue;stroke:pink;stroke-width:5;fill-opacity:0.1;stroke-opacity:0.9" />
  Sorry, your browser does not support inline SVG.
</svg>`)
