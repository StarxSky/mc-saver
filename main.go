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

	"acovia.net/mc-saver/parse"
	"acovia.net/record"
)

var (
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

	defaultDimensionConfig = parse.DimensionConfig{
		Range: []parse.RangeConfig{
			{
				From: parse.Coordinate{
					X: -1,
					Y: -1,
				},
				To: parse.Coordinate{
					X: 0,
					Y: 0,
				},
			},
		},
	}

	defaultConfig = parse.Config{
		Dimension: map[string]parse.DimensionConfig{
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
	if flag.Arg(1) != "" {
		configFilePath = flag.Arg(1)
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

	if len(flag.Arg(1)) != 0 {
		if absPath, err := filepath.Abs(flag.Arg(1)); err != nil {
			record.Error("load abs path:", err)
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

	zipWriter, _, end, err := initZipWriter()
	if err != nil {
		record.Error("init zip writer:", err)
	}

	defer end(nil)

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

func addDms() {
	var dimensionNamespaceID string

	if len(flag.Arg(1)) != 0 {
		dimensionNamespaceID = flag.Arg(1)
	} else {
		record.Error("dimension namespace id is missing.")
	}

	config, err := parse.LoadRootSaveRule(configFilePath)
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

	if len(flag.Arg(1)) != 0 {
		dimensionNamespaceID = flag.Arg(1)
	} else {
		record.Error("dimension namespace id is missing.")
	}

	config, err := parse.LoadRootSaveRule(configFilePath)
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
