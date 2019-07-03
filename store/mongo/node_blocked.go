package mongo

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/types"
)

// AddBlockedNode is exported
func (s *MgoStore) AddBlockedNode(node *types.Node) error {
	return s.insert(cBlockedNode, node)
}

// RemoveBlockedNode is exported
func (s *MgoStore) RemoveBlockedNode(id string) error {
	query := bson.M{"id": id}
	_, err := s.removeAll(cBlockedNode, query)
	return err
}

// GetBlockedNode is exported
func (s *MgoStore) GetBlockedNode(id string) (*types.Node, error) {
	var ret *types.Node
	query := bson.M{"id": id}
	err := s.one(cBlockedNode, query, &ret)
	return ret, err
}

// ListBlockedNodes is exported
func (s *MgoStore) ListBlockedNodes(pager types.Pager) ([]*types.Node, error) {
	ret := []*types.Node{}
	err := s.all(cBlockedNode, nil, pager, &ret, "-join_at")
	return ret, err
}

// CountBlockedNodes is exported
func (s *MgoStore) CountBlockedNodes() int {
	return s.count(cBlockedNode, nil)
}
