package main

import (
	"acovia.net/record"
	"acovia.net/minecraft/save"
)

func addSimple() {

	if len(arg) < 4 {
		record.Error("syntax error, usage: mc-saver add-simple <dimension> <x> <y>")
	}

	numberList, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("convert string to number:", err)
	}

	newCoordinate := save.Coordinate{
		X: numberList[0],
		Y: numberList[1],
	}

	dimension, _ := config.Dimension[arg[1]]
	dimension.Simple = append(dimension.Simple, newCoordinate)

	config.Dimension[arg[1]] = dimension
	
	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delSimple() {
	if len(arg) < 3 {
		record.Error("syntax error, usage: mc-saver del-simple <dimension> <number> [number]...")
	}

	var delList []int

	dimension, ok := config.Dimension[arg[1]]
	if !ok {
		record.Error(arg[1]+":", "dimension not found.")
	}

	delList, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("convert string to number:", err)
	}

	dimension.Simple = delSliceElement(dimension.Simple, delList...)
	config.Dimension[arg[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modSimple() {
	if len(arg) < 5 {
		record.Error("syntax error, usage: mc-saver mod-simple <dimension> <number> <x> <y>")
	}

	numberList, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("convert string to number:", err)
	}

	newCoordinate := save.Coordinate{
		X: numberList[1],
		Y: numberList[2],
	}

	if numberList[0] < 0 || numberList[0] >= len(config.Dimension[arg[1]].Simple) {
		record.Error("number out of range:", numberList[0])
	}

	config.Dimension[arg[1]].Simple[numberList[0]] = newCoordinate

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}