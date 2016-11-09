package main

import (
	"fmt"
	"os"

	"github.com/agnivade/funnel"
	"github.com/spf13/viper"
)

// TODO: add http stats endpoint conditionally
const (
	AppName = "funnel"
)

func main() {
	// Setting the config file name and the locations to search for the config
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("/etc/" + AppName + "/")
	v.AddConfigPath("$HOME/." + AppName)
	v.AddConfigPath(".")

	// Read config
	cfg, reloadChan, err := funnel.GetConfig(v)
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
		ReloadChan:    reloadChan,
	}
	c.Start(os.Stdin)
}
