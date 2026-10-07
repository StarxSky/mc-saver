package main

import (
	"acovia.net/minecraft/save"
	"acovia.net/record"
	"fmt"
)

func listRange() {
	if len(arg) < 1 {
		record.Error("syntax error, usage: mc-saver list-range <dimension>...")
	}

	_, ok := config.Dimension[arg[0]]
	if !ok {
		record.Error(arg[0]+":", "dimension not found")
	}

	for _, id := range arg {
		rangeConfig := config.Dimension[id].Range
		if len(rangeConfig) == 0 {
			fmt.Printf("no range config for %v\n", id)
			continue
		}
		for i, v := range rangeConfig {
			fmt.Printf("- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
		}
	}
}

func addRange() {
	if len(arg) < 5 {
		record.Error("syntax error, usage: mc-saver add-range <dimension> <from_x> <from_y> <to_x> <to_y>")
	}

	numberList, err := convertIntArray(arg[1:])
	if err != nil {
		record.Error("parse command line args:", err)
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

	dimension, _ := config.Dimension[arg[0]]
	dimension.Range = append(dimension.Range, newRangeConfig)
	config.Dimension[arg[0]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delRange() {

	if len(arg) < 2 {
		record.Error("syntax error, usage: mc-saver del-range <dimension> <number>...")
	}

	_, ok := config.Dimension[arg[0]]
	if !ok {
		record.Error(arg[0]+":", "dimension not found")
	}

	delList, err := convertIntArray(arg[1:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	dimension, _ := config.Dimension[arg[0]]
	dimension.Range, err = deleteSliceElements(dimension.Range, delList...)
	if err != nil {
		record.Error("delete element:", err)
	}

	config.Dimension[arg[0]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modRange() {
	if len(arg) < 6 {
		record.Error("syntax error, usage: mc-saver mod-range <dimension> <number> <from_x> <from_y> <to_x> <to_y>")
	}

	_, ok := config.Dimension[arg[0]]
	if !ok {
		record.Error(arg[0]+":", "dimension not found")
	}

	numberList, err := convertIntArray(arg[1:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	if numberList[0] < 0 || numberList[0] >= len(config.Dimension[arg[0]].Range) {
		record.Error("number out of range:", numberList[0])
	}

	config.Dimension[arg[0]].Range[numberList[0]] = save.RangeConfig{
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
