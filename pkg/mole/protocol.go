package mole

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"net"
)

var (
	// HEADER define the protocl header flag
	HEADER = []byte("MOLE")
)

var (
	// agent -> master (new connection)
	cmdJoin = "join"

	// agent ->  master (reuse persistent connection)
	cmdLeave = "leave"

	// agent -> master (reuse persistent connection)
	// agent tell master it's still alive, master could use this to implement a `TTL Mechanism` about agent online status
	cmdHeartbeat = "heartbeat"

	// master -> agent
	// TODO: agent -> master (notify back with confirm that agent has save the shutdown flag and will not rejoin any more)
	cmdShutdown = "shutdown"

	// master -> agent (with new workerID) (reuse persistent connection)
	// agent -> master (notify back with the same workerID that conn established) (new connection)
	cmdNewWorker = "new"
)

func newCmd(cmd, aid, wid string) []byte {
	buf := bytes.NewBuffer(nil)
	gob.NewEncoder(buf).Encode(command{cmd, aid, wid})
	return Encode(buf.Bytes())
}

// TODO replace this struct by fixed-size [32]byte
// so we don't need to use gob to encode/decode the command
type command struct {
	Cmd      string // cmdJoin, cmdLeave, cmdNewWorker, cmdHeartbeat
	AgentID  string // require on cmdJoin / cmdLeave / cmdHeartbeat
	WorkerID string // require on cmdNewWorker
}

func (cmd *command) valid() error {
	switch cmd.Cmd {
	case cmdJoin, cmdLeave, cmdHeartbeat, cmdShutdown:
		if cmd.AgentID == "" {
			return errors.New("protocol: agent id required")
		}
	case cmdNewWorker:
		if cmd.WorkerID == "" {
			return errors.New("protocol: worker id required")
		}
	default:
		return errors.New("protocol: unknown command")
	}
	return nil
}

// Encode serialize the protocol command to binary bytes
func Encode(msg []byte) []byte {
	ret := make([]byte, 0)

	ret = append(ret, HEADER...) // write header, 4 bytes

	lenBytes := int2bytes(len(msg))
	ret = append(ret, lenBytes...) // write length of msg body, 4 bytes

	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, msg)
	ret = append(ret, buf.Bytes()...) // write msg body

	return ret
}

// decoder represents a protocol command stream decoder
type decoder struct {
	r     io.Reader // read from
	store []byte    // store the read out bytes
}

// newDecoder returns a new protocol decoder that reads from r.
func newDecoder(r io.Reader) *decoder {
	if v, ok := r.(net.Conn); ok {
		r = newConnTestReader(v) // if net.Conn, wrap it with conn tester before each read()
	}
	return &decoder{
		r:     r,
		store: make([]byte, 0),
	}
}

// Buffered return current buffered bytes
func (d *decoder) Buffered() []byte {
	return d.store
}

// Decode reads the next protocol-encoded command from reader
// Note that Decode() is not concurrency safe.
func (d *decoder) Decode() (*command, error) {
	var (
		headerN = len(HEADER) + 4    // header + length
		buf     = make([]byte, 1024) // each piece read out
		n       int
		err     error
	)

	// if previous buffered data length over than one header+length
	// consume them firstly
	if len(d.Buffered()) > headerN {
		goto READ_HEADER
	}

READ_MORE:
	// read one piece of data
	n, err = d.r.Read(buf)
	if err != nil {
		return nil, err
	}

	// accumulated append to p.store
	d.store = append(d.store, buf[:n]...)

	// read more if short than header
	if len(d.Buffered()) < headerN {
		goto READ_MORE
	}

READ_HEADER:
	// scan the read out buffered data...
	// read the header and lenghtN firstly
	var (
		hl      = d.store[:headerN] // header and length bytes
		header  = hl[:len(HEADER)]  // header bytes
		length  = hl[len(HEADER):]  // length bytes
		lengthN = bytes2int(length) // length int
	)

	// ensure protocol HEADER
	if !bytes.Equal(header, HEADER) {
		// slice down the consumed abnormal data in d.store
		d.store = d.store[headerN:]
		return nil, errors.New("NOT MOLE PROTOCOL")
	}

	// if readout data not contains a full body, continue read
	if len(d.Buffered()) < headerN+lengthN {
		goto READ_MORE
	}

	// read the body out
	body := d.store[headerN : headerN+lengthN]

	// slice down the consumed data in d.store
	d.store = d.store[headerN+lengthN:]

	var (
		cmd    *command
		buffer = bytes.NewBuffer(body)
	)

	if err := gob.NewDecoder(buffer).Decode(&cmd); err != nil {
		return nil, err
	}

	return cmd, nil
}

// utils to convert between int and bytes
//
//
func int2bytes(n int) []byte { // 4 bytes
	x := int32(n)
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, x)
	return buf.Bytes()
}

func bytes2int(bs []byte) int {
	var x int32
	buf := bytes.NewBuffer(bs)
	binary.Read(buf, binary.BigEndian, &x)
	return int(x)
}
