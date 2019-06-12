package main

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestInfo(c *check.C) {
	startAt := time.Now()
	info, err := s.client.Info()
	c.Assert(err, check.IsNil)
	c.Assert(info.Version, check.Not(check.Equals), "")
	c.Assert(len(info.Listens), check.Equals, 2)
	c.Assert(info.Uptime, check.Not(check.Equals), "")
	c.Assert(info.StoreTyp, check.Not(check.Equals), "")
	costPrintln("TestInfo() passed", startAt)
}
