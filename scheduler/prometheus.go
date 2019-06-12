package scheduler

import (
	"fmt"
	"math"
	"time"

	promodel "github.com/prometheus/common/model"

	"github.com/bbklab/adbot/pkg/prometheus"
)

var (
	metricTraffic = "inf_shadowsocks_service_traffic"
)

type promClient struct {
	client *prometheus.Client
}

func newPromClient(address string) (*promClient, error) {
	client, err := prometheus.NewClient(address)
	if err != nil {
		return nil, err
	}

	return &promClient{client: client}, nil
}

func (c *promClient) removeSeries(matches []string) error {
	start, _ := time.Parse(time.RFC3339, "2018-01-01T00:00:00Z")
	return c.client.RemoveSeries(matches, start, time.Now(), time.Second*10)
}

// ss node traffic
//

func (c *promClient) ssNodeTrafficsMonth(nid string) (float64, promodel.Matrix, error) {
	var (
		start, end = thisMonth()
		step       = time.Hour * 6 // (24/6)*30=120 samples
		timeout    = time.Second * 10
	)
	return c.ssNodeTrafficsRangeQuery(nid, start, end, step, timeout)
}

func (c *promClient) ssNodeTrafficsToday(nid string) (float64, promodel.Matrix, error) {
	var (
		start, end = thisDay()
		step       = time.Minute * 10 // 24*(60/10)=144 samples
		timeout    = time.Second * 10
	)
	return c.ssNodeTrafficsRangeQuery(nid, start, end, step, timeout)
}

func (c *promClient) ssNodeTrafficsRangeQuery(nid string, start, end time.Time, step, timeout time.Duration) (float64, promodel.Matrix, error) {
	var (
		expr = fmt.Sprintf("%s{node_id=\"%s\"}", metricTraffic, nid)
	)

	result, err := c.client.QueryRange(expr, start, end, step, timeout)
	if err != nil {
		return 0, promodel.Matrix{}, err
	}

	data, ok := result.(promodel.Matrix)
	if !ok {
		return 0, promodel.Matrix{}, fmt.Errorf("prometheus.QueryRange() on [%s] got unexpected value type %q", expr, data.Type())
	}

	if data.Len() < 1 { // empty list
		return 0, data, nil
	}

	// try to query the latest value and merge into the range query result
	endRes, err := c.client.QueryAt(expr, end, timeout)
	if err == nil {
		if endData, ok := endRes.(promodel.Vector); ok {
			mergeVecIntoMatrix(&data, &endData)
		}
	}

	var total float64
	for _, serie := range data { // node_id maybe have multi series of sample stream (multi services on one node)
		total += caculateRangeSeriesTraffic(serie.Values)
	}
	return total, data, nil

}

// note: actually no need, prometheus GCed in 31 days
//func (c *promClient) removeSSNodeTrafficsSeries(nid string) error {
//query := fmt.Sprintf("%s{node_id=\"%s\"}", metricTraffic, nid)
//return c.removeSeries([]string{query})
//}

// ss service traffic
//

func (c *promClient) ssServiceTrafficsMonth(sid string) (float64, promodel.Matrix, error) {
	var (
		start, end = thisMonth()
		step       = time.Hour * 6 // (24/6)*30=120 samples
		timeout    = time.Second * 10
	)
	return c.ssServiceTrafficsRangeQuery(sid, start, end, step, timeout)
}

func (c *promClient) ssServiceTrafficsToday(sid string) (float64, promodel.Matrix, error) {
	var (
		start, end = thisDay()
		step       = time.Minute * 10 // 24*(60/10)=144 samples
		timeout    = time.Second * 10
	)
	return c.ssServiceTrafficsRangeQuery(sid, start, end, step, timeout)
}

func (c *promClient) ssServiceTrafficsRangeQuery(sid string, start, end time.Time, step, timeout time.Duration) (float64, promodel.Matrix, error) {
	var (
		expr = fmt.Sprintf("%s{service_id=\"%s\"}", metricTraffic, sid)
	)

	result, err := c.client.QueryRange(expr, start, end, step, timeout)
	if err != nil {
		return 0, promodel.Matrix{}, err
	}

	data, ok := result.(promodel.Matrix)
	if !ok {
		return 0, promodel.Matrix{}, fmt.Errorf("prometheus.QueryRange() on [%s] got unexpected value type %q", expr, data.Type())
	}

	if data.Len() < 1 { // empty list
		return 0, data, nil
	}

	// try to query the latest value and merge into the range query result
	endRes, err := c.client.QueryAt(expr, end, timeout)
	if err == nil {
		if endData, ok := endRes.(promodel.Vector); ok {
			mergeVecIntoMatrix(&data, &endData)
		}
	}

	var values = data[0].Values // node_id + service_id should have only one serie
	return caculateRangeSeriesTraffic(values), data, nil
}

