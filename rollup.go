package funnel

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"vbom.ml/util/sortorder"
)

func renameFileTimestamp(cfg *Config) error {
	t := time.Now()
	err := os.Rename(
		path.Join(cfg.DirName, cfg.ActiveFileName),
		path.Join(cfg.DirName, t.Format("15_04_05.00000-2006_01_02")+".log"),
	)
	return err
}

func renameFileSerial(cfg *Config) error {
	// Read all the files from log dir
	files, err := ioutil.ReadDir(cfg.DirName)
	if err != nil {
		return err
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
			// Get the index from the file name
			num := strings.TrimPrefix(fileName, cfg.ActiveFileName+".")
			intNum, err := strconv.Atoi(num)
			if err != nil {
				continue
			}
			// Now increase it by 1 and rename
			intNum++
			err = os.Rename(
				path.Join(cfg.DirName, fileName),
				path.Join(cfg.DirName, cfg.ActiveFileName+"."+strconv.Itoa(intNum)),
			)
			if err != nil {
				return err
			}
		}
	}

	// Rename active file to file.1
	err = os.Rename(
		path.Join(cfg.DirName, cfg.ActiveFileName),
		path.Join(cfg.DirName, cfg.ActiveFileName+".1"),
	)
	if err != nil {
		return err
	}
	return nil
}
