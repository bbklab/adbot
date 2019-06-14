package master

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/api"
	"github.com/bbklab/adbot/master/ha"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/pkg/pidfile"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

var (
	defaultUnixSock = "/var/run/adbot/adbot.sock"
	defaultPidFile  = "/var/run/adbot/adbot.pid"
)

// Master is the runtime adbot master
type Master struct {
	cfg           *types.MasterConfig
	apiserver     *api.Server
	clusterMaster *mole.Master
	tcpMux        *tcpMux // dispatch tcp Conn to clusterMaster & apiServer
	cmpg          ha.Campaigner
	unixSock      string // unix sock file
	pidFile       string // pid file
}

// New initialize a runtime adbot master
func New(cfg *types.MasterConfig) *Master {
	// set pid file & unix sock file
	var unixSock = cfg.UnixSock
	if unixSock == "" {
		unixSock = defaultUnixSock
	}

	var pidFile = cfg.PidFile
	if pidFile == "" {
		pidFile = defaultPidFile
	}

	// save pid firstly (with pid conflicts checking)
	if err := pidfile.New(pidFile); err != nil {
		log.Fatalf("save pid file error: %v", err)
	}

	// db store setup
	if err := store.Setup(cfg.Store); err != nil {
		log.Fatalln("db store setup error:", err)
	}

	// tcp mux
	tcpMux := newTCPMux(cfg.Listen)
	hl := tcpMux.NewHTTPListener()
	ml := tcpMux.NewMoleListener()

	// newly create a https listener if user provide key/cert files
	var (
		flag  = flagHTTP | flagMole
		tlshl net.Listener
	)
	if cfg.RequireServeTLS() {
		tlshl = tcpMux.NewHTTPSListener()
		flag |= flagHTTPS
	}
	// set tcpmxuer serve flag, so it knows what protocols it's serving
	tcpMux.SetServeFlag(flag)

	// mole master
	master := mole.NewMaster(ml)

	// api server
	// typically the api server doesn't need the master configs parameter `cfg`,
	// we pass it to api server only for debug dumping via api
	apiServer := api.NewServer(unixSock, hl, tlshl, cfg)

	// init runtime scheduler (top level)
	scheduler.Init(master, cfg.PublicKeyData)

	cmpg := ha.NewDummyCampaigner()

	return &Master{
		cfg:           cfg,
		apiserver:     apiServer,
		clusterMaster: master,
		tcpMux:        tcpMux,
		cmpg:          cmpg,
		unixSock:      unixSock,
		pidFile:       pidFile,
	}
}

// Run launch the initialized adbot master
func (m *Master) Run() {

	m.initDBGlobalSettings()

	go m.exitTrap()

	// start tcpmuxer / mole master / api server
	go func() {
		log.Println("master tcpmuxer serving ...")
		if err := m.tcpMux.ListenAndServe(); err != nil {
			log.Fatalln("master tcpmuxer serve error:", err)
		}
	}()

	go func() {
		log.Println("master mole serving ...")
		if err := m.clusterMaster.Serve(); err != nil {
			log.Fatalln("master mole serve error:", err)
		}
	}()

	go func() {
		log.Println("master api serving ...")
		if err := m.apiserver.Run(); err != nil {
			log.Fatalln("master api serve error:", err)
		}
	}()

	// wait m.apiserver ready to prevent following m.apiserver.SetLeader() panic
	for {
		if m.apiserver.IsReady() {
			break
		}
		time.Sleep(time.Millisecond * 300)
	}
	log.Printf("master api server ready")

	// start the leadership election
ELECT:

	log.Printf("[HA] start to campaign the leader election ...")

	electedCh, errCh, err := m.cmpg.WaitElection()
	if err != nil {
		log.Errorf("[HA] master wait election error: [%v], retry ...", err)
		time.Sleep(time.Second * 3)
		goto ELECT
	}

	for {
		select {

		case err := <-errCh: // election met error, eg: `lock server` is unreachable
			log.Errorf("[HA] electing the leader met error: [%v], retry ...", err)
			time.Sleep(time.Second * 3)
			goto ELECT

		case isElected := <-electedCh: // the `lock server` is reachable and return the election result normally

			if !isElected {

				log.Warnf("[HA] I lost the leadership! campaign for it again")
				m.apiserver.SetLeader(false)     // then Api -> 303, agents won't join on me and Api won't serving
				scheduler.SetLeader(false)       // then scheduler knows the current role, it's background loops(guarders) will be disabled
				m.clusterMaster.CloseAllAgents() // then all agents will re-detect the leader master

			} else {

				log.Printf("[HA] Hooray! I won the election! I'm now the leader")
				log.Printf("master is working on some initializations, this may take a while ...")
				m.initDBNodesStatus()            // mark all db nodes as `offline` except the `deleting` ones
				m.initDBAdbDevicesStatus()       // mark all db adb devices as `offline`
				m.initDBAdbOrderCallbackStatus() // mark all of ongoing adb order callback as `aborted`
				m.apiserver.SetLeader(true)      // then Api -> 200, agents will join on me
				scheduler.SetLeader(true)        // then scheduler knows the current role, it's background loops(guarders) will be enabled

				// launch background `watcher like` goroutines
				m.launchUserSessionsCleaner()
				m.launchAdbDeviceGuarder()
				m.launchAdbEventWatcher()
				log.Printf("master in serving now.")
			}
		}
	}
}

