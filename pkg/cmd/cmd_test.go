package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	check "gopkg.in/check.v1"
)

var _ = check.Suite(new(cmdSuit))

type cmdSuit struct{}

func TestRunCmd(t *testing.T) {
	check.TestingT(t)
}

func (s *cmdSuit) TestRunCmd(c *check.C) {
	tempdir := os.TempDir() + "/" + random(32)
	os.Mkdir(tempdir, 0644)
	defer os.RemoveAll(tempdir)

	testData := []struct {
		cmd     string // will be run with `sh -c`
		stdout  string // regex match check
		stderr  string // regex match check
		errcode int    // 0
		errmsg  string // regex match check
	}{
		{
			fmt.Sprintf("touch %s/abc", tempdir),
			"",
			"",
			0,
			"",
		},
		{
			fmt.Sprintf("ls %s/abc", tempdir),
			"abc",
			"",
			0,
			"",
		},
		{
			fmt.Sprintf("ls %s/def", tempdir),
			"",
			"def",
			2, // OS error code   2:  No such file or directory
			"",
		},
		{
			"echo $ABC",
			"DEF",
			"",
			0,
			"",
		},
	}

	for _, data := range testData {
		so, se, err := RunCmd(map[string]string{"ABC": "DEF"}, "sh", "-c", data.cmd)
		if data.stdout != "" {
			match, _ := regexp.MatchString(data.stdout, so)
			c.Assert(match, check.Equals, true)
		}
		if data.stderr != "" {
			match, _ := regexp.MatchString(data.stderr, se)
			c.Assert(match, check.Equals, true)
		}
		c.Assert(DetectExitCode(err), check.Equals, data.errcode)
		if data.errcode == 0 {
			c.Assert(err, check.Equals, nil)
			continue
		}
		c.Assert(err, check.Not(check.Equals), nil)
		if data.errmsg != "" {
			match, _ := regexp.MatchString(data.errmsg, err.Error())
			c.Assert(match, check.Equals, true)
		}
	}
}

func (s *cmdSuit) TestRunCmdTimeout(c *check.C) {
	so, se, err := RunCmdTimeout(nil, time.Second*3, "sleep", "5s")
	c.Assert(err, check.Not(check.Equals), nil)
	c.Assert(err.Error(), check.Matches, "signal: terminated")
	c.Assert(so, check.Equals, "")
	c.Assert(se, check.Equals, "")

	so, se, err = RunCmdTimeout(nil, time.Second*3, "sh", "-c", "ping 127.0.0.1")
	c.Assert(err, check.Not(check.Equals), nil)
	c.Assert(err.Error(), check.Matches, "signal: terminated")
	c.Assert(so, check.Not(check.Equals), "")
	c.Assert(se, check.Equals, "")
}

func random(size int) string {
	id := make([]byte, 32)
	io.ReadFull(rand.Reader, id)
	return hex.EncodeToString(id)[:size]
}
