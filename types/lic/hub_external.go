// Package lic ...
//
// note: this file will be copied to product codes to decode the hub addresses
// we introduce this into product side by copying this file instead of importing this package
// because of security concern about personal identifiers
package lic

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/bbklab/adbot/pkg/utils"
)

// Hub is exported
type Hub struct {
	Locator  string `json:"locator"`
	Resolver string `json:"resolver"`
}

// String is exported
func (hub *Hub) String() string {
	return fmt.Sprintf("%s -> %s", hub.Resolver, hub.Locator)
}

// LookupURLs is exported
func (hub *Hub) LookupURLs() []string {
	var ret []string
	for _, txt := range hub.LookupTXT() {
		ret = append(ret, txt)
	}
	return ret
}

// LookupTXT is exported
// note: dns udp lookup always failed, we retry the querying for max 3 times by default
func (hub *Hub) LookupTXT() []string {
	for i := 1; i <= 3; i++ {
		res, err := utils.LookupHostTXT(hub.Locator, hub.Resolver)
		if err == nil {
			return res
		}
		time.Sleep(time.Millisecond * 500)
	}
	return nil
}

// EncodeHubs is exported
func EncodeHubs(hubs []*Hub) string {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(hubs)
	if err != nil {
		return ""
	}
	encoded := hex.EncodeToString(buf.Bytes())
	obfuscated := utils.Obfuscate(encoded)
	return base64.StdEncoding.EncodeToString([]byte(obfuscated))
}

// DecodeHubs is exported
func DecodeHubs(data string) ([]*Hub, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	deobfuscated := utils.Deobfuscate(string(decoded))

	decoded, err = hex.DecodeString(deobfuscated)
	if err != nil {
		return nil, err
	}

	var hubs []*Hub
	err = gob.NewDecoder(bytes.NewBuffer(decoded)).Decode(&hubs)
	if err != nil {
		return nil, err
	}

	return hubs, nil
}
