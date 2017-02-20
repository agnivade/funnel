// +build !nats

package outputs

// This is the nats output writer
import (
	"log/syslog"

	"github.com/agnivade/funnel"
	"github.com/nats-io/go-nats"
	"github.com/spf13/viper"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("nats", newNATSOutput)
}

func newNATSOutput(v *viper.Viper, logger *syslog.Writer) (funnel.OutputWriter, error) {
	o := nats.Options{
		Url:      "nats://" + v.GetString("target.host") + ":" + v.GetString("target.port"),
		User:     v.GetString("target.user"),
		Password: v.GetString("target.password"),
	}

	c, err := o.Connect()
	if err != nil {
		return nil, err
	}

	n := &natsOutput{
		client:  c,
		logger:  logger,
		subject: v.GetString("target.subject"),
	}
	return n, nil
}

// natsOutput contains the stuff to publish to nats server
type natsOutput struct {
	client  *nats.Conn
	logger  *syslog.Writer
	subject string
}

// Implementing the OutputWriter interface

func (n *natsOutput) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	err := n.client.Publish(n.subject, p[:len(p)-1])
	return len(p), err
}

func (n *natsOutput) Flush() error {
	return n.client.Flush()
}

func (n *natsOutput) Close() error {
	// Closing the client
	n.client.Close()
	return nil
}
