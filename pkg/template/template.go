package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	units "github.com/docker/go-units"

	"github.com/bbklab/adbot/pkg/color"
)

// NewParser initialize a template.Template with some built-in template functions
func NewParser(tmpl string) (*template.Template, error) {
	return template.New("").Funcs(basicFunctions).Parse(tmpl)
}

var basicFunctions = template.FuncMap{
	"json": func(data interface{}) string {
		var (
			buf = bytes.NewBuffer(nil)
			enc = json.NewEncoder(buf)
		)
		enc.SetEscapeHTML(false)
		enc.Encode(data)
		return strings.TrimSpace(buf.String())
	},
	"split":      strings.Split,
	"join":       strings.Join,
	"title":      strings.Title,
	"lower":      strings.ToLower,
	"upper":      strings.ToUpper,
	"size":       sizeOf,
	"tformat":    timeFormat,
	"dformat":    durationFormat,
	"dsecformat": durationFormatSec,
	"boolformat": boolFormat,
	"count":      countOf,
	"comb":       combine, // similar to `join`, but for two string
	"hostof":     hostOf,  // host of a address
	"portof":     portOf,  // port of a address
	"divide":     divide,
	"multiply":   multiply,
	"rune":       runeStr,
	"red":        red,
	"green":      green,
	"yellow":     yellow,
	"cyan":       cyan,
	"magenta":    magenta,
	"grey":       grey,
}

func sizeOf(data interface{}) string {
	switch val := data.(type) {
	case string:
		return units.HumanSize(float64(len(val)))
	case int64:
		return units.HumanSize(float64(val))
	case uint64:
		return units.HumanSize(float64(val))
	case float64:
		return units.HumanSize(val)
	}
	return ""
}

func countOf(obj interface{}) int {
	objv := reflect.ValueOf(obj)

	switch objv.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map:
		return objv.Len()
	default:
		return 0
	}
}

func combine(a, b interface{}, sep string) string {
	return fmt.Sprintf("%v%s%v", a, sep, b)
}

func hostOf(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

func portOf(addr string) string {
	_, port, _ := net.SplitHostPort(addr)
	return port
}

func divide(a, b, p int) string {
	expr := "%0." + strconv.Itoa(p) + "f"
	return fmt.Sprintf(expr, float64(a)/float64(b))
}

func multiply(a, b int) int {
	return a * b
}

func runeStr(r int) string {
	return string(r)
}

func red(obj interface{}) string {
	return color.Red(fmt.Sprintf("%v", obj))
}

func green(obj interface{}) string {
	return color.Green(fmt.Sprintf("%v", obj))
}

func yellow(obj interface{}) string {
	return color.Yellow(fmt.Sprintf("%v", obj))
}

func cyan(obj interface{}) string {
	return color.Cyan(fmt.Sprintf("%v", obj))
}

func magenta(obj interface{}) string {
	return color.Magenta(fmt.Sprintf("%v", obj))
}

func grey(obj interface{}) string {
	return color.Grey(fmt.Sprintf("%v", obj))
}

func timeFormat(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	dur := time.Since(t)
	if dur > 0 {
		return units.HumanDuration(dur) + " ago"
	}
	return units.HumanDuration(-dur) + " later"
}

func durationFormatSec(sec int64) string {
	return durationFormat(time.Second * time.Duration(sec))
}

func durationFormat(dur time.Duration) string {
	return units.HumanDuration(dur)
}

func boolFormat(b *bool) string {
	if b == nil {
		return "UNKNOWN"
	}
	if *b {
		return "YES"
	}
	return "NO"
}
