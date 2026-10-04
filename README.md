# mc-saver

A Minecraft world backup tool written in Go. Based on a JSON rule file, it packs the *selected parts* of a world save — chunk region files and core data files — into a dated ZIP archive.

中文版：[README.zh-CN.md](README.zh-CN.md)

## Features

- **Selective backups** — back up only dimensions and chunks that you care about instead of the whole world, so archives are smaller and faster.
- **Per-dimension configuration** — selectively configure and back up the overworld (`overworld`), the Nether (`the_nether`), the End (`the_end`), and any custom dimension.
- **Flexible region selection** — use rectangular `range` rules, or `simple` rules to specify individual region coordinates.
- **Config management from the command line** — add, modify, delete and list dimensions, `range`/`simple` rules and `file` rules without editing JSON by hand.

## Download

Download the corresponding binary executable from the [Releases](https://github.com/fovlin/mc-saver/releases) page.

## Example usage

```bash
./mc-saver gencfg
# Generates a default config file - save-rule.json

./mc-saver run
# Starts a backup with default values, which are:
# config file - save-rule.json
# world directory - world
# output - world-$time.zip

./mc-saver -c config.json run level level.zip
# Starts a backup with config file config.json, world directory level, and output file level.zip

./mc-saver -l run
# Starts a backup with default values in legacy mode; worlds from before 1.21.11 should use this mode

./mc-saver add-dms minecraft:custom
# Adds a dimension with the default range rule to the config file

./mc-saver add-range minecraft:custom -2 -2 2 2
# Adds a range rule (from region -2,-2 to region 2,2) to minecraft:custom

./mc-saver add-simple minecraft:custom 8 8
# Adds the single region 8,8 to minecraft:custom

./mc-saver list-dms
# Lists the dimension ids stored in the config file
```

### Syntax

```
mc-saver [-c <config file>] [-l] [-color] <command> [args...]
```

Flags must be placed before the subcommand; anything after the subcommand is treated as a positional argument.

### Commands

Backup and utility commands:

| Command | Description |
| --- | --- |
| `run [world] [output]` | Back up a world. Defaults: world directory `world`, output `.`. |
| `gencfg [config file]` | Generate a default config file (default `save-rule.json`) and exit; fails if the file already exists. |
| `help` | Print the built-in help text and exit. |

Config commands operate on the config file: it is loaded before the command runs and written back afterwards.

| Command | Description |
| --- | --- |
| `add-dms <dimension>` | Add a dimension with the default range rule (`-1,-1` to `0,0`). |
| `del-dms <dimension>` | Delete a dimension. |
| `list` | Print every dimension rule, followed by the file rules. |
| `list-dms` | Print only the dimension ids. |
| `list-range <dimension>` | Print the range rules of a dimension, with their indices. |
| `list-simple <dimension>` | Print the simple rules of a dimension, with their indices. |
| `list-config <dimension>` | Print both the range rules and the simple rules of a dimension. |
| `list-file` | Print the file rules, with their indices. |
| `add-range <dimension> <from_x> <from_y> <to_x> <to_y>` | Append a rectangular range rule to a dimension. |
| `del-range <dimension> <index>...` | Delete the range rules at `<index>...`, as shown by `list-range`. |
| `mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>` | Replace the range rule at `<index>` with the given rectangle. |
| `add-simple <dimension> <x> <y>` | Append a single region coordinate to a dimension. |
| `del-simple <dimension> <index>...` | Delete the simple rules at `<index>...`, as shown by `list-simple`. |
| `mod-simple <dimension> <index> <x> <y>` | Replace the simple rule at `<index>`. |
| `add-file <name> [name...]` | Append one or more file rules. |
| `del-file <index>...` | Delete the file rules at `<index>...`, as shown by `list-file`. |
| `mod-file <index> <name>` | Replace the file rule at `<index>`. |

A `<dimension>` is a namespace id in the form `<namespace>:<id>`, e.g. `minecraft:overworld`. `add-range` and `add-simple` create the dimension if it does not exist yet; the other dimension commands report an error instead.

Indices are 0-based, as printed by the `list*` commands, and refer to the list as it is before the command runs, so several rules can be deleted in one call. `mod-*` reports an error when the index is not in the list, while `del-*` simply ignores such an index.

### Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-c` | `save-rule.json` | Path to the JSON backup rule file. |
| `-l` | off | Back up using the legacy single-folder layout (`DIM-1`/`DIM1`). Also makes `gencfg` write a legacy-friendly `file` list. |
| `-color` | off | Enable colored output. |

## Configuration

The rule file is JSON with two top-level fields: `dimension` and `file`. It is meant to be maintained with the config commands above; editing it by hand is **not recommended**, because a typo in a field name is silently ignored and can leave rules out of the backup. If you do edit it, run `list` afterwards to see what the tool actually reads.

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
			]
		},
		"minecraft:the_nether": {
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

Keyed by dimension namespace ID (`<namespace>:<id>`). Each dimension rule supports two selection methods:

- `range` — an array of rectangles. Each entry is `{ "from": [x, z], "to": [x, z] }`; every region file whose coordinates fall inside the rectangle (inclusive) is backed up.

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
      ]
    }
  }
}
```

- `simple` — an array of individual region coordinates, e.g. `[x, z]`, for precisely specifying individual regions.

```json
{
  "dimension": {
    "minecraft:overworld": {
      "simple": [
        {
          "x": 0,
          "y": 0
        }
      ]
    }
  }
}
```

Mix `range` and `simple` rules.

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
          "x": 0,
          "y": 0
        }
      ]
    }
  }
}
```

Region coordinates follow Minecraft's region file naming (`r.<x>.<z>.mca`, one region covers 512×512 blocks).

### `file`

A list of files or folders at the world root to include. Files are added directly; folders are traversed recursively.

## Notes

- Flags must be written before the subcommand.
- Every command except `help` and `gencfg` loads the config file first; if the file does not exist, an empty config is created automatically.
- Running without a command prints an error; use `help` for the built-in usage text.
- When using legacy world mode, `gencfg -l` writes a legacy-friendly default `file` list (`playerdata` instead of `players`, plus `advancements`). If you reuse an existing config, update the `file` field with `del-file` / `add-file`.
