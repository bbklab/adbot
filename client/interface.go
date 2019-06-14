package client

import (
	"io"
	"time"

	maxminddb "github.com/oschwald/maxminddb-golang"

	"github.com/bbklab/adbot/debug"
	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/types"
)

// Client offers a common interface to access adbot api services and
// hides all the http request based operations behind it.
type Client interface {
	SetHeader(name, value string) // with extra header
	Reset() error

	Peer() string     // print who we're talking to
	PeerAddr() string // similar as above, but only unix or tcp address
	QueryLeader() (string, bool)

	AnyUsers() (bool, error)
	ListUsers() ([]*types.User, error)
	UserProfile() (*types.User, error)
	UserSessions() ([]*types.UserSessionWrapper, error)
	CreateUser(user *types.User) (*types.User, error)
	ChangeUserPassword(req *types.ReqChangePassword) error

	ListNodes(lbsFilter label.Labels, online *bool, cldsvr string) ([]*types.NodeWrapper, error)
	InspectNode(id string) (*types.NodeWrapper, error)
	WatchNodeStats(id string) (io.ReadCloser, error)
	RunNodeCmd(id, cmd string) (io.ReadCloser, error)
	OpenNodeTerminal(id string, input io.Reader, output io.Writer) error
	WatchNodeEvents(id string) (io.ReadCloser, error)
	CloseNode(id string) error

	UpsertNodeLabels(id string, lbs label.Labels) (label.Labels, error)
	RemoveNodeLabels(id string, all bool, keys []string) (label.Labels, error)

	CurrentGeoMetadata() (map[string]maxminddb.Metadata, error)
	UpdateGeoData() (prev, current map[string]maxminddb.Metadata, cost time.Duration, err error)

	ListAdbNodes() ([]*types.AdbNode, error)
	InspectAdbNode(id string) (*types.AdbNode, error)

	ListAdbDevices() ([]*types.AdbDeviceWrapper, error)
	InspectAdbDevice(id string) (*types.AdbDeviceWrapper, error)
	SetAdbDeviceBill(id string, val int) error
	SetAdbDeviceAmount(id string, val int) error
	SetAdbDeviceWeight(id string, val int) error
	BindAdbDeviceAlipay(id string, alipay *types.AlipayAccount) error
	RevokeAdbDeviceAlipay(id string) error

	ReportAdbEvent(ev *adbot.AdbEvent) error // public, used by adb node to report adb events
	WatchAdbEvents() (io.ReadCloser, error)

	GetSettings() (*types.Settings, error)
	UpdateSettings(req *types.UpdateSettingsReq) (*types.Settings, error)
	SetGlobalAttrs(attrs label.Labels) (label.Labels, error)
	RemoveGlobalAttrs(all bool, keys []string) (label.Labels, error)
	ResetSettings() error
	GenAdvertiseQrCode() ([]byte, error)

	Ping() error
	Version() (*types.Version, error)
	Info() (*types.SummaryInfo, error)
	DebugDump(name string) ([]byte, *debug.Info, *types.MasterConfig, map[string]interface{}, error)
	PProfData(name string, seconds int) (io.ReadCloser, error)

	Panic() error

	Login(req *types.ReqLogin) (string, error) // return the access token
	Logout() error
}
