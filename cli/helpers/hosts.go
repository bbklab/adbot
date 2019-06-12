package helpers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/bbklab/adbot/pkg/flock"
	"github.com/bbklab/adbot/types"
)

var (
	defaultConfigs = &LocalConfigs{
		Current: "default",
		Hosts: map[string]*AdbotHost{
			"default": {
				Name: "default",
				Addr: "unix:///var/run/adbot/adbot.sock",
			},
		},
	}
)

// LocalConfigs is exported
type LocalConfigs struct {
	Current string                `json:"current"` // current adbot daemon host
	Hosts   map[string]*AdbotHost `json:"hosts"`
}

func (cfgs *LocalConfigs) current() *AdbotHost {
	if len(cfgs.Hosts) == 0 {
		return nil
	}
	return cfgs.Hosts[cfgs.Current]
}

func (cfgs *LocalConfigs) hideSensitive() {
	for _, host := range cfgs.Hosts {
		host.hideSensitive()
	}
}

func (cfgs *LocalConfigs) uncoverSensitive() {
	for _, host := range cfgs.Hosts {
		host.uncoverSensitive()
	}
}

// AdbotHostWrapper is exported
type AdbotHostWrapper struct {
	*AdbotHost
	Current bool `json:"current"`
}

// AdbotHost is exported
type AdbotHost struct {
	Name     string `json:"name"`
	Addr     string `json:"addr"` // eg: unix:///var/run/adbot/adbot.sock  http://ip:port
	User     string `json:"user"`
	Password string `json:"passwrod"`
	Token    string `json:"token"` // optional
}

// Valid is exported
func (h *AdbotHost) Valid() error {
	if h.Name == "" {
		return errors.New("adbot host name required")
	}
	if h.Addr == "" {
		return errors.New("adbot host address required")
	}
	if h.User == "" {
		return errors.New("adbot host username required")
	}
	if h.Password == "" {
		return errors.New("adbot host password required")
	}
	return nil
}

// ReqLogin is exported
func (h *AdbotHost) ReqLogin() *types.ReqLogin {
	return &types.ReqLogin{
		UserName: h.User,
		Password: types.Password(h.Password),
	}
}

func (h *AdbotHost) hideSensitive() {
	h.User = base64.StdEncoding.EncodeToString([]byte(h.User))
	h.Password = base64.StdEncoding.EncodeToString([]byte(h.Password))
}

func (h *AdbotHost) uncoverSensitive() {
	bs, err := base64.StdEncoding.DecodeString(h.User)
	if err == nil {
		h.User = string(bs)
	}

	bs, err = base64.StdEncoding.DecodeString(h.Password)
	if err == nil {
		h.Password = string(bs)
	}
}

// CurrentAdbotHost is exported
func CurrentAdbotHost() (*AdbotHost, error) {
	cfgs, err := LoadLocalConfigs()
	if err != nil {
		return nil, err
	}

	curr := cfgs.current()
	if curr == nil {
		return nil, errors.New("current adbot hosts not avaliable")
	}

	return curr, nil
}

// ListAdbotHosts is exported
func ListAdbotHosts() ([]*AdbotHostWrapper, error) {
	cfgs, err := LoadLocalConfigs()
	if err != nil {
		return nil, err
	}

	var ret = make([]*AdbotHostWrapper, 0, len(cfgs.Hosts))
	for _, host := range cfgs.Hosts {
		ret = append(ret, &AdbotHostWrapper{
			AdbotHost: host,
			Current:   host.Name == cfgs.Current,
		})
	}

	sort.Sort(AdbotHostWrapperSorter(ret))
	return ret, nil
}

// LoadLocalConfigs safely load local configs
func LoadLocalConfigs() (*LocalConfigs, error) {
	l, err := getConfigFileLock()
	if err != nil {
		return nil, err
	}
	defer l.Close()

	err = l.Lock()
	if err != nil {
		return nil, err
	}
	defer l.Unlock()

	cfgs, err := loadConfigs()
	if err != nil {
		return nil, err
	}

	return cfgs, nil
}

