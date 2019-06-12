package main

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestVersion(c *check.C) {
	startAt := time.Now()

	version, err := s.client.Version()
	c.Assert(err, check.IsNil)
	c.Assert(version.GitCommit, check.Not(check.Equals), "")
	c.Assert(version.BuildTime, check.Not(check.Equals), "")
	c.Assert(version.GoVersion, check.Not(check.Equals), "")

	costPrintln("TestVersion() passed", startAt)
}
