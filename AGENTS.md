# resvg

`resvg` is a Go package that renders an SVG as an `image.RGBA`. It wraps the C
API of the Rust [resvg](https://github.com/linebender/resvg) crate through cgo.
`usql` uses it to render charts, and `vegagoja` uses it to render Vega
visualizations.

The package does not build resvg. It links a prebuilt static archive, and the
repository holds one archive for each supported platform. The archives are
large, so each one is in a Go module of its own under `libresvg/`. A build
downloads only the archive for the platform that it builds for.

## Standing rules

These rules hold in every `xo` repository, for every coding agent (dbmeta
D110).

1. Stage changes for review. Commit and push only when Ken says so.
2. Load the `simple-english` skill before you write any text that a person
   reads: project documentation, a code comment, an error message or a commit
   message. Follow it for that text.
3. Load the `go-pedantry` skill before you write or review Go code. Follow it
   where it does not conflict with a rule in this file. A rule here wins.

`CLAUDE.md` holds one line that imports this file, so that Claude Code and
every other agent read the same rules. Edit this file, not that one.

## Which document to read

| If you are | Read |
| --- | --- |
| asking why something is the way it is | the index at the top of [docs/PLAN.md](docs/PLAN.md) |
| looking for work that is known and not done | [docs/BACKLOG.md](docs/BACKLOG.md) |
| updating to a new resvg release | [README.md](README.md), under Building Artifacts, then [libresvg/README.md](libresvg/README.md), under Releasing |
| adding, removing or renaming a platform | [libresvg/README.md](libresvg/README.md), then `gen.sh`. See D8 and D13 |
| changing the generated `link_*.go` or `libresvg/<target>` files | `gen.sh`, and not the files. See Generated files below |
| tagging a release | [libresvg/README.md](libresvg/README.md), under Releasing. See D9 |
| changing `TestRender` or adding a test SVG | D10 in [docs/PLAN.md](docs/PLAN.md), then Testing below |
| changing the CI workflow | D11 and D12 in [docs/PLAN.md](docs/PLAN.md) |
| writing a document, a code comment, an error message or a commit message | the `simple-english` skill. Load it first |
| writing or reviewing Go code | the `go-pedantry` skill. Load it first, then Go conventions below |
| installing or updating an agent skill | [CONTRIBUTING.md](CONTRIBUTING.md), under Agent skills |
| looking for what the library does and how to call it | [README.md](README.md) |

A document that is not in that table does not exist. If you cannot find where
something is written down, it is not written down. Ask Ken. Do not decide it
yourself and do not write it as though it were settled.

A decision is named by its number. A bare number, such as D3, is a decision in
[docs/PLAN.md](docs/PLAN.md) of this repository. A decision of another
repository names that repository, such as dbmeta D110.
`TestEveryDecisionReferenceExists` fails on a bare number that
`docs/PLAN.md` does not hold.

## Layout

- `resvg.go` is the whole package. The cgo preamble at its top holds the C
  helpers `version`, `parse` and `render`. The rest is the `Resvg` and `Tree`
  types, the `With*` options, the scale modes, the errors, and the
  `image.RegisterFormat` hook for `svg`.
- `link_<target>.go` holds one blank import of the platform module for that
  target, behind a build constraint. `gen.sh` writes these files.
- `libresvg/resvg.h` is the header of the resvg C API. `gen.sh` copies it from
  the resvg source tree.
- `libresvg/<target>/` is a Go module for one platform. It holds
  `libresvg.a`, and `gen.sh` writes its `go.mod`, `doc.go` and `lib.go`. The
  `lib.go` holds the `#cgo LDFLAGS` for that archive.
- `version.txt` holds the resvg release that the archives come from.
  `TestVersion` compares it with `RESVG_VERSION` in `libresvg/resvg.h`.
- `gen.sh` builds the archives with `cross` and writes the generated files.
- `test.sh` runs `go test` with a static link, which Windows needs.
- `testdata/` holds the SVG files that `TestRender` renders, and
  `testdata/fonts/DejaVuSans.ttf`, the font that the tests load.

There are seven targets: `darwin_amd64`, `darwin_arm64`, `linux_amd64`,
`linux_amd64_musl`, `linux_arm`, `linux_arm64` and `windows_amd64`. The
`TARGETS` map in `gen.sh` is the list.

## Generated files

`gen.sh -m` writes `link_<target>.go` and
`libresvg/<target>/{go.mod,doc.go,lib.go}` from the archives already in the
tree. It does not build anything. Do not edit
these files by hand. Change `gen.sh` and run:

```bash
./gen.sh -m
```

CI runs the same command and fails when `git diff` shows a change in `go.mod`,
`libresvg` or `link_*.go`.

`gen.sh` without `-m` clones resvg into `$HOME/src/charts` and builds every
archive in a container with `cross`. Do not run it unless Ken asks for a new
build.

The `require` block in the root `go.mod` is written by hand. The `replace`
block overrides it in this repository, so no build here and no CI job can
find a wrong version in it. The `linux_amd64_musl` line must stay, because
`go mod tidy` does not find a module that only a custom build tag selects.
See D9 and D13.

## Go conventions

These are the conventions that `resvg.go` follows. Follow them in new code.

- Configuration is a functional option. An option is `func(*Resvg)`, it is
  named `With<Field>`, and it sets one unexported field. `New` applies the
  options and then builds the C options once. Nothing changes a `Resvg`
  after `New`, so a `Resvg` is safe to share between goroutines. See D3.
- An error is a constant of the string type `Error`, with the `Err` prefix,
  such as `ErrInvalidWidth`. The text is lower case. A code from the resvg C
  API becomes an `ErrNo`, and its text starts with `resvg:`. See D6.
- A value that holds C memory has a `Close` method and a finalizer. `Close`
  takes the write lock, and a call that reads the C memory takes the read
  lock. `Close` can run more than once, and every call after it returns
  `ErrClosed`. See D7.
- Free each `C.CString` with `C.free` in the same function.
- Pass a `[]byte` to C as the cgo type `_GoBytes_`, as `parse` and `render` do.
- Every exported identifier has a doc comment that starts with its name.
- Receivers are `r` for `Resvg`, `t` for `Tree`, `mode` for `ScaleMode`, and
  `err` for `Error` and `ErrNo`.
- The package depends on the standard library and on the `libresvg` platform
  modules, and on nothing else.

## Testing

Run the tests from the repository root:

```bash
gofmt -l .
go vet -cgocall=false ./...
go test -race ./...
```

`gofmt -l .` must print nothing. `go vet` runs with `-cgocall=false`, because
its cgocall analyzer reports the two `_GoBytes_` calls, which are safe. See D12.

On Windows, run `./test.sh`, which links statically. CI does not use the race
detector on Windows.

On Alpine, and on any other musl system, add `-tags musl`, and install
`libunwind-dev` first. CI runs this in an Alpine container. See D13.

The repository has no `.golangci.yml`, and CI does not run golangci-lint.

The tests in `resvg_test.go` are in package `resvg`. The examples in
`example_test.go` are in package `resvg_test`, and each one writes a PNG file
to the repository root. `TestRender` renders every SVG in `testdata/` and
writes `<name>.svg.png` beside it. `.gitignore` ignores all of these files.

`TestRender` does not compare pixels. It loads the embedded DejaVu Sans font
with system fonts off, and it makes sure that each image has a size and at
least one pixel that is not transparent. See D10.

`.gitignore` also ignores `*.svg`. If you add an SVG to `testdata/`, add it
with `git add -f`, or CI never sees it.

`skills_test.go` and `docs_test.go` test the agent setup and the documents.
They read files and do not render anything.

## Writing documentation

Only `README.md`, `AGENTS.md`, `CLAUDE.md` and `CONTRIBUTING.md` go in the
repository root, and `TestTheRootHoldsFourDocuments` fails on any other
Markdown file there. A new document goes in `docs/`. Add it to the table at
the top of this file and to the list in `README.md`.
`TestEveryDocumentIsInBothTables` fails when you do not.
`libresvg/README.md` is beside the archives that it describes. D14 proposes
that it stays there, and Ken has not answered yet.

A decision goes in [docs/PLAN.md](docs/PLAN.md) with the next number. This is
a small library, so its decisions stay in one file (dbmeta D111). Its heading
is `### D<n>. <Title>. <Status>.`, and it has a row in the index at the top of
the file, with the same title and status. The status is `Decided` or
`Proposed`. If a decision changes an earlier one, write `, amends D<n>` after
its status, and `, amended by D<m>` after the status of the earlier one, as
in `Decided, amends D8`.
`TestThePlanIndexIsComplete` and `TestAnAmendmentPointsBothWays` enforce this.

Write `Decided` only when Ken chose it. If you do not know, write `Proposed`
and add the question to the end of `docs/PLAN.md`.

Work that is known and not done goes in [docs/BACKLOG.md](docs/BACKLOG.md),
with where it came from. When an item is done, delete it, and record anything
that was decided in `docs/PLAN.md`.
