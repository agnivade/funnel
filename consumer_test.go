package funnel

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"syscall"
	"testing"
)

func TestRollover(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	f, err := os.Open("testdata/file_84lines")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
	c.Start(f)

	// testing results
	files := readTestDir(t, dir)

	sep := []byte{'\n'}
	if len(files) != 3 {
		t.Errorf("Incorrect no. of files created. Expected 3, Got %d", len(files))
	}
	for i, file := range files {
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
}

func TestHugeLine(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	// This file also contains arabic, indian and tibetan characters
	// to test any ascii-utf8 codec incompatibility
	targetBytes, err := ioutil.ReadFile("testdata/file_bigline")
	if err != nil {
		t.Fatal(err)
		return
	}
	r := bytes.NewReader(targetBytes)
	c.Start(r)

	// testing results
	files := readTestDir(t, dir)
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		// removing the newline character at the end
		cmp := bytes.Compare(data, targetBytes)
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", string(targetBytes), string(data))
		}
	}
}

func TestNewLines(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	// This file also contains arabic, indian and tibetan characters
	// to test any ascii-utf8 codec incompatibility
	targetBytes, err := ioutil.ReadFile("testdata/file_newline")
	if err != nil {
		t.Fatal(err)
		return
	}
	r := bytes.NewReader(targetBytes)
	c.Start(r)

	// testing results
	files := readTestDir(t, dir)
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		cmp := bytes.Compare(data, targetBytes)
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", string(targetBytes), string(data))
		}
	}
}

func TestEmptyString(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	r := strings.NewReader("")
	c.Start(r)

	// testing results
	files := readTestDir(t, dir)
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		cmp := bytes.Compare(data, []byte(""))
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", "", string(data))
		}
	}
}

func TestRolloverSerial(t *testing.T) {
	dir, c := setupTest(t)
	c.Config.FileRenamePolicy = "serial"
	defer os.RemoveAll(dir)

	f, err := os.Open("testdata/file_84lines")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
	c.Start(f)

	// testing results
	files := readTestDir(t, dir)

	sep := []byte{'\n'}
	if len(files) != 3 {
		t.Errorf("Incorrect no. of files created. Expected 3, Got %d", len(files))
	}
	for i, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		numLines := bytes.Count(data, sep)
		// First file will have 4 lines
		// rest will have 40 lines
		if i == 0 {
			if numLines != 4 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 4, Got %d", i, numLines)
			}
		} else {
			if numLines != 40 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 40, Got %d", i, numLines)
			}
		}
	}
}

func TestSendInterruptSerial(t *testing.T) {
	// TODO
}

// Commenting this unless I can fix the flakiness
// the line2 sometimes gets written after the config change. Need to fix that.
/*func TestConfigReload(t *testing.T) {
	dir, c := setupTest(t)
	c.Config.FileRenamePolicy = "serial"
	defer os.RemoveAll(dir)

	// Setting up this reader, writer pipe because we want to write some,
	// reload config and then again write some
	rdr, wtr := io.Pipe()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		c.Start(rdr)
		wg.Done()
	}()
	line1 := "hello this is a line\n"
	line2 := "hello this is another line\n"
	wtr.Write([]byte(line1))
	wtr.Write([]byte(line2))

	c.ReloadChan <- &Config{
		DirName:          dir,
		ActiveFileName:   "out.log",
		RotationMaxLines: 40,
		RotationMaxBytes: 1000000,
		PrependValue:     "[agniva]",
		FileRenamePolicy: "serial",
		MaxAge:           int64(1 * 60 * 60),
		MaxCount:         500,
		Target:           "file",
	}

	wtr.Write([]byte("trying again with a line\n"))
	wtr.Write([]byte("trying again with another line"))
	wtr.Close()

	wg.Wait()
	// testing results
	files := readTestDir(t, dir)
	if len(files) != 2 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	data, err := ioutil.ReadFile(path.Join(dir, "out.log.1"))
	if err != nil {
		t.Fatal(err)
		return
	}
	cmp := bytes.Compare(data, []byte("[agniva]trying again with a line\n[agniva]trying again with another line"))
	if cmp != 0 {
		t.Errorf("Incorrect string found. Expected- %s, Found- %s", "[agniva]trying again with a line\n[agniva]trying again with another line", string(data))
	}
	data2, err := ioutil.ReadFile(path.Join(dir, "out.log.2"))
	if err != nil {
		t.Fatal(err)
		return
	}
	cmp2 := bytes.Compare(data2, []byte(line1+line2))
	if cmp2 != 0 {
		t.Errorf("Incorrect string found. Expected- %s, Found- %s", line1+line2, string(data2))
	}
}*/

// Benchmarking different file creation and status flags to check write speed
func benchmarkFileIO(b *testing.B, flags int) {
	dir, _ := ioutil.TempDir("", "test")
	f, _ := os.OpenFile(path.Join(dir, "testspeed"),
		flags,
		0644)
	defer f.Sync()
	defer f.Close()
	defer os.Remove(f.Name())
	defer os.RemoveAll(dir)

	for n := 0; n < b.N; n++ {
		f.Write(randStringBytes(50))
	}
}

func BenchmarkFileIO_Append(b *testing.B) {
	benchmarkFileIO(b, os.O_CREATE|os.O_WRONLY|os.O_APPEND)
}

func BenchmarkFileIO_Normal(b *testing.B) {
	benchmarkFileIO(b, os.O_CREATE|os.O_WRONLY)
}

func BenchmarkFileIO_Sync(b *testing.B) {
	benchmarkFileIO(b, os.O_CREATE|os.O_WRONLY|os.O_SYNC)
}

func BenchmarkFileIO_DSync(b *testing.B) {
	benchmarkFileIO(b, os.O_CREATE|os.O_WRONLY|syscall.O_DSYNC)
}

func BenchmarkFileIO_AppendSync(b *testing.B) {
	benchmarkFileIO(b, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC)
}

func BenchmarkFileIO_AppendDSync(b *testing.B) {
	benchmarkFileIO(b, os.O_CREATE|os.O_WRONLY|os.O_APPEND|syscall.O_DSYNC)
}

// Internal helper functions
func setupTest(t *testing.T) (string, *Consumer) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
		return "", nil
	}

	c := &Consumer{
		Config: &Config{
			DirName:                  dir,
			ActiveFileName:           "out.log",
			RotationMaxLines:         40,
			RotationMaxBytes:         1000000,
			FlushingTimeIntervalSecs: 5,
			FileRenamePolicy:         "timestamp",
			MaxAge:                   int64(1 * 60 * 60),
			MaxCount:                 500,
			Target:                   "file",
		},
		LineProcessor: &NoProcessor{},
		ReloadChan:    make(chan *Config),
	}
	return dir, c
}

func readTestDir(t *testing.T, dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return files
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) []byte {
	b := make([]byte, n)
	for i := 0; i < n-1; i++ {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	b[n-1] = '\n'
	return b
}
