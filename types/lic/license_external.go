// Package lic ...
//
// note: this file will be copied to product codes to decode the license & signature file
// we introduce this into product side by copying this file instead of importing this package
// because of security concern about personal identifiers
package lic

import (
	"bytes"
	"crypto"
	"crypto/aes"
	stdcipher "crypto/cipher"
	"crypto/rc4"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bbklab/adbot/pkg/utils"
	"github.com/bbklab/adbot/pkg/validator"
)

var (
	// EmptyLicense define the fixed empty license, note: at least 1 node to make the integration test cases passed!
	EmptyLicense = &License{ID: "756729c4a925400b", MaxNodes: 1}
)

// nolint
var (
	RC4Key = bytes.Repeat([]byte(`rc4 key something should be very secret and as long as 256 bytes`), 4) // 64*4 bytes
	AESKey = []byte(`size 16/24/32 -> AES-128/192/256`)                                                  // 32 bytes -> AES-256
)

// nolint
var (
	// used in the product side
	LicenseOutdatedDuration = time.Hour * 24 // define the default outdated duration after the license is created
)

// nolint
var (
	ErrLicenseNotProvided    = errors.New("no license provided")
	ErrLicenseExpired        = errors.New("the license has been expired")
	ErrLicenseOutdated       = errors.New("the license has been outdated, re-apply pls")
	ErrLicenseTimeahead      = errors.New("system time is ahead than licence created")
	ErrLicenseNodesOverQuota = errors.New("nodes count over quota")
	ErrLicenseModuleInactive = errors.New("the license modules inactive")
	ErrLicenseDigestNotMatch = errors.New("the license digest not matched")
	ErrSignatureInvalid      = errors.New("unrecognized signature data")
	ErrSignatureNotVerified  = errors.New("the signature not verified")
)

// nolint
var (
	ProductInf    = "inf"
	ProductPaybot = "paybot"
	ProductAdbot  = "adbot"
)

// nolint
var (
	ProductsList = []string{ProductInf, ProductPaybot, ProductAdbot}
)

// LicenseWrapper wrap the license with ActivedAt & ActiveRemote
type LicenseWrapper struct {
	*License
	Expired           bool          `json:"expired"`             // is expired
	CustomerName      string        `json:"customer_name"`       // related customer name
	ModuleName        string        `json:"module_name"`         // related module name
	LifeTime          time.Duration `json:"life_time"`           // life time ExpiredAt-CreatedAt
	LifeTimeHumanable string        `json:"life_time_humanable"` // humanable type of of .LifeTime
	Report            *Report       `json:"report"`              // received report
}

// License is a db license
type License struct {
	ID        string    `json:"id" bson:"id"`                 // license uniq id
	Product   string    `json:"product" bson:"product"`       // product name
	Customer  string    `json:"customer" bson:"customer"`     // ref: customer id
	Salt      string    `json:"-" bson:"salt"`                // meaningless random transformation
	Module    Module    `json:"module" bson:"module"`         // user provide, modules value
	Nonce     int64     `json:"-" bson:"nonce"`               // meaningless random transformation
	MaxNodes  int       `json:"max_nodes" bson:"max_nodes"`   // user provide, max nodes value
	Random    int64     `json:"-" bson:"random"`              // meaningless random transformation
	ExpiredAt time.Time `json:"expired_at" bson:"expired_at"` // user provide, expire time (if expired can't create any new objects)
	CreatedAt time.Time `json:"created_at" bson:"created_at"` // create time

	// we use `Raw` to store the `ORIGINAL License Data` instead of store the object `License`
	// because some db store maybe lead to some loss of precision for some fields
	// eg: mongodb maybe store `time.Time` 2018-05-07 21:01:41.120885122 +0800 CST  to  2018-05-07 21:01:41.12 +0800 CST
	// thus may lead to difference between the first-time License and re-generated License
	// so we just store the Fixed Raw string (gob+base64) into all types of database.
	// note: only set once on the first creation time
	Raw string `json:"raw" bson:"raw"` // gob & base64 the rest of License Fields & Values
}

// LicenseTemplate is exported
var LicenseTemplate = ` ID:           {{.ID}}
 Product:      {{.Product}}
 Module:       {{.Module}}
 MaxNodes:     {{.MaxNodes}}
 ExpiredAt:    {{tformat .ExpiredAt}}
 CreatedAt:    {{tformat .CreatedAt}}
`

