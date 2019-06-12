package extensions

import (
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/client"
)

var (
	mcmux sync.RWMutex  // protect following
	mc    client.Client // package global reference on leader master api client (renew on each agent.Join)

	once sync.Once
	nid  string // package global reference on adbot agent local id
)

// RenewMasterAPIClient renew leader master api client
func RenewMasterAPIClient(c client.Client, currentNid string) {
	if c == nil {
		log.Fatalln("master api client can't be nil")
	}

	log.Printf("renew master api client to %s", c.Peer())
	mcmux.Lock()
	mc = c
	mcmux.Unlock()

	once.Do(func() {
		nid = currentNid
	})
}

// GetMasterAPIClient is exported
func GetMasterAPIClient() client.Client {
	mcmux.RLock()
	client := mc
	mcmux.RUnlock()
	return client
}

// GetNodeID is exported
func GetNodeID() string {
	return nid
}
