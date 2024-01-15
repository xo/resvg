# resvg

`resvg` is a wrapper around [rust's resvg c-api crate][resvg], used by
[`usql`][usql] to render SVGs. Useful for rendering SVGs.

[Overview][] | [Quickstart][] | [Examples][] | [Schemes][] | [Installing][] | [Using][] | [About][]

[Overview]: #overview "Overview"
[Installing]: #installing "Installing"
[Using]: #using "Using"
[About]: #about "About"

[![Unit Tests][resvg-ci-status]][resvg-ci]
[![Go Reference][goref-resvg-status]][goref-resvg]
[![Discord Discussion][discord-status]][discord]

[resvg-ci]: https://github.com/xo/resvg/actions/workflows/test.yml
[resvg-ci-status]: https://github.com/xo/resvg/actions/workflows/test.yml/badge.svg
[goref-resvg]: https://pkg.go.dev/github.com/xo/resvg
[goref-resvg-status]: https://pkg.go.dev/badge/github.com/xo/resvg.svg
[discord]: https://discord.gg/yJKEzc7prt "Discord Discussion"
[discord-status]: https://img.shields.io/discord/829150509658013727.svg?label=Discord&logo=Discord&colorB=7289da&style=flat-square "Discord Discussion"

## Overview

Install in the usual Go fashion:

```sh
$ go get github.com/xo/resvg@latest
```

Then use like the following:

```go
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
```

### Using on Windows

When using this library with Windows, the Go binary must be built statically:

```sh
$ go build -ldflags '-extldflags "-static"'
```

Otherwise, the `resvg.dll` (not included) will need to be located next to the
built Go binary, or in the Windows system path.

## Building Artifacts

Notes for building the `libresvg` artifacts:

```sh
$ mkdir -p libresvg/$(go env GOOS)_$(go env GOARCH)
$ git clone https://github.com/RazrFalcon/resvg.git && cd resvg/crates/c-api
$ cargo build --release
$ cp ../../target/release/libresvg.a ../../../libresvg/$(go env GOOS)_$(go env GOARCH)
$ cd ../../../ && rm -rf resvg
```

To get the needed static dependencies (for Windows or other platforms):

```sh
$ cd resvg/crates/c-api
$ RUSTFLAGS="--print=native-static-libs" cargo build --release
```

## TODO

- Expand API to support the avaliable `resvg_options`

## About

`resvg` was written primarily to support these projects:

- [vegagoja][vegagoja] - a package to render Vega visualizations
- [usql][usql] - a universal command-line interface for SQL databases

[resvg]: https://github.com/RazrFalcon/resvg
[usql]: https://github.com/xo/usql
[vegagoja]: https://github.com/xo/vegagoja
