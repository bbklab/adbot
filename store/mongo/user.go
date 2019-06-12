package mongo

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/types"
)

//
// User
//

// AddUser is exported
func (s *MgoStore) AddUser(user *types.User) error {
	return s.insert(cUser, user)
}

// UpdateUser is exported
func (s *MgoStore) UpdateUser(id string, update interface{}) error {
	query := bson.M{"$or": []bson.M{{"id": id}, {"name": id}}}
	return s.update(cUser, query, update)
}

// GetUser is exported
func (s *MgoStore) GetUser(id string) (*types.User, error) {
	var ret *types.User
	query := bson.M{"$or": []bson.M{{"id": id}, {"name": id}}}
	err := s.one(cUser, query, &ret)
	return ret, err
}

// ListUsers is exported
func (s *MgoStore) ListUsers(pager types.Pager) ([]*types.User, error) {
	ret := []*types.User{}
	err := s.all(cUser, nil, pager, &ret, "-created_at")
	return ret, err
}

// CountUsers is exported
func (s *MgoStore) CountUsers() int {
	return s.count(cUser, nil)
}

//
// User Session
//

// UpsertUserSession is exported
func (s *MgoStore) UpsertUserSession(sess *types.UserSession) error {
	query := bson.M{"id": sess.ID}
	return s.upsert(cUserSession, query, sess) // insert or update the whole session
}

// RemoveUserSession is exported
func (s *MgoStore) RemoveUserSession(id string) error {
	query := bson.M{"id": id}
	_, err := s.removeAll(cUserSession, query)
	return err
}

// GetUserSession is exported
func (s *MgoStore) GetUserSession(id string) (*types.UserSession, error) {
	var ret *types.UserSession
	query := bson.M{"id": id}
	err := s.one(cUserSession, query, &ret)
	return ret, err
}

// ListUserSessions is exported
// note: if userID is empty, will list all of db user sessions
func (s *MgoStore) ListUserSessions(userID string) ([]*types.UserSession, error) {
	var filter bson.M
	if userID != "" {
		filter = bson.M{"user_id": userID}
	}
	ret := []*types.UserSession{}
	err := s.all(cUserSession, filter, nil, &ret, "-last_active_at")
	return ret, err
}

// CountUserSessions is exported
// note: if userID is empty, will count all of db user sessions
func (s *MgoStore) CountUserSessions(userID string) int {
	var filter bson.M
	if userID != "" {
		filter = bson.M{"user_id": userID}
	}
	return s.count(cUserSession, filter)
}
