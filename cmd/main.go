package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func main() {
	initProgram()

	configFilePath = strings.ReplaceAll(configFilePath, "\\", "/")
	function, ok := cmdMap[cmd]
	if !ok {
		record.Error(errors.New("'" + cmd + "' command not found"))
	}

	function()
}

func help() {
	fmt.Printf("%v", helpInfo)
}

func initProgram() {
	flag.StringVar(&argConfigPath, "c", argConfigPath, "config file path")
	flag.BoolFunc("l", "legacy world mode", func(s string) error {
		useLegacyMode = true
		return nil
	})
	flag.BoolVar(&record.EnableColor, "color", false, "enable color output")
	flag.Parse()
	args = flag.Args()

	if len(args) < 1 {
		record.Error(errors.New("command is missing"))
	}

	cmd = args[0]

	subCmdArgs = loadSubCmdArgs()
	configFilePath = loadConfigFilePath()

	initConfig()
}

func initConfig() {
	var err error
	if ok, _ := noInitConfigCmd[cmd]; ok {
		return
	}

	config, err = save.LoadConfig(configFilePath)
	if err != nil {
		record.Error("load config:", err)
	}

	_, err = os.Stat(configFilePath)
	if !os.IsNotExist(err) {
		return
	}
	
	record.Info("config file:", configFilePath, "not found, generate a empty config file")
	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func gencfg() {
	if len(subCmdArgs) < 1 {
		record.Error("syntax error, usage: mc-saver run <world> [output]")
	}

	if len(subCmdArgs) > 1 {
		configFilePath = subCmdArgs[1]
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

func saveWorld() {
	var err error

	absPath, err := filepath.Abs(worldDirPath)
	if err != nil {
		record.Error("load abs path:", err)
	}

	worldDirPath = absPath

	if len(subCmdArgs) >= 3 {
		outputPath = path.Clean(subCmdArgs[2])
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

	if useLegacyMode {
		if err := save.SaveOldAllFile(root, config, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	} else {
		if err := save.SaveAllFile(root, config, zipWriter, saveFile); err != nil {
			if err := end(err); err != nil {
				record.Error(err)
			}
		}
	}

	if err := end(nil); err != nil {
		record.Error("close file writer:", err)
	}
}

func addDms() {
	if len(subCmdArgs) < 2 {
		record.Error("syntax error, usage: mc-saver add-dms <world> <dimension>...")
	}

	for _, v := range subCmdArgs[1:] {
		_, ok := config.Dimension[v]
		if ok {
			record.Warn(v+":", "dimension existed, skip")
			continue
		}
		config.Dimension[v] = defaultDimensionConfig
	}

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delDms() {
	if len(subCmdArgs) < 2 {
		record.Error("syntax error, usage: mc-saver del-dms <world> <dimension>...")
	}

	for _, v := range subCmdArgs[1:] {
		_, ok := config.Dimension[v]
		if !ok {
			record.Warn(v+":", "dimension not found, skip")
			continue
		}
		delete(config.Dimension, v)
	}

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modDms() {
	if len(subCmdArgs) < 3 {
		record.Error("syntax error, usage: mc-saver mod-dms <world> <old_dimension> <dimension>")
	}

	if _, ok := config.Dimension[subCmdArgs[1]]; !ok {
		record.Error(subCmdArgs[1]+":", "dimension not found")
	}

	if subCmdArgs[1] == subCmdArgs[2] {
		record.Error("dimension no change")
	}

	config.Dimension[subCmdArgs[2]] = config.Dimension[subCmdArgs[1]]
	delete(config.Dimension, subCmdArgs[1])

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func listConfig() {
	if len(subCmdArgs) < 2 {
		record.Error("syntax error, usage: mc-saver list-config <world> <dimension>...")
	}
	listRange()
	listSimple()
}

func list() {
	if len(subCmdArgs) < 1 {
		record.Error("syntax error, usage: mc-saver list-config <world> <dimension>...")
	}

	if len(config.Dimension) == 0 {
		fmt.Println("no dimension config at all")
	}

	for id, rule := range config.Dimension {
		fmt.Printf("range config for %v:\n", id)
		if len(rule.Range) == 0 {
			fmt.Printf("no range config for %v\n", id)
		}
		for i, v := range rule.Range {
			fmt.Printf("- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
		}

		fmt.Printf("simple config for %v:\n", id)
		if len(rule.Simple) == 0 {
			fmt.Printf("no simple config for %v\n", id)
		}
		for i, v := range rule.Simple {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}

	fmt.Println("file config:")
	listFile()
}

func listDms() {
	for id := range config.Dimension {
		fmt.Printf("- %v\n", id)
	}
}
