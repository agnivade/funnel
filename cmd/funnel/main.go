package main

import (
	"fmt"
	"log/syslog"
	"os"

	"github.com/agnivade/funnel"
	_ "github.com/agnivade/funnel/outputs"
	"github.com/spf13/viper"
)

// TODO: add http stats endpoint conditionally
const (
	AppName = "funnel"
)

func main() {
	logger, err := syslog.New(syslog.LOG_ERR, AppName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// Verifying whether the app has a piped stdin or not
	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		fmt.Println("No pipe found to consume data from.")
		os.Exit(1)
	}

	// Setting the config file name and the locations to search for the config
	v := viper.New()
	v.SetConfigName("config")
	v.AddConfigPath("/etc/" + AppName + "/")
	v.AddConfigPath("$HOME/." + AppName)
	v.AddConfigPath(".")

	// Read config
	// The outputWriter is nil if its file output
	cfg, reloadChan, outputWriter, err := funnel.GetConfig(v, logger)
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
		Logger:        logger,
		Writer:        outputWriter,
	}
	c.Start(os.Stdin)
}
