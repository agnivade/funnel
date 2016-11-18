package outputs

// This is kafka output writer
import (
	"fmt"

	"github.com/agnivade/funnel"
	"github.com/spf13/viper"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("kafka", newKafkaOutput)
}

func newKafkaOutput(v *viper.Viper) (funnel.OutputWriter, error) {
	return &kafkaOutput{}, nil
}

// kafkaOutput contains the stuff to write to kafka
type kafkaOutput struct {
}

// Implementing the OutputWriter interface

func (k *kafkaOutput) Write(p []byte) (n int, err error) {
	fmt.Printf("%s\n", p)
	return 2, nil
}

func (k *kafkaOutput) Flush() error {
	fmt.Println("flushing")
	return nil
}

func (k *kafkaOutput) Close() error {
	fmt.Println("closing")
	return nil
}