// IsExpired is exported
func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiredAt)
}

// IsOutdated is exported
func (l *License) IsOutdated() bool {
	return time.Now().After(l.CreatedAt.Add(LicenseOutdatedDuration))
}

// SetRaw is exported
func (l *License) SetRaw() error {
	l.Raw = ""

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(l)
	if err != nil {
		return err
	}

	l.Raw = base64.StdEncoding.EncodeToString(buf.Bytes())
	return nil
}

// ReadFromRaw is exported
func (l *License) ReadFromRaw() error {
	if l.Raw == "" {
		return errors.New("license.Raw missing")
	}

	licBase64Decoded, err := base64.StdEncoding.DecodeString(l.Raw)
	if err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewBuffer(licBase64Decoded)).Decode(&l)
}

// LoadFromBytes read raw license text and decrypt into license struct
// note: we replace all of explicit error messages by fixed step error message
func (l *License) LoadFromBytes(bs []byte) error {
	licPemDecoded, _ := pem.Decode(bs)
	if licPemDecoded == nil {
		return errors.New("step 1 error")
	}
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return errors.New("step 2-1 error")
	}
	nonce := AESKey[:12]
	aesgcm, err := stdcipher.NewGCM(block)
	if err != nil {
		return errors.New("step 2-2 error")
	}
	licAesDecrypted, err := aesgcm.Open(nil, nonce, licPemDecoded.Bytes, nil)
	if err != nil {
		return errors.New("step 2-3 error")
	}
	licBase64Decoded, err := base64.StdEncoding.DecodeString(string(licAesDecrypted))
	if err != nil {
		return errors.New("step 3 error")
	}
	cipher, err := rc4.NewCipher(RC4Key)
	if err != nil {
		return errors.New("step 4 error")
	}
	var licRc4Decrypted = make([]byte, len(licBase64Decoded))
	cipher.XORKeyStream(licRc4Decrypted, licBase64Decoded)
	licHexDecoded, err := hex.DecodeString(string(licRc4Decrypted))
	if err != nil {
		return errors.New("step 5 error")
	}
	// note: last step we got the License.Raw instead of the License
	var licRaw string
	err = gob.NewDecoder(bytes.NewBuffer([]byte(licHexDecoded))).Decode(&licRaw)
	if err != nil {
		return errors.New("step 6 error")
	}
	// extra steps, decode the License.Raw Text to License struct
	l.Raw = licRaw
	err = l.ReadFromRaw()
	if err != nil {
		return errors.New("step 7 error")
	}
	return nil
}

// NewLicenseReq is exported
//
type NewLicenseReq struct {
	Product  string `json:"product"`
	Customer string `json:"customer"`
	Nodes    int    `json:"nodes"`
	Days     int    `json:"days"`
	Modules  string `json:"modules"`
}

// Valid is exported
func (req *NewLicenseReq) Valid() error {
	if !utils.SliceContains(ProductsList, req.Product) {
		return fmt.Errorf("license product name un-recognized")
	}
	if err := validator.String(req.Customer, 1, 1024, nil); err != nil {
		return fmt.Errorf("license customer %v", err)
	}
	if err := validator.Int(req.Nodes, 1, 10000000); err != nil {
		return fmt.Errorf("license max nodes %v", err)
	}
	if err := validator.Int(req.Days, 1, 36500); err != nil {
		return fmt.Errorf("license max days %v", err)
	}
	if _, err := ParseModule(req.Modules); err != nil {
		return fmt.Errorf("license modules %v", err)
	}
	return nil
}

// LicenseSignature is a parsed runtime license signature from SignatureText
//
type LicenseSignature struct {
	Digest    string `json:"digest"`    // digest of the full license data text ...
	Signature string `json:"signature"` // signature of the digest
}

// Valid is exported
func (ls *LicenseSignature) Valid() error {
	if err := validator.String(ls.Digest, 1, 2048, nil); err != nil {
		return fmt.Errorf("license signature digest %v", err)
	}
	if err := validator.String(ls.Signature, 1, 2048, nil); err != nil {
		return fmt.Errorf("license signature data %v", err)
	}
	return nil
}

