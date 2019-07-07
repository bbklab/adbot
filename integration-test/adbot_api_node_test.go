package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/pkg/label"
)

func (s *ApiSuite) TestNodeInspect(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node.Status, check.Equals, "online")
	c.Assert(node.ErrMsg, check.Equals, "")
	c.Assert(node.SysInfo, check.NotNil)
	c.Assert(node.SysInfo.Hostname, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.OS, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.Kernel, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.Uptime, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.UnixTime, check.Not(check.Equals), 0)
	c.Assert(node.SysInfo.CPU.Processor, check.Not(check.Equals), 0)
	c.Assert(node.SysInfo.Memory.Total, check.Not(check.Equals), 0)

	node, err = s.client.InspectNode("node.id.that.is.not.exists")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*not found.*")

	costPrintln("TestNodeInspect() passed", startAt)
}

func (s *ApiSuite) TestNodeList(c *check.C) {
	startAt := time.Now()

	node := s.getAssertOnlineNode(c)

	c.Assert(node.Status, check.Equals, "online")
	c.Assert(node.ErrMsg, check.Equals, "")
	c.Assert(node.SysInfo, check.NotNil)
	c.Assert(node.SysInfo.Hostname, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.OS, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.Kernel, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.Uptime, check.Not(check.Equals), "")
	c.Assert(node.SysInfo.UnixTime, check.Not(check.Equals), 0)
	c.Assert(node.SysInfo.CPU.Processor, check.Not(check.Equals), 0)
	c.Assert(node.SysInfo.Memory.Total, check.Not(check.Equals), 0)

	costPrintln("TestNodeList() passed", startAt)
}

func (s *ApiSuite) TestNodeCreate(c *check.C) {
	// TODO
}

func (s *ApiSuite) TestNodeWatchEvent(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)

	stream, err := s.client.WatchNodeEvents(node.ID)
	c.Assert(err, check.IsNil)
	defer stream.Close()

	_, err = s.client.WatchNodeEvents("node.id.that.is.not.exists")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*No such node.*online.*")

	costPrintln("TestNodeWatchEvent() passed", startAt)
}

func (s *ApiSuite) TestNodeExec(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node, check.NotNil)

	datas := map[string]string{
		"echo hello node":                  "hello node\n",
		`echo "abcdef" | grep -E -o "cde"`: "cde\n",
		"kkkkk": "/bin/sh: kkkkk: not found\nexit status 127\r\n",
	}
	for cmd, output := range datas {
		stream, err := s.client.RunNodeCmd(node.ID, cmd)
		c.Assert(err, check.IsNil)
		defer stream.Close()

		bs, err := ioutil.ReadAll(stream)
		c.Assert(err, check.IsNil)
		c.Assert(string(bs), check.Equals, output)
	}

	_, err = s.client.RunNodeCmd("node.id.that.is.not.exists", "hostname")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*No such node.*online.*")

	costPrintln("TestNodeExec() passed", startAt)
}

func (s *ApiSuite) TestNodeRemove(c *check.C) {
	// TODO
}

/*
func (s *ApiSuite) TestNodeClose(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)

	// save previous nb of fds & goroutines
	_, general, _, _, err := s.client.DebugDump("general")
	c.Assert(err, check.IsNil)
	c.Assert(general, check.NotNil)
	c.Assert(general.NumGoroutines, check.Not(check.Equals), 0)
	c.Assert(general.NumFds, check.Not(check.Equals), 0)
	// numGosPrev := general.NumGoroutines
	// numGosPrev, numFdsPrev := general.NumGoroutines, general.NumFds

	err = s.client.CloseNode(node.ID)
	c.Assert(err, check.IsNil)

	time.Sleep(time.Second * 10)

	// node should rejoin in quickly
	node, err = s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node.Status, check.Equals, "online")
	c.Assert(node.ErrMsg, check.Equals, "")
	c.Assert(node.LastActiveAt.After(time.Now().Add(-time.Second*5)), check.Equals, true) // join in 5s

	// FIXME this always be not precisely exactly
	// ensure no fds / goroutines leaks
	_, general, _, _, err = s.client.DebugDump("general")
	c.Assert(err, check.IsNil)
	c.Assert(general, check.NotNil)
	c.Assert(general.NumGoroutines, check.Not(check.Equals), 0)
	c.Assert(general.NumFds, check.Not(check.Equals), 0)
	// numGosCurr := general.NumGoroutines
	// numGosCurr, numFdsCurr := general.NumGoroutines, general.NumFds
	// c.Assert(numGosPrev >= numGosCurr, check.Equals, true)
	// c.Assert(numFdsPrev >= numFdsCurr, check.Equals, true)

	costPrintln("TestNodeClose() passed", startAt)
}
*/

