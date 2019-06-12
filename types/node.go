package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bbklab/adbot/pkg/geoip"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/ssh"
	"github.com/bbklab/adbot/pkg/validator"
)

// nolint
var (
	NodeWaittingFirstJoin = "waitting for node first join"
	NodeInstallJobAborted = "node installation job aborted while restart"
)

// nolint
var (
	NodeLabelKeyZone      = "_node_zone"      // preserved node label: node zone attribute
	NodeLabelKeyType      = "_node_type"      // preserved node label: node type attribute
	NodeLabelKeyProtected = "_node_protected" // preserved node label: node under protected attribute
)

// NodeWrapper wrap db node with related cluster name
type NodeWrapper struct {
	*Node
	Cluster string `json:"cluster"` // cluster name belongs to
}

// nolint
var (
	NodeStatusDeploying = "deploying"
	NodeStatusOnline    = "online"
	NodeStatusOffline   = "offline"
	NodeStatusFlagging  = "flagging"
	NodeStatusDeleting  = "deleting"
)

// Node is a db node
type Node struct {
	ID           string         `json:"id" bson:"id"`
	Status       string         `json:"status" bson:"status"`                 // online, offline, flagging, deploying
	Version      string         `json:"version" bson:"version"`               // version (timer updated)
	ErrMsg       string         `json:"error" bson:"error"`                   // the error message abount offline (not empty if offline)
	RemoteAddr   string         `json:"remote_addr" bson:"remote_addr"`       // the remote addr of online (conected) node (only updated by node join callback)
	GeoInfo      *geoip.GeoInfo `json:"geoinfo" bson:"geoinfo"`               // the region info detected via GeoIP of remote ip (lang=en-US)
	GeoInfoZh    *geoip.GeoInfo `json:"geoinfo_zh" bson:"geoinfo_zh"`         // same as above, but for lang=zh-CN
	SysInfo      *SysInfo       `json:"sysinfo" bson:"sysinfo"`               // the collected node sysinfo (timer updated)
	SSHConfig    *ssh.Config    `json:"ssh_config" bson:"ssh_config"`         // ssh configs of this node (used for initilization)
	Labels       label.Labels   `json:"labels" bson:"labels"`                 // node labels
	JoinAt       time.Time      `json:"join_at" bson:"join_at"`               // the join time of online(connected) node
	LastActiveAt time.Time      `json:"last_active_at" bson:"last_active_at"` // the latest active node heartbeat(ping) received
	Latency      time.Duration  `json:"latency" bson:"latency"`               // network latency while master ping node
	InstJob      string         `json:"inst_job" bson:"inst_job"`             // related node install async job id
}

// Hidden set the ssh config sensitive fields as invisible
func (a *Node) Hidden() {
	if a.SSHConfig != nil {
		a.SSHConfig.User = SensitiveHolder
		a.SSHConfig.Password = SensitiveHolder
		a.SSHConfig.PrivKey = SensitiveHolder
	}
}

// Name return a human readable name for node
func (a *Node) Name() string {
	if a.SysInfo != nil {
		return fmt.Sprintf("%s[%s,%s]", a.ID, a.SysInfo.Hostname, a.RemoteIP())
	}
	return fmt.Sprintf("%s[%s]", a.ID, a.RemoteIP())
}

// AttrZone return current node zone
func (a *Node) AttrZone() string {
	return a.Labels.Get(NodeLabelKeyZone)
}

// AttrType return current node type
func (a *Node) AttrType() string {
	return a.Labels.Get(NodeLabelKeyType)
}

// AttrProtected return current node protected
func (a *Node) AttrProtected() string {
	return a.Labels.Get(NodeLabelKeyProtected)
}

// IsProtected check if a node is under protected
func (a *Node) IsProtected() bool {
	return strings.ToLower(a.AttrProtected()) == "true"
}

// Valid verify the node fields
func (a *Node) Valid() error {
	if err := validator.String(a.ID, 1, 32, nil); err != nil {
		return fmt.Errorf("agent ID %v", err)
	}
	if a.SysInfo == nil {
		return errors.New("agent sysinfo required")
	}
	return nil
}

// RemoteIP return current node's remote ip address
func (a *Node) RemoteIP() string {
	fields := strings.SplitN(a.RemoteAddr, ":", 2)
	if len(fields) != 2 {
		return ""
	}
	return fields[0]
}

// IsOnGoing is exported
func (a *Node) IsOnGoing() (string, bool) {
	return a.Status, a.Status == NodeStatusDeploying || a.Status == NodeStatusDeleting
}

