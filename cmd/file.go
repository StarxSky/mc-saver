package main

import (
	"fmt"
	"strconv"

	"acovia.net/record"
)

func listFile() {
	if len(config.File) == 0 {
		fmt.Println("no file config")
	}

	for i, v := range config.File {
		fmt.Printf("- %v: %q\n", i, v)
	}
}

func addFile() {
	if len(args) < 3 {
		record.Error("syntax error, usage: mc-saver add-file <world> <file_name>...")
	}

	config.File = append(config.File, subCmdArgs[1:]...)

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delFile() {
	if len(args) < 3 {
		record.Error("syntax error, usage: mc-saver del-file <world> <number>...")
	}

	indexSet, err := convertIntArray(subCmdArgs[1:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	config.File, err = deleteSliceElements(config.File, indexSet...)
	if err != nil {
		record.Error("delete element:", err)
	}

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modFile() {
	if len(args) < 4 {
		record.Error("syntax error, usage: mc-saver mod-file <world> <number> <file_name>")
	}

	index, err := strconv.ParseInt(subCmdArgs[1], 10, 64)
	if err != nil {
		record.Error("parse command line args:", err)
	}

	if index < 0 || int(index) >= len(config.File) {
		record.Error("index out of range:", index)
	}

	config.File[index] = subCmdArgs[2]

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}