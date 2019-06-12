package color

import (
	"fmt"
	"math/rand"
)

var (
	colorMap = map[string]string{
		"cyan":           "36",
		"yellow":         "33",
		"green":          "32",
		"magenta":        "35",
		"red":            "31",
		"blue":           "34",
		"grey":           "37",
		"intenseCyan":    "36;1",
		"intenseYellow":  "33;1",
		"intenseGreen":   "32;1",
		"intenseMagenta": "35;1",
		"intenseRed":     "31;1",
		"intenseBlue":    "34;1",
		"intenseGrey":    "37;1",
	}
	colorVals []string
	seqIdx    int
)

func init() {
	for _, val := range colorMap {
		colorVals = append(colorVals, val)
	}
}

// Cyan is exported
func Cyan(data interface{}) string { return color(data, colorMap["cyan"]) }

// Yellow is exported
func Yellow(data interface{}) string { return color(data, colorMap["yellow"]) }

// Green is exported
func Green(data interface{}) string { return color(data, colorMap["green"]) }

// Magenta is exported
func Magenta(data interface{}) string { return color(data, colorMap["magenta"]) }

// Red is exported
func Red(data interface{}) string { return color(data, colorMap["red"]) }

// Blue is exported
func Blue(data interface{}) string { return color(data, colorMap["blue"]) }

// Grey is exported
func Grey(data interface{}) string { return color(data, colorMap["grey"]) }

// IntenseCyan is exported
func IntenseCyan(data interface{}) string { return color(data, colorMap["intenseCyan"]) }

// IntenseYellow is exported
func IntenseYellow(data interface{}) string { return color(data, colorMap["intenseYellow"]) }

// IntenseGreen is exported
func IntenseGreen(data interface{}) string { return color(data, colorMap["intenseGreen"]) }

// IntenseMagenta is exported
func IntenseMagenta(data interface{}) string { return color(data, colorMap["intenseMagenta"]) }

// IntenseRed is exported
func IntenseRed(data interface{}) string { return color(data, colorMap["intenseRed"]) }

// IntenseBlue is exported
func IntenseBlue(data interface{}) string { return color(data, colorMap["intenseBlue"]) }

// IntenseGrey is exported
func IntenseGrey(data interface{}) string { return color(data, colorMap["intenseGrey"]) }

// RandColor is exported
func RandColor(data interface{}) string { return RandColorFunc()(data) }

// RandColorFunc is exported
func RandColorFunc() func(data interface{}) string {
	var idx = rand.Intn(len(colorVals))

	return func(data interface{}) string {
		return color(data, colorVals[idx])
	}
}

// SeqColorFunc is exported
func SeqColorFunc() func(data interface{}) string {
	if seqIdx >= len(colorVals)-1 {
		seqIdx = 0
	}

	f := func(data interface{}) string {
		return color(data, colorVals[seqIdx])
	}

	seqIdx++
	return f
}

func color(data interface{}, color string) string {
	return fmt.Sprintf("\033[%sm%v\033[0m", color, data)
}
