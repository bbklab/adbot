package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Client queries the prometheus metrics
type Client struct {
	api v1.API
}

// NewClient creates the prometheus client
func NewClient(addr string) (*Client, error) {
	cli, err := api.NewClient(api.Config{
		Address: addr,
	})
	if err != nil {
		return nil, err
	}

	c := new(Client)
	c.api = v1.NewAPI(cli)

	err = c.ping()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) ping() error {
	_, err := c.api.Config(context.Background())
	return err
}

// Query is exported
func (c *Client) Query(expr string, timeout time.Duration) (model.Value, error) {
	return c.QueryAt(expr, time.Now(), timeout)
}

// QueryAt is exported
func (c *Client) QueryAt(expr string, at time.Time, timeout time.Duration) (model.Value, error) {
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()
	return c.api.Query(ctx, expr, at)
}

// QueryRange is exported
func (c *Client) QueryRange(expr string, start, end time.Time, step, timeout time.Duration) (model.Value, error) {
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()
	return c.api.QueryRange(ctx, expr, v1.Range{Start: start, End: end, Step: step})
}

// ListSeries is exported
func (c *Client) ListSeries(matches []string, start, end time.Time, timeout time.Duration) ([]model.LabelSet, error) {
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()
	return c.api.Series(ctx, matches, start, end)
}

// RemoveSeries is exported
func (c *Client) RemoveSeries(matches []string, start, end time.Time, timeout time.Duration) error {
	ctx, done := context.WithTimeout(context.Background(), timeout)
	defer done()
	return c.api.DeleteSeries(ctx, matches, start, end)
}
