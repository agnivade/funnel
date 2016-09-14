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
