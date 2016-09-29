package main

import (
	"github.com/agnivade/funnel"
	"os"
)

// TODO: read from config - add tests for config too

// TODO: add flushing policies
// TODO: add rollup policies

// TODO: add line processor

// files - config reader, rollup manager (gzip, deleting)
func main() {
	// Read config
	c := &funnel.Consumer{
		DirName:        "log",
		ActiveFileName: "out.log",
	}
	c.Start(os.Stdin)
}
