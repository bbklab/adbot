package mongo

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/types"
)

//
// AdbOrder
//

// AddAdbOrder is exported
func (s *MgoStore) AddAdbOrder(order *types.AdbOrder) error {
	return s.insert(cAdbOrder, order)
}

// UpdateAdbOrder is exported
func (s *MgoStore) UpdateAdbOrder(id string, update interface{}) error {
	query := bson.M{"id": id}
	return s.update(cAdbOrder, query, update)
}

// RemoveAdbOrder is exported
func (s *MgoStore) RemoveAdbOrder(id string) error {
	query := bson.M{"id": id}
	_, err := s.removeAll(cAdbOrder, query)
	return err
}

// GetAdbOrder is exported
func (s *MgoStore) GetAdbOrder(id string) (*types.AdbOrder, error) {
	var ret *types.AdbOrder
	query := bson.M{"id": id}
	err := s.one(cAdbOrder, query, &ret)
	return ret, err
}

// ListAdbOrders is exported
func (s *MgoStore) ListAdbOrders(pager types.Pager, filter interface{}) ([]*types.AdbOrder, error) {
	ret := []*types.AdbOrder{}
	err := s.all(cAdbOrder, filter, pager, &ret, "-created_at")
	return ret, err
}

// CountAdbOrders is exported
func (s *MgoStore) CountAdbOrders(filter interface{}) (int, int) {
	ret := []struct {
		ID  string `bson:"id"`
		Fee int    `bson:"fee"`
	}{}
	selectQuery := bson.M{"id": 1, "fee": 1}
	err := s.selectAll(cAdbOrder, filter, selectQuery, &ret)
	if err != nil {
		return 0, 0
	}
	var sum, feesum = len(ret), 0
	for _, o := range ret {
		feesum += o.Fee
	}
	return sum, feesum
}
