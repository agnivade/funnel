package outputs

// This is redis output writer
import (
	"log/syslog"

	"github.com/agnivade/funnel"
	"github.com/spf13/viper"
	"gopkg.in/redis.v5"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("redis", newRedisOutput)
}

func newRedisOutput(v *viper.Viper, logger *syslog.Writer) (funnel.OutputWriter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     v.GetString("target.host"),
		Password: v.GetString("target.password"),
		DB:       0,
	})

	r := &redisOutput{
		c:       client,
		logger:  logger,
		pubChan: v.GetString("target.channel"),
	}
	return r, nil
}

// redisOutput contains the stuff to publis to redis
type redisOutput struct {
	c       *redis.Client
	logger  *syslog.Writer
	pubChan string
}

// Implementing the OutputWriter interface

func (r *redisOutput) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Doing len(p)-1 to strip off the ending newline
	// Publishing to the channel
	err := r.c.Publish(r.pubChan, string(p[:len(p)-1])).Err()
	return len(p), err
}

func (r *redisOutput) Flush() error {
	return nil
}

func (r *redisOutput) Close() error {
	// Closing the client
	return r.c.Close()
}
