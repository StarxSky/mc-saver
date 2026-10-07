package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"acovia.net/record"
)

func initZipWriter() (*zip.Writer, *os.File, func(error) error, error) {

	archiveFilePath, err := formatOutPutPath(outputPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("format output path: %v", err)
	}

	file, err := os.Create(archiveFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create temp archive: %v", err)
	}

	zipWriter := zip.NewWriter(file)

	end := func(err error) error {

		if err != nil {
			if removeErr := os.Remove(file.Name()); removeErr != nil {
				return errors.Join(err, removeErr)
			}
			return err
		}

		if err := zipWriter.Close(); err != nil {
			os.Remove(file.Name())
			return fmt.Errorf("close zip writer: %v", err)
		}

		if err := file.Close(); err != nil {
			os.Remove(file.Name())
			return fmt.Errorf("close file writer: %v", err)
		}
		record.Info("backup completed successfully!")
		return nil
	}
	return zipWriter, file, end, nil
}

func saveFile(fileName string, zipWriter *zip.Writer) error {

	fileReader, err := root.Open(fileName)
	if os.IsNotExist(err) {
		record.Warn("skip file:", err)
		return nil
	} else if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	defer fileReader.Close()

	fileInfo, err := root.Stat(fileName)
	if err != nil {
		return fmt.Errorf("read file info: %v", err)
	}

	zipFileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	headerName := path.Join(path.Base(root.Name()), fileName)

	zipFileHeader.Name = headerName
	zipFileHeader.Method = zip.Deflate

	file, err := zipWriter.CreateHeader(zipFileHeader)
	if err != nil {
		return fmt.Errorf("create file: %v", err)
	}

	if err = record.RunningInfo(func() error {
		_, err = io.Copy(file, fileReader)
		if err != nil {
			return fmt.Errorf("write file: %v", err)
		}
		return nil
	}, "adding: ", headerName); err != nil {
		return err
	}
	return nil
}

func saveConfig() error {
	jsonData, err := json.MarshalIndent(config, "", "	")
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
