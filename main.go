package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"acovia.net/minecraft/saver"
	"acovia.net/record"
)

var (
	arg []string

	useLegacyMode  bool   = false
	configFilePath string = "save-rule.json"
	worldDirPath   string = "world"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":     run,
		"gencfg":  gencfg,
		"help":    help,
		"repl":    repl,
		"add-dms": addDms,
		"del-dms": delDms,
		"":        repl,
	}

	root *os.Root

	defaultDimensionConfig = saver.DimensionConfig{
		Range: []saver.RangeConfig{
			{
				From: saver.Coordinate{
					X: -1,
					Y: -1,
				},
				To: saver.Coordinate{
					X: 0,
					Y: 0,
				},
			},
		},
	}

	defaultConfig = saver.Config{
		Dimension: map[string]saver.DimensionConfig{
			"minecraft:overworld":  defaultDimensionConfig,
			"minecraft:the_nether": defaultDimensionConfig,
			"minecraft:the_end":    defaultDimensionConfig,
		},
		File: []string{
			"level.dat",
			"data",
			"datapacks",
			"players",
		},
	}
)

func main() {
	initProgram()
	arg = flag.Args()

	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")

	function, ok := cmdMap[arg[0]]
	if !ok {
		record.Error(errors.New("'" + arg[0] + "' command not found."))
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
		record.Error("config file not found, please run 'mc-saver gencfg' to generate a default config.")
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
	if len(arg) > 1 {
		configFilePath = arg[1]
	}

	if useLegacyMode {
		defaultConfig.File = []string{
			"level.dat",
			"data",
			"datapacks",
			"advancements",
			"playerdata",
		}
	}

	if _, err := os.Stat(configFilePath); !os.IsNotExist(err) {
		record.Error("generate config file:", "'"+configFilePath+"'", "already exists", configFilePath)
	}

	jsonData, err := json.MarshalIndent(defaultConfig, "", "	")
	if err != nil {
		record.Error("load default config struct:", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		record.Error(err)
	}
	record.Info("created config file:", configFilePath)
	os.Exit(0)
}

func run() {
	var err error

	if len(arg) > 1 {
		if absPath, err := filepath.Abs(arg[1]); err != nil {
			record.Error("load abs path:", err)
		} else {
			worldDirPath = absPath
		}
	}

	if len(arg) > 2 {
		outputPath = path.Clean(flag.Arg(2))
	}

	worldDirPath = strings.ReplaceAll(worldDirPath, "\\", "/")
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	root, err = os.OpenRoot(worldDirPath)
	if err != nil {
		record.Error("open world directory:", err)
	}

	zipWriter, _, end, err := initZipWriter()
	if err != nil {
		record.Error("init zip writer:", err)
	}

	defer end(nil)

	if useLegacyMode {
		if err := saver.SaveOldAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	} else {
		if err := saver.SaveAllFile(root, configFilePath, zipWriter, addFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	}
}

func addDms() {
	var dimensionNamespaceID string

	if len(arg) > 1 {
		dimensionNamespaceID = arg[1]
	} else {
		record.Error("dimension namespace id is missing.")
	}

	config, err := saver.LoadConfig(configFilePath)
	if err != nil {
		record.Error("load config:", err)
	}

	config.Dimension[dimensionNamespaceID] = defaultDimensionConfig

	jsonData, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		record.Error("encode json data:", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		record.Error(err)
	}
}

func delDms() {
	var dimensionNamespaceID string

	if len(arg) > 1 {
		dimensionNamespaceID = arg[1]
	} else {
		record.Error("dimension namespace id is missing.")
	}

	config, err := saver.LoadConfig(configFilePath)
	if err != nil {
		record.Error("load config:", err)
	}

	delete(config.Dimension, dimensionNamespaceID)

	jsonData, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		record.Error("encode json data:", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		record.Error(err)
	}
}
