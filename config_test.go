package funnel

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

func init() {
	viper.AddConfigPath("./testdata/")
}

// Test whether values are being read properly or not
func TestSanity(t *testing.T) {
	viper.SetConfigName("goodconfig")

	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
		return
	}

	tests := []interface{}{
		"testdir",
		"testfile",
		100,
		uint64(4509),
		5,
		"",
		"timestamp",
	}

	cfgValue := reflect.ValueOf(cfg).Elem()

	// Iterating through the properties to check everything is good
	for i := 0; i < cfgValue.NumField(); i++ {
		v := cfgValue.Field(i).Interface()
		if v != tests[i] {
			t.Errorf("Incorrect value from config. Expected %s, Got %s", tests[i], v)
		}
	}
}

func TestBadFile(t *testing.T) {
	viper.SetConfigName("badsyntaxconfig")

	_, err := GetConfig()
	if err == nil {
		t.Error("Expected error in config file, got none")
	}
}

func TestInvalidConfigValue(t *testing.T) {
	viper.SetConfigName("invalidvalueconfig")

	_, err := GetConfig()
	if err == nil {
		t.Error("Expected error in config file, got none")
	}
	if serr, ok := err.(*ConfigValueError); ok {
		if serr.Key != LoggingDirectory {
			t.Errorf("Incorrect error key detected. Expected %s, Got %s", LoggingDirectory, serr.Key)
		}
	}
}

// TODO: We need to pass individual viper instances for this
// marking for later
func TestNoConfigFile(t *testing.T) {
}

func TestEnvVars(t *testing.T) {
	viper.SetConfigName("goodconfig")
	envValue := "env_var_value"
	os.Setenv("LOGGING_DIRECTORY", envValue)

	cfg, err := GetConfig()
	if err != nil {
		t.Fatal(err)
		return
	}
	if cfg.DirName != envValue {
		t.Errorf("Failed to set value from env var. Expected %s, Got %s", envValue, cfg.DirName)
	}
}
