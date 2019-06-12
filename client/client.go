package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	schemaUnix = "unix"
	schemaHTTP = "http"
)

// AdbotClient is the adbot api server client
// used by CLI, agent, integration test ...
type AdbotClient struct {
	addrs    []string          // use provided raw address list
	url      *url.URL          // the first avaliable address (url parsed) from the obove list
	client   *http.Client      // http client (note: should be reused to prevent *http.Transport (goroutine/fd) leaks)
	wsDialer *websocket.Dialer // web socket dialer
	headers  map[string]string // extra headers
}

// New create a adbot client smartly by range all given address list
// and pick up the first avaliable healthy client
func New(addrs []string) (Client, error) {
	var c = &AdbotClient{addrs: addrs, headers: make(map[string]string)}
	runtime.SetFinalizer(c, func(c *AdbotClient) { c.close() }) // set Finalizer on GC() to prevent leaks
	return c, c.Reset()
}

// Reset re-lookup current `addrs` and reset internal `url` and reuse `api` `wsDialer`.
// note: reuse `api` would help to prevent *http.Transport (goroutine/memory/fd) leaks
// so if caller want to re-connect to current avaliable adbot api server
// it's recommanded to use existing *AdbotClient.Reset() instead of New()
// See More:
//   https://github.com/golang/go/issues/16005
func (c *AdbotClient) Reset() error {
	var (
		jar  = new(Jar) // the same one cookie jar used by http client & ws dailer, for auth login session store // note: no need currently
		url  *url.URL
		err  error
		errs []string
	)

	// reuse internal `client` `wsDialer` (prevent *http.Transport leaks)
	if c.client == nil {
		c.client = &http.Client{
			Jar:           jar, // note: no need currently
			CheckRedirect: disableRedirectFunc,
			Transport: &http.Transport{
				Dial: c.dial,
			},
		}
	}

	if c.wsDialer == nil {
		c.wsDialer = &websocket.Dialer{
			NetDial: c.dial,
			Jar:     jar, // note: no need currently
		}
	}

	// reset internal `url` by re-lookup current `addrs`
	// range raw address list and pick up the first avaliable one

	c.url = nil

	for _, addr := range normalizeEndpoints(c.addrs) {
		url, err = url.Parse(addr)
		if err != nil {
			c.url = nil // note: reset c.url as nil
			errs = append(errs, fmt.Sprintf("%s:%s", addr, err.Error()))
			continue
		}

		// set the peer url, complete the new client
		c.url = url

		// verify the peer url
		if err = c.Ping(); err != nil {
			c.url = nil // note: reset c.url as nil
			errs = append(errs, fmt.Sprintf("%s:%s", addr, err.Error()))
			continue
		}

		// Note: if unix client, we skip leader master chekcing,
		// so the CLI would working fine on standby masters
		if c.url.Scheme == schemaUnix {
			break // got!
		}

		// ensure we're talking to the leader master
		if info, ok := c.QueryLeader(); !ok {
			c.url = nil // note: reset c.url as nil
			errs = append(errs, fmt.Sprintf("%s: not the leader, current leader is %s", addr, info))
			continue
		}

		// got!
		break
	}

	if c.url == nil {
		return fmt.Errorf("without any avaliable adbot api endpoints: %v", strings.Join(errs, ",  "))
	}
	return nil
}

// close releaes all resources of *AdbotClient
//  - clear the *http.Transport cached connections
//  - set the *http.Client as nil
func (c *AdbotClient) close() {
	if c.client != nil {
		c.client.Transport.(*http.Transport).CloseIdleConnections()
		c.client = nil
	}
}

// SetHeader make each request with extra headers
func (c *AdbotClient) SetHeader(name, value string) {
	if name != "" && value != "" {
		c.headers[name] = value
	}
}

// Peer return the peer server URL
func (c *AdbotClient) Peer() string {
	if c.url == nil {
		return ""
	}
	return c.url.String()
}

// PeerAddr return the peer server host address or unix path
func (c *AdbotClient) PeerAddr() string {
	if c.url == nil {
		return ""
	}
	switch c.url.Scheme {
	case schemaUnix:
		return c.url.Path
	case schemaHTTP:
		return c.url.Host
	}
	return ""
}

func (c *AdbotClient) dial(_, _ string) (net.Conn, error) {
	switch c.url.Scheme {
	case schemaUnix:
		return net.DialTimeout("unix", c.url.Path, time.Second*5)
	case schemaHTTP:
		return net.DialTimeout("tcp", c.url.Host, time.Second*5)
	}
	return nil, errors.New("unsupported scheme")
}

func (c *AdbotClient) bind(r io.Reader, val interface{}) error {
	return json.NewDecoder(r).Decode(&val)
}

func (c *AdbotClient) sendRequest(method, path string, data interface{}, timeout time.Duration, user, password string) (*http.Response, error) {
	var (
		buf   = bytes.NewBuffer(nil) // request Body holder
		ctype string                 // request Content-Type
	)

	// fill up the body buffer
	switch v := data.(type) {

	case nil: // empty request body

	case []byte:
		if _, err := buf.Write(v); err != nil {
			return nil, err
		}
		ctype = "application/octet-stream"

	default:
		if err := json.NewEncoder(buf).Encode(v); err != nil {
			return nil, err
		}
		ctype = "application/json"

	}

	var host = c.url.Host
	if c.url.Scheme == schemaUnix {
		host = "what-ever"
	}

	path = "http://" + host + path
	req, err := http.NewRequest(method, path, buf)
	if err != nil {
		return nil, err
	}

	// with Content-Type header if detected
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}

	// with extra headers
	for key, val := range c.headers {
		req.Header.Add(key, val)
	}

	// sends context-aware HTTP requests to avoid block forever
	// See: https://godoc.org/golang.org/x/net/context/ctxhttp#Do
	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		req = req.WithContext(ctx)
		defer cancel()
	}

	if user != "" && password != "" {
		req.SetBasicAuth(user, password)
	}

	return c.client.Do(req)
}

// disable http redirect if adbot client met 30x, just return the server's http response
// See: https://godoc.org/net/http#Client
func disableRedirectFunc(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

//
// Cookie Jar
//

// Jar implement http.CookieJar
type Jar struct {
	sync.RWMutex
	cookies []*http.Cookie
}

// SetCookies implement http.CookieJar
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.Lock()
	jar.cookies = cookies
	jar.Unlock()
}

// Cookies implement http.CookieJar
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	jar.RLock()
	defer jar.RUnlock()
	return jar.cookies
}

// APIError represents adbot server api error
type APIError struct {
	Code    int
	Message string
}

// Error implement error
func (e *APIError) Error() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Message)
}

// normalize adbot api endpoints like followings
//  - unix:///var/run/adbot/adbot.sock -> keep
//  - http://192.168.1.10:80             -> keep
//  - 192.168.1.10:88                    -> add prefix http://
//  - 192.168.1.10                       -> add prefix http://  and  suffix :80
func normalizeEndpoints(addrs []string) []string {
	ret := make([]string, len(addrs))
	for idx, addr := range addrs {
		if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "unix://") {
			ret[idx] = addr
			continue
		}
		_, _, err := net.SplitHostPort(addr)
		if err == nil {
			ret[idx] = "http://" + addr
			continue
		}
		ret[idx] = "http://" + addr + ":80"
	}
	return ret
}
