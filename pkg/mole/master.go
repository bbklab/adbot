package mole

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/pubsub"
	"github.com/bbklab/adbot/pkg/utils"
)

var (
	pub      *pubsub.Publisher // new connected worker conn publisher
	evpub    *pubsub.Publisher // node event publisher
	onceJoin sync.Once
	onceDie  sync.Once
)

// NodeJoinCallBack is the call back function while node joined
type NodeJoinCallBack func(id string, firstJoin bool) error

// NodeDieCallBack is the call back function while node died
type NodeDieCallBack func(id string) error

// Master is a runtime cluster object
//
//
type Master struct {
	sync.RWMutex                          // protect agents map
	agents       map[string]*ClusterAgent // agents held all of joined agents
	listener     net.Listener             // specified listener
	authToken    string                   // TODO auth token
	cbNodeJoin   NodeJoinCallBack         // node join call back func
	cbNodeDie    NodeDieCallBack          // node die call back func
}

// NewMaster initialize a runtime cluster master
func NewMaster(l net.Listener) *Master {
	return &Master{
		listener:  l,
		authToken: "xxx",
		agents:    make(map[string]*ClusterAgent),
	}
}

// RegisterNodeJoinCallBack is exported
func (m *Master) RegisterNodeJoinCallBack(cb NodeJoinCallBack) {
	onceJoin.Do(func() {
		m.cbNodeJoin = cb
	})
}

// RegisterNodeDieCallBack is exported
func (m *Master) RegisterNodeDieCallBack(cb NodeDieCallBack) {
	onceDie.Do(func() {
		m.cbNodeDie = cb
	})
}

// ExecNodeJoinCallBack is exported
func (m *Master) ExecNodeJoinCallBack(id string, firstJoin bool) {
	if m.cbNodeJoin == nil {
		return
	}
	if err := m.cbNodeJoin(id, firstJoin); err != nil {
		log.Errorf("agent join callback error: %v", err)
	}
}

// ExecNodeDieCallBack is exported
func (m *Master) ExecNodeDieCallBack(id string) {
	if m.cbNodeDie == nil {
		return
	}
	if err := m.cbNodeDie(id); err != nil {
		log.Errorf("agent die callback error: %v", err)
	}
}

// Serve listen for node's new connections with cmd: cmdJoin, cmdNewWorker
func (m *Master) Serve() error {
	// init global publisher for new-connected worker connections
	pub = pubsub.NewPublisher(time.Second*5, 1024)
	evpub = pubsub.NewPublisher(time.Second*5, 1024)

	// serving tcp
	for {
		conn, err := m.listener.Accept()
		if err != nil {
			log.Errorln("master Accept error: %v", err)
			return err
		}

		go m.handle(conn) // handle node's new connection (cmdJoin, cmdNewWorker)
	}
}

// handle on node's each of new connection with cmd: cmdJoin, cmdNewWorker
func (m *Master) handle(conn net.Conn) {
	// enable TCP KEEPALIVE
	setKeepAlive(conn, &kaopt{time.Second * 30, 3, time.Second * 10})

	cmd, err := newDecoder(conn).Decode()
	if err != nil {
		log.Errorf("master decode protocol error: %v", err)
		return
	}

	if err := cmd.valid(); err != nil {
		log.Errorf("master received invalid command: %v", err)
		return
	}

	switch cmd.Cmd {

	case cmdJoin:
		log.Printf("agent %s joined", cmd.AgentID)
		firstJoin := m.AddAgent(cmd.AgentID, conn)           // this is the persistent control connection
		m.ExecNodeJoinCallBack(cmd.AgentID, firstJoin)       // run node join call back if we have one
		evpub.Publish(newNodeEvent(cmd.AgentID, NodeEvJoin)) // pub node event

	case cmdNewWorker:
		log.Debugf("agent %s launched a new worker connection %s", cmd.AgentID, cmd.WorkerID)
		ca := &clusterWorker{
			agentID:       cmd.AgentID,
			workerID:      cmd.WorkerID,
			conn:          conn, // this is the worker connection
			establishedAt: time.Now(),
		}
		// note:
		// if current net.Conn hit none of subscribers, we just close it to avoid leaks.
		// because maybe some of reverse Dial() already timeout and evict the subscriber, so
		// current net.Conn will never be picked up by any consumers.
		if n := pub.NumHitTopic(ca); n == 0 {
			log.Warnf("agent %s launched a staled worker connection %s, close it.", cmd.AgentID, cmd.WorkerID)
			conn.Close()
		} else {
			pub.Publish(ca)
			evpub.Publish(newNodeEvent(cmd.AgentID, NodeEvNewWorker)) // pub node event
		}
		m.FreshAgent(cmd.AgentID, true)

	case cmdLeave:
		log.Println("agent leaved", cmd.AgentID)
		m.CloseAgent(cmd.AgentID)

	}
}

