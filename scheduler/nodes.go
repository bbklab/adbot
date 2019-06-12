package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/adbot"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

var (
	// ErrMsgNoSuchNodeOnline define the errmsg that node is not online
	ErrMsgNoSuchNodeOnline = "No such node %s online"
)

//
// cluster node internal call utils
//

// NodePing ping check node once by real time
func NodePing(id string) (time.Duration, error) {
	startAt := time.Now()
	nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/ping", id), nil)

	resp, err := ProxyNode(id, nodeReq, time.Second*10)
	if err != nil {
		return time.Duration(0), err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return time.Duration(0), fmt.Errorf("%d - %s", code, string(bs))
	}

	return time.Since(startAt), nil
}

// NodeInfo query node's sysinfo once by real time
func NodeInfo(id string) (*types.SysInfo, error) {
	nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/sysinfo", id), nil)

	resp, err := ProxyNode(id, nodeReq, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("node:%s - %d - %s", id, code, string(bs))
	}

	var info *types.SysInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	return info, err
}

// NodeVersion query node's version once by real time
func NodeVersion(id string) (string, error) {
	nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/version", id), nil)

	resp, err := ProxyNode(id, nodeReq, time.Second*10)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("%d - %s", code, string(bs))
	}

	var version *types.Version
	if err = json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return "", err
	}
	return version.Version, nil
}

// NodeStats redirect node live stream *types.SysInfo output
func NodeStats(id string) (io.ReadCloser, error) {
	nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/stats", id), nil)

	resp, err := ProxyNode(id, nodeReq, 0)
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("node:%s - %d - %s", id, code, string(bs))
	}

	return resp.Body, nil
}

// DoNodeExec exec remote node cmd and redirect live stream cmd output
func DoNodeExec(id string, cmd *types.NodeCmd) (io.ReadCloser, error) {
	cmdbs, _ := json.Marshal(cmd)
	nodeReq, _ := http.NewRequest("POST", fmt.Sprintf("http://%s/api/exec", id), bytes.NewBuffer(cmdbs))

	resp, err := ProxyNode(id, nodeReq, 0)
	if err != nil {
		return nil, err
	}

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("node:%s - %d - %s", id, code, string(bs))
	}

	return resp.Body, nil
}

//
// Adb Device
//

// DoNodeQueryAdbDevices query node's adb devices info
func DoNodeQueryAdbDevices(id string) (map[string]*adbot.AndroidSysInfo, error) {
	nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/adbot/devices", id), nil)

	resp, err := ProxyNode(id, nodeReq, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("node:%s - %d - %s", id, code, string(bs))
	}

	var dvcsinfo map[string]*adbot.AndroidSysInfo
	err = json.NewDecoder(resp.Body).Decode(&dvcsinfo)
	return dvcsinfo, err
}

//
// Terminal
//

// DoNodeTerminalResizing send window resizing requests to specified node's terminal
func DoNodeTerminalResizing(id, wid string, width, height int) error {
	nodeReq, _ := http.NewRequest("PATCH", fmt.Sprintf("http://%s/api/terminal?wid=%s&width=%d&height=%d", id, wid, width, height), nil)
	resp, err := ProxyNode(id, nodeReq, 0)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code != 200 {
		bs, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("node:%s - %d - %s", id, code, string(bs))
	}

	return nil
}

