package funnel

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

type Consumer struct {
	DirName string

	// internal stuff
	currFile *os.File
	writer   *bufio.Writer
	feed     chan string

	done         chan struct{}
	rolloverChan chan struct{}
	signal_chan  chan os.Signal
	numLines     int
}

func (c *Consumer) Start() {
	c.setupSignalHandling()
	c.done = make(chan struct{})
	c.rolloverChan = make(chan struct{})

	// make the dir along with parents
	if err := os.MkdirAll(c.DirName, 0775); err != nil {
		fmt.Println(err)
		return
	}

	if err := c.newFile(); err != nil {
		fmt.Println(err)
		return
	}

	c.feed = make(chan string)
	go c.startFeed()

	scanner := bufio.NewScanner(os.Stdin)
	c.numLines = 0
	for scanner.Scan() {
		c.numLines++
		if c.rollOverCondition() {
			c.rolloverChan <- struct{}{}
		}
		c.feed <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("scanner stopped- ", err)
	}
	// work is done, signalling done channel
	c.done <- struct{}{}
	// quitting from signal handler
	close(c.signal_chan)
}

func (c *Consumer) CleanUp() {
	// flush writer
	if c.writer != nil {
		c.writer.Flush()
	}

	// close file handle
	if c.currFile != nil {
		err := c.currFile.Sync()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = c.currFile.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// Rename the currfile to a rolled up one
	if err := c.rename(); err != nil {
		fmt.Println(err)
		return
	}
}

func (c *Consumer) newFile() error {
	f, err := os.OpenFile(path.Join(c.DirName, "out.log"),
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
	return c.numLines%40 == 0
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
		path.Join(c.DirName, "out.log"),
		path.Join(c.DirName, t.Format("15_04_05.000-2006_01_02")+".log"),
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *Consumer) startFeed() {
	// Will flush the writer every 5 sec
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case line := <-c.feed:
			_, err := fmt.Fprintln(c.writer, line)
			if err != nil {
				fmt.Println(err)
				return
			}
		case <-c.rolloverChan:
			if err := c.rollOver(); err != nil {
				fmt.Println(err)
				return
			}
		case <-c.done:
			ticker.Stop()
			if err := c.writer.Flush(); err != nil {
				fmt.Println(err)
			}
			return
		case <-ticker.C:
			if err := c.writer.Flush(); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func (c *Consumer) setupSignalHandling() {
	c.signal_chan = make(chan os.Signal, 1)
	signal.Notify(c.signal_chan,
		os.Interrupt, syscall.SIGPIPE)

	// Block until a signal is received.
	// Or EOF happens
	go func() {
		for _ = range c.signal_chan {
			fmt.Println("Caught signal")
		}
	}()
}
