package main

import (
	"acovia.net/minecraft/save"
	"acovia.net/record"
	"fmt"
)

func listSimple() {
	if len(arg) < 2 {
		record.Error("syntax error, usage: mc-saver list-simple <dimension>...")
	}

	_, ok := config.Dimension[arg[1]]
	if !ok {
		record.Error(arg[1]+":", "dimension not found")
	}

	for _, id := range arg[1:] {
		simpleConfig := config.Dimension[id].Simple
		if len(simpleConfig) == 0 {
			fmt.Printf("no simple config for %v\n", id)
			continue
		}
		for i, v := range simpleConfig {
			fmt.Printf("- %v: (%v, %v)\n", i, v.X, v.Y)
		}
	}
}

func addSimple() {

	if len(arg) < 4 {
		record.Error("syntax error, usage: mc-saver add-simple <dimension> <x> <y>")
	}

	indexSet, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[0],
		Y: indexSet[1],
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
		record.Error("syntax error, usage: mc-saver del-simple <dimension> <number>...")
	}

	dimension, ok := config.Dimension[arg[1]]
	if !ok {
		record.Error(arg[1]+":", "dimension not found")
	}

	var indexSet []int

	indexSet, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	dimension.Simple, err = deleteSliceElements(dimension.Simple, indexSet...)
	if err != nil {
		record.Error("delete element:", err)
	}

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

	_, ok := config.Dimension[arg[1]]
	if !ok {
		record.Error(arg[1]+":", "dimension not found")
	}

	indexSet, err := convertIntArray(arg[2:])
	if err != nil {
		record.Error("parse command line args:", err)
	}

	newCoordinate := save.Coordinate{
		X: indexSet[1],
		Y: indexSet[2],
	}

	if indexSet[0] < 0 || indexSet[0] >= len(config.Dimension[arg[1]].Simple) {
		record.Error("index out of range:", indexSet[0])
	}

	config.Dimension[arg[1]].Simple[indexSet[0]] = newCoordinate

	err = saveConfig()
	if err != nil {
		record.Error("save config:", err)
	}
}
