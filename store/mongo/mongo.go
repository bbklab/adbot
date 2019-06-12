package mongo

import (
	"time"

	"github.com/bbklab/adbot/types"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	cUser        = "user"
	cUserSession = "user_session"
	cNode        = "node" // node
	cBlockedNode = "blocked_node"
	cAdbDevice   = "adb_device" // adb device
	cAdbOrder    = "adb_order"  // adb order
	cSettings    = "settings"
	cPing        = "ping"
)

// Setup is exported
func Setup(typ string, cfg *types.MongodbConfig) (*MgoStore, error) {
	var (
		s   = &MgoStore{typ: typ}
		err error
	)

	// parse mongo url & set default dial timeout
	s.dial, err = mgo.ParseURL(cfg.MgoURL)
	if err != nil {
		return nil, err
	}
	if s.dial.Timeout == 0 {
		s.dial.Timeout = time.Second * 5
	}
	if s.dial.Database == "" {
		s.dial.Database = "adbot"
	}

	// dial
	s.sess, err = mgo.DialWithInfo(s.dial)
	if err != nil {
		return nil, err
	}

	// ping verify
	if err = s.Ping(); err != nil {
		return nil, err
	}

	// ensure indexes for all base collections
	if err = s.ensureIndexes(); err != nil {
		return nil, err
	}

	return s, nil
}

// MgoStore is an implemention of store.Store interface
type MgoStore struct {
	typ  string
	dial *mgo.DialInfo // dial configs
	sess *mgo.Session  // first (initial) mgo session, afterwards mgo session should always be Cloned from this one
}

// Type is exported
func (s *MgoStore) Type() string {
	return s.typ
}

// Ping is exported
func (s *MgoStore) Ping() error {
	err := s.sess.Ping()
	if err != nil {
		return err
	}
	return s.upsert(cPing, nil, bson.M{"time": time.Now().Unix()})
}

// ErrNotFound is exported
func (s *MgoStore) ErrNotFound(err error) bool {
	return err == mgo.ErrNotFound
}

//
// shorthands on various frequently used mgo ops
//

// find all of matched objects into result with optional pager parameter
// note: result must be a slice address, otherwise panic
func (s *MgoStore) all(coll string, query interface{}, pager types.Pager, result interface{}, sorts ...string) error {
	return s.exec(func(db *mgo.Database) error {
		q := db.C(coll).Find(query)
		if len(sorts) > 0 {
			q = q.Sort(sorts...)
		}
		if pager != nil {
			if n := pager.Limit(); n > 0 {
				q = q.Limit(n)
			}
			if n := pager.Offset(); n > 0 {
				q = q.Skip(n)
			}
		}
		return q.All(result)
	})
}

// count the total number of matched resulsts
func (s *MgoStore) count(coll string, query interface{}) int {
	var n int
	s.exec(func(db *mgo.Database) error {
		n, _ = db.C(coll).Find(query).Count()
		return nil
	})
	return n
}

// find the first of matched objects into result
func (s *MgoStore) one(coll string, query interface{}, result interface{}) error {
	return s.exec(func(db *mgo.Database) error {
		return db.C(coll).Find(query).One(result)
	})
}

// similar as above, but with select one to given fields
func (s *MgoStore) selectOne(coll string, query, selectQuery interface{}, result interface{}) error {
	return s.exec(func(db *mgo.Database) error {
		return db.C(coll).Find(query).Select(selectQuery).One(result)
	})
}

// similar as above, but with select all to given fields
func (s *MgoStore) selectAll(coll string, query, selectQuery interface{}, result interface{}) error {
	return s.exec(func(db *mgo.Database) error {
		return db.C(coll).Find(query).Select(selectQuery).All(result)
	})
}

// removeAll remove all of matched objects
func (s *MgoStore) removeAll(coll string, query bson.M) (int, error) {
	var n int
	err := s.exec(func(db *mgo.Database) error {
		info, err := db.C(coll).RemoveAll(query)
		if err != nil {
			return err
		}
		if info != nil {
			n = info.Removed
		}
		return nil
	})
	if err != nil {
		return -1, err
	}
	return n, nil
}

func (s *MgoStore) insert(coll string, value interface{}) error {
	return s.exec(func(db *mgo.Database) error {
		return db.C(coll).Insert(value)
	})
}

func (s *MgoStore) update(coll string, query bson.M, update interface{}) error {
	return s.exec(func(db *mgo.Database) error {
		return db.C(coll).Update(query, update)
	})
}

func (s *MgoStore) upsert(coll string, query bson.M, update interface{}) error {
	return s.exec(func(db *mgo.Database) error {
		_, err := db.C(coll).Upsert(query, update)
		return err
	})
}

// exec clone a new session and pick up new session's database to run given execHandler
// close the newly cloned session afterwards
func (s *MgoStore) exec(handler execHandler) error {
	ss := s.sess.Clone()
	defer ss.Close()
	db := ss.DB(s.dbname())
	return handler(db)
}

func (s *MgoStore) dbname() string {
	return s.dial.Database // mostly should be "adbot"
}

type execHandler func(db *mgo.Database) error

//
// Indexes
//
func (s *MgoStore) ensureIndexes() error {
	for col, idxes := range indexes {
		for _, idx := range idxes {
			err := s.sess.DB(s.dbname()).C(col).EnsureIndex(idx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var indexes = map[string][]mgo.Index{
	cUser: {
		{
			Key:    []string{"id"},
			Unique: true,
		},
		{
			Key:    []string{"name"},
			Unique: true,
		},
		{
			Key: []string{"created_at"},
		},
		{
			Key: []string{"last_login_at"},
		},
	},
	cUserSession: {
		{
			Key:    []string{"id"},
			Unique: true,
		},
		{
			Key: []string{"user_id"},
		},
		{
			Key: []string{"last_active_at"},
		},
	},
	cNode: {
		{
			Key:    []string{"id"},
			Unique: true,
		},
		{
			Key: []string{"status"},
		},
		{
			Key: []string{"remote_addr"},
		},
		{
			Key: []string{"cloudsvr_id"},
		},
		{
			Key: []string{"inst_job"},
		},
		{
			Key: []string{"join_at"},
		},
	},
	cBlockedNode: {
		{
			Key:    []string{"id"},
			Unique: true,
		},
		{
			Key: []string{"join_at"},
		},
	},
	cAdbDevice: {
		{
			Key:    []string{"id"},
			Unique: true,
		},
		{
			Key: []string{"node_id"},
		},
		{
			Key: []string{"status"},
		},
		{
			Key:    []string{"alipay.user_id"}, // alipay user id must be uniq
			Unique: true,
		},
	},
	cAdbOrder: {
		{
			Key:    []string{"id"}, // adb order id
			Unique: true,
		},
		{
			Key: []string{"node_id"},
		},
		{
			Key: []string{"device_id"},
		},
		{
			Key:    []string{"out_order_id"}, // out side order id
			Unique: true,
		},
		{
			Key: []string{"qrtype"},
		},
		{
			Key: []string{"fee"},
		},
		{
			Key: []string{"status"},
		},
		{
			Key: []string{"created_at"},
		},
		{
			Key: []string{"paid_at"},
		},
	},
}
