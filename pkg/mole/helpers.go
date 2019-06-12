package mole

import (
	"io"
	"net"
	"time"

	"github.com/felixge/tcpkeepalive"
)

type connTestReader struct {
	c net.Conn
}

func newConnTestReader(c net.Conn) io.Reader {
	return &connTestReader{c: c}
}

// implement io.Reader interface
// test the connection before each read
func (r *connTestReader) Read(p []byte) (int, error) {
	if err := testConn(r.c); err != nil {
		return -1, err
	}
	return r.c.Read(p)
}

// note: the given net.Conn shouldn't be read simultaneously by another reader
// because the SetReadDeadline() may have effect on any currently-blocked read calls
func testConn(c net.Conn) error {
	c.SetReadDeadline(time.Now().Add(time.Second * 10))
	defer c.SetReadDeadline(time.Time{})

	var empty = make([]byte, 0)
	_, err := c.Read(empty)
	if err != nil {
		return err
	}

	return nil
}

// the builtin net.Conn doesn't provide full access to keepalive options
// on default, the TCP peer need to wait for a long time to detect the broken connection
// here we call setKeepAlive to shorten the time to `idle` + `count` * `intv`
// FIXME
// not effective as expected
// actually always wait for 15 minutes while the peer host lose power or unplug network cable
func setKeepAlive(c net.Conn, opt *kaopt) error {
	return tcpkeepalive.SetKeepAlive(c, opt.idle, opt.count, opt.intv)
}

// Linux OS default tcp keepalive parameters
//  - /proc/sys/net/ipv4/tcp_keepalive_time
//  - /proc/sys/net/ipv4/tcp_keepalive_probes
//  - /proc/sys/net/ipv4/tcp_keepalive_intvl
type kaopt struct {
	idle  time.Duration // tcp_keepalive_time
	count int           // tcp_keepalive_probes
	intv  time.Duration // tcp_keepalive_intvl
}
