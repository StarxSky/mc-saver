package main

import (
	"acovia.net/record"
	"strconv"
)
func addFile() {
	if len(arg) < 2 {
		record.Error("syntax error, usage: mc-saver add-file <file_name> [file_name]...")
	}

	config.File = append(config.File, arg[1:]...)

	err := saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delFile() {

	if len(arg) < 2 {
		record.Error("syntax error, usage: mc-saver del-file <number> [number]...")
	}

	numberList, err := convertIntArray(arg[1:])
	if err != nil {
		record.Error("convert string to number:", err)
	}

	config.File = delSliceElement(config.File, numberList...)

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
		record.Error("convert string to number", err)
	}

	if number < 0 || int(number) >= len(config.File) {
		record.Error("number out of range:", number)
	}

	config.File[number] = arg[2]
	
	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}