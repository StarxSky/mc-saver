package main

import (
	"fmt"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func listRange() {
	if len(subCmdArgs) < 2 {
		record.Error("syntax error, usage: mc-saver list-range <world> <dimension>...")
	}

	initWorldConfig()
	_, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		record.Error(subCmdArgs[1]+":", "dimension not found")
	}

	for _, id := range subCmdArgs[1:] {
		rangeConfig := config.Dimension[id].Range
		if len(rangeConfig) == 0 {
			fmt.Printf("no range config for %v\n", id)
			continue
		}
		fmt.Printf("range config for %v:\n", id)
		for i, v := range rangeConfig {
			fmt.Printf("- %v: from: (%v, %v) to: (%v, %v)\n", i, v.From.X, v.From.Y, v.To.X, v.To.Y)
		}
	}
}

func addRange() {
	if len(subCmdArgs) < 6 {
		record.Error("syntax error, usage: mc-saver add-range <world> <dimension> <from_x> <from_y> <to_x> <to_y>")
	}

	initWorldConfig()
	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	newRangeConfig := save.RangeConfig{
		From: save.Coordinate{
			X: indexSet[0],
			Y: indexSet[1],
		},
		To: save.Coordinate{
			X: indexSet[2],
			Y: indexSet[3],
		},
	}

	dimension, _ := config.Dimension[subCmdArgs[1]]
	dimension.Range = append(dimension.Range, newRangeConfig)
	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delRange() {
	if len(subCmdArgs) < 3 {
		record.Error("syntax error, usage: mc-saver del-range <world> <dimension> <number>...")
	}

	initWorldConfig()
	_, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		record.Error(subCmdArgs[1]+":", "dimension not found")
	}

	delList, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	dimension, _ := config.Dimension[subCmdArgs[1]]
	dimension.Range, err = deleteSliceElements(dimension.Range, delList...)
	if err != nil {
		record.Error("delete element:", err)
	}

	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modRange() {
	if len(subCmdArgs) < 7 {
		record.Error("syntax error, usage: mc-saver mod-range <world> <dimension> <number> <from_x> <from_y> <to_x> <to_y>")
	}

	initWorldConfig()
	_, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		record.Error(subCmdArgs[1]+":", "dimension not found")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	if indexSet[0] < 0 || indexSet[0] >= len(config.Dimension[subCmdArgs[1]].Range) {
		record.Error("number out of range:", indexSet[0])
	}

	config.Dimension[subCmdArgs[1]].Range[indexSet[0]] = save.RangeConfig{
		From: save.Coordinate{
			X: indexSet[1],
			Y: indexSet[2],
		},
		To: save.Coordinate{
			X: indexSet[3],
			Y: indexSet[4],
		},
	}

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}
