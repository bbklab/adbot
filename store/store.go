package store

import (
	"errors"

	"github.com/bbklab/adbot/store/mongo"
	"github.com/bbklab/adbot/types"
)

var (
	store Store
)

var (
	// ErrNotImplemented is exported
	ErrNotImplemented = errors.New("db store not implemented yet")
	// ErrNotSupportedStore is exported
	ErrNotSupportedStore = errors.New("unsupported db store")
)

// Store represents the backend storage interface
type Store interface {
	AddUser(user *types.User) error
	UpdateUser(id string, update interface{}) error
	GetUser(id string) (*types.User, error)
	ListUsers(pager types.Pager) ([]*types.User, error)
	CountUsers() int

	UpsertUserSession(sess *types.UserSession) error
	RemoveUserSession(id string) error
	GetUserSession(id string) (*types.UserSession, error)
	ListUserSessions(userID string) ([]*types.UserSession, error)
	CountUserSessions(userID string) int

	AddNode(node *types.Node) error
	UpdateNode(id string, update interface{}) error
	RemoveNode(id string) error
	GetNode(id string) (*types.Node, error)
	ListNodes(pager types.Pager) ([]*types.Node, error)
	CountNodes() int

	// adb node
	AddAdbDevice(dvc *types.AdbDevice) error
	UpdateAdbDevice(id string, update interface{}) error
	RemoveAdbDevice(id string) error
	GetAdbDevice(id string) (*types.AdbDevice, error)
	ListAdbDevices(pager types.Pager, filter interface{}) ([]*types.AdbDevice, error)
	CountAdbDevices(filter interface{}) int

	// adb order
	AddAdbOrder(order *types.AdbOrder) error
	UpdateAdbOrder(id string, update interface{}) error
	RemoveAdbOrder(id string) error
	GetAdbOrder(id string) (*types.AdbOrder, error)
	ListAdbOrders(pager types.Pager, filter interface{}) ([]*types.AdbOrder, error)
	CountAdbOrders(filter interface{}) (int, int) // count orders, fees

	UpsertSettings(update interface{}) error
	GetSettings() (*types.Settings, error)

	ErrNotFound(error) bool
	Type() string
	Ping() error
}

// Setup is exported
func Setup(cfg *types.StoreConfig) error {
	var err error

	switch typ := cfg.Type; typ {
	case "memory":
		err = ErrNotImplemented
	case "mongo", "mongodb":
		store, err = mongo.Setup(typ, cfg.MongodbConfig)
	default:
		err = ErrNotSupportedStore
	}

	return err
}

// DB pick up the initialized db store
func DB() Store {
	if store == nil {
		panic("db store not initilized yet")
	}
	return store
}
