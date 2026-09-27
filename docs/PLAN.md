# Plan

This file holds the plan for `resvg`: its purpose, its current state and its
direction. Then it holds every decision, and at the end, the questions that
are open for Ken.

## Index of decisions

Each decision is below, under Decisions. This table lists each one with its
title and its status.

| Decision | Title | Status |
| --- | --- | --- |
| [D1](#d1-the-package-links-a-prebuilt-resvg-archive-that-the-repository-holds-decided) | The package links a prebuilt resvg archive that the repository holds | Decided |
| [D2](#d2-versiontxt-names-the-resvg-release-and-a-test-compares-it-with-the-header-decided) | version.txt names the resvg release, and a test compares it with the header | Decided |
| [D3](#d3-options-passed-to-new-set-the-configuration-and-it-does-not-change-after-new-decided) | Options passed to New set the configuration, and it does not change after New | Decided |
| [D4](#d4-a-width-a-height-and-a-scale-mode-set-the-size-of-the-image-decided) | A width, a height and a scale mode set the size of the image | Decided |
| [D5](#d5-the-package-registers-itself-as-the-svg-image-format-decided) | The package registers itself as the svg image format | Decided |
| [D6](#d6-an-error-is-a-constant-of-a-string-type-decided) | An error is a constant of a string type | Decided |
| [D7](#d7-resvg-and-tree-have-close-and-are-safe-for-concurrent-use-decided) | Resvg and Tree have Close, and are safe for concurrent use | Decided |
| [D8](#d8-each-platform-archive-is-a-go-module-of-its-own-decided-amended-by-d9-and-d13) | Each platform archive is a Go module of its own | Decided, amended by D9 and D13 |
| [D9](#d9-a-platform-module-has-the-version-of-the-resvg-release-that-it-holds-decided-amends-d8) | A platform module has the version of the resvg release that it holds | Decided, amends D8 |
| [D10](#d10-testrender-judges-an-image-by-its-structure-and-not-by-its-pixels-decided) | TestRender judges an image by its structure and not by its pixels | Decided |
| [D11](#d11-ci-tests-each-platform-on-a-runner-of-that-platform-decided) | CI tests each platform on a runner of that platform | Decided |
| [D12](#d12-go-vet-runs-without-the-cgocall-analyzer-decided) | go vet runs without the cgocall analyzer | Decided |
| [D13](#d13-musl-is-a-separate-target-that-the-musl-build-tag-selects-decided-amends-d8) | musl is a separate target that the musl build tag selects | Decided, amends D8 |
| [D14](#d14-libresvgreadmemd-stays-beside-the-archives-proposed) | libresvg/README.md stays beside the archives | Proposed |
| [D15](#d15-resvg-is-set-up-for-coding-agents-as-every-xo-repository-is-decided) | resvg is set up for coding agents as every xo repository is | Decided |

## Purpose

`resvg` renders an SVG as a Go `image.RGBA`. It wraps the C API of the Rust
resvg crate through cgo, so that a Go program can render an SVG without a
browser and without a separate process. It was written for `usql`, which
renders charts with it, and for `vegagoja`, which renders Vega
visualizations.

## Current state

The latest release is v0.11.0, at commit 1b38dcf. It links resvg v0.48.1.

The package builds for seven targets: `darwin_amd64`, `darwin_arm64`,
`linux_amd64`, `linux_amd64_musl`, `linux_arm`, `linux_arm64` and
`windows_amd64`. Each archive is in a module of its own under `libresvg/`,
tagged `libresvg/<target>/v0.48.1` (D8, D9). A linux/amd64 build of `usql`
downloads 11 MiB of module zips for `resvg`, against 184 MiB before the
split.

CI runs the tests on linux/amd64, linux/arm64, macOS on Apple silicon and
Windows, and in an Alpine container for musl (D11, D13). It does not run
`linux_arm` or `darwin_amd64`, because GitHub has no runner for either.

The two issues that were ever opened, #1 and #2, are closed. There are no
open pull requests.

## Direction

The TODO in `README.md` says "Rewrite as pure Go". Ken wrote it on
2024-01-18, in commit 8eb2d69. Nothing in the history started that work, and
no decision holds it. It is the first open question below.

Until then, the work is to follow each resvg release: build the archives with
`gen.sh`, tag the platform modules, and tag the root module. The steps are in
`libresvg/README.md`, under Releasing. The work that is known and not done is
in [BACKLOG.md](BACKLOG.md).

## Decisions

Each decision has a number, a title and a status. The status is `Decided` or
`Proposed`. `Decided` means that Ken chose it: he made the commit under his
name, and a release tag carries it. `Proposed` means that nobody asked Ken. A
decision that changes an earlier one says so in its status, and the earlier
one says so too. Decisions D1 to D13 were rebuilt from the git history and the
code on 2026-09-27, and each one names the commits that it comes from.

### D1. The package links a prebuilt resvg archive that the repository holds. Decided.

The package does not build resvg. The first commit, f3672c3 on 2024-01-13,
committed one archive, `libresvg/libresvg.a`. Commit 0794e91 moved it to
`libresvg/linux_amd64/`, and each later target added its own archive.
`gen.sh` arrived in ead81ac and builds every archive with `cross`.

The archives are in git because `go get` has no build step. A consumer gets
what the repository holds and never runs `gen.sh`, so without a committed
archive, cgo has nothing to link. `libresvg/README.md` gives this reason,
under "Why these are committed".

On Windows, a program links the archive statically, with
`-ldflags '-extldflags "-static"'`. Commit e776103 wrote that in `README.md`,
and `test.sh` runs the tests that way.

Some targets were dropped because their builds failed. Commit 76ec2fe, on
2024-02-04, commented out `freebsd_amd64`, `netbsd_amd64` and
`windows_arm64`. Commit 667bd90, on 2024-04-06, dropped `solaris_amd64`.
They are still in `gen.sh` as comments.

### D2. version.txt names the resvg release, and a test compares it with the header. Decided.

`gen.sh` checks out the latest tag of `linebender/resvg` and writes it to
`version.txt`, and it copies `resvg.h` from the same checkout.
`TestVersion` compares that file with `Version()`, which returns
`RESVG_VERSION` from `libresvg/resvg.h`. A header and a file that disagree
fail the test. The test does not read the version of the archive. Commit
ead81ac added both, first released in v0.2.2.

The release moved with the upstream releases: v0.41.0 in 667bd90, v0.47.0 in
71be162 and v0.48.1 in 869a1fe.

### D3. Options passed to New set the configuration, and it does not change after New. Decided.

Commit 8eb2d69, in v0.2.0, added the options. Each one is a function named
`With<Field>` that sets one field of a `Resvg`. `New` applies them in order.

Commit 1d547d8, in v0.9.0, made `New` build the C options at once. Before, a
`sync.Once` built them at the first render. There was never a way to change
the options after `New`, so the delay gave nothing, and it hid that a
`Resvg` is safe to share. D7 depends on this.

### D4. A width, a height and a scale mode set the size of the image. Decided.

Commit f196668, in v0.5.0, added `WithWidth`, `WithHeight` and
`WithTransform`. Commit dbc018b, in v0.6.0, replaced the best fit option with
`WithScaleMode` and six scale modes: `ScaleNone`, `ScaleMinWidth`,
`ScaleMinHeight`, `ScaleMaxWidth`, `ScaleMaxHeight` and `ScaleBestFit`.

With `ScaleNone`, a width and a height stretch the image to that size, and
`Example_distort` shows it. `WithTransform` replaces the scale that the mode
computes, and the size of the image still comes from the mode. `TestScale`
holds the expected size for each mode.

### D5. The package registers itself as the svg image format. Decided.

Commit aafcded, in v0.3.0, calls `image.RegisterFormat` for `svg` from
`init`, for data that starts with `<?xml` or with `<svg`. After a blank
import of the package, `image.Decode` renders an SVG. `Decode` and
`DecodeConfig` use `Default`, which is `New()` with no options.

### D6. An error is a constant of a string type. Decided.

Commit dbc018b, in v0.6.0, "standardizes the errors". An error of the
package is a constant of the type `Error`, such as `ErrInvalidWidth`, so no
importer can change it. A code from the resvg C API becomes an `ErrNo`, and
its text starts with `resvg:`.

### D7. Resvg and Tree have Close, and are safe for concurrent use. Decided.

Commit 1d547d8, in v0.9.0, added `Close` to `Resvg`, and the `Tree` type with
its own `Close`. Before, only a finalizer freed the C options, and the Go
garbage collector does not see the memory behind them.

A read lock guards each call that reads the C memory, and `Close` takes the
write lock. `Close` waits for a call in progress, and each call after it
returns `ErrClosed`. `Close` can run more than once. A finalizer still frees
a value that nobody closed.

`Resvg.Parse` returns a `Tree`, so that a caller can call `Config` and
`Render` more than once without a second parse. `Render` and `ParseConfig`
use it.

`Default` must not be closed. The package level `Render`, `Decode` and
`DecodeConfig` use it, and so does the `svg` image format (D5).

### D8. Each platform archive is a Go module of its own. Decided, amended by D9 and D13.

Commit 48dac20, on 2026-09-23 and in v0.9.0, moved each archive into a Go
module of its own under `libresvg/<target>/`. It copies the pattern of
`xo/blitz`. Before, the root module held all six archives, 184 MiB, and every
consumer downloaded all six.

The root module imports each platform module from `link_<target>.go`, behind
a build constraint on the target. The go command downloads a module only
when a package of that module is in the build, so a build downloads one
archive. Each `lib.go` holds the `#cgo LDFLAGS` for its archive, and
`resvg.go` keeps only the `#cgo CFLAGS`. `gen.sh -m` writes all of these
files.

The `require` block in the root `go.mod` names the platform modules, and a
`replace` block points each one at the working tree. The `replace` block
has no effect on a consumer.

`go get`, `go mod tidy` and `go mod download` with no package fetch every
platform module. That is how the go command works, and it is not a fault of
the split. Commit b75420f wrote that down after `usql` measured it.

A platform that has no target still fails at the link step, with undefined
references to the resvg functions. The split did not change that.

### D9. A platform module has the version of the resvg release that it holds. Decided, amends D8.

D8 tagged the platform modules v0.1.0. Commit 0bdac9a, in v0.9.1, tagged
them again with the version in `version.txt`, v0.48.1. There is then one
version to follow and not two, and a new resvg release is always a higher
version.

The root module has its own version, which follows the API of this package.
The old `libresvg/*/v0.1.0` tags stay. A version that the module proxy
fetched is fixed forever, so a mistake gets a new version, and a pushed tag
never moves. `gen.sh` prints the tag commands at the end of a build.

### D10. TestRender judges an image by its structure and not by its pixels. Decided.

Commit ad4494e, in v0.10.0, removed the golden PNG files. `TestRender`
compared each render byte for byte with a committed PNG, and the test failed
in CI because each runner had other fonts.

The test now loads `testdata/fonts/DejaVuSans.ttf` with system fonts off, as
the upstream resvg tests do. Even with one font, anti-aliased pixels can
differ between amd64 and arm64. So the test makes sure that the render
succeeds, that the image has a size, that the size is the same after a PNG
encode and decode, and that at least one pixel is not transparent. The
pixels are the job of upstream resvg. This package proves the cgo boundary.

The same commit added `WithSansSerifFamily`, because the test SVGs use the
generic `sans-serif` family.

### D11. CI tests each platform on a runner of that platform. Decided.

Commit cbfd061, in v0.10.0, gave `.github/workflows/test.yml` the job shape of
`xo/blitz`. One job runs for each target that GitHub has a runner for:
linux/amd64, linux/arm64, macOS on Apple silicon, and Windows. `linux_arm` and
`darwin_amd64` have no runner, so CI does not run them, and `gen.sh` still
builds them.

The linux/amd64 job also makes sure that `gen.sh -m` changes nothing, and
that no module is over the 500 MiB limit of the go command. A hygiene job
makes sure that no file is over 100 MiB, and that no archive is executable.
Each job uploads the images that `TestRender` wrote.

### D12. go vet runs without the cgocall analyzer. Decided.

Commit 8a4a3a7, in v0.11.0, runs `go vet -cgocall=false`. An analyzer is one
test that `go vet` runs. The cgocall analyzer reports `C.parse(data, ...)` and
`C.render(..., img.Pix)`, because it cannot prove that a `[]byte` holds no Go
pointer. A `[]byte` never holds one, and `_GoBytes_` is the documented way to
pass one to C. Every other analyzer of `go vet` still runs.

### D13. musl is a separate target that the musl build tag selects. Decided, amends D8.

Commit 4a0ab19, in v0.11.0, added `libresvg/linux_amd64_musl`, built for
`x86_64-unknown-linux-musl`. It closed issue #1, which said that `usql` did
not build on Alpine or with Nix (xo/usql#494).

Go has no build constraint for the C library of a Linux host, so musl and
glibc are both `linux && amd64`. A build selects musl with `-tags musl`, and
the tag of `linux_amd64` became `linux && amd64 && !musl`. This amends D8,
where the target alone selected the module.

`go mod tidy` does not find a module that only a custom tag selects, so the
`linux_amd64_musl` line in the `require` block is written by hand. A musl
build needs `libunwind-dev` at link time. The archive was tested in an
Alpine 3.20 container, and a CI job runs the tests there.

### D14. libresvg/README.md stays beside the archives. Proposed.

The xo standard puts every document other than the four root documents in
`docs/` (dbmeta D110). `libresvg/README.md` is a document outside `docs/`.
It describes the files in its folder, and GitHub shows it on the page of
that folder. `README.md`, `gen.sh`, `go.mod` and the comments in `resvg.go`
point at it.

This decision keeps it in `libresvg/`. The agent that set up the repository
chose this and did not ask Ken. It is the second open question below.

### D15. resvg is set up for coding agents as every xo repository is. Decided.

Ken decided on 2026-09-27 that every repository in the `xo` namespace is set
up for coding agents the same way. The standard is dbmeta D110. This small
library keeps its decisions in this file and does not split them into one
file each, as dbmeta D111 says.

- `AGENTS.md` is new and holds the rules. There was no `CLAUDE.md` before.
- `CLAUDE.md` holds one line, `@AGENTS.md`. It is an ordinary file.
- `CONTRIBUTING.md` is new. Its Agent skills section holds the command that
  installs each skill.
- `.gitignore` ignores `.claude/settings.local.json`.
- `.gitattributes` holds `* text=auto eol=lf`.
- `docs/PLAN.md` and `docs/BACKLOG.md` are new.

The `.gitignore` line and `.gitattributes` were in the working tree before
this work started, and nobody had committed them.

Commit fe0f1d5, on 2026-09-22, committed `.claude/skills/go-pedantry` and
`.claude/skills/simple-english` as symbolic links to `.agents/skills`. A
Windows checkout writes a link as a small text file, and Claude Code then
loads no skill. They are now ordinary folders. The working tree already held
the copies, and the two `npx skills@1.7.0 add ... --copy` commands in
`CONTRIBUTING.md` changed nothing in them or in `skills-lock.json`.

`TestSkillsAreCopies` fails on a link, on a missing copy, on two copies that
differ and on a skill that `skills-lock.json` does not name.
`TestClaudeImportsAgents` fails unless `CLAUDE.md` is an ordinary file that
holds only `@AGENTS.md`. `TestTheRootHoldsFourDocuments` fails on any other
Markdown file in the root. `TestThePlanIndexIsComplete` and
`TestAnAmendmentPointsBothWays` hold the index and the amendments of this
file. The tests use only the standard library.

## Open questions for Ken

1. Is "Rewrite as pure Go" still the direction? If it is, it needs a
   decision and a first step. If it is not, the TODO section of `README.md`
   can go.
2. Does `libresvg/README.md` stay in `libresvg/`, or does it move to
   `docs/`? D14 keeps it where it is, and it is Proposed.
3. Do any of the dropped targets come back? They are `freebsd_amd64`,
   `netbsd_amd64`, `solaris_amd64` and `windows_arm64` (D1), and musl on
   linux/arm64 (D13), which Alpine on ARM needs.
4. Does the repository get a `.golangci.yml` and a lint job in CI? Other xo
   repositories have both. This one has neither.
5. Can the `master` branch on GitHub go? It stops at 10768cd, 18 commits
   behind `main`, and `main` holds all of it.
