package main

import (
	"fmt"

	"acovia.net/minecraft/save"
	"acovia.net/record"
)

func listSimple() {
	if len(subCmdArgs) < 2 {
		record.Error("syntax error, usage: mc-saver list-simple <world> <dimension>...")
	}

	for _, id := range subCmdArgs[1:] {
		_, ok := config.Dimension[id]
		if !ok {
			record.Error(id+":", "dimension not found")
		}

		simpleConfig := config.Dimension[id].Simple
		if len(simpleConfig) == 0 {
			fmt.Printf("no simple config for %v\n", id)
			continue
		}
		fmt.Printf("simple config for %v:\n", id)
		for i, v := range simpleConfig {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}
}

func addSimple() {
	if len(subCmdArgs) < 4 {
		record.Error("syntax error, usage: mc-saver add-simple <world> <dimension> <x> <y>")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[0],
		Y: indexSet[1],
	}

	dimension, _ := config.Dimension[subCmdArgs[1]]
	dimension.Simple = append(dimension.Simple, newCoordinate)

	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func delSimple() {
	if len(subCmdArgs) < 3 {
		record.Error("syntax error, usage: mc-saver del-simple <world> <dimension> <number>...")
	}

	dimension, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		record.Error(subCmdArgs[1]+":", "dimension not found")
	}

	var indexSet []int

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	dimension.Simple, err = deleteSliceElements(dimension.Simple, indexSet...)
	if err != nil {
		record.Error("delete element:", err)
	}

	config.Dimension[subCmdArgs[1]] = dimension

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}

func modSimple() {
	if len(subCmdArgs) < 5 {
		record.Error("syntax error, usage: mc-saver mod-simple <world> <dimension> <number> <x> <y>")
	}

	_, ok := config.Dimension[subCmdArgs[1]]
	if !ok {
		record.Error(subCmdArgs[1]+":", "dimension not found")
	}

	indexSet, err := convertIntArray(subCmdArgs[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[1],
		Y: indexSet[2],
	}

	if indexSet[0] < 0 || indexSet[0] >= len(config.Dimension[subCmdArgs[1]].Simple) {
		record.Error("index out of range:", indexSet[0])
	}

	config.Dimension[subCmdArgs[1]].Simple[indexSet[0]] = newCoordinate

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}
