package orderparam

import (
	"encoding/json"
	"net/url"
	"sort"
	"strings"
)

// Params is exported
type Params struct {
	allParams   map[string]string
	keyOrdering []string
}

// New is exported
func New() *Params {
	return &Params{
		allParams:   make(map[string]string),
		keyOrdering: make([]string, 0),
	}
}

// MarshalJSON implement json.Marshaler
func (o *Params) MarshalJSON() ([]byte, error) {
	type Param struct {
		Key string `json:"key"`
		Val string `json:"val"`
	}

	var Params = []*Param{}

	for _, key := range o.Keys() {
		Params = append(Params, &Param{key, o.Get(key)})
	}

	return json.Marshal(Params)
}

// Get is exported
func (o *Params) Get(key string) string {
	return o.allParams[key]
}

// Keys return the ordered keys of params map
func (o *Params) Keys() []string {
	sort.Sort(o)
	return o.keyOrdering
}

// Set is exported
func (o *Params) Set(key, value string) {
	o.keyOrdering = append(o.keyOrdering, key)
	o.allParams[key] = value
}

// Del is exported
func (o *Params) Del(key string) {
	delete(o.allParams, key)
	for idx, val := range o.keyOrdering {
		if val == key {
			o.keyOrdering = append(o.keyOrdering[:idx], o.keyOrdering[idx+1:]...) // remove from slice
		}
	}
}

// SetIgnoreNull is exported
func (o *Params) SetIgnoreNull(key, value string) {
	if key == "" || value == "" {
		return
	}
	o.Set(key, value)
}

// implement sort.Interface
func (o *Params) Len() int {
	return len(o.keyOrdering)
}

func (o *Params) Less(i int, j int) bool {
	return o.keyOrdering[i] < o.keyOrdering[j]
}

func (o *Params) Swap(i int, j int) {
	o.keyOrdering[i], o.keyOrdering[j] = o.keyOrdering[j], o.keyOrdering[i]
}

// Escape is exported
func Escape(s string) string {
	return strings.NewReplacer([]string{
		"+", "%20",
		"*", "%2A",
		"~", "%7E",
	}...).Replace(url.QueryEscape(s))
}
