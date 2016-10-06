package main

import (
	"fmt"
	"os"

	"github.com/agnivade/funnel"
)

// TODO: add rollup policies

// TODO: add line processor

// TODO: add http stats endpoint conditionally

// files - config reader, rollup manager (gzip, deleting)

func main() {
	// Read config
	cfg, err := funnel.GetConfig()
	if err != nil {
		fmt.Println("Error in config file: ", err)
		os.Exit(1)
	}

	// Initialise consumer
	c := &funnel.Consumer{
		Config: cfg,
	}
	c.Start(os.Stdin)
}
