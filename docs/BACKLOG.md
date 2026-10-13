# Backlog

This file holds the work that is known and not done. Each item names where it
came from. When an item is done, delete it, and record anything that was
decided in [PLAN.md](PLAN.md).

The repository had no open issue, no open pull request and no TODO comment in
the code on 2026-09-27. The items below came from the documents, the code, the
CI workflow and the git history.

## Defects

### The package level Render does not close the Resvg that it makes

`Render(data, opts...)` calls `New(opts...)` and returns without `Close`. The
C options then wait for the finalizer, and the Go garbage collector does not
see that memory. Commit 1d547d8 added `Close` for this reason (D7), and this
function does not use it. `TestRender` and every example call it. The fix is
one deferred `Close`.

Source: found in `resvg.go` while writing `AGENTS.md`.

### testdata/folder.svg is not in git

`testRender` in `resvg_test.go` gives `testdata/folder.svg` a scale mode and a
width of its own. The file is in the working tree of Ken's checkout, and git
does not hold it, because `.gitignore` ignores `*.svg`. `TestRender` renders
it on that machine and never in CI. Either commit it with `git add -f`, or
remove the special case.

Source: found with `git ls-files testdata` while writing `AGENTS.md`.

## Documentation

### The example in README.md is not the example that runs

The example in `README.md` calls `resvg.Render`, and its `Printf` has no
newline. `Example` in `example_test.go` calls `image.Decode`, and the
`resvg.Render` form is now `Example_render`. The README copy is not tested.

Source: found by comparing `README.md` with `example_test.go`.

### README.md has a stray line

The "Manually" section under Building Artifacts ends with the single word
"Use". It came in with commit ead81ac on 2024-01-18.

Source: `README.md`.

## Platforms

### An unsupported platform fails at the link step

On a platform with no target, such as FreeBSD, the C code in `resvg.go` still
compiles, no platform module matches, and the link fails with undefined
references to the resvg functions. A file with a build constraint for every
other platform can give a clear error first. The split in D8 did not change
this.

Source: `libresvg/README.md`, "An unsupported platform still fails at the
linker, not before it".

### freebsd_amd64 and windows_arm64 have never run

The archives build and link into a Go test binary with `zig cc`, and nobody ran
either binary. No FreeBSD or Windows ARM machine was available. Run
`go test` on a FreeBSD host and on Windows ARM. The FreeBSD library list in
`gen.sh` comes from `rustc` and is untested on a real host.

Source: D16 and `libresvg/README.md`.

### windows_arm64 and freebsd_amd64 are not in CI

GitHub has a `windows-11-arm` runner that can test `windows_arm64`. It needs a
clang based C compiler and no `-race`. FreeBSD has no hosted runner. A job with
a FreeBSD VM action can test `freebsd_amd64`.

Source: D16.

### The ZIGBUILD targets need cargo-zigbuild and zig

`gen.sh` needs `cargo install cargo-zigbuild` and `zig` for two targets, and
the README does not list them as requirements for building the archives.

Source: found while adding D16.

### Targets with no decision

`linux_riscv64`, `linux_ppc64le` and `linux_s390x` were never tried, and the
models disagreed on them. `netbsd_amd64` and `solaris_amd64` were dropped for
failed builds, and the models said to skip them. `dragonfly_amd64` and
`openbsd_amd64` never had a Rust target in `gen.sh`. They are comments in the
`TARGETS` map of `gen.sh`.

Source: `gen.sh`, commits 76ec2fe and 667bd90, D1 and D16. This is open
question 3 in [PLAN.md](PLAN.md).

### Nix is not tested

Issue #1 said that `usql` did not build on musl or with Nix
(xo/usql#494). The musl target (D13) fixed Alpine. Nobody built `usql` with
Nix afterward.

Source: GitHub issue #1.

## CI and release

### CI does not run linux_arm or darwin_amd64

GitHub has no runner for either target, so CI builds neither and tests
neither. `gen.sh` still builds and ships both archives.

Source: `.github/workflows/test.yml`, and D11.

### No test reads the require block

The `replace` block in `go.mod` overrides the `require` block in this
repository, so a wrong version there reaches consumers and no test fails. A
release step can build a copy of the module with the `replace` block
removed.

Source: `libresvg/README.md`, under Releasing, and D9.

### The repository has no lint configuration

There is no `.golangci.yml`, and CI does not run golangci-lint. Other xo
repositories have both.

Source: found while writing `AGENTS.md`. This is open question 4 in
[PLAN.md](PLAN.md).

### The master branch on GitHub is stale

`origin/master` stops at 10768cd, 18 commits behind `main`, and `main` holds
all of it. The CI workflow still runs on a push to `master`.

Source: `git branch -r`. This is open question 5 in [PLAN.md](PLAN.md).

## Code

### A pure Go rewrite

The TODO section of `README.md` says "Rewrite as pure Go". Nothing started it.

Source: `README.md`, commit 8eb2d69. This is open question 1 in
[PLAN.md](PLAN.md).

### The background fill sets one pixel at a time

`Tree.Render` fills the background with `SetRGBA` in a loop over each pixel.
`draw.Draw` with an `image.Uniform` does the same fill in one call.

Source: found in `resvg.go` while writing `AGENTS.md`.

### Font data goes to C through C.CString

`buildOpts` passes each font from `WithFonts` to
`resvg_options_load_font_data` through `C.CString(string(font))`. That copies
the font twice and adds a terminating NUL that the call does not use.
`C.CBytes` copies it once.

Source: found in `resvg.go` while writing `AGENTS.md`.
