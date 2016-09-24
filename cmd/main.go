package main

import (
	"github.com/agnivade/funnel"
)

func main() {
	// Read config
	c := &funnel.Consumer{
		DirName:        "log",
		ActiveFileName: "out.log",
	}
	c.Start()
	defer c.CleanUp()
}

// TODO: if exit with log file exists, do not rollover that file

// TODO: handle the return errors on all cases
