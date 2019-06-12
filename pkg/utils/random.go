package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	mrand "math/rand"
	"sync"
	"time"
)

func init() {
	randomStringRand = mrand.New(mrand.NewSource(time.Now().UnixNano()))
	mrand.Seed(time.Now().UnixNano())
}

var (
	// LowerAlpha represents all lower case alphas
	LowerAlpha = []rune("abcdefghijklmnopqrstuvwxyz")
	// UpperAlpha represents all upper case alphas
	UpperAlpha = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	// Digits represents all digital numbers
	Digits = []rune("0123456789")
	// Specials represents some of special chars
	// note:
	// ~#+&% will lead to aliyun api Signature abnormal complains
	// <>    will be escaped to \u003c and \u003e which is confusing for display
	// ;     will lead to qingcloud password abnormal complains
	// |     will lead to tencent password abnormal complains
	Specials = []rune(`!@$^-{}[]:,./=?*`)
)

var (
	randomStringMu   sync.Mutex
	randomStringRand *mrand.Rand
)

// RandomStringRange will return a string of length size that will only
// contain runes inside validRunes, panics if the validRunes is empty
func RandomStringRange(size int, validRunes []rune) string {
	randomStringMu.Lock()
	defer randomStringMu.Unlock()

	runes := make([]rune, size)
	for i := range runes {
		runes[i] = validRunes[randomStringRand.Intn(len(validRunes))]
	}

	return string(runes)
}

// RandomString is exported
func RandomString(size int) string {
	id := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		return "adbot"
	}
	return hex.EncodeToString(id)[:size]
}

// RandomNumber is exported
func RandomNumber(size int) string {
	key := make([]byte, size)
	rand.Read(key)
	for i := range key {
		key[i] = key[i]%10 + '0'
	}
	return string(key)
}

// RandomIntRange is exported
func RandomIntRange(min, max int) int {
	return mrand.Intn(max-min) + min
}