// SysInfo represents system informations collected by node
type SysInfo struct {
	Hostname     string                 `json:"hostname" bson:"hostname"`
	OS           string                 `json:"os" bson:"os"`
	Kernel       string                 `json:"kernel" bson:"kernel"`
	Uptime       string                 `json:"uptime" bson:"uptime"`         // legacy, for compatible
	UptimeInt    int64                  `json:"uptime_int" bson:"uptime_int"` // by seconds
	UnixTime     int64                  `json:"unixtime" bson:"unixtime"`
	LoadAvgs     LoadAvgInfo            `json:"loadavgs" bson:"loadavgs"`
	CPU          CPUInfo                `json:"cpu" bson:"cpu"`
	Memory       MemoryInfo             `json:"memory" bson:"memory"`
	Swap         SwapInfo               `json:"swap" bson:"swap"`
	User         UserInfo               `json:"user" bson:"user"`         // current user on node (run as)
	IPs          map[string][]string    `json:"ips" bson:"ips"`           // inet name -> ips
	Disks        map[string]*DiskInfo   `json:"disks" bson:"disks"`       // dev name -> disk info
	DisksIO      map[string]*DiskIOInfo `json:"disksio" bson:"disksio"`   // dev name -> disk io
	Traffics     map[string]*NetTraffic `json:"traffics" bson:"traffics"` // inet name -> inet traffics
	Docker       DockerInfo             `json:"docker" bson:"docker"`
	BBREnabled   bool                   `json:"bbr_enabled" bson:"bbr_enabled"`   // if tcp bbr enabled (require kernel>=4.9)
	WithMaster   bool                   `json:"with_master" bson:"with_master"`   // is deployed on the same host with master
	Manufacturer string                 `json:"manufacturer" bson:"manufacturer"` // hardware manufacturer
}

// LoadAvgInfo is exported
type LoadAvgInfo struct {
	One     float64 `json:"one" bson:"one"`
	Five    float64 `json:"five" bson:"five"`
	Fifteen float64 `json:"fifteen" bson:"fifteen"`
}

// SwapInfo is exported
type SwapInfo struct {
	Total int64 `json:"total" bson:"total"`
	Used  int64 `json:"used" bson:"used"`
	Free  int64 `json:"free" bson:"free"`
}

// CPUInfo is exported
type CPUInfo struct {
	Processor int64  `json:"processor" bson:"processor"`
	Physical  int64  `json:"physical" bson:"physical"`
	Used      uint64 `json:"used" bson:"used"` // 1 - idle
}

// MemoryInfo is exported
type MemoryInfo struct {
	Total  int64 `json:"total" bson:"total"`
	Used   int64 `json:"used" bson:"used"`
	Cached int64 `json:"cached" bson:"cached"`
}

// UserInfo is exported
type UserInfo struct {
	UID  string `json:"uid" bson:"uid"`
	GID  string `json:"gid" bson:"gid"`
	Name string `json:"name" bson:"name"`
	Sudo bool   `json:"sudo" bson:"sudo"` // if un-privileged user could sudo, we're expecting /etc/sudoers options like: `user ALL=(ALL) NOPASSWD: ALL`
}

// HasPrivileges is exported
func (u UserInfo) HasPrivileges() bool {
	return u.UID == "0" || u.Sudo
}

// DiskInfo is exported
type DiskInfo struct {
	DevName string `json:"dev_name" bson:"dev_name"`
	MountAt string `json:"mount_at" bson:"mount_at"`
	Total   uint64 `json:"total" bson:"total"`
	Used    uint64 `json:"used" bson:"used"`
	Free    uint64 `json:"free" bson:"free"`
	Inode   uint64 `json:"inode" bson:"inode"`
	Ifree   uint64 `json:"ifree" bson:"ifree"`
}

// DiskIOInfo is exported
type DiskIOInfo struct {
	DevName    string `json:"dev_name" bson:"dev_name"`
	ReadBytes  uint64 `json:"read_bytes" bson:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes" bson:"write_bytes"`
}

// NetTraffic is exported
type NetTraffic struct {
	Name      string    `json:"name" bson:"name"`
	Mac       string    `json:"mac" bson:"mac"`
	RxBytes   uint64    `json:"rx_bytes" bson:"rx_bytes"`     // receive bytes
	TxBytes   uint64    `json:"tx_bytes" bson:"tx_bytes"`     // send bytes
	RxPackets uint64    `json:"rx_packets" bson:"rx_packets"` // receive packets
	TxPackets uint64    `json:"tx_packets" bson:"tx_packets"` // send packets
	RxRate    uint64    `json:"rx_rate" bson:"rx_rate"`       // by Bytes/s
	TxRate    uint64    `json:"tx_rate" bson:"tx_rate"`       // by Bytes/s
	Time      time.Time `json:"time" bson:"time"`             // collect time
}

// DockerInfo is exported
type DockerInfo struct {
	Version              string            `json:"version" bson:"version"`
	NumImages            int64             `json:"num_images" bson:"num_images"`
	NumContainers        int64             `json:"num_containers" bson:"num_containers"`
	NumRunningContainers int64             `json:"num_running_containers" bson:"num_running_containers"`
	Driver               string            `json:"driver" bson:"driver"` // devicemapper, overlay
	DriverStatus         map[string]string `json:"driver_status" bson:"driver_status"`
}

// NodeCmd is exported
type NodeCmd struct {
	Command string `json:"command"`
}

// NodeCPUSorter is exported
//
type NodeCPUSorter []*Node

func (s NodeCPUSorter) Len() int      { return len(s) }
func (s NodeCPUSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s NodeCPUSorter) Less(i, j int) bool {
	return s[i].SysInfo.CPU.Processor >= s[j].SysInfo.CPU.Processor
}
