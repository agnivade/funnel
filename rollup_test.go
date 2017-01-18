package funnel

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

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
	_, err = renameFileTimestamp(cfg)
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

	dateRegex := "[0-9]{4}-[0-9]{2}-[0-9]{2}"
	timeRegex := "[0-9]{2}-[0-9]{2}-[0-9]{2}.[0-9]{5}"
	regexStr := fmt.Sprintf("%s_%s%s", dateRegex, timeRegex, ".log")
	for _, file := range files {
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
	numFiles := 13
	if err := populateFiles(cfg, numFiles); err != nil {
		t.Fatal(err)
		return
	}

	// Rename the files
	fileName, err := renameFileSerial(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	if fileName != cfg.ActiveFileName+".1" {
		t.Errorf("Incorrect active file name received. Expected %s, Got %s", cfg.ActiveFileName+".1", fileName)
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

func TestRenameFileSerialGzip(t *testing.T) {
	cfg := setupRollupTest(t)
	cfg.Gzip = true
	defer os.RemoveAll(cfg.DirName)

	// Create a whole lot of files
	err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName)).Run()
	if err != nil {
		t.Fatal(err)
		return
	}
	numFiles := 12
	for i := 1; i <= numFiles; i++ {
		err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName+"."+strconv.Itoa(i)+".gz")).Run()
		if err != nil {
			t.Fatal(err)
			return
		}
	}

	// Rename the files
	fileName, err := renameFileSerial(cfg)
	if err != nil {
		t.Fatal(err)
		return
	}
	if fileName != cfg.ActiveFileName+".1" {
		t.Errorf("Incorrect active file name received. Expected %s, Got %s", cfg.ActiveFileName+".1", fileName)
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

	for i := 2; i <= len(fileNames); i++ {
		if fileNames[i-1] != cfg.ActiveFileName+"."+strconv.Itoa(i)+".gz" {
			t.Errorf("Incorrect file created. Expected %s, Got %s", cfg.ActiveFileName+"."+strconv.Itoa(i)+".gz", fileNames[i])
		}
	}
}

func TestGzipFile(t *testing.T) {
	content := []byte("gzip test content")
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	gzipFile(tmpfile.Name())

	// check that the file is now deleted
	_, err = os.Open(tmpfile.Name())
	if os.IsExist(err) {
		t.Error(tmpfile.Name() + " exists. Expected to be deleted")
	}

	// check that a gzip file is there at the same place
	_, err = os.Open(tmpfile.Name() + ".gz")
	if os.IsNotExist(err) {
		t.Error(tmpfile.Name() + ".gz does not exist. Expected to exist")
	}
	os.Remove(tmpfile.Name() + ".gz")
}

func TestMaxFiles(t *testing.T) {
	cfg := setupRollupTest(t)
	defer os.RemoveAll(cfg.DirName)

	if err := populateFiles(cfg, 13); err != nil {
		t.Fatal(err)
		return
	}

	deleteOldFiles(cfg)

	files, err := ioutil.ReadDir(cfg.DirName)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(files) != cfg.MaxCount {
		t.Errorf("Incorrect no. of files created. Expected %d, Got %d", cfg.MaxCount, len(files))
	}
}

func TestOldFiles(t *testing.T) {
	cfg := setupRollupTest(t)
	defer os.RemoveAll(cfg.DirName)
	cfg.MaxCount = 100
	cfg.MaxAge = int64(2)

	if err := populateFiles(cfg, 13); err != nil {
		t.Fatal(err)
		return
	}

	// Sleeping for 5 seconds to let it exceed the max age
	time.Sleep(time.Second * 5)
	deleteOldFiles(cfg)

	files, err := ioutil.ReadDir(cfg.DirName)
	if err != nil {
		t.Fatal(err)
		return
	}
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
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
		MaxAge:                   int64(1 * 24 * 60 * 60),
		MaxCount:                 5,
	}
	return cfg
}

func populateFiles(cfg *Config, numFiles int) error {
	for i := numFiles; i > 0; i-- {
		err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName+"."+strconv.Itoa(i))).Run()
		if err != nil {
			return err
		}
	}
	err := exec.Command("touch", path.Join(cfg.DirName, cfg.ActiveFileName)).Run()
	if err != nil {
		return err
	}
	return nil
}
