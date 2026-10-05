package main

import (
	"acovia.net/record"
	"fmt"
	"strconv"
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
	if len(arg) < 2 {
		record.Error("syntax error, usage: mc-saver add-file <file_name>...")
	}

	config.File = append(config.File, arg[1:]...)

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delFile() {

	if len(arg) < 2 {
		record.Error("syntax error, usage: mc-saver del-file <number>...")
	}

	indexSet, err := convertIntArray(arg[1:])
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
	if len(arg) < 3 {
		record.Error("syntax error, usage: mc-saver mod-file <number> <file_name>")
	}

	number, err := strconv.ParseInt(arg[1], 10, 64)
	if err != nil {
		record.Error("parse command line args:", err)
	}

	if number < 0 || int(number) >= len(config.File) {
		record.Error("index out of range:", number)
	}

	config.File[number] = arg[2]

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}
