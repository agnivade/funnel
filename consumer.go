package funnel

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

type Consumer struct {
	Config *Config

	// internal stuff
	currFile *os.File
	writer   *bufio.Writer
	feed     chan string

	// channel signallers
	done         chan struct{}
	rolloverChan chan struct{}
	signalChan   chan os.Signal
	wg           sync.WaitGroup

	// variable to track write progress
	numLines      int64
	fileSizeBytes int64
}

func (c *Consumer) Start(inputStream io.Reader) {
	c.setupSignalHandling()
	c.done = make(chan struct{})
	c.rolloverChan = make(chan struct{})

	// make the dir along with parents
	if err := os.MkdirAll(c.Config.DirName, 0775); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if err := c.newFile(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	c.feed = make(chan string)
	go c.startFeed()

	reader := bufio.NewReader(inputStream)
	c.numLines = 0
	for {
		// This will return a line until delimiter
		// If delimiter is not found, it returns the line with error
		// so line will always be available
		// Then we check for error and quit
		line, err := reader.ReadString('\n')
		if c.rollOverCondition() {
			c.rolloverChan <- struct{}{}
			c.numLines = 0
		}
		c.feed <- line
		c.numLines++
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
			}
			break
		}
	}
	// work is done, signalling done channel
	c.wg.Add(1)
	c.done <- struct{}{}
	c.wg.Wait()
	// quitting from signal handler
	close(c.signalChan)
}

func (c *Consumer) cleanUp() {
	// close file handle
	if c.currFile != nil {
		if err := c.currFile.Sync(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		if err := c.currFile.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
	// Rename the currfile to a rolled up one
	if err := c.rename(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func (c *Consumer) newFile() error {
	f, err := os.OpenFile(path.Join(c.Config.DirName, c.Config.ActiveFileName),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND|os.O_EXCL,
		0644)
	if err != nil {
		return err
	}
	c.currFile = f
	c.writer = bufio.NewWriter(c.currFile)
	return nil
}

func (c *Consumer) rollOverCondition() bool {
	return c.numLines > 0 && c.numLines%c.Config.RotationMaxLines == 0
}

func (c *Consumer) rollOver() error {
	// flush writer
	err := c.writer.Flush()
	if err != nil {
		return err
	}

	// close file handle
	if err := c.currFile.Sync(); err != nil {
		return err
	}

	if err := c.currFile.Close(); err != nil {
		return err
	}

	if err := c.rename(); err != nil {
		return err
	}

	if err := c.newFile(); err != nil {
		return err
	}
	return nil
}

func (c *Consumer) rename() error {
	t := time.Now()
	err := os.Rename(
		path.Join(c.Config.DirName, c.Config.ActiveFileName),
		path.Join(c.Config.DirName, t.Format("15_04_05.000-2006_01_02")+".log"),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *Consumer) startFeed() {
	// Will flush the writer at some intervals
	ticker := time.NewTicker(time.Duration(c.Config.FlushingTimeIntervalSecs) * time.Second)
	for {
		select {
		case line := <-c.feed:
			//TODO: process the line
			_, err := fmt.Fprint(c.writer, line)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		case <-c.rolloverChan:
			if err := c.rollOver(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		case <-c.done:
			ticker.Stop()
			if err := c.writer.Flush(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			c.cleanUp()
			c.wg.Done()
			return
		case <-ticker.C:
			if err := c.writer.Flush(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

func (c *Consumer) setupSignalHandling() {
	c.signalChan = make(chan os.Signal, 1)
	signal.Notify(c.signalChan,
		os.Interrupt, syscall.SIGPIPE)

	// Block until a signal is received.
	// Or EOF happens
	go func() {
		for _ = range c.signalChan {
		}
	}()
}
