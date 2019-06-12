package main

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestPing(c *check.C) {
	startAt := time.Now()
	err := s.client.Ping()
	c.Assert(err, check.IsNil)
	costPrintln("TestPing() passed", startAt)
}
