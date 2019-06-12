package main

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ApiSuite) TestPanic(c *check.C) {
	startAt := time.Now()
	err := s.client.Panic()
	c.Assert(err, check.IsNil) // make panic
	err = s.client.Ping()
	c.Assert(err, check.IsNil) // make sure server handled the panic and still alive
	costPrintln("TestPanic() passed", startAt)
}
