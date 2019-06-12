package mongo

import "github.com/bbklab/adbot/types"

// UpsertSettings is exported
func (s *MgoStore) UpsertSettings(update interface{}) error {
	return s.upsert(cSettings, nil, update)
}

// GetSettings is exported
func (s *MgoStore) GetSettings() (*types.Settings, error) {
	var ret *types.Settings
	err := s.one(cSettings, nil, &ret)
	return ret, err
}
