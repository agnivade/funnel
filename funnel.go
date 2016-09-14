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
}

func (c *Consumer) Start() {
	// make the dir along with parents
	err := os.MkdirAll(c.DirName, 0775)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.newFile()

	c.setupSignalHandling()

	c.feed = make(chan string)
	go c.startFeed()

	scanner := bufio.NewScanner(os.Stdin)
	lines := 0
	for scanner.Scan() {
		lines++
		if c.rollOverCondition() {
			c.rollOver()
		}
		c.feed <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	// close channel
	close(c.feed)
}

func (c *Consumer) CleanUp() {
	// flush writer
	c.writer.Flush()

	// close file handle
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
	// Rename the currfile to a rolled up one
	c.rename()
}

func (c *Consumer) newFile() {
	f, err := os.OpenFile(path.Join(c.DirName, "out.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC|os.O_APPEND|os.O_EXCL,
		0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.currFile = f
	c.writer = bufio.NewWriter(c.currFile)
}

func (c *Consumer) rollOverCondition() bool {
	return false
}

func (c *Consumer) rollOver() {
	// flush writer
	c.writer.Flush()

	// close file handle
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

	c.rename()

	c.newFile()
}

func (c *Consumer) rename() {
	t := time.Now()
	err := os.Rename(
		path.Join(c.DirName, "out.log"),
		path.Join(c.DirName, t.Format("15_04_05-2006_01_02")+".log"),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (c *Consumer) startFeed() {
	for line := range c.feed {
		_, err := fmt.Fprintln(c.writer, line)
		if err != nil {
			fmt.Println(err)
			return
		}
		c.writer.Flush()
	}
}

func (c *Consumer) setupSignalHandling() {
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		os.Interrupt, syscall.SIGPIPE)

	// Block until a signal is received.
	go func(signal_chan chan os.Signal) {
		<-signal_chan
		c.CleanUp()
		os.Exit(1)
	}(signal_chan)
}
