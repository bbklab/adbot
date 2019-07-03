package api

import (
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/unixsock"
	"github.com/bbklab/adbot/types"
)

const (
	// APIPREFIX define the http api prefix
	APIPREFIX = "/api"
)

// Server is the runtime http api server
type Server struct {
	ls      [2]net.Listener // plain listeners:  http/unix
	tlsl    net.Listener    // tls listener: https
	startAt time.Time
	mux     *httpmux.Mux        // httpmux.Mux reference
	cfg     *types.MasterConfig // typically only used for debuging dump

	classifyOnce sync.Once
	classified   map[string][]httpmux.HandleFunc // classified http handlers

	sync.RWMutex      // protect leader flag
	leader       bool // if elected as leader
	ready        bool // if s.Run() ready
}

// NewServer is exported
func NewServer(unixSock string, httpListener, httpsListener net.Listener, cfg *types.MasterConfig) *Server {
	// create/overwrite unix socket (ignore conflicts, as we've verify conflicts in pidfile previous)
	unixListener, err := unixsock.New(unixSock)
	if err != nil {
		log.Fatalln(err)
	}

	return &Server{
		ls:         [2]net.Listener{unixListener, httpListener},
		tlsl:       httpsListener,
		cfg:        cfg,
		classified: make(map[string][]httpmux.HandleFunc),
		leader:     false,
		ready:      false,
	}
}

// Run is exported
func (s *Server) Run() error {
	s.startAt = time.Now()

	mux := httpmux.New(APIPREFIX)
	s.mux = mux // save mux reference

	// enable this midware after implement a real HA Campaigner
	// note: we must set this midware at first, as we expect to redirect all traffics
	// to current leader if we stand by, so all the rest midwares and handlers won't take effect
	// set leadership midware to verify we're the leader
	// s.mux.SetGlobalPreMidware(s.checkLeaderShipMW)

	// check licenses
	s.mux.SetGlobalPreMidware(s.checkLicenseMW)

	// set cors http headers
	s.mux.SetGlobalPreMidware(s.corsMW)
	s.mux.SetGlobalPreMidware(s.bypassMW)

	// audit every request afterwards
	// note: here we use SetAuditLog instead of SetGlobalPostMidware to install our log midware
	// to ensure this midware can't be ctx.Abort() by any PreMidwares or any other http handlers
	// s.mux.SetGlobalPostMidware(s.auditMW) // deprecated: could be aborted by previous MWs or handlers
	s.mux.SetAuditLog(s.auditMW)

	// set api auth midware
	s.mux.SetGlobalPreMidware(s.checkAuthLoginMW)

	// apply global db settings
	if err := s.applyRuntimeSettings(); err != nil {
		return err
	}

	// setup our routes
	s.setupRoutes(mux)

	// classify all handlers
	s.classify()

	// setup httpd & serve apis on all listeners
	server := http.Server{
		Handler: mux,
	}

	// listen on plain http / unix
	errCh := make(chan error, 1)
	for _, l := range s.ls {
		go func(l net.Listener) {
			errCh <- server.Serve(l)
		}(l)
	}

	// listen on https
	if s.tlsl != nil {
		go func() {
			errCh <- server.ServeTLS(s.tlsl, s.cfg.TLSCert, s.cfg.TLSKey)
		}()
	}

	s.setReady()

	return <-errCh
}

// SetLeader mark current leader flag
func (s *Server) SetLeader(flag bool) {
	s.Lock()
	s.leader = flag
	s.Unlock()

	// apply runtime settings on each leadership changing
	s.applyRuntimeSettings()
}

func (s *Server) isLeader() bool {
	s.RLock()
	defer s.RUnlock()
	return s.leader
}

// IsReady return if the apiserver.Run() ready
func (s *Server) IsReady() bool {
	s.RLock()
	defer s.RUnlock()
	return s.ready
}

func (s *Server) setReady() {
	s.Lock()
	s.ready = true
	s.Unlock()
}
