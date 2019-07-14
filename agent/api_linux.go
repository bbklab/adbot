package agent

import (
	"io"
	"net/http"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/kr/pty"

	"github.com/bbklab/adbot/pkg/httpmux"
)

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
