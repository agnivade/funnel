package funnel

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
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
	c.Start(f)
	f.Close()
	c.CleanUp()

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
		t.Errorf("Incorrect no. of files created. Expected 3, Got %d", i+1)
	}
}

func TestHugeLine(t *testing.T) {
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
	target_string := "this ishis hdg s ghdg sdogsdog sdogj osdos dgsodgosj dg sio jsdj gsdgs jdgo of  ojoijg ao jgoarg 38t3p0t8 gh39gh3q9 g huah gaerhg nolaeo ijaoghwtoj joigoawjg awjgpa goarigwo hw ghwogij wogjwg vmaijvowh goawjg wofpojefoi oweijfj owjowjf jfewfj weifjq qOF Qfjoq24 hgawjawo jawwp jwrgjwpgokweo aw jawoif oawjf waefj oawej oaewij fawew jawi ga jgaajg parjgpawogjwaogjwepg awgj waoigjawp gweo gwjag pwagpowaejgp jaweg wagwepgwaepogjwap ogjaw gw agjaw jsjdg sdsdgs doj sgo s"
	reader := strings.NewReader(target_string)
	if err != nil {
		t.Fatal(err)
		return
	}
	c.Start(reader)
	c.CleanUp()

	// testing results
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
		return
	}

	var i int
	var file os.FileInfo
	for i, file = range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		cmp := strings.Compare(string(data[:len(data)-1]), target_string)
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", target_string, string(data))
		}
	}
	if i != 0 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", i+1)
	}

}

func TestJustEOF(t *testing.T) {

}

func TestSendInterrupt(t *testing.T) {

}