// LoadFromBytes load digest & signature from raw bytes
func (ls *LicenseSignature) LoadFromBytes(sig []byte) error {
	sigDecoded, _ := pem.Decode(sig)
	if sigDecoded == nil {
		return errors.New("signature pem decode error")
	}

	sigFields := bytes.SplitN(sigDecoded.Bytes, []byte("___"), 2)
	if len(sigFields) != 2 {
		return ErrSignatureInvalid
	}

	ls.Digest = string(sigFields[0])
	ls.Signature = string(sigFields[1])

	return ls.Valid()
}

// Verify verify the digest & signature
func (ls *LicenseSignature) Verify(pubKeyFileOrText string, data []byte) error {
	// load public key
	pub, err := utils.LoadRSAPublicKey(pubKeyFileOrText)
	if err != nil {
		return err
	}

	// verify digest
	digest := sha256.Sum256(data)
	if !bytes.Equal(digest[:], []byte(ls.Digest)) {
		return ErrLicenseDigestNotMatch
	}

	// verify the digest signature with public key
	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], []byte(ls.Signature))
	if err != nil {
		return ErrSignatureNotVerified
	}

	return nil
}

// Module represents the product module
//
type Module uint32

// nolint
var (
	ErrNoValidModules     = errors.New("without any valid modules found")
	ErrUnRecognizedModule = "un-recognized module name %s"
)

// nolint
const (
	ModuleAll Module = 0 // all modules
)

// nolint
const (
	ModuleCloudSvr    Module = 1 << iota // 1: cloudsvr
	ModuleNetProbe                       // 2: node network probe
	ModuleDNSServe                       // 4: node dns serving
	ModuleShadowSocks                    // 8: shadowsocks
)

// nolint
// mappings between Module and it's text-name
var (
	ModuleNamesMap = map[Module]string{
		ModuleAll:         "all",
		ModuleCloudSvr:    "cloudsvr",
		ModuleNetProbe:    "netprobe",
		ModuleDNSServe:    "dns",
		ModuleShadowSocks: "shadowsocks",
	}
	NameModulesMap = map[string]Module{
		"all":         ModuleAll,
		"cloudsvr":    ModuleCloudSvr,
		"netprobe":    ModuleNetProbe,
		"dns":         ModuleDNSServe,
		"shadowsocks": ModuleShadowSocks,
	}
	ZhNameModulesMap = map[string]string{
		"all":         "全部模块",
		"cloudsvr":    "云节点管理",
		"netprobe":    "分布式拨测",
		"dns":         "分布式DNS",
		"shadowsocks": "网络代理",
	}
)

// String is for easily readable on Module
func (m Module) String() string {
	if m == ModuleAll {
		return ModuleNamesMap[m]
	}
	var buf = make([]string, 0, 0)
	for key, val := range ModuleNamesMap {
		if m&key != 0 {
			buf = append(buf, val)
		}
	}
	return strings.Join(buf, ",")
}

// ParseModule parse string to Module, eg: cloudsvr,netprobe
func ParseModule(s string) (Module, error) {
	var m Module

	fields := strings.Split(s, ",")
	if len(fields) == 0 {
		return m, ErrNoValidModules
	}

	for _, val := range fields {
		module, ok := NameModulesMap[val]
		if !ok {
			return m, fmt.Errorf(ErrUnRecognizedModule, val)
		}
		if module == ModuleAll { // if any of module is ModuleAll, return ModuleAll
			return ModuleAll, nil
		}
		m |= module
	}

	if m == 0 {
		return m, ErrNoValidModules
	}

	return m, nil
}

// ModulesEqual check if two given module texts are the same
func ModulesEqual(a, b string) bool {
	ma, erra := ParseModule(a)
	mb, errb := ParseModule(b)
	return erra == nil && errb == nil && ma == mb
}

// HasCloudSvr is exported
func (m Module) HasCloudSvr() bool { return m == ModuleAll || m&ModuleCloudSvr != 0 }

// HasNetProbe is exported
func (m Module) HasNetProbe() bool { return m == ModuleAll || m&ModuleNetProbe != 0 }

// HasDNSServe is exported
func (m Module) HasDNSServe() bool { return m == ModuleAll || m&ModuleDNSServe != 0 }

// HasShadowSocks is exported
func (m Module) HasShadowSocks() bool { return m == ModuleAll || m&ModuleShadowSocks != 0 }
