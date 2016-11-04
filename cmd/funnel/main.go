package main

import (
	"fmt"
	"os"

	"github.com/agnivade/funnel"
)

// TODO: add http stats endpoint conditionally
// TODO: test deleteoldfiles with modTime

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
