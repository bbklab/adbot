package scheduler

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
	"github.com/bbklab/adbot/version"
)

// NodeJoinCallBack is executed while node joined onto master
// note: called within mole master logic
//
func NodeJoinCallBack(id string, firstJoin bool) error {
	var (
		node = Node(id)
		err  error
	)

	if node == nil {
		return errors.New("can't pick up the newly joined node " + id)
	}

	// defer to notify the node install finished
	defer DelNodeJoinNotifier(id)

	// get node system info & version
	ver, _ := NodeVersion(id)
	info, _ := NodeInfo(id)
	remoteAddr := node.RemoteAddr()

	exists, _ := store.DB().GetNode(id)

	if exists == nil { // not exists
		// note: to make fit with manual installed nodes or local-cluster container nodes
		// because this type of nodes are not installed by ssh, so the db node won't exists before join
		// we insert a new db node while this type of nodes first joined to master
		err = store.DB().AddNode(&types.Node{
			ID:           id,
			Status:       types.NodeStatusOnline, // newly joined node as online
			Version:      version.GetVersion(),
			ErrMsg:       "",
			RemoteAddr:   remoteAddr,
			GeoInfo:      GetGeoInfoEn(remoteAddr),
			GeoInfoZh:    GetGeoInfoZh(remoteAddr),
			SysInfo:      info,
			SSHConfig:    nil, // because not by ssh
			Labels:       label.New(nil),
			JoinAt:       node.JoinAt(),
			LastActiveAt: node.LastActiveAt(),
		})
	} else { // exists, db updating newly joined node info
		setUpdator := bson.M{
			"status":         types.NodeStatusOnline,
			"version":        ver,
			"error":          "",
			"remote_addr":    remoteAddr,
			"geoinfo":        GetGeoInfoEn(remoteAddr),
			"geoinfo_zh":     GetGeoInfoZh(remoteAddr),
			"sysinfo":        info,
			"join_at":        node.JoinAt(),
			"last_active_at": node.LastActiveAt(),
		}
		err = store.DB().UpdateNode(id, bson.M{"$set": setUpdator})
	}
	if err != nil {
		return err
	}

	// if first join, start node refresher loop
	if firstJoin {
		go runNodeReFreshLoop(node)
		go runAdbNodeReFreshLoop(node)
	}
	return nil
}

// NodeDieCallBack is executed while node die
// note: called within mole master logic
//
func NodeDieCallBack(id string) error {
	return nil
}

// NodeOfflineCallBack is executed while node offline
// FIXME merge with above NodeDieCallBack
// note: called within runNodeReFreshLoop()
// note: triggered if trap node die event
func NodeOfflineCallBack(id string) {
}

// db node refresher until node is removed
// launched by node join call back only on the first join in the runtime
//
// - monitor node runtime events and refresh db node status
// - periodically collect and refresh db node sysinfo
// until the db node is removed
func runNodeReFreshLoop(node *mole.ClusterAgent) {
	var (
		id       = node.ID()
		loopName = fmt.Sprintf("node %s db refresher loop", id)
	)

	RegisterGoroutine("nodes_refresher", id)
	defer DeRegisterGoroutine("nodes_refresher", id)

	log.Printf("starting %s ...", loopName)
	defer log.Warnf("stopped %s, node maybe removed", loopName)

	// node event notifier
	nodeEvSub := node.Events()
	defer node.EvictEventSubscriber(nodeEvSub)

	// periodically timer notifier
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	// async refresher notifier
	nodeRefresh := make(chan struct{}, 1)
	AddNodeRefreshNotifier(id, nodeRefresh) // register node async refresh notifier
	defer DelNodeRefreshNotifier(id)        // avoid leaks while follings steps met errors

	// trigger node refresh once immediately before actually launch the loop
	go RefreshNodeAsync(id)

	var err error

	for {

		select {

		case <-ticker.C: // update db node sysinfo periodically
			time.Sleep(time.Second * time.Duration(rand.Int()%5+1)) // randomize the refresh loop to smooth potential burst load
			err = nodeRefreshOnce(id)

		case <-nodeRefresh: // async node refresh triggered
			err = nodeRefreshOnce(id)

		case ev := <-nodeEvSub: // update db status according by node runtime events

			switch ev.(*mole.NodeEvent).Type {

			case mole.NodeEvJoin, mole.NodeEvHeartbeat, mole.NodeEvNewWorker, mole.NodeEvRecovery, mole.NodeEvRejoin: // mark online
				err = MemoNodeStatus(id, types.NodeStatusOnline, nil, nil)

			case mole.NodeEvClose, mole.NodeEvDie: // mark offline
				NodeOfflineCallBack(id) // call offline callbacks
				err = MemoNodeStatus(id, types.NodeStatusOffline, nil, nil)

			case mole.NodeEvFlagging: // mark flagging
				err = MemoNodeStatus(id, types.NodeStatusFlagging, nil, nil)

			case mole.NodeEvShutdown: // removed
				return
			}
		}

		if err != nil {
			log.Errorf("node %s db refresher loop got error: %v", id, err)
		}

		if err == errNodeNotFound {
			return // db node maybe removed
		}
	}
}

