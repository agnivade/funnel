package funnel

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"vbom.ml/util/sortorder"
)

func renameFileTimestamp(cfg *Config) (error, string) {
	t := time.Now()
	err := os.Rename(
		path.Join(cfg.DirName, cfg.ActiveFileName),
		path.Join(cfg.DirName, t.Format("15_04_05.00000-2006_01_02")+".log"),
	)
	return err, t.Format("15_04_05.00000-2006_01_02") + ".log"
}

func renameFileSerial(cfg *Config) (error, string) {
	// Read all the files from log dir
	files, err := ioutil.ReadDir(cfg.DirName)
	if err != nil {
		return err, ""
	}

	// Extracting the file names
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	// Sorting the files in natural order
	sort.Sort(sortorder.Natural(fileNames))

	// Reverse traversing the slice
	for i := len(fileNames) - 1; i >= 0; i-- {
		fileName := fileNames[i]
		// Continuing if its the current file
		if fileName == cfg.ActiveFileName {
			continue
		}

		// Check if the log file is prefixed with the correct active file name
		if strings.HasPrefix(fileName, cfg.ActiveFileName) {
			suffix := ".gz"
			// Get the index from the file name
			num := strings.TrimPrefix(fileName, cfg.ActiveFileName+".")
			// Trim the suffix if ends with .gz
			if strings.HasSuffix(fileName, suffix) {
				num = strings.TrimSuffix(num, suffix)
			}
			intNum, err := strconv.Atoi(num)
			if err != nil {
				continue
			}
			// Now increase it by 1 and rename
			intNum++
			finalName := cfg.ActiveFileName + "." + strconv.Itoa(intNum)
			// If ends with gz, add the gz suffix
			if strings.HasSuffix(fileName, suffix) {
				finalName += suffix
			}
			err = os.Rename(
				path.Join(cfg.DirName, fileName),
				path.Join(cfg.DirName, finalName),
			)
			if err != nil {
				return err, ""
			}
		}
	}

	// Rename active file to file.1
	err = os.Rename(
		path.Join(cfg.DirName, cfg.ActiveFileName),
		path.Join(cfg.DirName, cfg.ActiveFileName+".1"),
	)
	if err != nil {
		return err, ""
	}
	return nil, cfg.ActiveFileName + ".1"
}

func gzipFile(sourcePath string) error {
	reader, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	// Remove the old file once done
	defer os.Remove(sourcePath)

	target := sourcePath + ".gz"
	// Open new gzip stream
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = path.Base(sourcePath)
	defer archiver.Close()

	// Write to the gzip stream
	_, err = io.Copy(archiver, reader)
	return err
}
