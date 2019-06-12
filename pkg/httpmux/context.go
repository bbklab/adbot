package httpmux

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

const (
	contentText   = "text/plain"
	contentJSON   = "application/json"
	contentStream = "application/octet-stream"
)

// Context for request scope
type Context struct {
	Req   *http.Request       // raw http request
	Res   http.ResponseWriter // raw http response writer
	Query Params              // query params
	Path  Params              // path params

	abort    bool                   // flag to quit following midwares & handlers
	kvs      map[string]interface{} // for handlers to get/set temporarily datas in the calling chain
	handlers []HandleFunc           // matched HandleFunc handlers
	startAt  time.Time              // this context creation time
}

// Params is a map of name/value pairs for path or query params.
type Params map[string]string

func newContext(r *http.Request, w http.ResponseWriter, m *Mux) *Context {
	// obtain query params
	qs := make(Params)
	r.ParseForm()
	for k, v := range r.Form {
		qs[k] = v[0]
	}

	ctx := &Context{
		Req:      r,
		Res:      w,
		Query:    qs,           // query parameters
		Path:     make(Params), // default empty path parameters
		abort:    false,
		kvs:      make(map[string]interface{}),
		handlers: make([]HandleFunc, 0),
		startAt:  time.Now(),
	}

	// obtain the best matched route's handlers
	route, _ := m.bestMatch(ctx.Req.Method, ctx.Req.URL.Path)
	if route != nil {
		ctx.handlers = route.handlers
	}

	return ctx
}

// withPathParams set context's parsed path parameters
func (ctx *Context) withPathParams(ps Params) {
	if ps != nil {
		ctx.Path = ps
	}
}

// StartAt return current context start time
func (ctx *Context) StartAt() time.Time {
	return ctx.startAt
}

// ClientIP return current request http client ip
func (ctx *Context) ClientIP() string {
	ip, _, _ := net.SplitHostPort(ctx.Req.RemoteAddr)
	return ip
}

// SetKey ...
func (ctx *Context) SetKey(key string, val interface{}) {
	ctx.kvs[key] = val
}

// GetKey ...
func (ctx *Context) GetKey(key string) interface{} {
	return ctx.kvs[key]
}

// Abort ...
func (ctx *Context) Abort() {
	ctx.abort = true
}

func (ctx *Context) isAbort() bool {
	return ctx.abort
}

// MatchedHandlers ...
func (ctx *Context) MatchedHandlers() []HandleFunc {
	return ctx.handlers
}

// Bind ...
func (ctx *Context) Bind(data interface{}) error {
	return json.NewDecoder(ctx.Req.Body).Decode(&data)
}

// DumpRequest ...
func (ctx *Context) DumpRequest() ([]byte, error) {
	return httputil.DumpRequest(ctx.Req, true)
}

// JSON ...
func (ctx *Context) JSON(code int, data interface{}) {
	bs, err := json.Marshal(data)
	if err != nil {
		ctx.Res.WriteHeader(http.StatusInternalServerError)
		ctx.Res.Write([]byte(err.Error()))
		return
	}

	ctx.Res.Header().Set("Content-Type", contentJSON+"; charset=UTF-8")
	ctx.Res.WriteHeader(code)
	ctx.Res.Write(bs)
}

// Data ...
func (ctx *Context) Data(code int, data []byte) {
	ctx.Res.Header().Set("Content-Type", contentStream)
	ctx.Res.WriteHeader(code)
	ctx.Res.Write(data)
}

// Text ...
func (ctx *Context) Text(code int, data string) {
	ctx.Res.Header().Set("Content-Type", contentText)
	ctx.Res.WriteHeader(code)
	ctx.Res.Write([]byte(data))
}

// Redirect ...
func (ctx *Context) Redirect(url string, code int) {
	if code == 0 {
		code = http.StatusFound
	}
	http.Redirect(ctx.Res, ctx.Req, url, code)
}

// Status ...
func (ctx *Context) Status(code int) {
	ctx.Res.WriteHeader(code)
}

// PaymentRequired ...
func (ctx *Context) PaymentRequired(data interface{}) {
	ctx.ShowError(http.StatusPaymentRequired, data)
}

// NotFound ...
func (ctx *Context) NotFound(data interface{}) {
	ctx.ShowError(http.StatusNotFound, data)
}

// Conflict ...
func (ctx *Context) Conflict(data interface{}) {
	ctx.ShowError(http.StatusConflict, data)
}

// Locked ...
func (ctx *Context) Locked(data interface{}) {
	ctx.ShowError(http.StatusLocked, data)
}

// Gone ...
func (ctx *Context) Gone(data interface{}) {
	ctx.ShowError(http.StatusGone, data)
}

// BadRequest ...
func (ctx *Context) BadRequest(data interface{}) {
	ctx.ShowError(http.StatusBadRequest, data)
}

// InternalServerError ...
func (ctx *Context) InternalServerError(data interface{}) {
	ctx.ShowError(http.StatusInternalServerError, data)
}

// Forbidden ...
func (ctx *Context) Forbidden(data interface{}) {
	ctx.ShowError(http.StatusForbidden, data)
}

// Unauthorized ...
func (ctx *Context) Unauthorized(data interface{}) {
	ctx.ShowError(http.StatusUnauthorized, data)
}

// TooManyRequests ...
func (ctx *Context) TooManyRequests(data interface{}) {
	ctx.ShowError(http.StatusTooManyRequests, data)
}

// AutoError ...
func (ctx *Context) AutoError(data interface{}) {
	var msg string
	switch v := data.(type) {
	case error:
		msg = v.Error()
	case string:
		msg = v
	default:
		msg = fmt.Sprintf("%v", v)
	}

	msg = strings.ToLower(msg)

	switch {
	case strings.Contains(msg, "conflict") || strings.Contains(msg, "collision") || strings.Contains(msg, "duplicate"):
		ctx.Conflict(data)
	case strings.Contains(msg, "not found") || strings.Contains(msg, "not exist") || strings.Contains(msg, "no such file or directory"):
		ctx.NotFound(data)
	case strings.Contains(msg, "deny") || strings.Contains(msg, "forbid"):
		ctx.Forbidden(data)
	case strings.Contains(msg, "invalid") || strings.Contains(msg, "not valid"):
		ctx.BadRequest(data)
	default:
		ctx.InternalServerError(data)
	}
}

// ShowError write http error response
func (ctx *Context) ShowError(code int, data interface{}) {
	var msg string

	switch v := data.(type) {
	case error:
		msg = v.Error()
	case string:
		msg = v
	default:
		ctx.JSON(code, HTTPError{Error: data})
		return
	}

	ctx.JSON(code, HTTPError{Error: msg})
}

// HTTPError is an internal http error body
type HTTPError struct {
	Error interface{} `json:"error"`
}
