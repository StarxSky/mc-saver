package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"

	"acovia.net/record"
)

func newZipWriter() (*zip.Writer, *os.File, func(error) error, error) {

	archiveFilePath, err := formatOutPutPath(outputPath)
	if err != nil {
		return nil, nil, nil, err
	}

	file, err := os.CreateTemp(path.Dir(archiveFilePath), "archive")
	if err != nil {
		return nil, nil, nil, err
	}

	if err = file.Chmod(0655); err != nil {
		return nil, nil, nil, err
	}

	zipWriter := zip.NewWriter(file)

	end := func(err error) error {

		if err != nil {
			os.Remove(file.Name())
			return err
		}

		if err := os.Rename(file.Name(), archiveFilePath); err != nil {
			os.Remove(file.Name())
			return err
		}

		if err := zipWriter.Close(); err != nil {
			os.Remove(file.Name())
			return err
		}

		if err := file.Close(); err != nil {
			os.Remove(file.Name())
			return err
		}

		record.Info("backup completed successfully!")

		return nil

	}
	return zipWriter, file, end, nil
}

func addFile(filePath string, zipWriter *zip.Writer) error {

	fileReader, err := root.Open(filePath)
	if err != nil {
		record.Warn("skip file:", err)
		return nil
	}
	defer fileReader.Close()

	fileInfo, err := root.Stat(filePath)
	if err != nil {
		return err
	}

	zipFileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	headerName := path.Join(path.Base(root.Name()), filePath)

	zipFileHeader.Name = headerName
	zipFileHeader.Method = zip.Deflate

	file, err := zipWriter.CreateHeader(zipFileHeader)

	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if err = record.RunningInfo(func() error {
		_, err = io.Copy(file, fileReader)
		if err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		return nil
	}, "adding: ", headerName); err != nil {
		return err
	}

	return nil

}