// AddAgent register an persistent tcp conn of a given agent id
// return the flag if the agent is the first join since boot up
func (m *Master) AddAgent(id string, conn net.Conn) bool {
	m.Lock()
	defer m.Unlock()

	var (
		firstJoin   = true
		prevHealthy *bool
	)

	// note: if we already have agent connection with the same id
	// close the previous staled connection and use the new one
	// mostly occurred while the same agent rejoined to the cluster
	// similar to CloseAgent()
	if agent, ok := m.agents[id]; ok {
		firstJoin = false
		prevHealthy = pbool(agent.Healthy())

		agent.conn.Close() // close the previous persistent conn, so previous watchAgentProtocol() quit
	}

	// register with new cluster agent
	ca := &ClusterAgent{
		id:           id,
		conn:         conn,
		joinAt:       time.Now(),
		lastActiveAt: time.Now(),
		healthy:      true,
	}

	m.agents[id] = ca

	// start agent protocol watch loop, quit while agent closed
	go m.watchAgentProtocol(id)

	// if the agent is the first join, run agent heartbeat watch loop until node removal
	// note: loop do NOT quit while agent closed, quit on agent removed
	if firstJoin {
		go m.watchAgentHeartbeat(id)
		return firstJoin
	}

	// the agent is rejoin to the cluster, emit events: `rejoin` and `recovery`
	evpub.Publish(newNodeEvent(id, NodeEvRejoin)) // rejoin
	if prevHealthy != nil && !*prevHealthy {
		evpub.Publish(newNodeEvent(id, NodeEvRecovery)) // recovery
	}
	return firstJoin
}

// CloseAllAgents just close all of registered agents once (but keep agents registered)
func (m *Master) CloseAllAgents() {
	for id := range m.Agents() {
		m.CloseAgent(id)
	}
}

// CloseAgent just close the persistent connection once (keep agent registered)
// if the agent re-connected to master, will produce a rejoin event
func (m *Master) CloseAgent(id string) {
	m.FreshAgent(id, false) // mark as unhealthy firstly, so the watchAgentHeartbeat() won't working

	m.Lock()
	defer m.Unlock()

	if agent, ok := m.agents[id]; ok {
		agent.conn.Close()                           // close the persistent conn, so watchAgentProtocol() quit
		evpub.Publish(newNodeEvent(id, NodeEvClose)) // pub node event
	}
}

// ShutdownAgent permanently shutdown the agent
//   - tell agent do NOT rejoin any more
//   - close persistent conn
//   - unregister
func (m *Master) ShutdownAgent(id string) error {
	m.Lock()
	defer m.Unlock()

	if agent, ok := m.agents[id]; ok {
		// tell the agent do NOT rejoin any more
		command := newCmd(cmdShutdown, id, "")
		_, err := agent.conn.Write(command)
		if err != nil {
			return fmt.Errorf("agent Shutdown().write error %v", err)
		}

		// close and unregister
		agent.conn.Close()                              // close the persistent conn, so watchAgentProtocol() quit
		delete(m.agents, id)                            // unregister
		m.ExecNodeDieCallBack(id)                       // run node die call back if we have one
		evpub.Publish(newNodeEvent(id, NodeEvShutdown)) // pub node event
	}
	return nil
}

// FreshAgent refresh the node's `lastActiveAt` & `healthy`, only called
// while received new worker connections or agent's heartbeat
func (m *Master) FreshAgent(id string, healthy bool) {
	m.Lock()
	defer m.Unlock()
	if agent, ok := m.agents[id]; ok {
		if healthy {
			agent.lastActiveAt = time.Now()
			agent.healthy = true
		} else {
			agent.healthy = false
		}
	}
}

