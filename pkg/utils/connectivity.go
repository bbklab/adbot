package utils

import (
	"errors"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
)

var (
	// ErrWaitConnectivityTimeout represents wait timeout
	ErrWaitConnectivityTimeout = errors.New("wait connectivity timeout")
)

// WaitTCPPort wait a tcp address `addr` connectivity until `maxWait`
func WaitTCPPort(addr string, maxWait time.Duration) error {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		return err
	}

	var (
		timeout = time.After(maxWait)
		ticker  = time.NewTicker(time.Second * 1)
	)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return ErrWaitConnectivityTimeout

		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, time.Second*3)
			if err == nil {
				conn.Close()
				return nil
			}

			log.Debugf("addr %s is not avaliable: %v, retrying ...", addr, err)
		}
	}
}
