package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/mole"
	"github.com/bbklab/adbot/scheduler"
	"github.com/bbklab/adbot/store"
	"github.com/bbklab/adbot/types"
)

// TODO
func (s *Server) listNodes(ctx *httpmux.Context) {
	// obtain label filter
	var (
		lbsQuery  = ctx.Query["labels"] // key1=val1,key2=val2,key3=val3...
		lbsFilter = label.New(nil)
	)
	if lbsQuery != "" {
		for _, pair := range strings.Split(lbsQuery, ",") {
			if strings.TrimSpace(pair) == "" {
				continue
			}

			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				ctx.BadRequest(fmt.Sprintf("[%s] is not valid label kv format", pair))
				return
			}

			key, val := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			lbsFilter.Set(key, val)
		}
	}

	// obtain online filter
	//   - empty      : query all
	//   - true/false : query specified online/offline nodes
	var (
		online     = ctx.Query["online"]
		onlineFlag *bool // query all
	)
	if online != "" {
		flag, _ := strconv.ParseBool(online)
		onlineFlag = &flag
	}

	// filter nodes & sort
	nodes, err := scheduler.FilterNodes(lbsFilter, onlineFlag, false)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	wraps := make([]*types.NodeWrapper, len(nodes))
	for idx, node := range nodes {
		wraps[idx] = s.wrapNode(node)
	}

	ctx.Res.Header().Set("Total-Records", strconv.Itoa(len(wraps)))
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
