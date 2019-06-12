package main

import (
	"io/ioutil"
	"strings"
	"time"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestMetrics(c *check.C) {
	startAt := time.Now()

	stream, err := s.client.Metrics()
	c.Assert(err, check.IsNil)
	c.Assert(stream, check.NotNil)
	defer stream.Close()

	bs, err := ioutil.ReadAll(stream)
	c.Assert(err, check.IsNil)
	c.Assert(strings.Contains(string(bs), "inf_node_count_total"), check.Equals, true)

	costPrintln("TestMetrics() passed", startAt)
}
