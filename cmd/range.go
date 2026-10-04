package main

import (
	"acovia.net/record"
	"acovia.net/minecraft/save"
)

func addRange() {
	if len(arg) < 6{
		record.Error("syntax error, usage: mc-saver add-range <dimension> <from_x> <from_y> <to_x> <to_y>")
	}

	numberList, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("convert string to number:", err)
	}

	newRangeConfig := save.RangeConfig{
		From: save.Coordinate{
			X: numberList[0],
			Y: numberList[1],
		},
		To: save.Coordinate{
			X: numberList[2],
			Y: numberList[3],
		},
	}

	dimension, _ := config.Dimension[arg[1]]
	dimension.Range = append(dimension.Range, newRangeConfig)
	config.Dimension[arg[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delRange() {

	if len(arg) < 3 {
		record.Error("syntax error, usage: mc-saver del-range <dimension> <number> [number]...")
	}

	delList, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("convert string to number:", err)
	}

	dimension, _ := config.Dimension[arg[1]]
	dimension.Range = delSliceElement(dimension.Range, delList...)

	config.Dimension[arg[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modRange() {
	if len(arg) < 7 {
		record.Error("syntax error, usage: mc-saver mod-range <dimension> <number> <from_x> <from_y> <to_x> <to_y>")
	}

	numberList, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("convert string to number", err)
	}

	if numberList[0] < 0 || numberList[0] >= len(config.Dimension[arg[1]].Range) {
		record.Error("number out of range:", numberList[0])
	}

	config.Dimension[arg[1]].Range[numberList[0]] = save.RangeConfig{
		From: save.Coordinate{
			X: numberList[1],
			Y: numberList[2],
		},
		To: save.Coordinate{
			X: numberList[3],
			Y: numberList[4],
		},
	}

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}