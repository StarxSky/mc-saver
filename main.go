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
	"strconv"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

var (
	arg    []string
	config save.Config = save.Config{
		Dimension: map[string]save.DimensionConfig{},
	}

	useLegacyMode  bool   = false
	configFilePath string = "save-rule.json"
	worldDirPath   string = "world"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":         run,
		"gencfg":      gencfg,
		"help":        help,
		"repl":        repl,
		"add-dms":     addDms,
		"del-dms":     delDms,
		"list":        list,
		"list-range":  listRange,
		"list-simple": listSimple,
		"list-config": listConfig,
		"list-dms":    listDms,
		"list-file":   listFile,
		"add-range":   addRange,
		"del-range":   delRange,
		"add-simple":  addSimple,
		"del-simple":  delSimple,
		"add-file":    addFile,
		"del-file":    delFile,
	}

	root *os.Root

	defaultDimensionConfig = save.DimensionConfig{
		Range: []save.RangeConfig{
			{
				From: save.Coordinate{
					X: -1,
					Y: -1,
				},
				To: save.Coordinate{
					X: 0,
					Y: 0,
				},
			},
		},
	}

	defaultConfig = save.Config{
		Dimension: map[string]save.DimensionConfig{
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
	if len(arg) == 0 {
		record.Error(errors.New("command is missing."))
	}
	function, ok := cmdMap[arg[0]]
	if !ok {
		record.Error(errors.New("'" + arg[0] + "' command not found."))
	}

	function()
}

func help() {
	helpOutput :=
		`mc-saver [-c <config_file>] [-l] [-color] <command> [args...]

backup command:

	run [world_path] [output_path]
		start the backup according to the config file.
		the first path is world path, default is "world".
		second path is output path, default is "world-$time.zip".

	gencfg [config_file]
		generate a default config file, default is "save-rule.json".
		throw error if the file is already existed.

	repl
		run the interactive wizard.

	help
		print this help text.

config command:

	the config file is loaded before the command runs, and written back
	after it finished. dimension is a namespace id like
	"minecraft:overworld".

	add-dms <dimension>
		add a dimension with the default range rule.

	del-dms <dimension>
		delete a dimension.

	list
		list all dimension rules and file rules.

	list-dms
		list dimension namespace ids.

	list-range <dimension>
		list range rules of a dimension.

	list-simple <dimension>
		list simple rules of a dimension.

	list-config <dimension>
		list both range rules and simple rules of a dimension.

	list-file
		list file rules.

	add-range <dimension> <from_x> <from_y> <to_x> <to_y>
		add a range rule to a dimension.

	del-range <dimension> <number>
		delete the range rule of the given index.

	add-simple <dimension> <x> <y>
		add a simple rule to a dimension.

	del-simple <dimension> <number>
		delete the simple rule of the given index.

	add-file <name> [name...]
		add one or more file rules.

	del-file <number>
		delete the file rule of the given index.

options:

	-c <path>
		specify the config file, default is "save-rule.json".

	-l
		legacy world mode, for worlds from before 1.21.11.

	-color
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

	arg = flag.Args()

	if len(arg) != 0 && arg[0] != "gencfg" {
		configTemp, err := save.LoadConfig(configFilePath)
		if err == nil {
			config = configTemp
		} else if !isExist(configFilePath) {
			saveConfig()
		} else {
			record.Error("load config:", err)
		}
	}
}

func repl() {
	scanner := bufio.NewScanner(os.Stdin)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		record.Error("config file not found, please run 'mc-save gencfg' to generate a default config.")
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
		if err := save.SaveOldAllFile(root, configFilePath, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	} else {
		if err := save.SaveAllFile(root, configFilePath, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	}
}

func addDms() {
	if len(arg) < 2 {
		record.Error("dimension namespace id is missing.")
	}

	config.Dimension[arg[1]] = defaultDimensionConfig

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delDms() {
	if len(arg) < 2 {
		record.Error("dimension namespace id is missing.")
	}

	delete(config.Dimension, arg[1])

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func listRange() {
	if len(arg) < 2 {
		record.Error("dimension is missing.")
	}

	if dimension, ok := config.Dimension[arg[1]]; ok {
		if len(dimension.Range) == 0 {
			fmt.Printf("no range config for %v.\n", arg[1])
		}
		for i, v := range dimension.Range {
			fmt.Printf("- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
		}
	} else {
		record.Error(arg[1]+":", "dimension not found.")
	}
}

func listSimple() {
	if len(arg) < 2 {
		record.Error("dimension is missing.")
	}

	if dimension, ok := config.Dimension[arg[1]]; ok {
		if len(dimension.Simple) == 0 {
			fmt.Printf("no simple config for %v.\n", arg[1])
		}
		for i, v := range dimension.Simple {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	} else {
		record.Error(arg[1]+":", "dimension not found.")
	}
}

func listConfig() {
	if len(arg) < 2 {
		record.Error("dimension is missing.")
	}
	fmt.Printf("range config for %v:\n", arg[1])
	listRange()
	fmt.Printf("simple config for %v:\n", arg[1])
	listSimple()
}

func list() {
	if len(config.Dimension) == 0 {
		fmt.Println("no dimension config at all.")
	}

	for id, rule := range config.Dimension {
		fmt.Printf("range config for %v:\n", id)
		if len(rule.Range) == 0 {
			fmt.Printf("no range config for %v.\n", id)
		}
		for i, v := range rule.Range {
			fmt.Printf("- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
		}

		fmt.Printf("simple config for %v:\n", id)
		if len(rule.Simple) == 0 {
			fmt.Printf("no simple config for %v.\n", id)
		}
		for i, v := range rule.Simple {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}

	fmt.Println("file config:")
	listFile()
}

func listFile() {
	if len(config.File) == 0 {
		fmt.Println("no file config.")
	}
	for i, v := range config.File {
		fmt.Printf("- %v: %q\n", i, v)
	}
}

func listDms() {
	for id := range config.Dimension {
		fmt.Printf("- %v\n", id)
	}
}

func addRange() {
	switch true {
	case len(arg) < 2:
		record.Error("dimension is missing.")

	case len(arg) < 6:
		record.Error("range is missing.")
	}

	fromX, err := strconv.ParseInt(arg[2], 10, 32)
	if err != nil {
		record.Error("parse input coordinate.", err)
	}
	fromY, err := strconv.ParseInt(arg[3], 10, 32)
	if err != nil {
		record.Error("parse input coordinate.", err)
	}
	toX, err := strconv.ParseInt(arg[4], 10, 32)
	if err != nil {
		record.Error("parse input coordinate.", err)
	}
	toY, err := strconv.ParseInt(arg[5], 10, 32)
	if err != nil {
		record.Error("parse input coordinate.", err)
	}

	newRange := save.RangeConfig{
		From: save.Coordinate{
			X: int(fromX),
			Y: int(fromY),
		},
		To: save.Coordinate{
			X: int(toX),
			Y: int(toY),
		},
	}

	dimension, ok := config.Dimension[arg[1]]
	if ok {
		dimension.Range = append(dimension.Range, newRange)
	} else {
		dimension = save.DimensionConfig{
			Range: []save.RangeConfig{
				newRange,
			},
		}
	}
	config.Dimension[arg[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delRange() {
	switch true {
	case len(arg) < 2:
		record.Error("dimension is missing.")

	case len(arg) < 3:
		record.Error("number is missing.")
	}

	number, err := strconv.ParseInt(arg[2], 10, 32)
	if err != nil {
		record.Error("parse input number.", err)
	}

	dimension, ok := config.Dimension[arg[1]]
	if ok {
		if int(number) >= len(dimension.Range) {
			record.Error(number, "is out of range", len(dimension.Range)-1)
		}
		dimension.Range = append(dimension.Range[:number], dimension.Range[number+1:]...)
	} else {
		record.Error(arg[1]+":", "dimension not found.")
	}

	config.Dimension[arg[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func addSimple() {
	switch true {
	case len(arg) < 2:
		record.Error("dimension is missing.")

	case len(arg) < 4:
		record.Error("coordinate is missing.")
	}

	x, err := strconv.ParseInt(arg[2], 10, 32)
	if err != nil {
		record.Error("parse input number.", err)
	}
	y, err := strconv.ParseInt(arg[3], 10, 32)
	if err != nil {
		record.Error("parse input number.", err)
	}

	newCoordinate := save.Coordinate{
		X: int(x),
		Y: int(y),
	}

	dimension, ok := config.Dimension[arg[1]]
	if ok {
		dimension.Simple = append(dimension.Simple, newCoordinate)
	} else {
		dimension = save.DimensionConfig{
			Simple: []save.Coordinate{
				newCoordinate,
			},
		}
	}

	config.Dimension[arg[1]] = dimension
	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delSimple() {
	switch true {
	case len(arg) < 2:
		record.Error("dimension is missing.")

	case len(arg) < 3:
		record.Error("number is missing.")
	}

	number, err := strconv.ParseInt(arg[2], 10, 32)
	if err != nil {
		record.Error("parse input number.", err)
	}

	dimension, ok := config.Dimension[arg[1]]
	if ok {
		if int(number) >= len(dimension.Simple) {
			record.Error(number, "is out of range", len(dimension.Simple)-1)
		}
		dimension.Simple = append(dimension.Simple[:number], dimension.Simple[number+1:]...)
	} else {
		record.Error(arg[1]+":", "dimension not found.")
	}

	config.Dimension[arg[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func addFile() {
	if len(arg) < 2 {
		record.Error("filename is missing.")
	}

	config.File = append(config.File, arg[1:]...)

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delFile() {
	if len(arg) < 2 {
		record.Error("number is missing.")
	}

	number, err := strconv.ParseInt(arg[1], 10, 32)
	if err != nil {
		record.Error("parse command line flag:", err)
	}
	config.File = append(config.File[:number], config.File[number+1:]...)

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}
