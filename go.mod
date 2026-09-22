module github.com/xo/resvg

go 1.25

// One module per platform, each holding just that target's static archive and
// the #cgo LDFLAGS naming it. They are required here so `go mod tidy` keeps
// them pinned for every platform, but only the one matching GOOS/GOARCH is
// ever downloaded: the imports live in build-tagged link_*.go files, and the
// go command fetches a module's zip only when a package it provides is
// actually in the build. A linux/amd64 build pulls ~37 MiB, not the ~184 MiB
// all six come to. See libresvg/README.md.
//
// Each platform module is tagged with the same version as the vendored resvg
// release it carries (see version.txt at the repository root), not an
// independent version of its own -- there is exactly one version number to
// track, not two.
require (
	github.com/xo/resvg/libresvg/darwin_amd64 v0.48.1
	github.com/xo/resvg/libresvg/darwin_arm64 v0.48.1
	github.com/xo/resvg/libresvg/linux_amd64 v0.48.1
	// linux_amd64_musl is gated behind a manual `musl` build tag, not a
	// GOOS/GOARCH pair go mod tidy enumerates on its own, so it has to be
	// kept here by hand -- tidy won't discover or re-add it itself. See
	// "musl (Alpine, static linux)" in libresvg/README.md.
	github.com/xo/resvg/libresvg/linux_amd64_musl v0.48.1
	github.com/xo/resvg/libresvg/linux_arm v0.48.1
	github.com/xo/resvg/libresvg/linux_arm64 v0.48.1
	github.com/xo/resvg/libresvg/windows_amd64 v0.48.1
)

// Local development only, and inert for anyone who depends on this module:
// the go command applies replace directives from the main module's go.mod
// and ignores them everywhere else. This is what makes `go build`, `go test`
// and `go mod tidy` work in this repository before the
// libresvg/<target>/vX.Y.Z tags exist, and what makes them use the archives
// in the working tree afterwards, rather than whatever was last published.
//
// See "Releasing" in libresvg/README.md: the require block above is
// maintained by hand, and CI cannot check it, because these lines override
// it here.
replace (
	github.com/xo/resvg/libresvg/darwin_amd64 => ./libresvg/darwin_amd64
	github.com/xo/resvg/libresvg/darwin_arm64 => ./libresvg/darwin_arm64
	github.com/xo/resvg/libresvg/linux_amd64 => ./libresvg/linux_amd64
	github.com/xo/resvg/libresvg/linux_amd64_musl => ./libresvg/linux_amd64_musl
	github.com/xo/resvg/libresvg/linux_arm => ./libresvg/linux_arm
	github.com/xo/resvg/libresvg/linux_arm64 => ./libresvg/linux_arm64
	github.com/xo/resvg/libresvg/windows_amd64 => ./libresvg/windows_amd64
)
