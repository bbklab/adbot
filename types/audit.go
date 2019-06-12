package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// AuditEntry represents an audit log entry
type AuditEntry struct {
	Verb           string            `json:"verb"`                  // action, just the lower-cased of HTTP method
	VerbStatus     VerbStatus        `json:"verb_status"`           // op status associated with the response
	RequestURI     string            `json:"uri"`                   // request URI
	Source         string            `json:"source"`                // clinet source ip
	ResponseCode   int               `json:"response_code"`         // response code
	ResponseSize   int64             `json:"response_size"`         // response size
	Cost           string            `json:"cost"`                  // time cost
	Time           time.Time         `json:"time"`                  // current audit time
	Annotations    map[string]string `json:"annotations,omitempty"` // unstructured key-val map to store any optional informations
	ResponseErrMsg string            `json:"response_errmsg"`       // response error message (optional)
}

// FormatString format the audit entry to text format
func (e *AuditEntry) FormatString() string {
	return fmt.Sprintf("%s,%s,%s,%s,%d,%d,%s,%s,%s",
		e.Time.Format(time.RFC3339),
		e.Verb, e.RequestURI, e.Source,
		e.ResponseCode, e.ResponseSize,
		e.Cost, e.VerbStatus, e.ResponseErrMsg,
	)
}

// FormatJSON format the audit entry to json format
func (e *AuditEntry) FormatJSON() string {
	bs, _ := json.Marshal(e)
	return string(bs)
}

// VerbStatus represents the action result status
type VerbStatus string

// nolint
var (
	VerbStatusSucc = VerbStatus("success") // 2xx
	VerbStatusFail = VerbStatus("failed")  // 4xx, 5xx
	VerbStatusUnkn = VerbStatus("unkn")    // 3xx
)
