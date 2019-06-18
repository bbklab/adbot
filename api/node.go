package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

func (s *Server) listNodes(ctx *httpmux.Context) {
	var (
		nodeID     = ctx.Query["node_id"]
		status     = ctx.Query["status"]
		remote     = ctx.Query["remote"] // remote ip search
		hostname   = ctx.Query["hostname"]
		withMaster = ctx.Query["with_master"]
		labels     = ctx.Query["labels"] // key1=val1,key2=val2,key3=val3...
		query      = bson.M{}
	)

	// build query
	if nodeID != "" {
		query["id"] = nodeID
	}
	if status != "" {
		query["status"] = status
	}
	if remote != "" {
		query["remote_addr"] = bson.M{"$regex": bson.RegEx{Pattern: remote}}
	}
	if hostname != "" {
		query["sysinfo.hostname"] = bson.M{"$regex": bson.RegEx{Pattern: hostname}}
	}
	if withMaster != "" {
		withMasterV, _ := strconv.ParseBool(withMaster)
		query["sysinfo.with_master"] = withMasterV
	}
	if labels != "" {
		pairs := strings.Split(labels, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				query[fmt.Sprintf("labels.%s", kv[0])] = kv[1]
			}
		}
	}

	// filter nodes & sort
	nodes, err := store.DB().ListNodes(getPager(ctx), query)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.NodeWrapper, len(nodes))
	for idx, node := range nodes {
		wraps[idx] = s.wrapNode(node)
	}

	n := store.DB().CountNodes(query)
	ctx.Res.Header().Set("Total-Records", strconv.Itoa(n))
	ctx.JSON(200, wraps)
}

func (s *Server) getNode(ctx *httpmux.Context) {
	var (
		id = ctx.Path["node_id"]
	)

	node, err := store.DB().GetNode(id)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	go scheduler.RefreshNodeAsync(id)

	ctx.JSON(200, s.wrapNode(node))
}

func (s *Server) watchNodeEvents(ctx *httpmux.Context) {
	var (
		id   = ctx.Path["node_id"]
		node = scheduler.Node(id)
	)

	if node == nil {
		ctx.NotFound(fmt.Sprintf(scheduler.ErrMsgNoSuchNodeOnline, id))
		return
	}

	notifier, ok := ctx.Res.(http.CloseNotifier)
	if !ok {
		ctx.InternalServerError("not a http close notifier")
		return
	}

	flusher, ok := ctx.Res.(http.Flusher)
	if !ok {
		ctx.InternalServerError("not a http flusher")
		return
	}

	// obtain a node event subscriber
	evsub := node.Events()

	// must: evict the subscriber befor page exit
	go func() {
		<-notifier.CloseNotify()
		node.EvictEventSubscriber(evsub)
	}()

	// write response header firstly
	ctx.Res.WriteHeader(200)
	ctx.Res.Header().Set("Content-Type", "text/event-stream")
	ctx.Res.Header().Set("Cache-Control", "no-cache")
	ctx.Res.Write(nil)
	flusher.Flush()

	// write node event to the client with sse format
	for ev := range evsub {
		ctx.Res.Write(ev.(*mole.NodeEvent).Format())
		flusher.Flush()
	}
}

func (s *Server) watchNodeStats(ctx *httpmux.Context) {
	var (
		id = ctx.Path["node_id"]
	)

	notifier, ok := ctx.Res.(http.CloseNotifier)
	if !ok {
		ctx.InternalServerError("not a http close notifier")
		return
	}

	flusher, ok := ctx.Res.(http.Flusher)
	if !ok {
		ctx.InternalServerError("not a http flusher")
		return
	}

	// obtain a node live stream stats subscriber
	stream, err := scheduler.NodeStats(id)
	if err != nil {
		ctx.AutoError(err)
		return
	}
	defer stream.Close()

	// must: close the stream befor page exit
	// to prevent fd & goroutine leaks
	go func() {
		<-notifier.CloseNotify()
		stream.Close()
	}()

	// write response header firstly
	ctx.Res.WriteHeader(200)
	ctx.Res.Header().Set("Content-Type", "text/event-stream")
	ctx.Res.Header().Set("Cache-Control", "no-cache")
	ctx.Res.Write(nil)
	flusher.Flush()

	// decode & rewrite & flush every stats entry
	// FIXME try directly io.Copy proxy the stream with Hijack and prevent fd leaks
	var dec = json.NewDecoder(stream)
	for {
		stats := new(*types.SysInfo)
		err := dec.Decode(&stats)
		if err != nil {
			break
		}
		bs, _ := json.Marshal(stats)
		ctx.Res.Write(append(bs, '\r', '\n', '\r', '\n'))
		flusher.Flush()
	}
}

func (s *Server) runNodeCmd(ctx *httpmux.Context) {
	var (
		id   = ctx.Path["node_id"]
		node = scheduler.Node(id)
	)

	if node == nil {
		ctx.NotFound(fmt.Sprintf(scheduler.ErrMsgNoSuchNodeOnline, id))
		return
	}

	var nodeCmd = new(types.NodeCmd)
	if err := ctx.Bind(nodeCmd); err != nil {
		ctx.BadRequest(err)
		return
	}

	// hijack to obtain the client's underlying conn
	hj := ctx.Res.(http.Hijacker)
	clientConn, _, err := hj.Hijack()
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer clientConn.Close() // must

	// request on remote node
	stream, err := scheduler.DoNodeExec(id, nodeCmd)
	if err != nil {
		clientConn.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return
	}
	defer stream.Close()

	// write response header firstly
	clientConn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	clientConn.Write([]byte("Content-Type: text/event-stream\r\n"))
	clientConn.Write([]byte("Cache-Control: no-cache\r\n"))
	clientConn.Write([]byte("\r\n"))

	// redirect node response stream to client
	go func() {
		io.Copy(clientConn, clientConn) // triggered if clientConn is closed, then we close the resp.Body
		stream.Close()                  // so the following io.Copy quit
	}()
	io.Copy(clientConn, stream) // the resp.Body may never got EOF, so we have to cares how to close the resp.Body
}

// just close the control connection of cluster node, if the node
// implemented rejoin logic, this just make the node rejoin again.
func (s *Server) closeNode(ctx *httpmux.Context) {
	var (
		id   = ctx.Path["node_id"]
		node = scheduler.Node(id)
	)

	if node == nil {
		ctx.NotFound(fmt.Sprintf(scheduler.ErrMsgNoSuchNodeOnline, id))
		return
	}

	scheduler.CloseNode(id)

	ctx.Status(204)
}

//
// utils
//

func (s *Server) wrapNode(node *types.Node) *types.NodeWrapper {
	if !unmaskSensitive {
		node.Hidden()
	}
	return &types.NodeWrapper{
		Node:     node,
		RemoteIP: node.RemoteIP(),
		HwInfo:   node.HardwareInfo(),
	}
}
