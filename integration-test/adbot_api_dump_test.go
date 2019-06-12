package main

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestDump(c *check.C) {
	startAt := time.Now()

	_, _, _, _, err := s.client.DebugDump("xxxx")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "400 - .*unsupported dump name.*")

	_, _, _, _, err = s.client.DebugDump("param with space")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "400 - .*unsupported dump name.*")

	_, general, _, _, err := s.client.DebugDump("general")
	c.Assert(err, check.IsNil)
	c.Assert(general, check.NotNil)
	c.Assert(general.UnixTime, check.Not(check.Equals), 0)
	c.Assert(general.NumGoroutines, check.Not(check.Equals), 0)
	c.Assert(general.NumFds, check.Not(check.Equals), 0)
	c.Assert(general.Os, check.Not(check.Equals), "")
	c.Assert(general.GoVersion, check.Not(check.Equals), "")

	_, _, config, _, err := s.client.DebugDump("config")
	c.Assert(err, check.IsNil)
	c.Assert(config, check.NotNil)
	c.Assert(config.Listen, check.Not(check.Equals), "")
	c.Assert(config.Store, check.NotNil)
	c.Assert(config.Store.Type, check.Not(check.Equals), "")

	goroutine, _, _, _, err := s.client.DebugDump("goroutine")
	c.Assert(err, check.IsNil)
	c.Assert(len(goroutine) > 100, check.Equals, true)

	_, _, _, application, err := s.client.DebugDump("application")
	c.Assert(err, check.IsNil)
	c.Assert(len(application) > 0, check.Equals, true)

	costPrintln("TestDump() passed", startAt)
}
