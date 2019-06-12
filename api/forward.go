package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/httpmux"
)

func (s *Server) forwardToLeaderHandle(leaderAddr string, ctx *httpmux.Context) {
	// dial leader
	dst, err := net.DialTimeout("tcp", leaderAddr, time.Second*10)
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer dst.Close()

	// note: rewrite request to disable default persistent connection in HTTP/1.1
	// as we won't expect the Browser Chrome/FF to reuse the forwarded connection
	// while we're visiting the non-forward handlers for debugging
	// typically put this header in server response to disable persistent connection
	// but it's difficult to inject the header to response during io.Copy
	// so we inject it to the client request side Headers which have the same effect
	ctx.Req.Header.Set("Connection", "close")

	// err = ctx.Req.WriteProxy(dst) // send original request
	err = ctx.Req.Write(dst) // send original request
	if err != nil {
		ctx.InternalServerError(err)
		return
	}

	log.Printf("Forwarding Request %s %s %s --> %s", ctx.Req.RemoteAddr, ctx.Req.Method, ctx.Req.URL, leaderAddr)

	// obtain the client underlying net.Conn
	src, _, err := ctx.Res.(http.Hijacker).Hijack()
	if err != nil {
		ctx.InternalServerError(err)
		return
	}
	defer src.Close()

	// io copy between src & dst
	errc := make(chan error, 2)
	cp := func(w io.WriteCloser, r io.Reader) {
		defer w.Close()
		_, err := io.Copy(w, r)
		errc <- err
	}

	go cp(dst, src)
	cp(src, dst) // note: hanging wait while copying the response

	err = <-errc
	if err != nil && err != io.EOF {
		err = fmt.Errorf("io copy error: %v", err)
		src.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n" + err.Error() + "\r\n"))
		return
	}
}