// AddAdbotHost is exported
func AddAdbotHost(host *AdbotHost) error {
	l, err := getConfigFileLock()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Lock()
	if err != nil {
		return err
	}
	defer l.Unlock()

	cfgs, err := loadConfigs()
	if err != nil {
		return err
	}

	if _, ok := cfgs.Hosts[host.Name]; ok {
		return errors.New("duplicated name on adbot host")
	}

	cfgs.Hosts[host.Name] = host
	return saveConfigs(cfgs)
}

// SetAdbotHostAuth is exported
func SetAdbotHostAuth(name, username, password, token string) error {
	l, err := getConfigFileLock()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Lock()
	if err != nil {
		return err
	}
	defer l.Unlock()

	cfgs, err := loadConfigs()
	if err != nil {
		return err
	}

	host, ok := cfgs.Hosts[name]
	if !ok {
		return fmt.Errorf("adbot host %s not exists", name)
	}

	host.User = username
	host.Password = password
	host.Token = token
	return saveConfigs(cfgs)
}

// RemoveAdbotHost is exported
func RemoveAdbotHost(name string) error {
	l, err := getConfigFileLock()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Lock()
	if err != nil {
		return err
	}
	defer l.Unlock()

	cfgs, err := loadConfigs()
	if err != nil {
		return err
	}

	if _, ok := cfgs.Hosts[name]; !ok {
		return nil
	}

	delete(cfgs.Hosts, name)
	if cfgs.Current == name {
		cfgs.Current = ""
	}
	return saveConfigs(cfgs)
}

// SwitchAdbotHost is exported
func SwitchAdbotHost(name string) (*AdbotHost, error) {
	l, err := getConfigFileLock()
	if err != nil {
		return nil, err
	}
	defer l.Close()

	err = l.Lock()
	if err != nil {
		return nil, err
	}
	defer l.Unlock()

	cfgs, err := loadConfigs()
	if err != nil {
		return nil, err
	}

	host, ok := cfgs.Hosts[name]
	if !ok {
		return nil, fmt.Errorf("adbot host %s not exists", name)
	}

	cfgs.Current = host.Name
	return host, saveConfigs(cfgs)
}

// ResetAdbotHost is exported
func ResetAdbotHost() error {
	l, err := getConfigFileLock()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Lock()
	if err != nil {
		return err
	}
	defer l.Unlock()

	return saveConfigs(defaultConfigs)
}

// internal
//
func getConfigFileLock() (*flock.FileLock, error) {
	if _, err := os.Stat(LocalConfigFileLock); os.IsNotExist(err) {
		if fd, _ := os.Create(LocalConfigFileLock); fd != nil {
			fd.Close()
		}
	}

	return flock.New(LocalConfigFileLock)
}

// loadConfigs load local configs unsafe
func loadConfigs() (*LocalConfigs, error) {
	bs, err := ioutil.ReadFile(LocalConfigFile)
	if err != nil {
		if os.IsNotExist(err) { // if file not exists, maybe first run CLI, save and return initial configs
			return defaultConfigs, saveConfigs(defaultConfigs)
		}
		return nil, err
	}

	var cfgs *LocalConfigs
	err = json.Unmarshal(bs, &cfgs)
	if err != nil {
		return nil, err
	}

	if cfgs.Hosts == nil {
		cfgs.Hosts = make(map[string]*AdbotHost)
	}

	cfgs.uncoverSensitive()
	return cfgs, err
}

// saveConfigs save given configs to disk unsafe
func saveConfigs(cfgs *LocalConfigs) error {
	var (
		dir = path.Dir(LocalConfigFile)
	)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
			return err
		}
	}

	cfgs.hideSensitive()
	bs, _ := json.MarshalIndent(cfgs, "", "    ")

	return ioutil.WriteFile(LocalConfigFile, append(bs, '\r', '\n'), os.FileMode(0644))
}

// sorter
//

// AdbotHostWrapperSorter is exported
type AdbotHostWrapperSorter []*AdbotHostWrapper

func (s AdbotHostWrapperSorter) Len() int      { return len(s) }
func (s AdbotHostWrapperSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s AdbotHostWrapperSorter) Less(i, j int) bool {
	if s[i].Current {
		return true
	}
	return false
}
