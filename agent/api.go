package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/kr/pty"

	"github.com/bbklab/adbot/agent/extensions"
	"github.com/bbklab/adbot/pkg/httpmux"
	"github.com/bbklab/adbot/types"
	"github.com/bbklab/adbot/version"
)

func (agent *Agent) ping(ctx *httpmux.Context) {
	ctx.Res.Write([]byte{'O', 'K'})
}

func (agent *Agent) version(ctx *httpmux.Context) {
	ctx.JSON(200, version.Version())
}

// collect node sysinfo
//
//
func (agent *Agent) sysinfo(ctx *httpmux.Context) {
	info, err := extensions.GatherSysInfo()
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	ctx.JSON(200, info)
}

func (agent *Agent) stats(ctx *httpmux.Context) {
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

	stopCh := make(chan struct{})

	// must: evict the subscriber befor page exit
	go func() {
		<-notifier.CloseNotify()
		close(stopCh)
	}()

	// write response header firstly
	ctx.Res.WriteHeader(200)
	ctx.Res.Header().Set("Content-Type", "text/event-stream")
	ctx.Res.Header().Set("Cache-Control", "no-cache")
	ctx.Res.Write(nil)
	flusher.Flush()

	for {
		select {

		case <-stopCh:
			return

		default:
			info, err := extensions.GatherSysInfo()
			if err != nil {
				log.Warnln("GatherSysInfo() error:", err)
				continue
			}
			bs, _ := json.Marshal(info)
			ctx.Res.Write(append(bs, '\r', '\n', '\r', '\n'))
			flusher.Flush()
			time.Sleep(time.Second)
		}
	}
}

// run node cmd
//
//
func (agent *Agent) runCmd(ctx *httpmux.Context) {
	var (
		nodeCmd = new(types.NodeCmd)
	)

	if err := ctx.Bind(nodeCmd); err != nil {
		ctx.BadRequest(err)
		return
	}

	// hijack to obtain the underlying net.Conn
	hj := ctx.Res.(http.Hijacker)
	conn, _, err := hj.Hijack()
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer conn.Close() // must

	// run command with progress reader
	stopCh := make(chan struct{})
	defer close(stopCh) // ensure the command will be stopped
	stream, err := extensions.RunCmd(nodeCmd.Command, stopCh)
	if err != nil {
		conn.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return
	}
	defer stream.Close()

	// write response header firstly
	conn.Write([]byte("HTTP/1.1 200 OK\r\n"))
	conn.Write([]byte("Content-Type: text/event-stream\r\n"))
	conn.Write([]byte("Cache-Control: no-cache\r\n"))
	conn.Write([]byte("\r\n"))

	// redirect progress output to client
	go func() {
		io.Copy(conn, conn) // triggered if conn is closed, so we close the stream
		stream.Close()      // so the following io.Copy quit
	}()
	io.Copy(conn, stream) // the stream may never got EOF, so we have to cares how to close the stream
}

func (agent *Agent) terminal(ctx *httpmux.Context) {
	var (
		id = ctx.Query["wid"]
	)

	if id == "" {
		ctx.BadRequest("terminal id can't be empty")
		return
	}

	log.Infof("node terminal %s start", id)
	defer log.Infof("node terminal %s end", id)

	var (
		cmd = exec.Command("env", "TERM=xterm", "/bin/sh", "-l")
	)

	fd, err := pty.Start(cmd)
	if err != nil {
		ctx.AutoError(err)
		return
	}
	defer cmd.Wait() // to prevent <defunct> process left

	termFDs.add(id, fd)
	defer termFDs.del(id)

	// hijack to obtain the underlying net.Conn
	hj := ctx.Res.(http.Hijacker)
	conn, _, err := hj.Hijack()
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer conn.Close() // must

	// io.Copy on both sides
	go func() {
		io.Copy(conn, fd) // NOTE: *os.File.Read() can't be interrupted by os.File.Close()
		conn.Close()
	}()
	io.Copy(fd, conn)

	// NOTE: *os.File.Read() can't be interrupted by os.File.Close()
	// so we use this workaround way to terminated the pty.Start(cmd) to prevent goroutine leaks
	// See: https://github.com/golang/go/issues/20110
	fd.Write([]byte("exit\n")) // tell the cmd `sh -l` to quit
}

func (agent *Agent) terminalQuery(ctx *httpmux.Context) {
	var (
		id = ctx.Query["wid"]
	)

	if termFDs.exists(id) {
		ctx.Status(200)
		return
	}

	ctx.Status(404)
}

func (agent *Agent) terminalResize(ctx *httpmux.Context) {
	var (
		id     = ctx.Query["wid"]
		width  = ctx.Query["width"]
		height = ctx.Query["height"]
	)

	widthN, _ := strconv.Atoi(width)
	heightN, _ := strconv.Atoi(height)

	if err := termFDs.resize(id, widthN, heightN); err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.Status(200)
}

//
// adb bot
//

func (agent *Agent) listAdbDevices(ctx *httpmux.Context) {
	devices, err := extensions.ListAdbDevices()
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, devices)
}

func (agent *Agent) checkAdbAlipayOrder(ctx *httpmux.Context) {
	var (
		dvcID   = ctx.Query["device_id"]
		orderID = ctx.Query["order_id"]
	)

	order, err := extensions.CheckAdbAlipayOrder(dvcID, orderID)
	if err != nil {
		ctx.AutoError(err)
		return
	}

	ctx.JSON(200, order)
}