// Agent obtain one specified agent
// the caller should check the returned ClusterAgent is not nil
// otherwise the agent hasn't connected to the cluster ever since boot time
func (m *Master) Agent(id string) *ClusterAgent {
	m.RLock()
	defer m.RUnlock()
	return m.agents[id]
}

// Agents list all registered agents
func (m *Master) Agents() map[string]*ClusterAgent {
	m.RLock()
	defer m.RUnlock()
	return m.unsafeAgents(nil)
}

// HealthAgents list all healhty agents
func (m *Master) HealthAgents() map[string]*ClusterAgent {
	m.RLock()
	defer m.RUnlock()
	return m.unsafeAgents(pbool(true))
}

// UnHealthAgents list all unhealhty agents
func (m *Master) UnHealthAgents() map[string]*ClusterAgent {
	m.RLock()
	defer m.RUnlock()
	return m.unsafeAgents(pbool(false))
}

// should be called under protect of mutex
func (m *Master) unsafeAgents(healthy *bool) map[string]*ClusterAgent {
	if healthy == nil {
		return m.agents
	}

	ret := make(map[string]*ClusterAgent)
	for id, agent := range m.agents {
		if agent.Healthy() == *healthy {
			ret[id] = agent
		}
	}
	return ret
}

// watch on agent's heartbeat protocol command, do NOT quit while agent closed, quit on agent removed
func (m *Master) watchAgentHeartbeat(id string) {
	log.Printf("starting to watch agent %s heartbeat ...", id)
	defer log.Warnf("stopped watch agent %s heartbeat, agent maybe removed", id)

	// timer heartbeat checker
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	// FIXME hard coded agent heartbeat interval & ttl
	var (
		agentHbIntvl    = time.Second * 10
		agentHbFlagging = 3
		agentHbTTL      = 5
	)

	for range ticker.C { // healthy timer checking
		ca := m.Agent(id)
		if ca == nil { // agent has been removed
			return
		}

		if !ca.Healthy() { // if already unhealthy (fresh set while being die or being closed), do not emit events any more
			continue
		}

		var (
			now              = time.Now()
			lastActive       = ca.LastActiveAt()
			flaggingDeadLine = lastActive.Add(agentHbIntvl * time.Duration((agentHbFlagging)))
			deadDeadLine     = lastActive.Add(agentHbIntvl * time.Duration((agentHbTTL)))
		)

		if now.After(deadDeadLine) {
			log.Warnf("agent %s dead", id)
			m.FreshAgent(id, false)                    // mark as unhealthy
			m.ExecNodeDieCallBack(id)                  // run node die call back if we have one
			evpub.Publish(newNodeEvent(id, NodeEvDie)) // pub node event
			continue
		}

		if now.After(flaggingDeadLine) {
			log.Warnf("agent %s healthy, but assumed to be flagging", id)
			evpub.Publish(newNodeEvent(id, NodeEvFlagging)) // pub node event
		}
	}
}

// watch on agent's protocol command via persistent connection, quit while agent closed.
// this method actually processing all of agent's protocol command (currently noly heartbeat)
func (m *Master) watchAgentProtocol(id string) {
	ca := m.Agent(id)
	if ca == nil {
		log.Warnf("skip watch agent %s protocol command as the agent is not exists", id)
		return
	}

	log.Printf("starting to watch agent %s protocol command ...", id)
	defer log.Warnf("stopped watch agent %s protocol command, agent maybe disconnected", id)

	// protocol decoder
	var dec = newDecoder(ca.conn)

	// decode agent's protocol command (currently only heartbeat)
	for {
		cmd, err := dec.Decode()
		if err != nil {
			// node disconnected, exit the loop
			//   - EOF  --> mostly peer agent close the conn
			//   - use of closed network connection  --> mostly master close the conn
			log.Errorf("master decode protocol error: %v", err)
			return
		}

		if err := cmd.valid(); err != nil {
			log.Errorf("master received invalid command: %v", err)
			continue
		}

		if cmd.Cmd == cmdHeartbeat {
			var id = cmd.AgentID
			log.Debugf("master received agent %s heartbeat", id)
			evpub.Publish(newNodeEvent(id, NodeEvHeartbeat)) // pub node event
			if tmpca := m.Agent(id); tmpca != nil && !tmpca.Healthy() {
				evpub.Publish(newNodeEvent(id, NodeEvRecovery)) // pub node event
			}
			m.FreshAgent(id, true)
		}
	}
}

