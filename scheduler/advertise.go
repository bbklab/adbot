package scheduler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// DetectHealthyAdvertiseAddrs detect the fatest healthy advertise address concurrency
// the parameter advertiseAddr may cantains multi addresses splited by comma ','
func DetectHealthyAdvertiseAddrs(advertiseAddr string) (string, error) {
	var (
		addrs = strings.Split(advertiseAddr, ",")
		ch    = make(chan string, len(addrs))

		errs []string
		mux  sync.Mutex // protect errs
	)

	for _, addr := range addrs {
		go func(addr string) {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/api/ping", addr), nil)
			if err != nil {
				mux.Lock()
				errs = append(errs, fmt.Sprintf("%s: %s", addr, err.Error()))
				mux.Unlock()
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				mux.Lock()
				errs = append(errs, fmt.Sprintf("%s: %s", addr, err.Error()))
				mux.Unlock()
				return
			}
			defer resp.Body.Close()
			if code := resp.StatusCode; code != 200 {
				bs, _ := ioutil.ReadAll(resp.Body)
				mux.Lock()
				errs = append(errs, fmt.Sprintf("%s: %d - %s", addr, code, string(bs)))
				mux.Unlock()
				return
			}
			ch <- addr
		}(addr)
	}

	select {
	case <-time.After(time.Second * 10):
		return "", fmt.Errorf("none of advertise addresses are avaliable: %v", strings.Join(errs, ",  "))
	case addr := <-ch:
		log.Println("detected the current leader is", addr)
		return addr, nil
	}
}
