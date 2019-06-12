package label

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Labels is exported
// note: the Labels is NOT safe for concurrent access
type Labels map[string]string

// New is exported
func New(m map[string]string) Labels {
	lbs := make(Labels, 0)

	if len(m) == 0 {
		return lbs
	}

	for k, v := range m {
		lbs.Set(k, v)
	}

	return lbs
}

// Parse parse text like `name=val  name=val ...` to Labels
func Parse(expr string) (Labels, error) {
	var (
		fields = strings.Fields(expr)
		lbs    = New(nil)
	)

	if len(fields) > 0 {
		for _, pair := range fields {

			if strings.TrimSpace(pair) == "" {
				continue
			}

			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("[%s] is not valid label kv format", pair)
			}

			key, val := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			lbs.Set(key, val)
		}
	}

	return lbs, nil
}

// Uniq make uniq on group labels
func Uniq(group []Labels) []Labels {
	var ret = make([]Labels, 0)
	seen := func(lbs Labels) bool {
		for _, val := range ret {
			if val.EqualsTo(lbs) {
				return true
			}
		}
		return false
	}
	for _, lbs := range group {
		if !seen(lbs) {
			ret = append(ret, lbs)
		}
	}
	return ret
}

// Len is exported
func (lbs Labels) Len() int {
	return len(lbs)
}

// Has is exported
func (lbs Labels) Has(key string) bool {
	_, exists := lbs[key]
	return exists
}

// Get is exported
func (lbs Labels) Get(key string) string {
	return lbs[key]
}

// Set is exported
func (lbs Labels) Set(key, val string) {
	lbs[key] = val
}

// Del is exported
func (lbs Labels) Del(key string) {
	delete(lbs, key)
}

// DelPair is exported
func (lbs Labels) DelPair(key, val string) {
	if lbs.Get(key) == val {
		delete(lbs, key)
	}
}

// Keys is exported
func (lbs Labels) Keys() []string {
	ss := make([]string, 0, len(lbs))

	for key := range lbs {
		ss = append(ss, key)
	}

	sort.Strings(ss)
	return ss
}

// Vals is exported
func (lbs Labels) Vals() []string {
	ss := make([]string, 0, len(lbs))

	for _, val := range lbs {
		ss = append(ss, val)
	}

	sort.Strings(ss)
	return ss
}

// String implement string interface
func (lbs Labels) String() string {
	if lbs.Len() == 0 {
		return "{}"
	}

	ss := make([]string, 0, len(lbs))
	for key, val := range lbs {
		ss = append(ss, fmt.Sprintf("%s=%q", key, val))
	}

	sort.Strings(ss)
	return fmt.Sprintf("{%s}", strings.Join(ss, ", "))
}

// Merge is a helper function to non-destructively merge on two labels.
func (lbs Labels) Merge(new Labels) Labels {
	ret := make(Labels, len(lbs))

	for k, v := range lbs {
		ret[k] = v
	}

	for k, v := range new {
		ret[k] = v
	}

	return ret
}

// Clone make a copy of original labels map
func (lbs Labels) Clone() Labels {
	copy := make(Labels, len(lbs))
	for k, v := range lbs {
		copy[k] = v
	}
	return copy
}

// EqualsTo is exported
func (lbs Labels) EqualsTo(new Labels) bool {
	return reflect.DeepEqual(lbs, new)
}

// ConflictTo check if the label has a key matches with another map but the value does't match
func (lbs Labels) ConflictTo(other Labels) bool {
	small, big := lbs, other
	if other.Len() <= lbs.Len() {
		small, big = other, lbs
	}

	if small.Len() == 0 {
		return false
	}

	for k, v := range small {
		val, ok := big[k]
		if ok && val != v {
			return true
		}
	}

	return false
}

// MatchOne verfies if the filter labels could be matched at least one key-val pair
func (lbs Labels) MatchOne(filter Labels) bool {
	if filter.Len() == 0 {
		return true
	}

	small, big := lbs, filter
	if filter.Len() <= lbs.Len() {
		small, big = filter, lbs
	}

	for k, v := range small {
		val, ok := big[k]
		if ok && val == v {
			return true
		}
	}

	return false
}

// MatchAll verfies if the filter labels all key-val pairs could be matched
func (lbs Labels) MatchAll(filter Labels) bool {
	if filter.Len() == 0 {
		return true
	}

	for k, v := range filter {
		val, ok := lbs[k]
		if !ok {
			return false
		}

		if val != v {
			return false
		}
	}

	return true
}