// DoWaitNodeTerminalWindow wait given node termial terminal fd prepared
func DoWaitNodeTerminalWindow(id, wid string, maxWait time.Duration) error {
	timeout := time.After(maxWait)

	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()

	for {
		select {

		case <-timeout:
			return fmt.Errorf("wait node terminal %s window fd timeout in %s", id, maxWait)

		case <-ticker.C:
			nodeReq, _ := http.NewRequest("HEAD", fmt.Sprintf("http://%s/api/terminal?wid=%s", id, wid), nil)
			resp, err := ProxyNode(id, nodeReq, 0)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
	}
}

// ProxyNodeHandle direct bridge between Client-Request and Node-Response
func ProxyNodeHandle(id string, req *http.Request, w http.ResponseWriter) {
	resp, err := ProxyNode(id, req, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyNode send given http.Request to specified node and obtain the node http.Response
// if there is an error, it will be of type *ProxyNodeError which could be used to
// distinguish from any application logic errors
func ProxyNode(id string, req *http.Request, timeout time.Duration) (*http.Response, error) {
	node := Node(id)
	if node == nil {
		return nil, &ProxyNodeError{id, fmt.Sprintf(ErrMsgNoSuchNodeOnline, id)}
	}

	// rewrite request (prevent fd leak)
	req.Close = true
	req.Header.Set("Connection", "close")
	req.Host = id

	// with timeout context
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		req = req.WithContext(ctx)
		defer cancel()
	}

	log.Debugf("proxying node request: %s", req.URL.String())

	resp, err := node.Client().Do(req)
	if err != nil {
		return nil, &ProxyNodeError{id, err.Error()}
	}

	return resp, nil
}

// ProxyNodeError implement error represents cluster node request errors
type ProxyNodeError struct {
	NodeID string `json:"node_id"`
	Errmsg string `json:"errmsg"`
}

func (e *ProxyNodeError) Error() string {
	return fmt.Sprintf("node:%s, error:%s", e.NodeID, e.Errmsg)
}

//
// followings methods
// redirect mostly of mole master's management to package exported methods
//

// Node pick up given runtime mole agent
func Node(id string) *mole.ClusterAgent {
	return sched.master.Agent(id)
}

// Nodes pick up all of runtime mole agents
func Nodes() map[string]*mole.ClusterAgent {
	return sched.master.Agents()
}

// HealthNodes pick up healthy runtime mole agents
func HealthNodes() map[string]*mole.ClusterAgent {
	return sched.master.HealthAgents()
}

// UnHealthNodes similar as above but for unhealthy
func UnHealthNodes() map[string]*mole.ClusterAgent {
	return sched.master.UnHealthAgents()
}

// CloseNode close the persistent connection of given mole agent
func CloseNode(id string) {
	sched.master.CloseAgent(id)
}

// ShutdownNode permanently shutdown the given mole agent
func ShutdownNode(id string) error {
	return sched.master.ShutdownAgent(id)
}

// PickupRandomNode pick up one healthy node by random
func PickupRandomNode() (*mole.ClusterAgent, error) {
	var (
		healthyNodes []*mole.ClusterAgent
	)

	for _, node := range HealthNodes() {
		healthyNodes = append(healthyNodes, node)
	}

	n := len(healthyNodes)
	if n == 0 {
		return nil, errors.New("no healthy cluster nodes avaliable")
	}

	return healthyNodes[rand.Intn(n)], nil
}

// FilterNodes pick up current db nodes by filters:
//  - labels
//  - online/offline
//  - with master
func FilterNodes(filter label.Labels, online *bool, excludeMaster bool) ([]*types.Node, error) {
	// query all db nodes
	nodes, err := store.DB().ListNodes(nil)
	if err != nil {
		return nil, err
	}

	// filter by labels
	var idx int
	for _, node := range nodes {
		labels := node.Labels
		if labels.MatchAll(filter) {
			nodes[idx] = node
			idx++
		}
	}
	nodes = nodes[:idx]

	// filter by online (ignore flagging nodes)
	if online != nil {
		idx = 0                               // reset idx
		var expected = types.NodeStatusOnline // the caller want online or offline nodes
		if !*online {
			expected = types.NodeStatusOffline
		}
		for _, node := range nodes {
			if node.Status == expected {
				nodes[idx] = node
				idx++
			}
		}
	}
	nodes = nodes[:idx]

	// filter by with master
	if excludeMaster {
		idx = 0
		for _, node := range nodes {
			if info := node.SysInfo; info != nil && !info.WithMaster {
				nodes[idx] = node
				idx++
			}
		}
	}
	nodes = nodes[:idx]

	return nodes, nil
}
