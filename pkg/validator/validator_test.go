package validator

import (
	"testing"

	check "gopkg.in/check.v1"
)

var _ = check.Suite(new(testSuit))

type testSuit struct{}

func TestAll(t *testing.T) {
	check.TestingT(t)
}

func (s *testSuit) TestStringValidate(c *check.C) {
	datas := map[*strValidator]string{
		{"", -1, -1, nil}:                "",
		{"", 10, -1, nil}:                "required, can't be empty",
		{"abc", 0, 0, nil}:               "",
		{"xx^x", 0, 0, NormalCharacters}: "character .* not allowed",
		{"abc", 0, 2, nil}:               "length can't larger than 2",
		{"abc", 0, 3, nil}:               "",
		{"a", 2, 0, nil}:                 "length can't smaller than 2",
		{"a", 1, 0, nil}:                 "",
		{"xyz", 3, 3, NormalCharacters}:  "",
		{"xyz", 1, 5, NormalCharacters}:  "",
		{"xyz", 1, 5, []rune("xym")}:     "character .* not allowed",
	}

	for obj, errmsg := range datas {
		err := String(obj.str, obj.min, obj.max, obj.chars)
		if errmsg == "" {
			c.Assert(err, check.IsNil)
		} else {
			c.Logf("%s,%d,%d,%s    %s", obj.str, obj.min, obj.max, string(obj.chars), errmsg)
			c.Assert(err, check.NotNil)
			c.Assert(err, check.ErrorMatches, errmsg)
		}
	}
}

func (s *testSuit) TestIntValidate(c *check.C) {
	datas := map[*intValidator]string{
		{10, 1, 10}: "",
		{1, 1, 10}:  "",
		{5, 1, 10}:  "",
		{0, 1, 10}:  "0 must be numberic between .*1,10.*",
		{11, 1, 10}: "11 must be numberic between .*1,10.*",
		{5, 10, 1}:  "5 must be numberic between .*10,1.*", // always got this error
	}

	for obj, errmsg := range datas {
		err := Int(obj.val, obj.min, obj.max)
		if errmsg == "" {
			c.Assert(err, check.IsNil)
		} else {
			c.Logf("%d,%d,%d    %s", obj.val, obj.min, obj.max, errmsg)
			c.Assert(err, check.NotNil)
			c.Assert(err, check.ErrorMatches, errmsg)
		}
	}
}