var errNodeNotFound = errors.New("db node not found")

func nodeRefreshOnce(id string) error {
	if !isLeader() { // ensure we're the leader, otherwise skip
		return nil
	}

	// ensure db node exists, otherwise tell the loop to exit
	_, err := store.DB().GetNode(id)
	if err != nil {
		if store.DB().ErrNotFound(err) {
			return errNodeNotFound // use specified type error to identify node not exists
		}
		return err
	}

	cost, err := NodePing(id)
	if err != nil {
		return err
	}
	info, err := NodeInfo(id)
	if err != nil {
		return err
	}

	return MemoNodeStatus(id, types.NodeStatusOnline, info, ptype.TimeDuration(cost))
}

// RefreshTrafficRate is exported
func RefreshTrafficRate(prev, curr *types.SysInfo) {
	if prev == nil || curr == nil {
		return
	}
	if len(prev.Traffics) == 0 || len(curr.Traffics) == 0 {
		return
	}
	for inet, currInetV := range curr.Traffics {
		prevInetV, ok := prev.Traffics[inet]
		if !ok {
			continue
		}
		timeSeconds := int(currInetV.Time.Sub(prevInetV.Time).Seconds())
		if timeSeconds <= 0 { // time backwards !
			currInetV.RxRate = 0
			currInetV.TxRate = 0
			continue
		}
		if currInetV.RxBytes <= prevInetV.RxBytes {
			currInetV.RxRate = 0
		} else {
			currInetV.RxRate = (currInetV.RxBytes - prevInetV.RxBytes) / uint64(timeSeconds)
		}
		if currInetV.TxBytes <= prevInetV.TxBytes {
			currInetV.TxRate = 0
		} else {
			currInetV.TxRate = (currInetV.TxBytes - prevInetV.TxBytes) / uint64(timeSeconds)
		}
	}
}

//
// Node Join Notifier Manager
//

// joinMgr is an ongoing node join notifiers temporarily store manager
type joinMgr struct {
	sync.RWMutex
	m map[string]chan struct{} // node id -> notify channel
}

func newJoinMgr() *joinMgr {
	return &joinMgr{m: make(map[string]chan struct{})}
}

// DelNodeJoinNotifier is exported
func DelNodeJoinNotifier(id string) {
	sched.joinMgr.Lock()
	if ch, ok := sched.joinMgr.m[id]; ok {
		delete(sched.joinMgr.m, id)
		close(ch)
	}
	sched.joinMgr.Unlock()
}

// AddNodeJoinNotifier is exported
func AddNodeJoinNotifier(id string, ch chan struct{}) {
	sched.joinMgr.Lock()
	sched.joinMgr.m[id] = ch
	sched.joinMgr.Unlock()
}

//
// Node ReFresh Notifier Manager
//

// refreshMgr is a runtime node manually refresh notifier manager
type refreshMgr struct {
	sync.Mutex
	m map[string]chan struct{} // node id -> notify channel
}

func newRefreshMgr() *refreshMgr {
	return &refreshMgr{m: make(map[string]chan struct{})}
}

// DelNodeRefreshNotifier is exported
func DelNodeRefreshNotifier(id string) {
	sched.refreshMgr.Lock()
	if ch, ok := sched.refreshMgr.m[id]; ok {
		delete(sched.refreshMgr.m, id)
		close(ch)
	}
	sched.refreshMgr.Unlock()
}

// AddNodeRefreshNotifier is exported
func AddNodeRefreshNotifier(id string, ch chan struct{}) {
	sched.refreshMgr.Lock()
	sched.refreshMgr.m[id] = ch
	sched.refreshMgr.Unlock()
}

// RefreshNodeAsync is exported
func RefreshNodeAsync(id string) {
	sched.refreshMgr.Lock()
	defer sched.refreshMgr.Unlock()

	ch, ok := sched.refreshMgr.m[id]
	if !ok {
		return
	}

	// send avoid block
	select {
	case ch <- struct{}{}:
	default:
	}
}
