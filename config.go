package funnel

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// XXX: Move it to constants.go if needed
const (
	// config keys
	LoggingDirectory         = "logging.directory"
	LoggingActiveFileName    = "logging.active_file_name"
	RotationMaxLines         = "rotation.max_lines"
	RotationMaxFileSizeBytes = "rotation.max_file_size_bytes"
	FlushingTimeIntervalSecs = "flushing.time_interval_secs"
	PrependValue             = "misc.prepend_value"
	FileRenamePolicy         = "rollup.file_rename_policy"
	MaxAge                   = "rollup.max_age"
	MaxCount                 = "rollup.max_count"
	Gzip                     = "rollup.gzip"
)

var (
	// ErrInvalidFileRenamePolicy is raised for invalid values to file rename policy
	ErrInvalidFileRenamePolicy = errors.New(FileRenamePolicy + " can only be timestamp or serial")
	// ErrInvalidMaxAge is raised for invalid value in max age - life bad suffixes or no integer value at all
	ErrInvalidMaxAge = errors.New(MaxAge + " must end with either d or h and start with a number")
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
	MaxAge           int64
	MaxCount         int
	Gzip             bool
}

func init() {

}

// GetConfig returns the config struct which is then passed
// to the consumer
func GetConfig(v *viper.Viper) (*Config, chan *Config, error) {
	// Set default values. They are overridden by config file values, if provided
	setDefaults(v)
	// Create a chan to signal any config reload events
	reloadChan := make(chan *Config)

	// Find and read the config file
	err := v.ReadInConfig()
	// Return the error only if config file is present
	if err != nil && v.ConfigFileUsed() != "" {
		return nil, reloadChan, err
	}

	// Read from env vars
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Validate
	if err := validateConfig(v); err != nil {
		return nil, reloadChan, err
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			if err := validateConfig(v); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
			reloadChan <- getConfigStruct(v)
		}
	})

	// return struct
	return getConfigStruct(v), reloadChan, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault(LoggingDirectory, "log")
	v.SetDefault(LoggingActiveFileName, "out.log")
	v.SetDefault(RotationMaxLines, 100)
	v.SetDefault(RotationMaxFileSizeBytes, 1000000)
	v.SetDefault(FlushingTimeIntervalSecs, 5)
	v.SetDefault(PrependValue, "")
	v.SetDefault(FileRenamePolicy, "timestamp")
	v.SetDefault(MaxAge, "30d")
	v.SetDefault(MaxCount, 100)
	v.SetDefault(Gzip, false)
}

func validateConfig(v *viper.Viper) error {
	// Validate strings
	for _, key := range []string{
		LoggingDirectory,
		LoggingActiveFileName,
		PrependValue,
		FileRenamePolicy,
		MaxAge,
	} {
		// If a string value got successfully converted to integer,
		// then its incorrect
		if _, err := strconv.Atoi(v.GetString(key)); err == nil {
			return &ConfigValueError{key}
		}

		// File rename policy has to be either timestamp or serial
		if key == FileRenamePolicy &&
			(v.GetString(key) != "timestamp" && v.GetString(key) != "serial") {
			return ErrInvalidFileRenamePolicy
		}
	}

	// Validate integers
	for _, key := range []string{
		RotationMaxLines,
		RotationMaxFileSizeBytes,
		FlushingTimeIntervalSecs,
		MaxCount,
	} {
		// If an integer value was a string, it would come as zero,
		// hence its invalid
		if v.GetInt(key) == 0 {
			return &ConfigValueError{key}
		}
	}

	// Validate MaxAge
	maxAge := v.GetString(MaxAge)
	unit := maxAge[len(maxAge)-1:]
	_, err := strconv.Atoi(maxAge[0 : len(maxAge)-1])
	if err != nil {
		return ErrInvalidMaxAge
	}

	if unit != "d" && unit != "h" {
		return ErrInvalidMaxAge
	}

	return nil
}

func getConfigStruct(v *viper.Viper) *Config {
	return &Config{
		DirName:                  v.GetString(LoggingDirectory),
		ActiveFileName:           v.GetString(LoggingActiveFileName),
		RotationMaxLines:         v.GetInt(RotationMaxLines),
		RotationMaxBytes:         uint64(v.GetInt64(RotationMaxFileSizeBytes)),
		FlushingTimeIntervalSecs: v.GetInt(FlushingTimeIntervalSecs),
		PrependValue:             v.GetString(PrependValue),
		FileRenamePolicy:         v.GetString(FileRenamePolicy),
		MaxAge:                   getMaxAgeValue(v.GetString(MaxAge)),
		MaxCount:                 v.GetInt(MaxCount),
		Gzip:                     v.GetBool(Gzip),
	}
}

func getMaxAgeValue(maxAge string) int64 {
	unit := maxAge[len(maxAge)-1:]
	magnitude, _ := strconv.Atoi(maxAge[0 : len(maxAge)-1])

	if unit == "d" {
		return int64(magnitude) * 24 * 60 * 60
	} else {
		return int64(magnitude) * 60 * 60
	}
}
