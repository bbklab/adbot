package mole

import (
	"encoding/json"
	"fmt"
	"time"
)

var (
	// NodeEvJoin represent node event: join
	NodeEvJoin = "join"
	// NodeEvHeartbeat represent node event: heartbeat
	NodeEvHeartbeat = "heartbeat"
	// NodeEvNewWorker represent node event: new
	NodeEvNewWorker = "new"
	// NodeEvClose represent node event: close
	NodeEvClose = "close"
	// NodeEvShutdown represent node event: shutdown
	NodeEvShutdown = "shutdown"
	// NodeEvFlagging represent node event: flagging
	NodeEvFlagging = "flagging" // at least three heartbeats lost
	// NodeEvDie represent node event: die
	NodeEvDie = "die" // all TTL(five) heartbeats lost
	// NodeEvRecovery represent node event: recovery
	NodeEvRecovery = "recovery" // dead -> come back
	// NodeEvRejoin represent node event: rejoin
	NodeEvRejoin = "rejoin" // rejoin
)

func newNodeEvent(id, typ string) *NodeEvent {
	return &NodeEvent{
		ID:   id,
		Type: typ,
		Time: time.Now(), // note: abnormal timezone if running under a container without proper timezone settings
	}
}

// NodeEvent represents a node event
type NodeEvent struct {
	ID   string    `json:"id"`
	Type string    `json:"type"`
	Time time.Time `json:"time"`
}

// Format format node events to SSE text
func (ev *NodeEvent) Format() []byte {
	bs, _ := json.Marshal(ev)
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", ev.Type, string(bs)))
}
