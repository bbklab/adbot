package agent

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/agent/extensions"
	"github.com/bbklab/adbot/client"
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/types"
)

const (
	// APIPREFIX define the http api prefix
	APIPREFIX = "/api"
	// FILEUUID define the local uuid file
	FILEUUID = "/etc/.adbot.uuid"
)

var (
	once sync.Once // run once on first succeed join
)

// Agent is a runtime agent ...
type Agent struct {
	config *types.AgentConfig

	clusterNode *mole.Agent   // runtime mole agent, reset on each Join() (note: be aware of the persistent conn leaks)
	client      client.Client // adbot api server client, reset on each Join() (note: be aware of the http.Transport leaks)
}

// New new an Agent
func New(cfg *types.AgentConfig) *Agent {
	return &Agent{
		config: cfg,
	}
}

// Run serving protocol & Api with underlying mole
func (agent *Agent) Run() error {
	var (
		delayMin = time.Second      // min retry delay 1s
		delayMax = time.Second * 60 // max retry delay 60s
		delay    = delayMin         // retry delay
	)
	for {
		err := agent.Join()
		if err != nil {
			log.Errorln("agent Join() error:", err)
			delay *= 2
			if delay > delayMax {
				delay = delayMax // reset delay to max
			}
			log.Warnln("agent ReJoin in", delay.String())
			time.Sleep(delay)
			continue
		}

		l := agent.newListener()

		go func(l net.Listener) {
			err := agent.serveProtocol()
			if err != nil {
				log.Errorln("agent serveProtocol() error:", err)
				l.Close() // close the listener -> the serveAPI() return with error -> Rejoin triggered.
			}
		}(l)

		log.Println("agent Joined succeed, ready ...")
		delay = delayMin // reset dealy to min
		err = agent.serveAPI(l)
		if err != nil {
			log.Errorln("agent serveAPI() error:", err)
		}

		log.Warnln("agent Rejoin ...")
		time.Sleep(time.Second)
	}
}

// Join join current agent to master
// reset runtime mole agent (note: be aware of the persistent conn leaks)
// reset adbot api server client (note: be aware of the http.Transport leaks)
func (agent *Agent) Join() error {
	// load or save an agent id
	id, err := getAgentID()
	if err != nil {
		log.Fatalf("can't obtain agent id: %v", err)
	}

	// setup client (the smart client would automatic detect the healthy leader address)
	if agent.client == nil {
		agent.client, err = client.New(agent.config.JoinAddrs)
	} else {
		err = agent.client.Reset()
	}
	if err != nil {
		return err
	}
	log.Infof("talking to the healthy master %s", agent.client.Peer())

	// query if self allow to join
	if err := agent.isJoinReady(id); err != nil {
		return fmt.Errorf("agent %s can't join: %v", id, err)
	}

	// setup mole agent & join
	agent.clusterNode = mole.NewAgent(id, agent.client.PeerAddr())
	if err = agent.clusterNode.Join(); err != nil {
		return err
	}

	// renew pkg `extension` adbot api client immediately
	extensions.RenewMasterAPIClient(agent.client, id)

	// joined
	once.Do(func() {
	})

	// call GC() once
	runtime.GC()

	return nil
}

func (agent *Agent) isJoinReady(id string) error {
	return agent.client.NodeJoinCheck(id)
}

func getAgentID() (string, error) {
	_, err := os.Stat(FILEUUID)
	// initilization, save given id or new generated id
	if os.IsNotExist(err) {
		id := os.Getenv("ADBOT_AGENT_ID")
		if id == "" {
			id = utils.RandomString(16)
			log.Warnf("without any given agent id from env, will using a random agent id: %s", id)
		}
		err = ioutil.WriteFile(FILEUUID, []byte(id), os.FileMode(0400))
		return id, err
	}

	// load previous saved id
	bs, err := ioutil.ReadFile(FILEUUID)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(bs)), nil
}

func (agent *Agent) newListener() net.Listener {
	return agent.clusterNode.NewListener()
}

func (agent *Agent) serveProtocol() error {
	log.Println("agent protocol in serving ...")
	return agent.clusterNode.ServeProtocol()
}

func (agent *Agent) serveAPI(l net.Listener) error {
	log.Println("agent api in serving ...")

	mux := httpmux.New(APIPREFIX)
	agent.setupRoutes(mux)

	httpd := &http.Server{
		Handler: mux,
	}
	return httpd.Serve(l)
}
