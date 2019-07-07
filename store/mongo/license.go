package mongo

// UpsertLicense is exported
func (s *MgoStore) UpsertLicense(text string) error {
	data := map[string]string{"license": text}
	return s.upsert(cLicense, nil, data)
}

// RemoveLicense is exported
func (s *MgoStore) RemoveLicense() error {
	_, err := s.removeAll(cLicense, nil)
	return err
}

// GetLicense is exported
func (s *MgoStore) GetLicense() (string, error) {
	ret := make(map[string]string)
	err := s.one(cLicense, nil, &ret)
	return ret["license"], err
}
