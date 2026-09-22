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
  linux_amd64_musl/...      opt-in via `go build -tags musl`, see below
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

The original six targets come to ~184 MiB (linux_amd64_musl, added later, is
opt-in and not part of that count -- see below). Shipping them all in the
root module means every consumer downloads all six just to link one.

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
| linux_amd64_musl | 36.8 MiB |
| linux_arm64 | 37.1 MiB |
| linux_arm | 33.1 MiB |
| windows_amd64 | 19.4 MiB |
| **one platform (per consumer build)** | **~20-37 MiB** |
| all six original targets, as one module | ~184 MiB |

The one rough edge: `go get`, `go mod tidy`, and a bare `go mod download`
(with no package arguments) all walk every GOOS/GOARCH, so any of them
fetches all six to record `go.sum` hashes or populate the module cache. This
is expected, not a sign the split isn't working -- measuring the win by
running one of those and looking at what landed in the module cache will
show all six and looks like nothing changed. `go build`, `go install`, and
CI both before and after fetch one, which is where the win actually is:
confirmed against a real consumer (usql, `-tags charts`) with a cold module
cache -- a linux/amd64 build pulled 11 MiB of zips (`resvg` +
`libresvg/linux_amd64`, extracting to ~37 MiB) against v0.8.0's 184 MiB
monolith, and `go list -deps` for all six GOOS/GOARCH pairs each resolved to
their own submodule and no other.

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

## musl (Alpine, static linux)

`linux_amd64` is built for the `x86_64-unknown-linux-gnu` Rust target --
glibc. Alpine and other musl-based distributions have a different libc, and
before this target existed, linking against `linux_amd64`'s archive from a
musl host failed (github.com/xo/resvg#1, github.com/xo/usql#494).

`linux_amd64_musl` is built for `x86_64-unknown-linux-musl` instead, and is
opt-in: pass `-tags musl` to `go build`/`go test` on linux/amd64 to select
it in place of `linux_amd64`. It can't be automatic -- musl and glibc are
both `linux && amd64`, and Go has no build constraint for which libc a
linux host uses, so `linux_amd64`'s own tag excludes `musl` explicitly
(`linux && amd64 && !musl`) to keep the two from ever both matching.

Building against it needs `libunwind-dev` (or your distribution's
equivalent) installed at link time -- musl doesn't bundle an unwinder into
libc the way glibc does via `libgcc_s`, so it isn't implicit here the way
it is for the other Linux targets. No extra `-lm`: musl folds libm into
libc, confirmed by linking and running the full test suite in an actual
Alpine 3.20 container (`apk add gcc musl-dev libunwind-dev`), not just by
cross-linking from a glibc host -- a glibc host's linker fails on this
archive with undefined `atan2`/`cos`/`sin`/etc. regardless of what's listed
in `SYSTEM_LIBS`, because it's the wrong libc to link musl-targeted object
code against in the first place, not a missing flag.

Only `linux/amd64` is covered for musl today; `linux/arm64` musl (relevant
to Alpine on ARM) isn't built, same as the gap in CI coverage for
`linux_arm`.

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

A platform module's version is the same as the vendored resvg release it
carries (`version.txt` at the repository root), never a version of its own.
That's the operational standard here: one version number to track, not two,
and semver ordering keeps it correct regardless of what came before -- a new
resvg release is always a higher version than the last one, so there's
nothing to reconcile. `gen.sh` prints the exact commands for this at the end
of a build.

Nested modules are tagged with a path prefix, and the root module's
`require` block has to name versions that already exist. So the order is:

1. Tag each platform module: `libresvg/darwin_amd64/v0.48.1`, and so on,
   using the version from `version.txt`.
2. Push those tags and let the proxy see them.
3. Update the `require` block in the root `go.mod` to match.
4. Tag the root module, with its own independent version -- the root
   module's version tracks this package's own API, not upstream resvg.

Once a version is pushed and fetched by anyone through the public module
proxy, its content is cached forever against that exact version string: the
proxy will never re-fetch it, even if the underlying git tag is later moved
or deleted. Treat every pushed tag as immutable -- if a mistake ships,
publish a new version rather than trying to fix the old one in place.

The `replace` block in the root `go.mod` points at the platform modules in
the working tree, so `go build`, `go test`, and `go mod tidy` all work here
before any of that happens, and afterward use the archives in the tree
rather than whatever was last published. It has no effect on consumers: the
go command applies replace directives from the main module only and ignores
them in dependencies.

The flip side is that CI cannot check the `require` block, because the
replaces override it locally. Step 3 is the one to get right by hand.
