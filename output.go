package funnel

import (
	"bufio"
	"io"
	"log/syslog"

	"github.com/spf13/viper"
)

// OutputWriter interface is to be implemented by all remote target writers
// It embeds the io.Writer and has basic function calls which should be needed by all writers
// Can be modified later if needed
type OutputWriter interface {
	io.Writer
	Flush() error
	Close() error
}

// UnregisteredOutputError holds the error if some target was passed from the config
// which was not registered
type UnregisteredOutputError struct {
	target string
}

func (e *UnregisteredOutputError) Error() string {
	return "Output " + e.target + " was not registered from any module"
}

// OutputFactory is a function type which holds the output registry
type OutputFactory func(v *viper.Viper, logger *syslog.Writer) (OutputWriter, error)

var registeredOutputs = make(map[string]OutputFactory)

// RegisterNewWriter is called by the init function from every output
// Adds the constructor to the registry
func RegisterNewWriter(name string, factory OutputFactory) {
	registeredOutputs[name] = factory
}

// GetOutputWriter gets the constructor by extracting the target.
// Then returns the corresponding output writer by calling the constructor
func GetOutputWriter(v *viper.Viper, logger *syslog.Writer) (OutputWriter, error) {
	target := v.GetString(Target)
	if target == "file" {
		return nil, nil
	}
	w, ok := registeredOutputs[target]
	if !ok {
		return nil, &UnregisteredOutputError{target}
	}
	return w(v, logger)
}

// FileOutput is just an embed type which adds the Close method to buffered writer to satisfy the OutputWriter interface
// XXX: Might need to implement this in a better way
type FileOutput struct {
	*bufio.Writer
}

// Close function is just a no-op. Its never called.
func (f *FileOutput) Close() error {
	return nil
}
