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

	RotationMaxLines int64
	RotationMaxBytes int64

	FlushingTimeIntervalSecs int
}

func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + APP_NAME + "/")
	viper.AddConfigPath("$HOME/." + APP_NAME)
	viper.AddConfigPath(".")

	// Set default values. They are overridden by config file values, if provided
	setDefaults()

	// TODO: check if file is present or not, only then read from it
	// Find and read the config file
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		return nil, err
	}

	// validate

	// return struct
	return &Config{
		DirName:                  viper.GetString("logging.directory"),
		ActiveFileName:           viper.GetString("logging.active_file_name"),
		RotationMaxLines:         viper.GetInt64("rotation.lines"),
		RotationMaxBytes:         viper.GetInt64("rotation.file_size_bytes"),
		FlushingTimeIntervalSecs: viper.GetInt("flushing.time_interval_secs"),
	}, nil
}

func setDefaults() {
	viper.SetDefault("logging.directory", "log")
	viper.SetDefault("logging.active_file_name", "out.log")
	viper.SetDefault("rotation.lines", 100)
	viper.SetDefault("rotation.file_size_bytes", 1000000)
	viper.SetDefault("flushing.time_interval_secs", 5)
}
