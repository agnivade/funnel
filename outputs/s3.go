// +build !disables3

package outputs

// This is the aws s3 output writer
import (
	"bytes"
	"log/syslog"
	"strings"
	"time"

	"github.com/agnivade/funnel"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("s3", newS3Output)
}

func newS3Output(v *viper.Viper, logger *syslog.Writer) (funnel.OutputWriter, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(v.GetString("target.region"))})
	if err != nil {
		return nil, err
	}

	svc := s3.New(sess)

	s3o := &s3Output{
		svc:    svc,
		logger: logger,
		bucket: v.GetString("target.bucket"),
	}
	return s3o, nil
}

// s3Output contains the stuff to put objects to s3
type s3Output struct {
	svc    *s3.S3
	logger *syslog.Writer
	buffer bytes.Buffer
	bucket string
}

// Implmenting the OutputWriter interface
func (s3o *s3Output) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	return s3o.buffer.Write(p)
}

func (s3o *s3Output) Flush() error {
	t := time.Now()
	key := t.Format("15_04_05.00000-2006_01_02") + ".log"
	_, err := s3o.svc.PutObject(&s3.PutObjectInput{
		Body:   strings.NewReader(s3o.buffer.String()),
		Bucket: &s3o.bucket,
		Key:    &key,
	})
	// Resetting the buffer
	s3o.buffer.Reset()
	return err
}

func (s3o *s3Output) Close() error {
	return nil
}
