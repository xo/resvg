# libresvg

Prebuilt static archives, committed to the repository, one Go module per
platform.

```
libresvg/
  resvg.h                  header, from the resvg-capi crate
  darwin_amd64/go.mod       module github.com/xo/resvg/libresvg/darwin_amd64
  darwin_amd64/doc.go       generated
  darwin_amd64/lib.go       generated: the #cgo LDFLAGS for this target
  darwin_amd64/libresvg.a
  darwin_arm64/...
  linux_amd64/...
  linux_arm64/...
  linux_arm/...
  windows_amd64/...
```

Everything here is produced by `../gen.sh`.

## Why these are committed

`go get github.com/xo/resvg` has no build step. A consumer fetching the
module receives exactly what is in the repository and never runs `gen.sh`,
so if the archive isn't committed, cgo has nothing to link against and the
package fails to build for everyone downstream.

## Why each platform is its own module

The six targets come to ~184 MiB. Shipping them all in the root module means
every consumer downloads all six just to link one.

Splitting fixes that: each target gets its own `go.mod`, and the root module
imports it from a build-tagged `link_GOOS_GOARCH.go`. The go command
downloads a module's zip only when a package it provides is actually in the
build, so a linux/amd64 build fetches the linux_amd64 module and nothing
else. Directories holding a `go.mod` are excluded from the parent module's
zip, so the root module stays small rather than carrying the archives twice.

Measured on disk:

| | size |
|---|---|
| darwin_amd64 | 27.6 MiB |
| darwin_arm64 | 27.1 MiB |
| linux_amd64 | 36.7 MiB |
| linux_arm64 | 37.1 MiB |
| linux_arm | 33.1 MiB |
| windows_amd64 | 19.4 MiB |
| **one platform (per consumer build)** | **~20-37 MiB** |
| all six, as one module | ~184 MiB |

The one rough edge: `go get` and `go mod tidy` walk every GOOS/GOARCH, so the
first of either in a consumer's tree fetches all six to record `go.sum`
hashes. Ordinary builds afterward fetch one.

## An unsupported platform still fails at the linker, not before it

Splitting the archives out does not, by itself, change what happens on a
platform with no matching target -- FreeBSD, NetBSD, illumos, and so on.
`resvg.go`'s `#cgo CFLAGS` line has no build constraint, so its embedded C
still compiles there; with no `link_GOOS_GOARCH.go` matching, no submodule
gets pulled in, and the link fails with undefined references to
`resvg_parse_tree_from_data` and the rest of the C API. That's the same
failure mode as before the split, just relocated: previously every `#cgo
..., LDFLAGS` line lived in `resvg.go` itself and none of them matched
either. Giving unsupported platforms an explicit, earlier error (a
build-tag-gated stub, or similar) is a separate, larger change to the
package's build constraints and isn't part of this split.

## Directory names are interface

The names here are the module paths, and they're also the keys of the
`TARGETS` map in `gen.sh`. Renaming one is a breaking change for anything
that already resolved it.

## Regenerating the Go scaffolding without a rebuild

`../gen.sh -m` (optionally with `-t <targets>`) (re)writes
`go.mod`/`doc.go`/`lib.go` for the given targets, and the corresponding root
`link_GOOS_GOARCH.go` files, from whatever archive is already at
`libresvg/<target>/libresvg.a`. It does not touch the archive itself, clone
resvg, or invoke cargo/cross -- it's for applying scaffolding changes (a
renamed target, a changed system library) without waiting on a full cross-rs
rebuild. Omit `-t` to regenerate every target.

## Releasing

Nested modules are tagged with a path prefix, and the root module's
`require` block has to name versions that already exist. So the order is:

1. Tag each platform module: `libresvg/darwin_amd64/v0.1.0`, and so on.
2. Push those tags and let the proxy see them.
3. Update the `require` block in the root `go.mod` to match.
4. Tag the root module.

The `replace` block in the root `go.mod` points at the platform modules in
the working tree, so `go build`, `go test`, and `go mod tidy` all work here
before any of that happens, and afterward use the archives in the tree
rather than whatever was last published. It has no effect on consumers: the
go command applies replace directives from the main module only and ignores
them in dependencies.

The flip side is that CI cannot check the `require` block, because the
replaces override it locally. Step 3 is the one to get right by hand.
