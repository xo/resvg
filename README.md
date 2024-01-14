# resvg

`resvg` is a wrapper around rust's resvg c-api crate.

## Building artifacts

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
