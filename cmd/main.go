package main

import (
	"github.com/agnivade/funnel"
)

func main() {
	// Read config
	c := &funnel.Consumer{
		DirName: "log",
	}
	c.Start()
	defer c.CleanUp()
}

// take care to see that when the signal handler does not work,
// the go-routine is ended.

// problem - done is not getting called on interrupt

// problem - signal handler is not getting called on clean end

// TODO: if exit with log file exists, do not rollover that file

// TODO: handle the return errors on all cases
