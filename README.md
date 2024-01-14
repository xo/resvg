# resvg

`resvg` is a wrapper around [rust's resvg c-api crate][resvg].

## Windows

When using this library with Windows, the Go binary must be built statically:

```sh
$ go build -ldflags '-extldflags "-static"'
```

Otherwise, the `resvg.dll` (not included) will need to be located next to the
built Go binary, or in the Windows system path.

## Building Artifacts

```sh
$ mkdir -p libresvg/$(go env GOOS)_$(go env GOARCH)
$ git clone https://github.com/RazrFalcon/resvg.git && cd resvg/crates/c-api
$ cargo build --release
$ cp ../../target/release/libresvg.a ../../../libresvg/$(go env GOOS)_$(go env GOARCH)
$ cd ../../../ && rm -rf resvg
```

To get the needed static dependencies (for Windows):

```sh
$ cd resvg/crates/c-api
$ RUSTFLAGS="--print=native-static-libs" cargo build --release
```

[resvg]: https://github.com/RazrFalcon/resvg
