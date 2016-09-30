package funnel

import (
	"github.com/spf13/viper"
)

// XXX: Move it to constants.go if needed
const (
	APP_NAME = "funnel"
)

type Config struct {
	DirName        string
	ActiveFileName string
}

func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + APP_NAME + "/")
	viper.AddConfigPath("$HOME/." + APP_NAME)
	viper.AddConfigPath(".")

	// Find and read the config file
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		return nil, err
	}
	// if values not found, set default

	// validate

	// return struct
	return &Config{
		DirName:        viper.GetString("logging.directory"),
		ActiveFileName: viper.GetString("logging.active_file_name"),
	}, nil
}
