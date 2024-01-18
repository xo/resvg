# resvg

`resvg` is a wrapper around [rust's resvg c-api crate][resvg], used by
[`usql`][usql] to render SVGs. Useful for rendering SVGs.

[Overview][] | [Building Artifacts][] | [TODO][] | [About][]

[Overview]: #overview "Overview"
[Building Artifacts]: #building-artifacts "Building Artifacts"
[TODO]: #todo "TODO"
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

### Using gen.sh

Build Darwin images:

```sh
# get cross
$ git clone https://github.com/cross-rs/cross.git
$ cd cross && git submodule update --init --remote


# grab sdk
$ cd cross/docker/cross-toolchains/docker
$ export SDK='https://github.com/phracker/MacOSX-SDKs/releases/download/11.3/MacOSX11.3.sdk.tar.xz'
$ curl -O -J -L "$SDK"

# NOTE: 14.2 (what's generated below) doesn't seem to work with the
# NOTE: build-docker-image command below

# download sdk from apple at https://developer.apple.com/download/all/?q=xcode
# then use osxcross to generate the sdk
# https://github.com/tpoechtrager/osxcross?tab=readme-ov-file#packaging-the-sdk
$ git clone https://github.com/tpoechtrager/osxcross.git && cd osxcross
$ ./tools/gen_sdk_package_pbzx.sh ~/Downloads/Xcode_15.2.xip

# build containers
$ cd cross
$ mv /path/to/MacOSX11.3.sdk.tar.xz docker/
$ cargo build-docker-image x86_64-apple-darwin-cross --build-arg 'MACOS_SDK_FILE=MacOSX11.3.sdk.tar.xz'
$ cargo build-docker-image aarch64-apple-darwin-cross --build-arg 'MACOS_SDK_FILE=MacOSX11.3.sdk.tar.xz'

# add rust toolchains
$ rustup target add x86_64-apple-darwin
$ rustup target add aarch64-apple-darwin
```

Then use:

```sh
$ ./gen.sh
```

### Manually

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

Use

## TODO

- Rewrite as pure Go

## About

`resvg` was written primarily to support these projects:

- [vegagoja][vegagoja] - a package to render Vega visualizations
- [usql][usql] - a universal command-line interface for SQL databases

[resvg]: https://github.com/RazrFalcon/resvg
[usql]: https://github.com/xo/usql
[vegagoja]: https://github.com/xo/vegagoja