// ClusterAgent is a runtime agent object within master lifttime
//
//
type ClusterAgent struct {
	id           string   // agent id
	conn         net.Conn // persistent control connection
	joinAt       time.Time
	lastActiveAt time.Time
	healthy      bool
}

// MarshalJSON implement json.Marshaler
func (ca *ClusterAgent) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"id":          ca.ID(),
		"remote":      ca.RemoteAddr(),
		"joined_at":   ca.JoinAt(),
		"last_active": ca.LastActiveAt(),
		"healthy":     ca.Healthy(),
	}

	return json.Marshal(m)
}

// RemoteAddr show node's remote address
func (ca *ClusterAgent) RemoteAddr() string {
	if ca.conn != nil {
		return ca.conn.RemoteAddr().String()
	}
	return ""
}

// ID is exported
func (ca *ClusterAgent) ID() string {
	return ca.id
}

// JoinAt is exported
func (ca *ClusterAgent) JoinAt() time.Time {
	return ca.joinAt
}

// LastActiveAt is exported
func (ca *ClusterAgent) LastActiveAt() time.Time {
	return ca.lastActiveAt
}

// Healthy is exported
func (ca *ClusterAgent) Healthy() bool {
	return ca.healthy
}

// Events launch a node event subscriber
// note it's the caller's responsibility to call EvictEventSubscriber() to
// evict the subscriber to prevent memory leak
func (ca *ClusterAgent) Events() pubsub.Subcriber {
	return evpub.Subcribe(func(v interface{}) bool {
		if vv, ok := v.(*NodeEvent); ok {
			return vv.ID == ca.id
		}
		return false
	})
}

// EvictEventSubscriber evict a given node event subscriber
func (ca *ClusterAgent) EvictEventSubscriber(sub pubsub.Subcriber) {
	evpub.Evict(sub)
}

// Dial specifies the dial function for creating unencrypted TCP connections within the http.Client
func (ca *ClusterAgent) Dial(network, addr string) (net.Conn, error) {
	wid := utils.RandomNumber(10)

	// NOTE: should run subscriber firstly to avoid the situation that new worker is faster than broadcaster
	// and then Dial() will wait for an broadcast-ed event until timeout.

	// subcribe waitting for the worker id connection
	sub := pub.Subcribe(func(v interface{}) bool {
		if vv, ok := v.(*clusterWorker); ok {
			return vv.workerID == wid && vv.agentID == ca.id
		}
		return false
	})
	defer pub.Evict(sub) // evict the subcriber before exit

	// notify the agent to create a new worker connection
	// TODO if agent stopped, this will NOT fail fast to return error, find out why ?
	command := newCmd(cmdNewWorker, ca.id, wid)
	_, err := ca.conn.Write(command)
	if err != nil {
		return nil, fmt.Errorf("agent Dial().write: new worker command error %v", err)
	}

	select {
	case cw := <-sub:
		return cw.(*clusterWorker).conn, nil
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("agent Dial().wait: new worker conn %s timeout", wid)
	}
}

// Client obtain a http client for an agent with customized dialer
func (ca *ClusterAgent) Client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			// use our cluster node revert dial
			Dial: ca.Dial,

			// on case with request header: "Expect: 100-continue"
			// we should keep this non-zero, otherwise will causes the large body
			// to be sent immediately even after got an BadRequest header response
			ExpectContinueTimeout: time.Second * 5,

			// note: use context.WithTimeout to limit each request max timeout instead of this overall timeout
			// max duration to wait for peer's response headers after fully writing the request
			// this maybe useful while requesting onto one `deadlocked docker daemon Api`
			// if hit this timeout, we will met error: `net/http: timeout awaiting response headers`
			// ResponseHeaderTimeout: time.Second * 60,

			// note: followings options keep same as http.DefaultTransport
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// clusterWorker is a worker connection
//
//
type clusterWorker struct {
	agentID       string
	workerID      string
	conn          net.Conn
	establishedAt time.Time
}

func pbool(v bool) *bool {
	return &v
}
