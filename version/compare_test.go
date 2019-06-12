package version

import (
	"testing"

	check "gopkg.in/check.v1"
)

var _ = check.Suite(new(vcSuit))

type vcSuit struct{}

func TestVersionCompare(t *testing.T) {
	check.TestingT(t)
}

func (s *vcSuit) TestVersionCompare(c *check.C) {
	c.Assert(Equal("0.2-beta1", "0.2-beta1"), check.Equals, true)
	c.Assert(LessThanOrEqualTo("0.2-beta1", "0.2-beta1"), check.Equals, true)
	c.Assert(GreaterThanOrEqualTo("0.2-beta1", "0.2-beta1"), check.Equals, true)
	c.Assert(Equal("0.2-beta1", "0.2-beta2"), check.Equals, false)
	c.Assert(LessThan("0.2-beta1", "0.2-beta2"), check.Equals, true)
	c.Assert(LessThanOrEqualTo("0.2-beta1", "0.2-beta2"), check.Equals, true)
	c.Assert(GreaterThan("0.2-beta1", "0.2-beta2"), check.Equals, false)
	c.Assert(GreaterThanOrEqualTo("0.2-beta1", "0.2-beta2"), check.Equals, false)

	c.Assert(Equal("0.2.beta1", "0.2-beta1"), check.Equals, true)
	c.Assert(LessThanOrEqualTo("0.2.beta1", "0.2-beta1"), check.Equals, true)
	c.Assert(GreaterThanOrEqualTo("0.2.beta1", "0.2-beta1"), check.Equals, true)
	c.Assert(Equal("0.2.beta1", "0.2-beta2"), check.Equals, false)
	c.Assert(LessThan("0.2.beta1", "0.2-beta2"), check.Equals, true)
	c.Assert(LessThanOrEqualTo("0.2.beta1", "0.2-beta2"), check.Equals, true)
	c.Assert(GreaterThan("0.2.beta1", "0.2-beta2"), check.Equals, false)
	c.Assert(GreaterThanOrEqualTo("0.2.beta1", "0.2-beta2"), check.Equals, false)

	c.Assert(Equal("1.12", "1.12"), check.Equals, true)
	c.Assert(Equal("1.0.0", "1"), check.Equals, true)
	c.Assert(Equal("1.m", "1.0.0.abc"), check.Equals, true)

	c.Assert(Equal("1.0.2", "1.0.3"), check.Equals, false)
	c.Assert(LessThan("1.0.2", "1.0.3"), check.Equals, true)
	c.Assert(LessThanOrEqualTo("1.0.2", "1.0.3"), check.Equals, true)
}
