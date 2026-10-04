package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

func formatOutPutPath(archiveFilePath string) (string, error) {

	outputFileInfo, err := os.Stat(outputPath)
	switch true {

	case os.IsNotExist(err):
		err = os.MkdirAll(path.Dir(outputPath), 0755)
		if err != nil {
			return "", fmt.Errorf("create output directory: %v", err)
		}
		archiveFilePath = outputPath

	case !os.IsNotExist(err) && err != nil:
		return "", fmt.Errorf("read file info: %v", err)

	case outputFileInfo.IsDir():
		archiveFileName := path.Base(worldDirPath) + "-" + time.Now().Format(time.DateOnly) + ".zip"
		archiveFilePath = path.Join(outputPath, archiveFileName)
		archiveFilePath, err = addSubfixBeforeExt(archiveFilePath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}

	default:
		archiveFilePath, err = addSubfixBeforeExt(outputPath)
		if err != nil {
			return "", fmt.Errorf("add subfix: %v", err)
		}
	}

	return archiveFilePath, nil
}

func addSubfixBeforeExt(archiveFilePath string) (string, error) {
	basePath := path.Base(archiveFilePath)
	dirPath := path.Dir(archiveFilePath)
	nameArr := strings.FieldsFunc(basePath, isExtKeyWord)
	for number := 1; ; number++ {
		var subfix string = "-" + fmt.Sprint(number)
		basePath = nameArr[0] + subfix
		for _, ext := range nameArr[1:] {
			basePath += "." + ext
		}
		stat, err := os.Stat(path.Join(dirPath, basePath))
		if os.IsNotExist(err) {
			break
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("read file info: %v", err)
		} else if !stat.IsDir() || !os.IsNotExist(err) {
			continue
		}
	}
	return path.Join(dirPath, basePath), nil
}

func isExtKeyWord(char rune) bool {
	if char == rune("."[0]) {
		return true
	} else {
		return false
	}
}

func delSliceElement[T comparable](arr []T, index ...int) []T {
	var newArr []T
	for n, e := range arr {
		canAppend := true
		for _, i := range index {
			if n == i {
				canAppend = false
				break
			}
		}
		if canAppend {
			newArr = append(newArr, e)
		}
	}
	return newArr
}
