// +build !disableinfluxdb

package outputs

import (
	"encoding/json"
	"errors"
	"log/syslog"
	"time"

	"github.com/agnivade/funnel"
	influxdb "github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

// Registering the constructor function
func init() {
	funnel.RegisterNewWriter("influxdb", newInfluxDBOutput)
}

func newInfluxDBOutput(v *viper.Viper, logger *syslog.Writer) (funnel.OutputWriter, error) {
	var c influxdb.Client
	var err error
	if v.GetString("target.protocol") == "http" {
		c, err = influxdb.NewHTTPClient(influxdb.HTTPConfig{
			Addr:     v.GetString("target.host"),
			Username: v.GetString("target.username"),
			Password: v.GetString("target.password"),
		})
		if err != nil {
			return nil, err
		}
	} else if v.GetString("target.protocol") == "udp" {
		c, err = influxdb.NewUDPClient(influxdb.UDPConfig{
			Addr: v.GetString("target.host"),
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Invalid target protocol " + v.GetString("target.protocol"))
	}

	bp, err := getNewBatch(v.GetString("target.db"), v.GetString("target.time_precision"))
	if err != nil {
		return nil, err
	}

	return &influxDBOutput{
		client:    c,
		batchPts:  bp,
		database:  v.GetString("target.db"),
		precision: v.GetString("target.time_precision"),
		metric:    v.GetString("target.metric"),
		protocol:  v.GetString("target.protocol"),
	}, nil
}

type influxDBOutput struct {
	client    influxdb.Client
	batchPts  influxdb.BatchPoints
	database  string
	precision string
	metric    string
	protocol  string
}

type influxDBLine struct {
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
}

// Implementing the OutputWriter interface

func (i *influxDBOutput) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	// Unmarshalling the json line to struct
	line := influxDBLine{}
	err := json.Unmarshal(p, &line)
	if err != nil {
		return 0, err
	}
	// Constructing the new point
	pt, err := influxdb.NewPoint(i.metric, line.Tags, line.Fields, time.Now())
	if err != nil {
		return 0, err
	}
	// Adding to the batch
	i.batchPts.AddPoint(pt)
	return len(p), nil
}

func (i *influxDBOutput) Flush() error {
	// Flushing the current batch
	err := i.client.Write(i.batchPts)
	if err != nil {
		return err
	}
	// Creating new batch
	bp, err := getNewBatch(i.database, i.precision)
	if err != nil {
		return err
	}
	i.batchPts = bp
	return nil
}

func (i *influxDBOutput) Close() error {
	return i.client.Close()
}

// helper function to return new batch points for every batch
func getNewBatch(db, precision string) (influxdb.BatchPoints, error) {
	return influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  db,
		Precision: precision,
	})
}