// note: actually no need, prometheus GCed in 31 days
//func (c *promClient) removeSSServiceTrafficsSeries(nid, sid string) error {
//query := fmt.Sprintf("%s{node_id=\"%s\",service_id=\"%s\"}", metricTraffic, nid, sid)
//return c.removeSeries([]string{query})
//}

// ss account traffic
//

func (c *promClient) ssAccountTrafficsMonth(acctID string) (float64, promodel.Matrix, error) {
	var (
		start, end = thisMonth()
		step       = time.Hour * 6 // (24/6)*30=120 samples
		timeout    = time.Second * 10
	)
	return c.ssAccountTrafficsRangeQuery(acctID, start, end, step, timeout)
}

func (c *promClient) ssAccountTrafficsToday(acctID string) (float64, promodel.Matrix, error) {
	var (
		start, end = thisDay()
		step       = time.Minute * 10 // 24*(60/10)=144 samples
		timeout    = time.Second * 10
	)
	return c.ssAccountTrafficsRangeQuery(acctID, start, end, step, timeout)
}

func (c *promClient) ssAccountTrafficsRangeQuery(acctID string, start, end time.Time, step, timeout time.Duration) (float64, promodel.Matrix, error) {
	// query range
	var (
		expr = fmt.Sprintf("%s{account_id=\"%s\"}", metricTraffic, acctID)
	)
	result, err := c.client.QueryRange(expr, start, end, step, timeout)
	if err != nil {
		return 0, promodel.Matrix{}, err
	}

	data, ok := result.(promodel.Matrix)
	if !ok {
		return 0, promodel.Matrix{}, fmt.Errorf("prometheus.QueryRange() on [%s] got unexpected value type %q", expr, data.Type())
	}

	if data.Len() < 1 { // empty list
		return 0, data, nil
	}

	// try to query the latest value and merge into the range query result
	endRes, err := c.client.QueryAt(expr, end, timeout)
	if err == nil {
		if endData, ok := endRes.(promodel.Vector); ok {
			mergeVecIntoMatrix(&data, &endData)
		}
	}

	var total float64
	for _, serie := range data { // account_id maybe have multi series of sample stream (multi services on multi nodes)
		total += caculateRangeSeriesTraffic(serie.Values)
	}
	return total, data, nil

}

func (c *promClient) removeSSAccountTrafficsSeries(acctID string) error {
	query := fmt.Sprintf("%s{account_id=\"%s\"}", metricTraffic, acctID)
	return c.removeSeries([]string{query})
}

// utils
//
func mergeVecIntoMatrix(rangeRes *promodel.Matrix, endRes *promodel.Vector) {
	if rangeRes == nil || endRes == nil {
		return
	}

	// get end matched sample from `endRes`
	fgetMatchedEndRes := func(m promodel.Metric) *promodel.Sample {
		for _, sample := range *endRes {
			if sample.Metric.Equal(m) {
				return sample
			}
		}
		return nil
	}

	// put the matched end value into the range results list
	for idx, sampleStream := range *rangeRes {
		endSample := fgetMatchedEndRes(sampleStream.Metric)
		if endSample != nil { // rewrite current value list
			(*rangeRes)[idx].Values = append((*rangeRes)[idx].Values, promodel.SamplePair{
				Timestamp: endSample.Timestamp,
				Value:     endSample.Value,
			})
		}
	}
}

func thisMonth() (time.Time, time.Time) {
	end := time.Now()
	start := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.Local) // yyyy-mm-01 00:00:00
	return start, end
}

func thisDay() (time.Time, time.Time) {
	end := time.Now()
	start := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.Local) // yyyy-mm-dd 00:00:00
	return start, end
}

func caculateRangeSeriesTraffic(values []promodel.SamplePair) float64 {
	if len(values) == 0 {
		return 0
	}

	var (
		ret  float64
		prev = float64(-1) // not set
	)

	for _, value := range values {
		curr := float64(value.Value)

		if isNaNInf(curr) { // ignore abnormal value
			continue
		}
		if curr == 0 { // ignore zero value
			continue
		}
		if prev < 0 { // if prev not set
			prev = curr // set and continue
			continue
		}

		if flow := curr - prev; flow >= 0 {
			ret += flow // increasing
		} else {
			ret += curr // decreasing ? maybe service restarted, treat `current traffic` as `newly increased`
		}
		prev = curr
	}

	return ret
}

func isNaNInf(f float64) bool {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return true
	}
	return false
}
