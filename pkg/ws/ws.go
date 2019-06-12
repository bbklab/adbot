package ws

import "github.com/gorilla/websocket"

// WrappedWsConn is an implement for io.Reader and io.Writer
// so the caller could use the ws conn as general io.Reader / io.Writer
type WrappedWsConn struct {
	*websocket.Conn
	MessageType int
}

// NewWrappedWsConn create a WrappedWsConn
func NewWrappedWsConn(conn *websocket.Conn, msgType int) *WrappedWsConn {
	if conn == nil {
		panic("nil websocket conn")
	}
	return &WrappedWsConn{conn, msgType}
}

// Read implement io.Reader
func (conn *WrappedWsConn) Read(p []byte) (int, error) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return -1, err
	}

	copy(p, msg)
	return len(msg), nil
}

// Write implement io.Writer
func (conn *WrappedWsConn) Write(p []byte) (int, error) {
	return len(p), conn.WriteMessage(conn.MessageType, p)
}
