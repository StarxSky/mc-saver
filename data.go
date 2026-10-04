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

	case err != nil && !os.IsNotExist(err):
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
	nameArr := strings.FieldsFunc(archiveFilePath, isExtKeyWord)
	for number := 1; ; number++ {
		var subfix string = "-" + fmt.Sprint(number)
		archiveFilePath = nameArr[0] + subfix
		for _, ext := range nameArr[1:] {
			archiveFilePath += "." + ext
		}
		stat, err := os.Stat(archiveFilePath)
		if os.IsNotExist(err) {
			break
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("read file info: %v", err)
		} else if !stat.IsDir() || !os.IsNotExist(err) {
			continue
		}
	}
	return archiveFilePath, nil
}

func isExtKeyWord(char rune) bool {
	if char == rune("."[0]) {
		return true
	} else {
		return false
	}
}

func isExist(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}
