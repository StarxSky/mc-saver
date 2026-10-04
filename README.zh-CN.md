# mc-saver

一个用 Go 编写的 Minecraft 存档备份工具。它根据 JSON 规则文件，将世界存档中*选定的部分*——区块区域文件和核心数据文件——打包成带日期的 ZIP 压缩包。

English version: [README.md](README.md)

## 功能

- **选择性备份** — 只备份你关心的维度和区块，而不是整个世界，压缩包更小、速度更快。
- **按维度配置** — 可以选择性配置和备份主世界（`overworld`）、下界（`the_nether`）、末地（`the_end`）以及任意自定义维度。
- **灵活的区域选择** — 既可以用矩形 `range` 规则，也可以用 `simple` 规则逐个指定区域坐标。
- **命令行管理配置** — 无需手动编辑 JSON，即可增删和查看维度、`range`/`simple` 区域规则以及 `file` 文件规则。

## 下载

在 [Releases](https://github.com/fovlin/mc-saver/releases)页面下载对应的二进制可执行文件。

## 示例方法

```bash
./mc-saver repl
# 启动一个向导，引导用户进行备份

./mc-saver gencfg
# 生成一个默认的配置文件 - save-rule.json

./mc-saver run
# 以默认值开始一个备份，默认值包括：
# 配置文件 - save-rule.json
# 世界存档目录 - world
# 输出目录 - world-$time.zip

./mc-saver -c config.json run level level.zip
# 开始一个备份，指定配置文件 config.json，世界文件 level，输出文件 level.zip

./mc-saver -l run
# 以默认值开始一个备份，使用旧版模式，1.21.11 前的存档应使用此模式

./mc-saver add-dms minecraft:custom
# 向配置文件添加一个维度，使用默认的 range 规则

./mc-saver add-range minecraft:custom -2 -2 2 2
# 为 minecraft:custom 添加一条 range 规则（从区域 -2,-2 到区域 2,2）

./mc-saver add-simple minecraft:custom 8 8
# 为 minecraft:custom 添加单个区域 8,8

./mc-saver list-dms
# 列出配置文件中保存的维度 ID
```

### 语法

```
mc-saver [-c <配置文件>] [-l] [-color] <命令> [参数...]
```

参数必须放在子命令之前；子命令之后的内容一律按位置参数处理。

### 命令

备份与辅助命令：

| 命令 | 说明 |
| --- | --- |
| `run [存档目录] [输出路径]` | 备份存档。默认值：存档目录 `world`，输出路径 `.`。 |
| `gencfg [配置文件]` | 生成默认配置文件（默认为 `save-rule.json`）后退出；若文件已存在则报错。 |
| `repl` | 运行交互向导：依次提示输入存档目录和输出路径。 |
| `help` | 显示内置帮助文本后退出。 |

配置命令直接操作配置文件：命令执行前会加载配置，执行后写回文件。

| 命令 | 说明 |
| --- | --- |
| `add-dms <维度>` | 添加一个维度，使用默认 range 规则（`-1,-1` 到 `0,0`）。 |
| `del-dms <维度>` | 删除一个维度。 |
| `list` | 打印全部维度规则，随后打印文件规则。 |
| `list-dms` | 只打印维度 ID。 |
| `list-range <维度>` | 打印某个维度的 range 规则及其下标。 |
| `list-simple <维度>` | 打印某个维度的 simple 规则及其下标。 |
| `list-config <维度>` | 同时打印某个维度的 range 和 simple 规则。 |
| `list-file` | 打印文件规则及其下标。 |
| `add-range <维度> <起始X> <起始Y> <结束X> <结束Y>` | 为某个维度追加一条矩形 range 规则。 |
| `del-range <维度> <下标>` | 删除下标对应的 range 规则（下标见 `list-range`）。 |
| `add-simple <维度> <X> <Y>` | 为某个维度追加一个区域坐标。 |
| `del-simple <维度> <下标>` | 删除下标对应的 simple 规则（下标见 `list-simple`）。 |
| `add-file <名称> [名称...]` | 追加一个或多个文件规则。 |
| `del-file <下标>` | 删除下标对应的文件规则（下标见 `list-file`）。 |

`<维度>` 是形如 `<命名空间>:<ID>` 的命名空间 ID，例如 `minecraft:overworld`。`add-range` 和 `add-simple` 在维度不存在时会自动创建该维度；其他维度相关命令则会报错。

### 参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-c` | `save-rule.json` | 备份规则文件（JSON）的路径。 |
| `-l` | 关闭 | 按旧版单文件夹布局（`DIM-1`/`DIM1`）备份；同时让 `gencfg` 生成适配旧版的 `file` 列表。 |
| `-color` | 关闭 | 启用彩色输出。 |

## 配置说明

规则文件是 JSON 格式，包含两个顶层字段：`dimension` 和 `file`。既可以手动编辑，也可以用上面的配置命令维护。

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

以维度的命名空间 ID（`<命名空间>:<ID>`）为键。每个维度的规则支持两种选择方式：

- `range` — 矩形区域数组。每一项是 `{ "from": [x, z], "to": [x, z] }`，坐标落在矩形内（含边界）的所有区域文件都会被备份。

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

- `simple` — 单个区域坐标数组，例如 `[x, z]`，用于精确指定个别区域。

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

混用 `range` 和 `simple` 规则。
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

区域坐标与 Minecraft 的区域文件命名一致（`r.<x>.<z>.mca`，一个区域覆盖 512×512 方块）。

### `file`

世界根目录下需要包含的文件或文件夹列表。文件直接加入备份；文件夹会递归遍历其中的全部文件。

## 注意

- 参数必须写在子命令之前。
- 除 `gencfg` 外，所有命令都会先加载配置文件；如果文件不存在，会自动创建一个空配置。
- 不写子命令会直接报错。交互向导请使用 `repl`，内置用法说明请使用 `help`。
- 使用旧版世界模式时，`gencfg -l` 会生成适配旧版的默认 `file` 列表（用 `playerdata` 代替 `players`，并加入 `advancements`）。如果沿用已有的配置文件，请手动调整 `file` 字段。
