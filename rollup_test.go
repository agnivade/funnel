package funnel

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"vbom.ml/util/sortorder"
)

func TestRenameFileTimestamp(t *testing.T) {
	cfg := setupRollupTest(t)
	defer os.RemoveAll(cfg.DirName)

	// Creating the active file
	err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName)).Run()
	if err != nil {
		t.Fatal(err)
		return
	}

	// Rename the file
	err, _ = renameFileTimestamp(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}

	files, err := ioutil.ReadDir(cfg.DirName)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		regexStr := "[0-9]{2}_[0-9]{2}_[0-9]{2}.[0-9]{5}-[0-9]{4}_[0-9]{2}_[0-9]{2}.log"
		matched, err := regexp.MatchString(regexStr, file.Name())
		if err != nil {
			t.Fatal(err)
			return
		}
		if !matched {
			t.Errorf("Did not match. Expected \"%s\", Got \"%s\"", regexStr, file.Name())
		}
	}
}

func TestRenameFileSerial(t *testing.T) {
	cfg := setupRollupTest(t)
	defer os.RemoveAll(cfg.DirName)

	// Create a whole lot of files
	err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName)).Run()
	if err != nil {
		t.Fatal(err)
		return
	}
	numFiles := 12
	for i := 1; i <= numFiles; i++ {
		err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName+"."+strconv.Itoa(i))).Run()
		if err != nil {
			t.Fatal(err)
			return
		}
	}

	// Rename the files
	err, _ = renameFileSerial(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	files, err := ioutil.ReadDir(cfg.DirName)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(files) != numFiles+1 {
		t.Errorf("Incorrect no. of files created. Expected %d, Got %d", numFiles+1, len(files))
	}
	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	// Sorting the files in natural order
	sort.Sort(sortorder.Natural(fileNames))

	for i := 1; i <= len(fileNames); i++ {
		if fileNames[i-1] != cfg.ActiveFileName+"."+strconv.Itoa(i) {
			t.Errorf("Incorrect file created. Expected %s, Got %s", cfg.ActiveFileName+"."+strconv.Itoa(i), fileNames[i])
		}
	}
}

// Internal helper functions
func setupRollupTest(t *testing.T) *Config {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
		return nil
	}

	cfg := &Config{
		DirName:                  dir,
		ActiveFileName:           "out.log",
		RotationMaxLines:         40,
		RotationMaxBytes:         1000000,
		FlushingTimeIntervalSecs: 5,
		FileRenamePolicy:         "timestamp",
	}
	return cfg
}
