package mongo

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/types"
)

//
// Adb Device
//

// AddAdbDevice is exported
func (s *MgoStore) AddAdbDevice(dvc *types.AdbDevice) error {
	return s.insert(cAdbDevice, dvc)
}

// UpdateAdbDevice is exported
func (s *MgoStore) UpdateAdbDevice(id string, update interface{}) error {
	query := bson.M{"id": id}
	return s.update(cAdbDevice, query, update)
}

// RemoveAdbDevice is exported
func (s *MgoStore) RemoveAdbDevice(id string) error {
	query := bson.M{"id": id}
	_, err := s.removeAll(cAdbDevice, query)
	return err
}

// GetAdbDevice is exported
func (s *MgoStore) GetAdbDevice(id string) (*types.AdbDevice, error) {
	var ret *types.AdbDevice
	query := bson.M{"id": id}
	err := s.one(cAdbDevice, query, &ret)
	return ret, err
}

// ListAdbDevices is exported
func (s *MgoStore) ListAdbDevices(pager types.Pager, filter interface{}) ([]*types.AdbDevice, error) {
	ret := []*types.AdbDevice{}
	err := s.all(cAdbDevice, filter, pager, &ret, "id")
	return ret, err
}

// CountAdbDevices is exported
func (s *MgoStore) CountAdbDevices(filter interface{}) int {
	return s.count(cAdbDevice, filter)
}
