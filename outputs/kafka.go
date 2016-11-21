package outputs

// This is kafka output writer
import (
	"log/syslog"
	"time"

	"github.com/Shopify/sarama"
	"github.com/agnivade/funnel"
	"github.com/spf13/viper"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("kafka", newKafkaOutput)
}

func newKafkaOutput(v *viper.Viper, logger *syslog.Writer) (funnel.OutputWriter, error) {
	// Setting up the kafka config
	cfg := sarama.NewConfig()
	cfg.Producer.Compression = sarama.CompressionGZIP
	if v.IsSet("target.flush_frequency_secs") {
		cfg.Producer.Flush.Frequency = time.Duration(v.GetInt("target.flush_frequency_secs")) * time.Second
	}
	if v.IsSet("target.batch_size") {
		cfg.Producer.Flush.Messages = v.GetInt("target.batch_size")
	}
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = true
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.ClientID = v.GetString("target.clientID")

	brokers := v.GetStringSlice("target.brokers")
	topic := v.GetString("target.topic")
	p, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	// Creating the struct
	k := &kafkaOutput{
		producer: p,
		topic:    topic,
		logger:   logger,
		msgChan:  make(chan *sarama.ProducerMessage),
		done:     make(chan struct{}),
	}
	// Starting the producer loop
	go k.startProducerLoop()
	return k, nil
}

// kafkaOutput contains the stuff to write to kafka
type kafkaOutput struct {
	producer sarama.AsyncProducer
	topic    string
	logger   *syslog.Writer
	msgChan  chan *sarama.ProducerMessage
	done     chan struct{}
}

// Implementing the OutputWriter interface

func (k *kafkaOutput) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Send a msg to the channel
	k.msgChan <- &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(string(p[:len(p)-1])), // stripping off the trailing newline
	}
	return len(p), nil
}

func (k *kafkaOutput) Flush() error {
	return nil
}

func (k *kafkaOutput) Close() error {
	// Send done signal to exit from goroutine
	k.done <- struct{}{}
	// Close producer
	return k.producer.Close()
}

func (k *kafkaOutput) startProducerLoop() {
	for {
		select {
		case msg := <-k.msgChan:
			k.producer.Input() <- msg
		case err := <-k.producer.Errors():
			k.logger.Err(err.Error())
		case <-k.done:
			return
		}
	}
}
