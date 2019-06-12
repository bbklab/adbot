package mole

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	// FILESHUTDOWN is exported
	FILESHUTDOWN = fmt.Sprintf("/etc/.%s.shutdown", path.Base(os.Args[0]))

	errNotConnected       = errors.New("agent is not connected to master")
	errClosed             = errors.New("agent listener closed")
	errShutdown           = errors.New("agent has been shut down permanently")
	errMasterAddrRequired = errors.New("agent initialization require master address")
	errAgentIDRequired    = errors.New("agent initialization require agent id")
)

// ConnHandler is exported
type ConnHandler interface {
	HandleWorkerConn(c net.Conn) error
}

// Agent is a runtime mole agent
type Agent struct {
	id         string        // unique agent id
	masterAddr string        // master addr [host:port]
	conn       net.Conn      // control connection to master
	handler    ConnHandler   // worker connection handler
	hbch       chan struct{} // heartbeat stop chan
}

// NewAgent initialize a new runtime agent
func NewAgent(id string, masterAddr string) *Agent {
	if id == "" {
		log.Fatalln(errAgentIDRequired)
	}

	if masterAddr == "" {
		log.Fatalln(errMasterAddrRequired)
	}
	if _, _, err := net.SplitHostPort(masterAddr); err != nil {
		log.Fatalln(err)
	}

	if isAgentShutdown() {
		log.Fatalln(errShutdown)
	}

	return &Agent{
		id:         id,
		masterAddr: masterAddr,
	}
}

func isAgentShutdown() bool {
	finfo, _ := os.Stat(FILESHUTDOWN)
	return finfo != nil
}

func agentShutdown() error {
	return ioutil.WriteFile(FILESHUTDOWN, []byte(`this agent has been permanently shutdown by master`), os.FileMode(0644))
}

// Join join the initialized agent to master
func (a *Agent) Join() error {
	conn, err := net.DialTimeout("tcp", a.masterAddr, time.Second*10)
	if err != nil {
		return fmt.Errorf("agent Join error: %v", err)
	}

	// Disable IO Read TimeOut
	// note: after set this, maybe we can't aware the gone away master (eg: sudden power off) in a short time
	conn.SetReadDeadline(time.Time{})

	// Enable TCP KeepAlive on the socket connection
	setKeepAlive(conn, &kaopt{time.Second * 30, 3, time.Second * 10})

	// save the reference for the persistent connection
	a.conn = conn

	// set agent Finalizer on GC()
	runtime.SetFinalizer(a, func(a *Agent) { a.close() })

	// send join cmd
	command := newCmd(cmdJoin, a.id, "")
	_, err = conn.Write(command)
	return err
}

func (a *Agent) close() {
	if a.conn != nil {
		a.conn.Close()
	}
}

// ServeProtocol listening on the persistent connection to master
// and decode the protocol command
func (a *Agent) ServeProtocol() error {
	// ensure joined
	if a.conn == nil {
		return errNotConnected
	}
	defer a.close()

	// run heartbeat loop
	go a.runHeartbeatLoop()
	defer a.stopHeartbeatLoop()

	log.Printf("agent serve protocol started")
	defer log.Warnln("agent serve protocol stopped")

	// protocol decoder
	dec := newDecoder(a.conn)

	for {
		cmd, err := dec.Decode()
		if err != nil {
			return fmt.Errorf("agent decode protocol error: %v", err) // control conn closed, exit Serve() to trigger agent ReJoin
		}
		if err := cmd.valid(); err != nil {
			log.Errorf("agent received invalid command: %v", err)
			continue
		}

		// handle master command
		switch cmd.Cmd {

		case cmdShutdown:
			if err := agentShutdown(); err != nil {
				log.Errorf("agent shutdown error: %v", err)
			}

		case cmdNewWorker: // launch a new tcp connection as the worker connection
			connWorker, err := net.DialTimeout("tcp", a.masterAddr, time.Second*10)
			if err != nil {
				log.Errorf("agent dial master error: %v", err)
				continue
			}
			command := newCmd(cmdNewWorker, a.id, cmd.WorkerID)
			_, err = connWorker.Write(command)
			if err != nil {
				log.Errorf("agent notify back worker id error: %v", err)
				continue
			}

			log.Debugln("agent launched a new tcp worker connection with id:", cmd.WorkerID)

			// put the worker conn to agent connection pools
			go a.handler.HandleWorkerConn(connWorker)
		}
	}
}

// runHeartbeatLoop send cmdHeartbeat to master periodic
func (a *Agent) runHeartbeatLoop() {
	if a.hbch == nil {
		a.hbch = make(chan struct{})
	}

	log.Printf("agent heartbeat loop started")
	defer log.Warnln("agent heatbeat loop stopped")

	var (
		ticker = time.NewTicker(time.Second * 10)
		hb     = newCmd(cmdHeartbeat, a.id, "")
	)

	for {
		select {

		case <-ticker.C:
			if _, err := a.conn.Write(hb); err != nil {
				log.Errorf("agent heartbeat error: %v", err)
			} else {
				log.Debugf("agent heartbeat succeed")
			}

		case <-a.hbch:
			return
		}
	}
}

// stopHeartbeatLoop stop sending heatbeat to master
func (a *Agent) stopHeartbeatLoop() {
	if a.hbch != nil {
		close(a.hbch)
	}
}

// NewListener create a virtual in-memory net listener
// to process the new-worker connection (which is revert launched by agent)
func (a *Agent) NewListener() net.Listener {
	l := &AgentListener{
		pool: make(chan net.Conn),
	}
	a.handler = l
	return l
}

// AgentListener is a virtul in-memory net listener
type AgentListener struct {
	sync.RWMutex               // protect flag closed
	closed       bool          // flag on pool closed
	pool         chan net.Conn // worker connection pool
}

// Accept implement net.Listener interface
// the caller could process the cached worker connection in the pool via `AgentListener`
func (l *AgentListener) Accept() (net.Conn, error) {
	conn, ok := <-l.pool
	if !ok {
		return nil, errClosed
	}
	return conn, nil
}

// Close implement net.Listener interface
func (l *AgentListener) Close() error {
	l.Lock()
	if !l.closed {
		l.closed = true
		close(l.pool) // so the Accept() returned immediately
	}
	l.Unlock()
	return nil
}

// Addr implement net.Listener interface
func (l *AgentListener) Addr() net.Addr {
	return net.Addr(nil)
}

// HandleWorkerConn implement ConnHandler interface
// put the worker connection to the pool
func (l *AgentListener) HandleWorkerConn(conn net.Conn) error {
	l.RLock()
	defer l.RUnlock()
	if l.closed {
		return errClosed
	}
	l.pool <- conn
	return nil
}
