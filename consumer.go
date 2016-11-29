package funnel

import (
	"bufio"
	"io"
	"log/syslog"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"
)

// Consumer is the main struct which holds all the stuff
// necessary to run the code
type Consumer struct {
	Config        *Config
	LineProcessor LineProcessor
	Logger        *syslog.Writer
	Writer        OutputWriter

	// internal stuff
	currFile *os.File
	feed     chan string

	// channel signallers
	done         chan struct{}
	rolloverChan chan struct{}
	signalChan   chan os.Signal
	errChan      chan error
	wg           sync.WaitGroup
	ReloadChan   chan *Config

	// variable to track write progress
	linesWritten int
	bytesWritten uint64
}

// Start takes the input stream and begins reading line by line
// buffering the output to a file and flushing at set intervals
func (c *Consumer) Start(inputStream io.Reader) {
	c.setupSignalHandling()
	c.done = make(chan struct{})
	c.rolloverChan = make(chan struct{})
	c.errChan = make(chan error, 1)
	// Check if the target is file, only then create dirs and all
	if c.Config.Target == "file" {
		// Make the dir along with parents
		if err := os.MkdirAll(c.Config.DirName, 0775); err != nil {
			c.Logger.Err(err.Error())
			return
		}

		// Create the file
		if err := c.createNewFile(); err != nil {
			c.Logger.Err(err.Error())
			return
		}
	}

	// Create the line feed channel and start the feed goroutine
	c.feed = make(chan string)
	go c.startFeed()

	// Get the reader to the input stream and set initial counters
	reader := bufio.NewReader(inputStream)
	c.linesWritten = 0
	c.bytesWritten = 0

	// start a for-select loop to wait until main loop is done, or catch errors
outer:
	for {
		select {
		case err := <-c.errChan: // error channel to get any errors happening
			// elsewhere. After printing to stderr, it breaks from the loop
			c.Logger.Err(err.Error())
			break outer
		default:
			// This will return a line until delimiter
			// If delimiter is not found, it returns the line with error
			// so line will always be available
			// Then we check for error and quit
			line, err := reader.ReadString('\n')
			// Send to feed
			c.feed <- line

			// Update counters
			c.linesWritten++
			c.bytesWritten += uint64(len(line))

			// Check for rollover
			if c.rollOverCondition() {
				c.rolloverChan <- struct{}{}
				c.linesWritten = 0
				c.bytesWritten = 0
			}

			if err != nil {
				if err != io.EOF {
					c.Logger.Err(err.Error())
				}
				break outer
			}
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
	var err error
	// If target is a file, close the file handles
	if c.Config.Target == "file" {
		// Close file handle
		if err = c.currFile.Sync(); err != nil {
			c.Logger.Err(err.Error())
			return
		}

		if err = c.currFile.Close(); err != nil {
			c.Logger.Err(err.Error())
			return
		}

		// Rename the currfile to a rolled up one
		var fileName string
		if fileName, err = c.rename(); err != nil {
			c.Logger.Err(err.Error())
			return
		}

		if err = c.compress(fileName); err != nil {
			c.Logger.Err(err.Error())
			return
		}
	} else { // else call the Close function on the writer
		c.Writer.Close()
	}
}

func (c *Consumer) createNewFile() error {
	f, err := os.OpenFile(path.Join(c.Config.DirName, c.Config.ActiveFileName),
		os.O_CREATE|os.O_WRONLY|os.O_EXCL,
		0644)
	if err != nil {
		return err
	}
	c.currFile = f
	// Embedding buffered writer in another struct to satisfy the OutputWriter interface
	// This is because in the consume loop, functions are called directly on the writer
	c.Writer = &FileOutput{bufio.NewWriter(c.currFile)}
	return nil
}

func (c *Consumer) rollOverCondition() bool {
	// Return true if either lines written has exceeded
	// or bytes written has exceeded
	return c.linesWritten >= c.Config.RotationMaxLines ||
		c.bytesWritten >= c.Config.RotationMaxBytes
}

func (c *Consumer) rollOver() error {
	var err error
	// Flush writer
	if err = c.Writer.Flush(); err != nil {
		return err
	}

	// Do file related stuff only if the target is file
	if c.Config.Target == "file" {
		// Close file handle
		if err = c.currFile.Sync(); err != nil {
			return err
		}
		if err = c.currFile.Close(); err != nil {
			return err
		}

		var fileName string
		if fileName, err = c.rename(); err != nil {
			return err
		}

		if err = c.compress(fileName); err != nil {
			return err
		}

		if err = c.deleteFiles(); err != nil {
			return err
		}

		if err = c.createNewFile(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Consumer) rename() (string, error) {
	var fileName string
	var err error
	if c.Config.FileRenamePolicy == "timestamp" {
		fileName, err = renameFileTimestamp(c.Config)
		if err != nil {
			return "", err
		}
	} else {
		fileName, err = renameFileSerial(c.Config)
		if err != nil {
			return "", err
		}
	}
	return fileName, nil
}

func (c *Consumer) compress(fileName string) error {
	// Check config and compress if yes
	if c.Config.Gzip {
		err := gzipFile(path.Join(c.Config.DirName, fileName))
		return err
	}
	return nil
}

func (c *Consumer) deleteFiles() error {
	return deleteOldFiles(c.Config)
}

func (c *Consumer) startFeed() {
	// Will flush the writer at some intervals
	ticker := time.NewTicker(time.Duration(c.Config.FlushingTimeIntervalSecs) * time.Second)
	for {
		select {
		case line := <-c.feed: // Write to buffered writer
			err := c.LineProcessor.Write(c.Writer, line)
			if err != nil {
				c.errChan <- err
			}
		case <-c.rolloverChan: // Rollover file to new one
			if err := c.rollOver(); err != nil {
				c.errChan <- err
			}
		case cfg := <-c.ReloadChan: // reload channel to listen to any changes in config file
			if err := c.rollOver(); err != nil {
				c.errChan <- err
			}

			c.linesWritten = 0
			c.bytesWritten = 0
			c.LineProcessor = GetLineProcessor(cfg) // setting new line processor
			if c.Config.Target == "file" {
				// create new config dir
				if err := os.MkdirAll(cfg.DirName, 0775); err != nil {
					c.errChan <- err
					break
				}

				// close old config file
				if err := c.currFile.Close(); err != nil {
					c.errChan <- err
					break
				}

				// delete old config file
				if err := os.Remove(path.Join(c.Config.DirName, c.Config.ActiveFileName)); err != nil {
					if !os.IsNotExist(err) {
						c.errChan <- err
						break
					}
				}
			}
			c.Config = cfg // setting new config

			if c.Config.Target == "file" {
				// create new config file
				if err := c.createNewFile(); err != nil {
					c.errChan <- err
				}
			}
		case <-c.done: // Done signal received, close shop
			ticker.Stop()
			if err := c.Writer.Flush(); err != nil {
				c.Logger.Err(err.Error())
			}
			c.cleanUp()
			c.wg.Done()
			return
		case <-ticker.C: // If tick happens, flush the writer
			if err := c.Writer.Flush(); err != nil {
				c.errChan <- err
			}
		}
	}
}

func (c *Consumer) setupSignalHandling() {
	c.signalChan = make(chan os.Signal, 1)
	signal.Notify(c.signalChan,
		os.Interrupt, syscall.SIGTERM)

	// Block until a signal is received.
	go func() {
		for range c.signalChan {
			// work is done, signalling done channel
			c.wg.Add(1)
			c.done <- struct{}{}
			c.wg.Wait()
			// Everything taken care of, goodbye
			os.Exit(1)

		}
	}()
}
