# mc-saver

A Minecraft world backup tool written in Go. It reads a JSON rule file and packs the *selected parts* of a world save — chunk region files and core data files — into a dated ZIP archive.

## Features

- **Selective backups** — back up only the dimensions and chunks you care about, so archives are smaller and faster than a full copy.
- **Per-dimension rules** — the overworld, the Nether, the End, and any custom dimension.
- **Two ways to pick regions** — `range` rectangles, or `simple` for individual regions.
- **Config managed from the command line** — add, modify, delete and list dimensions, rules and files without hand-editing JSON.

## Install

Download a prebuilt binary from the [Releases](https://github.com/fovlin/mc-saver/releases) page, or build from source — see [DEVELOP.md](DEVELOP.md).

## Usage

```
mc-saver [-c <config file>] [-l] [-color] <command> [args...]
```

Flags must be written before the subcommand; everything after it is a positional argument.

```bash
mc-saver gencfg                           # write a default save-rule.json
mc-saver run                              # back up ./world into a dated zip in .
mc-saver -c config.json run level out.zip # explicit config, world directory and output
mc-saver -l run                           # legacy worlds (before 1.21.11)
```

### Commands

Backup and utility:

| Command | Description |
| --- | --- |
| `run [world] [output]` | Back up a world. Defaults: `world` and `.`. |
| `gencfg [config file]` | Write a default config (default `save-rule.json`); fails if the file already exists. |
| `help` | Print the built-in usage text. |

Config commands load the config file before running and write it back afterwards.

| Command | Description |
| --- | --- |
| `list` | Print every dimension rule, then the file rules. |
| `list-config <dimension>...` | Print both the range and the simple rules. |
| `list-dms` | Print the dimension ids. |
| `add-dms <dimension>...` | Add dimensions with the default range rule (`-1,-1` to `0,0`). |
| `del-dms <dimension>...` | Delete dimensions. |
| `mod-dms <old> <new>` | Rename a dimension, keeping its rules. |
| `list-range <dimension>...` | Print the range rules, with their indices. |
| `add-range <dimension> <from_x> <from_y> <to_x> <to_y>` | Append a rectangular range rule. |
| `del-range <dimension> <index>...` | Delete the range rules at `<index>...`. |
| `mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>` | Replace the range rule at `<index>`. |
| `list-simple <dimension>...` | Print the simple rules, with their indices. |
| `add-simple <dimension> <x> <y>` | Append a single region. |
| `del-simple <dimension> <index>...` | Delete the simple rules at `<index>...`. |
| `mod-simple <dimension> <index> <x> <y>` | Replace the simple rule at `<index>`. |
| `list-file` | Print the file rules, with their indices. |
| `add-file <name>...` | Append one or more file rules. |
| `del-file <index>...` | Delete the file rules at `<index>...`. |
| `mod-file <index> <name>` | Replace the file rule at `<index>`. |

A `<dimension>` is a namespace id like `minecraft:overworld`. `add-range` and `add-simple` create the dimension if it does not exist yet; the other dimension commands report an error instead.

Indices are 0-based, as printed by the `list*` commands, and refer to the list as it is before the command runs — so several rules can be changed in one call. An index outside the list is an error for both `mod-*` and `del-*`; nothing is written in that case.

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-c` | `save-rule.json` | Path to the JSON rule file. |
| `-l` | off | Back up using the legacy single-folder layout (`DIM-1`/`DIM1`). Also makes `gencfg` write a legacy-friendly `file` list. |
| `-color` | off | Enable colored output. |

## Configuration

The rule file has two top-level fields: `dimension` and `file`. Maintain it with the commands above; hand-editing is **not recommended**, because a misspelled field name is silently ignored and can quietly drop rules from the backup.

```json
{
	"dimension": {
		"minecraft:overworld": {
			"range": [
				{
					"from": {
						"x": -1,
						"y": -1
					},
					"to": {
						"x": 0,
						"y": 0
					}
				}
			],
			"simple": [
				{
					"x": 8,
					"y": 8
				}
			]
		},
		"minecraft:the_end": {
			"range": [
				{
					"from": {
						"x": -1,
						"y": -1
					},
					"to": {
						"x": 0,
						"y": 0
					}
				}
			]
		}
	},
	"file": [
		"level.dat",
		"data",
		"datapacks",
		"players"
	]
}
```

### `dimension`

Keyed by dimension namespace id. Each dimension takes either or both of:

- **`range`** — rectangles, inclusive: every region between `from` and `to` is backed up.
- **`simple`** — individual regions, for the few that fall outside your rectangles.

`x`/`y` are Minecraft's region indices (`r.<x>.<y>.mca`); one region covers 512×512 blocks. A region listed more than once — by overlapping `range` rules, or by a `range` and a `simple` — is written to the archive once per occurrence.

### `file`

Files or folders at the world root. Files are added as-is; folders are walked recursively.

## Notes

- Every command except `help` and `gencfg` loads the config first; if the file is missing, an empty one is created automatically.
- Legacy worlds: `gencfg -l` writes a legacy-friendly `file` list. For an existing config, adjust it with `add-file` / `del-file` (`players` → `playerdata`, plus `advancements`).
