package funnel

import (
	"os"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// Test whether values are being read properly or not
func TestSanity(t *testing.T) {
	v := viper.New()
	v.SetConfigName("goodconfig")
	v.AddConfigPath("./testdata/")

	cfg, _, err := GetConfig(v)
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
		int64(2592000),
		100,
		false,
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
	v := viper.New()
	v.SetConfigName("badsyntaxconfig")
	v.AddConfigPath("./testdata/")

	_, _, err := GetConfig(v)
	if err == nil {
		t.Error("Expected error in config file, got none")
	}
}

func TestInvalidConfigValue(t *testing.T) {
	v := viper.New()
	v.SetConfigName("invalidvalueconfig")
	v.AddConfigPath("./testdata/")

	_, _, err := GetConfig(v)
	if err == nil {
		t.Error("Expected error in config file, got none")
	}
	if serr, ok := err.(*ConfigValueError); ok {
		if serr.Key != LoggingDirectory {
			t.Errorf("Incorrect error key detected. Expected %s, Got %s", LoggingDirectory, serr.Key)
		}
	}
}

func TestNoConfigFile(t *testing.T) {
	v := viper.New()
	v.SetConfigName("noconfig")
	v.AddConfigPath("./testdata/")

	_, _, err := GetConfig(v)
	if err != nil {
		t.Error("Did not expect an error for config file not being present. Got - ", err)
	}

}

func TestEnvVars(t *testing.T) {
	v := viper.New()
	v.SetConfigName("goodconfig")
	v.AddConfigPath("./testdata/")
	envValue := "env_var_value"
	os.Setenv("LOGGING_DIRECTORY", envValue)

	cfg, _, err := GetConfig(v)
	if err != nil {
		t.Fatal(err)
		return
	}
	if cfg.DirName != envValue {
		t.Errorf("Failed to set value from env var. Expected %s, Got %s", envValue, cfg.DirName)
	}
}
