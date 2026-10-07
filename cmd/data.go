package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"acovia.net/minecraft/save"
)

var (
	args           []string
	cmd            string
	config         save.Config = save.NullConfig
	configFilePath string
	argConfigPath  string
	subCmdArgs     []string

	useLegacyMode  bool   = false
	worldDirPath   string = "world"
	configFileName string = "saver.json"
	outputPath     string = "."

	cmdMap map[string]func() = map[string]func(){
		"run":         run,
		"gencfg":      gencfg,
		"help":        help,
		"list":        list,
		"list-config": listConfig,
		"list-dms":    listDms,
		"add-dms":     addDms,
		"del-dms":     delDms,
		"mod-dms":     modDms,
		"list-range":  listRange,
		"add-range":   addRange,
		"del-range":   delRange,
		"mod-range":   modRange,
		"list-simple": listSimple,
		"add-simple":  addSimple,
		"del-simple":  delSimple,
		"mod-simple":  modSimple,
		"list-file":   listFile,
		"add-file":    addFile,
		"del-file":    delFile,
		"mod-file":    modFile,
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

	helpInfo = `mc-saver [-c <config_file>] [-l] [-color] <command> [args...]

backup command:

	run [world_path] [output_path]
		start the backup according to the config file.
		the first path is world path, default is "world".
		second path is output path, default is "world-$time.zip".

	gencfg [config_file]
		generate a default config file, default is "save-rule.json".
		throw error if the file is already existed.

	help
		print this help text.

config command:

	list
		list all dimension rules and file rules.

	list-config <dimension>...
		list both range rules and simple rules of a dimension.

	list-dms
		list dimension namespace ids.

	add-dms <dimension>...
		add a dimension with the default range rule.

	del-dms <dimension>...
		delete a dimension.

	mod-dms <old_dimension> <new_dimension>
		rename a dimension, keeping its rules.

	list-range <dimension>...
		list range rules of a dimension.

	add-range <dimension> <from_x> <from_y> <to_x> <to_y>
		add a range rule to a dimension.

	del-range <dimension> <index>...
		delete the range rules of the given indices.

	mod-range <dimension> <index> <from_x> <from_y> <to_x> <to_y>
		replace the range rule at the given index.

	list-simple <dimension>...
		list simple rules of a dimension.

	add-simple <dimension> <x> <y>
		add a simple rule to a dimension.

	del-simple <dimension> <index>...
		delete the simple rules of the given indices.

	mod-simple <dimension> <index> <x> <y>
		replace the simple rule at the given index.

	list-file
		list file rules.

	add-file <name>...
		add one or more file rules.

	del-file <index>...
		delete the file rules of the given indices.

	mod-file <index> <name>
		replace the file rule at the given index.

options:

	-c <path>
		specify the config file, default is "save-rule.json".

	-l
		legacy world mode, for worlds from before 1.21.11.

	-color
		enable color output.
`
)

func formatOutPutPath(archiveFilePath string) (string, error) {

	worldAbsPath, err := filepath.Abs(archiveFilePath)
	if err != nil {
		return "", fmt.Errorf("load absolute path: %w", err)
	}

	worldDirName := path.Base(worldAbsPath)

	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return "", fmt.Errorf("create output directory: %v", err)
		}
		archiveFilePath = outputPath

	case !os.IsNotExist(err) && err != nil:
		return "", fmt.Errorf("read file info: %v", err)

	case outputFileInfo.IsDir():
		archiveFileName := worldDirName + "-" + time.Now().Format(time.DateOnly) + ".zip"
		archiveFilePath = path.Join(outputPath, archiveFileName)
		archiveFilePath, err = addSubfixBeforeExt(archiveFilePath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}

	default:
		archiveFilePath, err = addSubfixBeforeExt(outputPath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}
	}

	return archiveFilePath, nil
}

func addSubfixBeforeExt(archiveFilePath string) (string, error) {
	basePath := path.Base(archiveFilePath)
	dirPath := path.Dir(archiveFilePath)
	nameArr := strings.FieldsFunc(basePath, isExtKeyWord)
	for number := 1; ; number++ {
		var subfix string = "-" + fmt.Sprint(number)
		basePath = nameArr[0] + subfix
		for _, ext := range nameArr[1:] {
			basePath += "." + ext
		}
		stat, err := os.Stat(path.Join(dirPath, basePath))
		if os.IsNotExist(err) {
			break
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("read file info: %v", err)
		} else if !stat.IsDir() || !os.IsNotExist(err) {
			continue
		}
	}
	return path.Join(dirPath, basePath), nil
}

func isExtKeyWord(char rune) bool {
	if char == rune("."[0]) {
		return true
	} else {
		return false
	}
}

func deleteSliceElements[T comparable](arr []T, index ...int) ([]T, error) {
	for _, v := range index {
		if v < 0 || v >= len(arr) {
			return nil, fmt.Errorf("index out of range: %v", v)
		}
	}

	var newArr []T
	for n, e := range arr {
		canAppend := true
		for _, i := range index {
			if n == i {
				canAppend = false
				break
			}
		}
		if canAppend {
			newArr = append(newArr, e)
		}
	}

	return newArr, nil
}

func convertIntArray(array []string) ([]int, error) {
	var intList []int
	for _, v := range array {
		number, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, err
		}
		intList = append(intList, int(number))
	}
	return intList, nil
}

func loadConfigFilePath() string {
	if len(argConfigPath) == 0 {
		return path.Join(worldDirPath, configFileName)
	}
	return argConfigPath
}

func loadSubCmdArgs() []string {
	if len(args) > 1 {
		return args[1:]
	}
	return nil
}