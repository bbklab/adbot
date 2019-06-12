package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/websocket"

	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/pkg/ws"
	"github.com/bbklab/adbot/scheduler"
)

var (
	sio *socketio.Server
)

func init() {
	var err error
	sio, err = newSocketIOServer()
	if err != nil {
		log.Fatalln("initialize socket io server error:", err)
	}
}

// node terminal via socket.io implemention
func (s *Server) openNodeTerminalNG(ctx *httpmux.Context) {
	sio.ServeHTTP(ctx.Res, ctx.Req)
}

func newSocketIOServer() (*socketio.Server, error) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	// register socket io server event handlers
	server.On("connection", func(so socketio.Socket) {
		var (
			path = strings.TrimSuffix(so.Request().URL.Path, "/")
			id   = strings.TrimSuffix(strings.TrimPrefix(path, "/api/nodes/"), "/terminal_ng")
			node = scheduler.Node(id)
		)

		if node == nil {
			so.Emit("error", fmt.Sprintf("no such node %s online", id))
			so.Emit("disconnection", "1")
			return
		}

		nodeConn, err := node.Dial("", "")
		if err != nil {
			so.Emit("error", fmt.Sprintf("revert dial node %s error: %v", id, err))
			so.Emit("disconnection", "1")
			return
		}
		if tcpConn, ok := nodeConn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}

		nt := &nodeTerminal{nid: id, wid: utils.RandomString(16), conn: nodeConn}
		so.On("auth", nt.onAuth)
		so.On("data", nt.onInput)
		so.On("resize", nt.resize)
		so.On("disconnection", nt.close)
	})

	return server, nil
}

type nodeTerminal struct {
	wid  string   // window id, uniq, for resizing later
	nid  string   // node id
	conn net.Conn // node conn
}

// name return the identity of current node terminal
func (nt *nodeTerminal) name() string {
	return fmt.Sprintf("terminal (node:%s, window:%s)", nt.nid, nt.wid)
}

// send http request to remote node http api endpoint
func (nt *nodeTerminal) onAuth(so socketio.Socket, msg string) {
	args := strings.Split(msg, ",")
	if len(args) != 3 {
		so.Emit("error", "invalid credentials")
		so.Emit("disconnection", "1")
		return
	}

	// send http request call to node terminal api, so the connection could get ready
	notifyCh := make(chan struct{})
	go func() {
		nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/terminal?wid=%s", nt.nid, nt.wid), nil)
		err := nodeReq.Write(nt.conn)
		if err != nil {
			so.Emit("error", fmt.Sprintf("revert call node %s terminal error: %v", nt.nid, err))
			so.Emit("disconnection", "1")
		}

		// wait node terminal window fd exists, so the first resizing take effective
		err = scheduler.DoWaitNodeTerminalWindow(nt.nid, nt.wid, time.Second*5)
		if err != nil {
			so.Emit("error", fmt.Sprintf("wait node %s terminal window fd error: %v", nt.nid, err))
			so.Emit("disconnection", "1")
		}

		// go on
		close(notifyCh)
	}()

	// wait connection got ready
	<-notifyCh

	// initialize resize the terminal window
	nt.resize(so, fmt.Sprintf("%s,%s", args[1], args[2]))

	// copying from node connection and write to client
	go func() {
		var (
			data = make([]byte, 512)
			bs   []byte
		)
		for {
			n, err := nt.conn.Read(data)
			if n > 0 {
				bs = data[:n]
				if err := so.Emit("data", string(bs)); err != nil {
					log.Warnf("%s: write to client error: %v", nt.name(), err) // sometimes start up blockied and got error: upgrading
				}
			}
			if err != nil {
				log.Errorf("%s: read from node conn error: %v", nt.name(), err)
				nt.close()
				return
			}
		}
	}()
}

// copying from client input and write to node connection
func (nt *nodeTerminal) onInput(msg string) {
	total := len(msg)
	tried := 0
	for total > 0 {
		if nt.conn == nil {
			log.Warnf("%s: remote node connection not establisthed", nt.name())
			time.Sleep(time.Millisecond * 50)
			tried++
			if tried > 30 {
				log.Errorf("%s: remote node connection not established", nt.name())
				nt.close()
				return
			}
			continue
		}

		n, err := nt.conn.Write([]byte(msg))
		if err != nil {
			nt.close()
			return
		}
		total -= n
	}
}

func (nt *nodeTerminal) resize(so socketio.Socket, msg string) {
	var (
		width, height int
		args          = strings.Split(msg, ",")
	)

	if len(args) != 2 {
		so.Emit("error", fmt.Sprintf("invalid resizing parameters: %s", msg))
		so.Emit("disconnection", "1")
		return
	}

	width, err := strconv.Atoi(args[0])
	if err == nil {
		height, err = strconv.Atoi(args[1])
	}
	if err != nil {
		so.Emit("error", fmt.Sprintf("invalid resizing parameters: %s", msg))
		so.Emit("disconnection", "1")
		return
	}

	// call node api to resizing
	err = scheduler.DoNodeTerminalResizing(nt.nid, nt.wid, width, height)
	if err != nil {
		log.Errorf("%s: resizing window error: %v", nt.name(), err)
		return
	}

	log.Infof("%s: resized window width: %d height: %d", nt.name(), width, height)
}

func (nt *nodeTerminal) close() {
	log.Warnf("%s: closed", nt.name())
	nt.conn.Close()
}

// Legacy only for cli node terminal
// node terminal via simple websocket implemention
func (s *Server) openNodeTerminal(ctx *httpmux.Context) {
	var (
		id       = ctx.Path["node_id"]
		node     = scheduler.Node(id)
		upgrader = websocket.Upgrader{}
	)

	if node == nil {
		ctx.NotFound(fmt.Sprintf(scheduler.ErrMsgNoSuchNodeOnline, id))
		return
	}

	// obtain ws connection of client
	wsConn, err := upgrader.Upgrade(ctx.Res, ctx.Req, nil)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer wsConn.Close() // must

	var (
		wsConnWrapper = ws.NewWrappedWsConn(wsConn, websocket.TextMessage)
	)

	// dial to the remote node
	nodeConn, err := node.Dial("", "")
	if err != nil {
		wsConnWrapper.Write([]byte("Dial Node error: " + err.Error()))
		return
	}
	if tcpConn, ok := nodeConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	defer nodeConn.Close()

	// send http request onto remote node http api endpoint
	nodeReq, _ := http.NewRequest("GET", fmt.Sprintf("http://%s/api/terminal?wid=%s", id, utils.RandomString(16)), nil)
	err = nodeReq.Write(nodeConn)
	if err != nil {
		wsConnWrapper.Write([]byte("Internal server error: " + err.Error()))
		return
	}

	// io.Copy between node terminal <--> client ws conn
	go func() {
		io.Copy(nodeConn, wsConnWrapper)
	}()

	io.Copy(wsConnWrapper, nodeConn)
}
