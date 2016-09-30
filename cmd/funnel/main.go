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
	cfg, err := GetConfig()
	if err != nil {
		// TODO: check if this is idiomatic or not
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// Initialise consumer
	c := &funnel.Consumer{
		Config: cfg,
	}
	c.Start(os.Stdin)
}
