# Contributing to resvg

Read three things before you change anything.

[`AGENTS.md`](AGENTS.md) holds the rules: the standing rules, the layout, the
Go conventions and the commands that test the code. It is written for an AI
coding agent, and everything in it applies to a person too. `CLAUDE.md` holds
one line that imports it, for Claude Code.

[`docs/PLAN.md`](docs/PLAN.md) holds every decision this project made, with
the commits that it came from. The index at the top lists each decision with
its status. Read the status, because some decisions amend an earlier one.

[`libresvg/README.md`](libresvg/README.md) says why each platform archive is a
Go module of its own, and how to release a new resvg version.

Do not decide an open question on your own. The open questions are at the end
of `docs/PLAN.md`. Ask Ken.

## Before you send a change

Run these commands in the repository root. All of them must pass, and
`gofmt -l .` must print nothing:

```sh
gofmt -l .
go vet -cgocall=false ./...
go test -race ./...
```

On Windows, run `./test.sh` in place of `go test`, because Windows links the
archive statically.

If you changed `gen.sh`, write the generated files again, and commit them with
your change:

```sh
./gen.sh -m
```

If you add an SVG to `testdata/`, add it with `git add -f`, because
`.gitignore` ignores `*.svg`.

## Agent skills

The repository carries two agent skills. A skill is a set of instructions that
a coding agent loads for a task. `simple-english` sets how prose is written,
and `go-pedantry` sets how Go is written.

`skills-lock.json` names the source of each skill. The `skills` command from
npm writes that file, and version 1.7.0 is the one measured. It writes each
skill into two folders. Codex and the other agents read
`.agents/skills/<name>`, and Claude Code reads `.claude/skills/<name>`.

To install or update a skill, run its command in the repository root:

```sh
npx skills@1.7.0 add AminBlg/SimpleEnglish --skill simple-english --agent codex claude-code --copy -y
npx skills@1.7.0 add oborchers/fractional-cto --skill go-pedantry --agent codex claude-code --copy -y
```

Keep `--copy`. Without it, the command writes `.claude/skills/<name>` as a
symbolic link. A Windows checkout writes a symbolic link as a small text file,
and Claude Code then loads no skill and reports nothing. `TestSkillsAreCopies`
fails on a link, and it fails when the two folders differ. See D15.

`.claude/settings.local.json` holds the Claude Code permissions of one person.
The root `.gitignore` ignores it.
