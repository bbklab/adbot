package mongo

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/types"
)

// AddNode is exported
func (s *MgoStore) AddNode(node *types.Node) error {
	return s.insert(cNode, node)
}

// UpdateNode is exported
func (s *MgoStore) UpdateNode(id string, update interface{}) error {
	query := bson.M{"id": id}
	return s.update(cNode, query, update)
}

// RemoveNode is exported
func (s *MgoStore) RemoveNode(id string) error {
	query := bson.M{"id": id}
	_, err := s.removeAll(cNode, query)
	return err
}

// GetNode is exported
func (s *MgoStore) GetNode(id string) (*types.Node, error) {
	var ret *types.Node
	query := bson.M{"id": id}
	err := s.one(cNode, query, &ret)
	return ret, err
}

// ListNodes is exported
func (s *MgoStore) ListNodes(pager types.Pager) ([]*types.Node, error) {
	ret := []*types.Node{}
	err := s.all(cNode, nil, pager, &ret, "-join_at")
	return ret, err
}

// CountNodes is exported
func (s *MgoStore) CountNodes() int {
	return s.count(cNode, nil)
}
