package saver

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

type RangeIterator struct {
	FromX int
	FromY int
	ToX   int
	ToY   int
}

type Config struct {
	Dimension map[string]DimensionConfig `json:"dimension"`
	File      []string                   `json:"file"`
}

type DimensionConfig struct {
	Range  []RangeConfig `json:"range"`
	Simple []Coordinate  `json:"simple"`
}

type RangeConfig struct {
	From Coordinate `json:"from"`
	To   Coordinate `json:"to"`
}

type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var (
	rootFile = []string{
		"region",
		"entities",
		"poi",
	}
)

type addFile func(fileName string, zipWriter *zip.Writer) error

func (iterator RangeIterator) run(function func(x int, y int) error) error {
	if iterator.FromX > iterator.ToX {
		invertedIntValue(&iterator.FromX, &iterator.ToX)
	}
	if iterator.FromY > iterator.ToY {
		invertedIntValue(&iterator.FromY, &iterator.ToY)
	}
	for x := iterator.FromX; x <= iterator.ToX; x += 1 {
		for y := iterator.FromY; y <= iterator.ToY; y += 1 {
			if err := function(x, y); err != nil {
				return err
			}
		}
	}
	return nil
}

func SaveAllFile(root *os.Root, configFilePath string, zipWriter *zip.Writer, addFile addFile) (err error) {

	if err := SaveDimensionFile(root, configFilePath, zipWriter, addFile); err != nil {
		return err
	}

	if err := SaveRootDataFile(root, configFilePath, zipWriter, addFile); err != nil {
		return err
	}

	return nil
}

func SaveDimensionFile(root *os.Root, configFilePath string, zipWriter *zip.Writer, addFile addFile) error {

	rootRule, err := LoadConfig(configFilePath)
	if err != nil {
		return err
	}

	for namespaceID, dimensionRule := range rootRule.Dimension {

		namespaceAndID := strings.FieldsFunc(namespaceID, IsNamespaceKeyWord)
		if len(namespaceAndID) != 2 {
			return errors.New("parse namespaceID: invalid namespace ID '" + namespaceID + "'")
		}

		namespace, dimensionID := namespaceAndID[0], namespaceAndID[1]
		dimensionRootDirPath := path.Join("dimensions", namespace, dimensionID)

		if dimensionRootDirStat, err := root.Stat(dimensionRootDirPath); err != nil {
			return fmt.Errorf("open dimension root directory: %w", err)
		} else if !dimensionRootDirStat.IsDir() {
			return fmt.Errorf("open dimension root directory: %v: %w", dimensionRootDirPath, errors.New("not a directory"))
		}

		for _, rangeRule := range dimensionRule.Range {
			for _, regionDataDir := range rootFile {
				iterator := NewRangeIterator(rangeRule.From.X, rangeRule.From.Y, rangeRule.To.X, rangeRule.To.Y)
				err := iterator.run(func(x int, y int) error {
					regionFileName := FormatRegionFilePath(dimensionRootDirPath, regionDataDir, x, y)
					err := addFile(regionFileName, zipWriter)
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
		}

		for _, regionDataDir := range rootFile {
			for _, simpleRule := range dimensionRule.Simple {
				regionFileName := FormatRegionFilePath(dimensionRootDirPath, regionDataDir, simpleRule.X, simpleRule.Y)
				err := addFile(regionFileName, zipWriter)
				if err != nil {
					return err
				}
			}
		}

		dimensionDataDirName := path.Join("dimensions", namespace, dimensionID, "data")
		_, err = root.Stat(dimensionDataDirName)
		if !os.IsNotExist(err) {
			dimensionDataRootDir, err := root.OpenRoot(dimensionDataDirName)
			if err != nil {
				return err
			}
			defer dimensionDataRootDir.Close()

			err = fs.WalkDir(dimensionDataRootDir.FS(), ".", func(subFilePath string, d fs.DirEntry, err error) error {

				if err != nil {
					return fmt.Errorf("open dimension directory: %w", err)
				}

				dataFileName := path.Join(dimensionDataDirName, subFilePath)

				if !d.IsDir() {
					err := addFile(dataFileName, zipWriter)
					if err != nil {
						return fmt.Errorf("%w", err)
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("open dimension directory: "+dimensionDataDirName+": %w", err)
			}
		}
	}
	return nil
}

func SaveRootDataFile(root *os.Root, configFilePath string, zipWriter *zip.Writer, addFile addFile) error {

	rootRule, err := LoadConfig(configFilePath)
	if err != nil {
		return err
	}

	for _, file := range rootRule.File {

		fileStat, err := root.Stat(file)
		if err != nil {
			return fmt.Errorf(":open file: %w", err)
		}

		switch fileStat.IsDir() {

		case false:
			err := addFile(file, zipWriter)
			if err != nil {
				return err
			}
		case true:
			rootDataRootDir, err := root.OpenRoot(file)
			if err != nil {
				return err
			}
			defer rootDataRootDir.Close()

			err = fs.WalkDir(rootDataRootDir.FS(), ".", func(subFilePath string, d fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("open dimension data directory: %w", err)
				}

				fullFilePath := path.Join(file, subFilePath)

				if !d.IsDir() {
					err := addFile(fullFilePath, zipWriter)
					if err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("open directory: %w", err)
			}
		}
	}
	return nil
}

func IsNamespaceKeyWord(char rune) bool {
	if char == rune(":"[0]) {
		return true
	} else {
		return false
	}
}

func FormatRegionFilePath(dimensionRootDirPath string, regionDataDir string, x int, y int) string {
	regionFileName := "r." + strconv.FormatInt(int64(x), 10) + "." + strconv.FormatInt(int64(y), 10) + ".mca"
	regionFilePath := path.Join(dimensionRootDirPath, regionDataDir, regionFileName)
	return regionFilePath
}

func LoadConfig(configFilePath string) (Config, error) {

	jsonData, err := os.ReadFile(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %v", err)
	}

	var rootRule Config
	err = json.Unmarshal(jsonData, &rootRule)
	if err != nil {
		return Config{}, fmt.Errorf("decode json data: %v", err)
	}

	return rootRule, nil

}

func invertedIntValue(a *int, b *int) {
	*a, *b = func(a int, b int) (int, int) {
		return b, a
	}(*a, *b)
}

func NewRangeIterator(fromX int, fromY int, toX int, toY int) RangeIterator {
	return RangeIterator{
		FromX: fromX,
		FromY: fromY,
		ToX:   toX,
		ToY:   toY,
	}
}
