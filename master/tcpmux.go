package master

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	errProtocolNotServing = []byte("protocol not serving by tcpmuxer")
)

type tcpMux struct {
	listen string
	flag   flag // identify which type of protocols tcpmuxer is serving

	poolHTTP  chan net.Conn
	poolMole  chan net.Conn
	poolHTTPS chan net.Conn
}

func newTCPMux(l string) *tcpMux {
	return &tcpMux{
		listen:    l,
		poolHTTP:  make(chan net.Conn, 1),
		poolMole:  make(chan net.Conn, 1),
		poolHTTPS: make(chan net.Conn, 1),
	}
}

func (m *tcpMux) SetServeFlag(flag flag) {
	m.flag = flag
}

func (m *tcpMux) ListenAndServe() error {
	l, err := net.Listen("tcp", m.listen)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorln("tcpMux Accept() error: %v", err)
			return err
		}

		go m.dispatch(conn)
	}
}

func (m *tcpMux) dispatch(conn net.Conn) {
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(time.Second * 30)
	}

	var (
		remote     = conn.RemoteAddr().String()
		header     = make([]byte, 4)
		headerCopy = bytes.NewBuffer(nil) // buffer to hold another copy of header
	)
	_, err := io.ReadFull(io.TeeReader(conn, headerCopy), header)
	if err != nil {
		log.Errorln("failed to read protocol header:", remote, err)
		return
	}

	bc := &bufConn{Conn: conn, reader: io.MultiReader(headerCopy, conn)}

	// dispatch to mole connection pool
	if bytes.Equal(header, []byte(`MOLE`)) {
		if !m.flag.serveMole() {
			goto NOTSERVING
		}
		m.poolMole <- bc
		return
	}

	// dispatch to https connection pool
	if header[0] == 22 {
		if !m.flag.serveHTTPS() {
			goto NOTSERVING
		}
		m.poolHTTPS <- bc
		return
	}

	// dispatch to http connection pool
	if !m.flag.serveHTTP() {
		goto NOTSERVING
	}
	m.poolHTTP <- bc
	return

NOTSERVING:
	conn.Write(errProtocolNotServing)
	conn.Close()
}

func (m *tcpMux) NewHTTPListener() net.Listener {
	return &muxListener{
		l:    m.listen,
		pool: m.poolHTTP,
	}
}

func (m *tcpMux) NewMoleListener() net.Listener {
	return &muxListener{
		l:    m.listen,
		pool: m.poolMole,
	}
}

func (m *tcpMux) NewHTTPSListener() net.Listener {
	return &muxListener{
		l:    m.listen,
		pool: m.poolHTTPS,
	}
}

// implement net.Listener interface
// the caller could process the cached worker connection in the pool via `muxListener`
type muxListener struct {
	l          string        // pass from tcpMux
	sync.Mutex               // protect flag closed
	closed     bool          // flag on pool closed
	pool       chan net.Conn // connection pool
}

func (l *muxListener) Accept() (net.Conn, error) {
	conn, ok := <-l.pool
	if !ok {
		return nil, errors.New("listener closed")
	}
	return conn, nil
}

func (l *muxListener) Close() error {
	l.Lock()
	if !l.closed {
		l.closed = true
		close(l.pool) // so the Accept() returned immediately
	}
	l.Unlock()
	return nil
}

func (l *muxListener) Addr() net.Addr {
	return l
}

func (l *muxListener) Network() string {
	return "tcp"
}

func (l *muxListener) String() string {
	return l.l
}

// implement net.Conn with customized Read() method
type bufConn struct {
	net.Conn
	reader io.Reader
}

func (bc *bufConn) Read(bs []byte) (int, error) {
	return bc.reader.Read(bs)
}

// serve flag
//
type flag uint32

const (
	flagHTTP  flag = 1 << iota // h: serving plain http
	flagHTTPS                  // t: serving tls http
	flagMole                   // m: serving mole
)

func (f flag) serveHTTP() bool {
	return f&flagHTTP != 0
}

func (f flag) serveHTTPS() bool {
	return f&flagHTTPS != 0
}

func (f flag) serveMole() bool {
	return f&flagMole != 0
}
