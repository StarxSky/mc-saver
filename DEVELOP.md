# Develop docs

## Contributing policy

The projects of fovlin are all handwritten, and all projects reject AI coding.
It is strictly forbidden to use AI audit results to submit as issues.
If you can't accept it, please leave.

## Layout

```text
mc-saver
├── cmd/          CLI (module acovia.net/mc-saver), one file per command family
├── save/         backup engine: config loading, world walking, zip writing
├── record/       logging helpers
├── build.sh      cross-compile the release targets
├── go.work       workspace tying the three modules together
├── README.md     user documentation
└── DEVELOP.md    this file
```

Build and run from the repo root:

```bash
go build -o mc-saver ./cmd
./mc-saver help
```

Patterns resolve per module: use `go build ./cmd` or `go vet ./cmd/...`. A bare
`./...` from the root does not match anything, because the root is not itself a module.

## CLI shape

```
mc-saver [-l] [-color] <command> <world> [args...]
```

The world directory is the first argument of every command, and the rule file is
`<world>/saver.json`. `initProgram` resolves the command; each command then resolves its
own target and loads the rule file through `initWorld`.

## Release

1. `bash build.sh` — compiles 6 targets (linux / darwin / windows × amd64 / arm64) and
   leaves one `.tar.gz` per target in `build/`.
2. `git tag vX.Y.Z && git push origin vX.Y.Z`
3. Attach the six `build/*.tar.gz` files to the GitHub release.

There is no automated test suite yet; smoke-test the commands by hand before tagging.

## Conventions

- User-facing messages (`record.Error` / `Info` / `Warn`, `fmt.Print*`) carry no trailing
  period. An underlying error is appended after a colon: `record.Error("load config:", err)`.
- Returned errors are lowercase, carry no punctuation, and wrap the cause with `%w`:
  `fmt.Errorf("open file: %w", err)`.
- Version tags use a `v` prefix (`v1.2.0`).
- Commands validate their arguments before loading the config, and validate every index
  before touching it: an out-of-range index aborts the command and leaves the file unchanged.

## Not a bug

Intentional behaviours, in case they look like defects:

- The rule file is not created implicitly: any command other than `gencfg` and `help`
  fails when `<world>/saver.json` is missing. `gencfg <world>` is the only way to create it.
- Everything named in the rule file must exist in the world. A configured dimension whose
  `dimensions/<namespace>/<id>` directory is missing, or a `file` entry that is missing,
  aborts the whole backup instead of being skipped — silently skipping would hide a stale
  config or a typo. Missing *region* files inside an existing directory are skipped with a
  warning, because those are world content rather than declared input.
- `-l` and `-color` are switches; `-l=false` is not part of the interface.
- The config file format is not backward compatible — regenerate it with `gencfg` after
  upgrading.
- A region listed more than once in the config (duplicate or overlapping rules, or a
  `range` and a `simple` covering the same region) is added to the archive once per
  occurrence, so the archive can contain duplicate entries.
- If the output name is already taken, a `-1` suffix is appended instead of failing or
  overwriting the existing archive.