// set initial default settings if no global settings set
func (m *Master) initDBGlobalSettings() {
	// if previous settings not exists, db save initial default settings
	if _, err := store.DB().GetSettings(); store.DB().ErrNotFound(err) {
		err = scheduler.MemoSettings(types.GlobalDefaultSettings)
		if err != nil {
			log.Fatalln("setup db global default settings error:", err)
		}
	}

	// if previous paygate secret not exists, db save a new one
	curr, _ := store.DB().GetSettings()
	if curr.GlobalAttrs.Get(types.GlobalAttrPaygateSecretKey) == "" {
		err := scheduler.UpsertSettingsAttr(label.Labels{types.GlobalAttrPaygateSecretKey: "0" + utils.RandomString(30) + "x"})
		if err != nil {
			log.Fatalf("setup db global settings attr %s error: %v", types.GlobalAttrPaygateSecretKey, err)
		}
	}
}

// initDBNodesStatus mark all of db nodes as `offline` except deleting nodes
func (m *Master) initDBNodesStatus() {
	nodes, err := store.DB().ListNodes(nil)
	if err != nil {
		log.Fatalln("db ListNodes() error:", err)
	}
	for _, node := range nodes {
		// mark node as offline
		scheduler.MemoNodeStatus(node.ID, types.NodeStatusOffline, nil, nil)
		scheduler.MemoNodeErrmsg(node.ID, types.NodeWaittingFirstJoin)
	}
}

// initDBAdbDevicesStatus mark all of adb devices as `offline`
func (m *Master) initDBAdbDevicesStatus() {
	dvcs, err := store.DB().ListAdbDevices(nil, nil)
	if err != nil {
		log.Fatalln("db ListAdbDevices() error:", err)
	}
	for _, dvc := range dvcs {
		scheduler.MemoAdbDeviceStatus(dvc.ID, types.AdbDeviceStatusOffline, "waitting for the first adb device status collection")
	}
}

// initDBAdbOrderCallbackStatus mark all of ongoing adb order callback as `aborted`
// and re-sending order callback on those adb orders
func (m *Master) initDBAdbOrderCallbackStatus() {
	query := bson.M{"callback_status": types.AdbOrderCallbackStatusOngoing}
	orders, err := store.DB().ListAdbOrders(nil, query)
	if err != nil {
		log.Fatalln("db ListAdbOrders() ongoing error:", err)
	}

	for _, order := range orders {
		scheduler.AppendAdbOrderCallbackHistory(order.ID, "callback aborted while restart")
		scheduler.MemoAdbOrderCallbackStatus(order.ID, types.AdbOrderCallbackStatusAborted)
	}

	go scheduler.BootupReCallbackAbortedAdbOrders(orders)
}

// launch user sessions cleaner
func (m *Master) launchUserSessionsCleaner() {
	if !scheduler.IsRegisteredGoRoutine("user_sessions_cleaner", "system") {
		go scheduler.CleanExpiredUserSessionsLoop()
	}
}

// launch adb device perday limit quota checker
func (m *Master) launchAdbDeviceGuarder() {
	if !scheduler.IsRegisteredGoRoutine("adb_devices_limit_checker", "system") {
		go scheduler.RunAdbDeviceLimitCheckerLoop()
	}
}

// launch adb device event watcher
func (m *Master) launchAdbEventWatcher() {
	if !scheduler.IsRegisteredGoRoutine("adb_events_watcher", "system") {
		go scheduler.RunAdbEventWatcherLoop()
	}
}

func (m *Master) exitTrap() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	for range ch {
		os.Remove(m.unixSock)
		os.Remove(m.pidFile)
		os.Exit(0)
	}
}
