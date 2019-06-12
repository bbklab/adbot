package scheduler

import (
	"sync"

	"github.com/bbklab/adbot/types"
)

// nolint
var (
	TypNode = "node"
)

// mainly used for metrics exporting
type cache struct {
	m map[string]map[string]interface{} //  resource type -> id -> value
	sync.RWMutex
}

func newCache() *cache {
	return &cache{
		m: make(map[string]map[string]interface{}),
	}
}

func (c *cache) set(typ, id string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.m[typ]; !ok {
		c.m[typ] = make(map[string]interface{})
	}
	c.m[typ][id] = value
}

func (c *cache) del(typ, id string) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.m[typ]; !ok {
		return
	}
	delete(c.m[typ], id)
}

// SetNodeCache update cached node
func SetNodeCache(node *types.Node) {
	if node == nil || node.ID == "" {
		return
	}
	sched.cache.set(TypNode, node.ID, node)
}

// DelNodeCache remove cached node
func DelNodeCache(nid string) {
	if nid == "" {
		return
	}
	sched.cache.del(TypNode, nid)
}