func (s *ApiSuite) TestNodeLabelSet(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)

	lbs1 := label.New(map[string]string{"prod": "true", "os": "centos"})
	lbs2 := label.New(map[string]string{"ssd": "true", "os": "ubuntu"})
	lbs3 := label.New(map[string]string{"zone": "bj", "admin": "xx@yy.zz"})

	// remove all firstly
	current, err := s.client.RemoveNodeLabels(node.ID, true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	current, err = s.client.UpsertNodeLabels(node.ID, lbs1)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(lbs1), check.Equals, true)

	current, err = s.client.UpsertNodeLabels(node.ID, lbs2)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(lbs1.Merge(lbs2)), check.Equals, true)

	current, err = s.client.UpsertNodeLabels(node.ID, lbs3)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(lbs1.Merge(lbs2).Merge(lbs3)), check.Equals, true)

	node, err = s.client.InspectNode(node.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node.Labels.EqualsTo(lbs1.Merge(lbs2).Merge(lbs3)), check.Equals, true)

	// remove all finally
	current, err = s.client.RemoveNodeLabels(node.ID, false, node.Labels.Keys())
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	_, err = s.client.UpsertNodeLabels("node.id.that.is.not.exists", lbs3)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*not found.*")

	costPrintln("TestNodeLabelSet() passed", startAt)
}

func (s *ApiSuite) TestNodeLabelRemove(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node, check.NotNil)

	lbs1 := label.New(map[string]string{"prod": "true", "os": "centos"})
	lbs2 := label.New(map[string]string{"ssd": "true", "os": "ubuntu"})

	// remove all firstly
	current, err := s.client.RemoveNodeLabels(node.ID, true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	current, err = s.client.UpsertNodeLabels(node.ID, lbs1)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(lbs1), check.Equals, true)

	current, err = s.client.UpsertNodeLabels(node.ID, lbs2)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(lbs1.Merge(lbs2)), check.Equals, true)

	current, err = s.client.RemoveNodeLabels(node.ID, false, lbs2.Keys())
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(map[string]string{"prod": "true"})), check.Equals, true)

	node, err = s.client.InspectNode(node.ID)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(map[string]string{"prod": "true"})), check.Equals, true)

	current, err = s.client.RemoveNodeLabels(node.ID, true, nil) // remove all labels
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	node, err = s.client.InspectNode(node.ID)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	// remove all finally
	current, err = s.client.RemoveNodeLabels(node.ID, true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	_, err = s.client.RemoveNodeLabels("node.id.that.is.not.exists", false, []string{"whatever"})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*not found.*")

	costPrintln("TestNodeLabelRemove() passed", startAt)
}

func (s *ApiSuite) TestNodeLabelSetConcurrency(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)

	// remove all firstly
	current, err := s.client.RemoveNodeLabels(node.ID, true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	var (
		count       = 500
		concurrency = 10
	)

	var (
		wg          sync.WaitGroup
		tokens      = make(chan struct{}, concurrency)
		lbsExpected = label.New(nil)
	)

	wg.Add(count)
	for idx := 1; idx <= count; idx++ {
		tokens <- struct{}{}

		lbs := label.New(map[string]string{fmt.Sprintf("lbs_%d", idx): strconv.Itoa(idx)})
		lbsExpected = lbsExpected.Merge(lbs)

		go func(lbs label.Labels) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			_, err := s.client.UpsertNodeLabels(node.ID, lbs)
			c.Assert(err, check.IsNil)
		}(lbs)
	}
	wg.Wait()

	// ensure all labels are set correctly
	node, err = s.client.InspectNode(node.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node.Labels.Len(), check.Equals, count)
	c.Assert(node.Labels.EqualsTo(lbsExpected), check.Equals, true)

	// remove all finally
	current, err = s.client.RemoveNodeLabels(node.ID, false, node.Labels.Keys())
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	costPrintln("TestNodeLabelSetConcurrency() passed", startAt)
}

func (s *ApiSuite) TestNodeLabelRemoveConcurrency(c *check.C) {
	startAt := time.Now()

	assertnode := s.getAssertOnlineNode(c)

	node, err := s.client.InspectNode(assertnode.ID)
	c.Assert(err, check.IsNil)

	// remove all firstly
	current, err := s.client.RemoveNodeLabels(node.ID, true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(current.EqualsTo(label.New(nil)), check.Equals, true)

	var (
		count       = 500
		concurrency = 10
	)
	var lbsAll = label.New(nil)
	for idx := 1; idx <= count; idx++ {
		lbs := label.New(map[string]string{fmt.Sprintf("lbs_%d", idx): strconv.Itoa(idx)})
		lbsAll = lbsAll.Merge(lbs)
	}
	current, err = s.client.UpsertNodeLabels(node.ID, lbsAll)
	c.Assert(err, check.IsNil)
	c.Assert(current.Len(), check.Equals, count)
	c.Assert(current.EqualsTo(lbsAll), check.Equals, true)

	// remove node label concurrency
	var (
		wg     sync.WaitGroup
		tokens = make(chan struct{}, concurrency)
	)

	wg.Add(count)
	for idx := 1; idx <= count; idx++ {
		tokens <- struct{}{}

		key := []string{fmt.Sprintf("lbs_%d", idx)}

		go func(key []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			_, err := s.client.RemoveNodeLabels(node.ID, false, key)
			c.Assert(err, check.IsNil)
		}(key)
	}
	wg.Wait()

	// ensure all labels are removed correctly
	node, err = s.client.InspectNode(node.ID)
	c.Assert(err, check.IsNil)
	c.Assert(node.Labels.Len(), check.Equals, 0)
	c.Assert(node.Labels.EqualsTo(label.New(nil)), check.Equals, true)

	costPrintln("TestNodeLabelRemoveConcurrency() passed", startAt)
}

func (s *ApiSuite) doLaunchSshContainer() (string, error) {
	return "", nil
}

func (s *ApiSuite) doDestroySshContainer(id string) error {
	return nil
}
