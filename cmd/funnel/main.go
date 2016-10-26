package main

import (
	"fmt"
	"os"

	"github.com/agnivade/funnel"
)

// TODO: add rollup policies
// gzip files or not
// delete files older than

// TODO: add http stats endpoint conditionally
// TODO: decide on using goroutines for gzipping and deleting files

// files -rollup manager (gzip, deleting)

func main() {
	// Read config
	cfg, err := funnel.GetConfig()
	if err != nil {
		fmt.Println("Error in config file: ", err)
		os.Exit(1)
	}

	// Get the line processor depending on the config
	lp := funnel.GetLineProcessor(cfg)

	// Initialise consumer
	c := &funnel.Consumer{
		Config:        cfg,
		LineProcessor: lp,
	}
	c.Start(os.Stdin)
}
