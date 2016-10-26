package funnel

import (
	"errors"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// XXX: Move it to constants.go if needed
const (
	AppName = "funnel"

	// config keys
	LoggingDirectory         = "logging.directory"
	LoggingActiveFileName    = "logging.active_file_name"
	RotationMaxLines         = "rotation.max_lines"
	RotationMaxFileSizeBytes = "rotation.max_file_size_bytes"
	FlushingTimeIntervalSecs = "flushing.time_interval_secs"
	PrependValue             = "misc.prepend_value"
	FileRenamePolicy         = "rollup.file_rename_policy"
	Gzip                     = "rollup.gzip"
)

// ConfigValueError holds the error value if a config key contains
// an invalid value
type ConfigValueError struct {
	Key string
}

func (e *ConfigValueError) Error() string {
	return "Invalid config value entered for - " + e.Key
}

// Config holds all the config settings
type Config struct {
	DirName        string
	ActiveFileName string

	RotationMaxLines int
	RotationMaxBytes uint64

	FlushingTimeIntervalSecs int

	PrependValue string

	FileRenamePolicy string
	Gzip             bool
}

// Setting the config file name and the locations to search for the config
func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + AppName + "/")
	viper.AddConfigPath("$HOME/." + AppName)
	viper.AddConfigPath(".")
}

// GetConfig returns the config struct which is then passed
// to the consumer
func GetConfig() (*Config, error) {
	// Set default values. They are overridden by config file values, if provided
	setDefaults()

	// Find and read the config file
	err := viper.ReadInConfig()
	// Return the error only if config file is present
	if err != nil && viper.ConfigFileUsed() != "" {
		return nil, err
	}

	// Read from env vars
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Validate
	if err := validateConfig(); err != nil {
		return nil, err
	}

	// return struct
	return &Config{
		DirName:                  viper.GetString(LoggingDirectory),
		ActiveFileName:           viper.GetString(LoggingActiveFileName),
		RotationMaxLines:         viper.GetInt(RotationMaxLines),
		RotationMaxBytes:         uint64(viper.GetInt64(RotationMaxFileSizeBytes)),
		FlushingTimeIntervalSecs: viper.GetInt(FlushingTimeIntervalSecs),
		PrependValue:             viper.GetString(PrependValue),
		FileRenamePolicy:         viper.GetString(FileRenamePolicy),
		Gzip:                     viper.GetBool(Gzip),
	}, nil
}

func setDefaults() {
	viper.SetDefault(LoggingDirectory, "log")
	viper.SetDefault(LoggingActiveFileName, "out.log")
	viper.SetDefault(RotationMaxLines, 100)
	viper.SetDefault(RotationMaxFileSizeBytes, 1000000)
	viper.SetDefault(FlushingTimeIntervalSecs, 5)
	viper.SetDefault(PrependValue, "")
	viper.SetDefault(FileRenamePolicy, "timestamp")
	viper.SetDefault(Gzip, false)
}

func validateConfig() error {
	// Validate strings
	for _, key := range []string{
		LoggingDirectory,
		LoggingActiveFileName,
		PrependValue,
		FileRenamePolicy,
	} {
		// If a string value got successfully converted to integer,
		// then its incorrect
		if _, err := strconv.Atoi(viper.GetString(key)); err == nil {
			return &ConfigValueError{key}
		}

		// File rename policy has to be either timestamp or serial
		if key == FileRenamePolicy &&
			(viper.GetString(key) != "timestamp" && viper.GetString(key) != "serial") {
			return errors.New(key + " can only be timestamp or serial")
		}
	}

	// Validate integers
	for _, key := range []string{
		RotationMaxLines,
		RotationMaxFileSizeBytes,
		FlushingTimeIntervalSecs,
	} {
		// If an integer value was a string, it would come as zero,
		// hence its invalid
		if viper.GetInt(key) == 0 {
			return &ConfigValueError{key}
		}
	}

	return nil
}
