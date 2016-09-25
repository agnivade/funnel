package funnel

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestRollover(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer os.RemoveAll(dir)
	c := &Consumer{
		DirName:        dir,
		ActiveFileName: "out.log",
	}
	f, err := os.Open("testdata/file_84lines")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
	c.Start(f)
	defer c.CleanUp()

	// testing results
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
		return
	}

	var i int
	var file os.FileInfo
	sep := []byte{'\n'}
	for i, file = range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		numLines := bytes.Count(data, sep)
		// First 2 files will have 40 lines
		// last one will have 4 lines
		if i < 2 {
			if numLines != 40 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 40, Got %d", i, numLines)
			}
		} else {
			if numLines != 4 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 4, Got %d", i, numLines)
			}
		}
	}
	if i != 2 {
		t.Errorf("Incorrect no. of files created. Expected 2, Got %d", i)
	}
}

func TestHugeLine(t *testing.T) {

}

func TestJustEOF(t *testing.T) {

}

func TestSendInterrupt(t *testing.T) {

}
