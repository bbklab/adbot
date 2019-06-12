package qrcode

import (
	"testing"

	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/pkg/utils"
)

var _ = check.Suite(new(testSuite))

type testSuite struct{}

func TestQRCode(t *testing.T) {
	check.TestingT(t)
}

func (s *testSuite) TestQRCodeEncodeDecode(c *check.C) {
	datas := []string{}

	for i := 1; i <= 100; i++ {
		text, err := utils.GenPassword(48 + i)
		c.Assert(err, check.Equals, nil)
		datas = append(datas, text)
	}

	for _, data := range datas {
		png, err := Encode(data)
		c.Assert(err, check.Equals, nil)
		c.Assert(png, check.Not(check.Equals), nil)

		// text, err := Decode(png)
		// c.Assert(err, check.Equals, nil)
		// c.Assert(text, check.Equals, data)
	}
}
