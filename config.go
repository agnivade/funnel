package funnel

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// XXX: Move it to constants.go if needed
const (
	APP_NAME = "funnel"

	// config keys
	LOGGING_DIRECTORY            = "logging.directory"
	LOGGING_ACTIVE_FILE_NAME     = "logging.active_file_name"
	ROTATION_MAX_LINES           = "rotation.max_lines"
	ROTATION_MAX_FILE_SIZE_BYTES = "rotation.max_file_size_bytes"
	FLUSHING_TIME_INTERVAL_SECS  = "flushing.time_interval_secs"
)

// ConfigValueError holds the error value if a config key contains
// an invalid value
type ConfigValueError struct {
	key string
}

func (e *ConfigValueError) Error() string {
	return "Invalid config value entered for - " + e.key
}

// Config holds all the config settings
type Config struct {
	DirName        string
	ActiveFileName string

	RotationMaxLines int
	RotationMaxBytes uint64

	FlushingTimeIntervalSecs int
}

// GetConfig returns the config struct which is then passed
// to the consumer
func GetConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/" + APP_NAME + "/")
	viper.AddConfigPath("$HOME/." + APP_NAME)
	viper.AddConfigPath(".")

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
		DirName:                  viper.GetString(LOGGING_DIRECTORY),
		ActiveFileName:           viper.GetString(LOGGING_ACTIVE_FILE_NAME),
		RotationMaxLines:         viper.GetInt(ROTATION_MAX_LINES),
		RotationMaxBytes:         uint64(viper.GetInt64(ROTATION_MAX_FILE_SIZE_BYTES)),
		FlushingTimeIntervalSecs: viper.GetInt(FLUSHING_TIME_INTERVAL_SECS),
	}, nil
}

func setDefaults() {
	viper.SetDefault(LOGGING_DIRECTORY, "log")
	viper.SetDefault(LOGGING_ACTIVE_FILE_NAME, "out.log")
	viper.SetDefault(ROTATION_MAX_LINES, 100)
	viper.SetDefault(ROTATION_MAX_FILE_SIZE_BYTES, 1000000)
	viper.SetDefault(FLUSHING_TIME_INTERVAL_SECS, 5)
}

func validateConfig() error {
	// Validate strings
	for _, key := range []string{
		LOGGING_DIRECTORY,
		LOGGING_ACTIVE_FILE_NAME,
	} {
		// If a string value got successfully converted to integer,
		// then its incorrect
		if _, err := strconv.Atoi(viper.GetString(key)); err == nil {
			return &ConfigValueError{key}
		}
	}

	// Validate integers
	for _, key := range []string{
		ROTATION_MAX_LINES,
		ROTATION_MAX_FILE_SIZE_BYTES,
		FLUSHING_TIME_INTERVAL_SECS,
	} {
		// If an integer value was a string, it would come as zero,
		// hence its invalid
		if viper.GetInt(key) == 0 {
			return &ConfigValueError{key}
		}
	}

	return nil
}
