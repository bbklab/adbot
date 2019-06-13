package main

import (
	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/types"
)

func (s *ApiSuite) getAssertOnlineNode(c *check.C) *types.Node {
	nodes, err := s.client.ListNodes(nil, ptype.Bool(true), "")
	c.Assert(err, check.IsNil)
	c.Assert(len(nodes) > 0, check.Equals, true)

	var ret *types.Node
	for _, node := range nodes {
		ret = node.Node
	}
	c.Assert(ret, check.NotNil)
	return ret
}

func (s *ApiSuite) getOnlineNodeIDs(c *check.C) []string {
	nodes, err := s.client.ListNodes(nil, ptype.Bool(true), "")
	c.Assert(err, check.IsNil)
	c.Assert(len(nodes) > 0, check.Equals, true)

	ids := make([]string, 0, 0)
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	return ids
}
