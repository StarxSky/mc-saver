package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"acovia.net/mc-saver/parse"
	"acovia.net/record"
)

var (
	useLegacyMode  bool   = false
	configFilePath string = "save-rule.json"
	worldDirPath   string = "world"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":    run,
		"gencfg": gencfg,
		"help":   help,
		"repl":   repl,
		"":       repl,
	}

	root *os.Root

	defaultConfig string = `{
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
}`
)

func main() {

	initProgram()

	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")

	function, ok := cmdMap[flag.Arg(0)]
	if !ok {
		record.Error(errors.New("'" + flag.Arg(0) + "' command not found."))
	}

	function()
}

func help() {
	helpOutput :=
		`command:

	run [option] [world_path] [output_path].
		start the backup according to the config file.
		the first path is world path, default is "world".
		second path is output path, default is "world-$time.zip".

	gencfg [output_file]
		generate a default config file.

options:

	-c <path>
		specify the config file.

	-color <bool>
		enable color output.
`
	fmt.Printf("%v", helpOutput)
}

func initProgram() {
	flag.StringVar(&configFilePath, "c", configFilePath, "config file path.")
	flag.BoolFunc("l", "legacy world mode.", func(s string) error {
		useLegacyMode = true
		return nil
	})
	flag.BoolFunc("color", "enable color output.", func(s string) error {
		record.EnableColor = true
		return nil
	})
	flag.Parse()
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Printf("generate a default config file? (y/n): ")
		if scanner.Scan() {
			if input := scanner.Text(); input != "y" && input != "Y" && len(input) != 0 {
				os.Exit(0)
			}
			err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
			if err != nil {
				record.Error(err)
			}
			record.Info("created default config file:", configFilePath)
			fmt.Printf("continue with default config file? (y/n): ")
			if scanner.Scan() {
				input := scanner.Text()
				if input != "y" && input != "Y" && len(input) != 0 {
					os.Exit(0)
				}
			}
		}
	} else if !os.IsNotExist(err) && err != nil {
		record.Error("verify config file:", err)
	}
	fmt.Printf("world directory path (default is world): ")
	if scanner.Scan() {
		input := scanner.Text()
		input = strings.ReplaceAll(input, "\\", "/")
		if len(input) != 0 {
			worldDirPath = input
		}
		if absPath, err := filepath.Abs(worldDirPath); err != nil {
			record.Error("check world directory path):", err)
		} else {
			worldDirPath = absPath
		}
	}
	fmt.Printf("output path (default is world-$time.zip): ")
	if scanner.Scan() {
		input := scanner.Text()
		input = strings.ReplaceAll(input, "\\", "/")
		if len(input) != 0 {
			outputPath = path.Clean(input)
		}
	}
	run()
}

func gencfg() {
	if flag.Arg(1) != "" {
		configFilePath = flag.Arg(1)
	}

	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		record.Error("generate config file:", "'"+configFilePath+"'", "already exists", configFilePath)
	}

	err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644)
	if err != nil {
		record.Error(err)
	}
	record.Info("created config file:", configFilePath)
	os.Exit(0)
}

func run() {
	var err error

	if len(flag.Arg(1)) != 0 {
		if absPath, err := filepath.Abs(flag.Arg(1)); err != nil {
			record.Error("check world directory:", err)
		} else {
			worldDirPath = absPath
		}
	}

	if len(flag.Arg(2)) != 0 {
		outputPath = path.Clean(flag.Arg(2))
	}

	worldDirPath = strings.ReplaceAll(worldDirPath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	root, err = os.OpenRoot(worldDirPath)
	if err != nil {
		record.Error("open world directory:", err)
	}

	zipWriter, _, end, err := newZipWriter()
	if err != nil {
		record.Error("create zip writer:", err)
	}

	defer end(err)

	if useLegacyMode {
		if err := parse.SaveOldAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	} else {
		if err := parse.SaveAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	}
}