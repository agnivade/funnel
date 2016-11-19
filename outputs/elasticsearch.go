package outputs

// This is the elasticsearch output writer
import (
	"fmt"
	"log/syslog"

	"github.com/agnivade/funnel"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("elasticsearch", newElasticSearchOutput)
}

// This is a wrapper over syslogger to satisfy the logger interface of
// elasticsearch client
type ESLogger struct {
	*syslog.Writer
}

// Printf calls the Err() of the syslog object instead
func (el *ESLogger) Printf(format string, v ...interface{}) {
	el.Err(fmt.Sprintf(format, v))
}

func newElasticSearchOutput(v *viper.Viper, logger *syslog.Writer) (funnel.OutputWriter, error) {
	// Creating elastic client
	c, err := elastic.NewClient(
		elastic.SetURL(v.GetStringSlice("target.nodes")...),
		elastic.SetGzip(true),
		elastic.SetErrorLog(&ESLogger{logger}))
	if err != nil {
		return nil, err
	}

	// Creating the struct
	e := &elasticOutput{
		bulkSvc:   c.Bulk(),
		index:     v.GetString("target.index"),
		indexType: v.GetString("target.type"),
		logger:    logger,
	}
	return e, nil
}

type elasticOutput struct {
	bulkSvc   *elastic.BulkService
	index     string
	indexType string
	logger    *syslog.Writer
}

// Implmenting the OutputWriter interface

func (e *elasticOutput) Write(p []byte) (n int, err error) {
	// Adding a document to the bulk request
	bulkReq := elastic.NewBulkIndexRequest().
		Doc(string(p)).
		Index(e.index).
		Type(e.indexType)
	e.bulkSvc.Add(bulkReq)
	return len(p), nil
}

func (e *elasticOutput) Flush() error {
	// Sends all bulked request to elasticsearch
	_, err := e.bulkSvc.Do(context.TODO())
	return err
}

func (e *elasticOutput) Close() error {
	return nil
}
