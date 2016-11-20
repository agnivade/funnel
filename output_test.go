package funnel

import (
	"log/syslog"
	"testing"

	"github.com/spf13/viper"
)

// get nil outputwriter with file
func TestNilOutputWriter(t *testing.T) {
	v := viper.New()
	v.Set(Target, "file")

	logger, _ := syslog.New(syslog.LOG_ERR, "test")

	ow, err := GetOutputWriter(v, logger)
	if ow != nil {
		t.Errorf("Expected nil outputwriter, Got %s", ow)
	}

	if err != nil {
		t.Errorf("Expected nil error, Got %s", err)
	}
}

// get unregistered
func TestUnregisteredOutputWriter(t *testing.T) {
	v := viper.New()
	v.Set(Target, "somethingnotthere")

	logger, _ := syslog.New(syslog.LOG_ERR, "test")

	ow, err := GetOutputWriter(v, logger)
	if ow != nil {
		t.Errorf("Expected nil outputwriter, Got %s", ow)
	}

	if _, ok := err.(*UnregisteredOutputError); !ok {
		t.Errorf("Expected error to be UnregisteredOutputError, Got %s", err)
	}
}

// get registered
func TestRegisteredOutputWriter(t *testing.T) {
	target := "test"
	v := viper.New()
	v.Set(Target, target)

	RegisterNewWriter(target, newTestOutput)

	logger, _ := syslog.New(syslog.LOG_ERR, "test")

	ow, err := GetOutputWriter(v, logger)
	if err != nil {
		t.Errorf("Expected nil error, Got %s", err)
	}
	if _, ok := ow.(*testOutput); !ok {
		t.Errorf("Expected outputwriter to be testOutput, Got %s", ow)
	}
}

// Dummy function and struct types to test out the output registration
func newTestOutput(v *viper.Viper, logger *syslog.Writer) (OutputWriter, error) {
	return &testOutput{}, nil
}

type testOutput struct {
}

// Implementing the OutputWriter interface
func (k *testOutput) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (k *testOutput) Flush() error {
	return nil
}

func (k *testOutput) Close() error {
	return nil
}
