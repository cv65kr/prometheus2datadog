package internal

import (
	"github.com/DataDog/datadog-go/v5/statsd"
)

type datadogClient struct {
	statsd *statsd.Client
}

func NewDatadogClient(addr string) (*datadogClient, error) {
	statsd, err := statsd.New(addr)
	if err != nil {
		return nil, err
	}

	return &datadogClient{statsd: statsd}, nil
}

func (c *datadogClient) Flush() error {
	return c.statsd.Flush()
}

func (c *datadogClient) CloseConnection() error {
	return c.statsd.Close()
}
