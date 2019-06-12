package label

import (
	"fmt"
	"reflect"
	"testing"

	check "gopkg.in/check.v1"
)

var _ = check.Suite(new(labelSuit))

type labelSuit struct{}

func TestLabel(t *testing.T) {
	check.TestingT(t)
}

func (s *labelSuit) TestLabelsBase(c *check.C) {
	lbs := New(nil)
	c.Assert(lbs.Len(), check.Equals, 0)
	c.Assert(lbs.Has("name"), check.Equals, false)
	c.Assert(lbs.Get("name"), check.Equals, "")

	lbs.Set("name", "zgz")
	c.Assert(lbs.Len(), check.Equals, 1)
	c.Assert(lbs.Has("name"), check.Equals, true)
	c.Assert(lbs.Has("addr"), check.Equals, false)
	c.Assert(lbs.Get("name"), check.Equals, "zgz")
	c.Assert(lbs.Get("addr"), check.Equals, "")

	lbs.Set("name", "bbk")
	c.Assert(lbs.Len(), check.Equals, 1)
	c.Assert(lbs.Has("name"), check.Equals, true)
	c.Assert(lbs.Has("addr"), check.Equals, false)
	c.Assert(lbs.Get("name"), check.Equals, "bbk")
	c.Assert(lbs.Get("addr"), check.Equals, "")

	lbs.Set("addr", "bj")
	c.Assert(lbs.Len(), check.Equals, 2)
	c.Assert(lbs.Has("name"), check.Equals, true)
	c.Assert(lbs.Has("addr"), check.Equals, true)
	c.Assert(lbs.Get("name"), check.Equals, "bbk")
	c.Assert(lbs.Get("addr"), check.Equals, "bj")
	c.Assert(lbs.String(), check.Equals, `{addr="bj", name="bbk"}`)

	lbs.Del("addr")
	c.Assert(lbs.Len(), check.Equals, 1)
	c.Assert(lbs.Has("name"), check.Equals, true)
	c.Assert(lbs.Has("addr"), check.Equals, false)
	c.Assert(lbs.Get("name"), check.Equals, "bbk")
	c.Assert(lbs.Get("addr"), check.Equals, "")

	lbs.Del("name")
	c.Assert(lbs.Len(), check.Equals, 0)
	c.Assert(lbs.Has("name"), check.Equals, false)
	c.Assert(lbs.Has("addr"), check.Equals, false)
	c.Assert(lbs.Get("name"), check.Equals, "")
	c.Assert(lbs.Get("addr"), check.Equals, "")

	// DelPair
	lbs = New(map[string]string{"name": "bbk", "addr": "bj"})
	lbs.DelPair("addr", "cn")
	c.Assert(lbs.Get("addr"), check.Equals, "bj")
	lbs.DelPair("addr", "bj")
	c.Assert(lbs.Has("addr"), check.Equals, false)
	lbs.DelPair("name", "zgz")
	c.Assert(lbs.Get("name"), check.Equals, "bbk")
	lbs.DelPair("name", "bbk")
	c.Assert(lbs.Has("name"), check.Equals, false)

	// Keys & Vals
	var (
		keys1 = []string{"0", "1", "2", "3", "4"}
		vals1 = []string{"a", "b", "c", "d", "e"} // ordered
		vals2 = []string{"c", "d", "a", "e", "b"} // disordered
	)
	for idx, val := range vals2 {
		lbs.Set(fmt.Sprintf("%d", idx), val)
		c.Assert(lbs.Len(), check.Equals, idx+1)
	}

	for i := 0; i < 50; i++ {
		c.Assert(reflect.DeepEqual(lbs.Keys(), keys1), check.Equals, true)
		c.Assert(reflect.DeepEqual(lbs.Vals(), vals1), check.Equals, true)
		c.Assert(reflect.DeepEqual(lbs.Vals(), vals2), check.Equals, false)
	}

}

func (s *labelSuit) TestLabelsGroupUniq(c *check.C) {
	testData := [][2][]Labels{
		{
			[]Labels{New(nil), New(nil), New(nil)},
			[]Labels{New(nil)},
		},
		{
			[]Labels{New(map[string]string{"name": "bbk", "addr": "bj"}), New(nil)},
			[]Labels{New(map[string]string{"name": "bbk", "addr": "bj"}), New(nil)},
		},
		{
			[]Labels{New(map[string]string{"name": "bbk"}), New(map[string]string{"name": "bbk"})},
			[]Labels{New(map[string]string{"name": "bbk"})},
		},
		{
			[]Labels{New(map[string]string{"name": "bbk"}), New(map[string]string{"name": "bbk"}), New(map[string]string{"zone": "bj"})},
			[]Labels{New(map[string]string{"name": "bbk"}), New(map[string]string{"zone": "bj"})},
		},
	}

	for _, data := range testData {
		var (
			lbsGroup1 = data[0]
			lbsGroup2 = data[1]
			lbsUniqed = Uniq(lbsGroup1)
		)
		c.Assert(reflect.DeepEqual(lbsGroup2, lbsUniqed), check.Equals, true)
	}
}

func (s *labelSuit) TestLabelsMerge(c *check.C) {
	testData := [][3]Labels{
		{
			New(nil),
			New(nil),
			New(nil),
		},
		{
			New(map[string]string{"name": "bbk", "addr": "bj"}),
			New(nil),
			New(map[string]string{"name": "bbk", "addr": "bj"}),
		},
		{
			New(map[string]string{"name": "bbk", "addr": "bj"}),
			New(map[string]string{"age": "99", "prod": "true"}),
			New(map[string]string{"name": "bbk", "addr": "bj", "age": "99", "prod": "true"}),
		},
		{
			New(map[string]string{"name": "bbk", "addr": "bj"}),
			New(map[string]string{"name": "zgz", "prod": "true"}),
			New(map[string]string{"name": "zgz", "addr": "bj", "prod": "true"}),
		},
	}

	for _, data := range testData {
		var (
			lbs1 = data[0]
			lbs2 = data[1]
			lbs3 = data[2]
			lbs4 = lbs1.Merge(lbs2)
		)
		c.Assert(lbs1.EqualsTo(lbs1), check.Equals, true) // keep unchanged
		c.Assert(lbs2.EqualsTo(lbs2), check.Equals, true) // keep unchanged
		c.Assert(lbs4.EqualsTo(lbs3), check.Equals, true)
	}
}

func (s *labelSuit) TestLabelsClone(c *check.C) {
	testData := []Labels{
		New(nil),
		New(map[string]string{"name": "bbk", "addr": "bj"}),
		New(map[string]string{"name": "bbk", "addr": "bj", "age": "99", "prod": "true"}),
	}

	for _, data := range testData {
		copy := data.Clone()
		c.Assert(data.EqualsTo(copy), check.Equals, true)
	}
}

func (s *labelSuit) TestLabelsConflicts(c *check.C) {
	testData := []struct {
		data   [2]Labels
		result bool
	}{
		{
			data: [2]Labels{
				New(nil),
				New(map[string]string{"name": "bbk", "addr": "bj"}),
			},
			result: false,
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"name": "bbk", "addr": "bj"}),
			},
			result: false,
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"name": "zgz", "addr": "bj"}),
			},
			result: true,
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"name": "zgz", "addr": "cn"}),
			},
			result: true,
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"age": "99", "prod": "true"}),
			},
			result: false,
		},
	}

	for _, data := range testData {
		var (
			lbs1   = data.data[0]
			lbs2   = data.data[1]
			result = data.result
		)
		c.Assert(lbs1.ConflictTo(lbs2), check.Equals, result)
	}
}

func (s *labelSuit) TestLabelsMatch(c *check.C) {
	testData := []struct {
		data   [2]Labels
		result [2]bool // matchAll matchOne
	}{
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(nil),
			},
			result: [2]bool{
				true,
				true,
			},
		},
		{
			data: [2]Labels{
				New(nil),
				New(map[string]string{"name": "bbk", "addr": "bj"}),
			},
			result: [2]bool{
				false,
				false,
			},
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"name": "bbk", "addr": "bj"}),
			},
			result: [2]bool{
				true,
				true,
			},
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk"}),
				New(map[string]string{"name": "bbk", "addr": "bj"}),
			},
			result: [2]bool{
				false, // match all
				true,  // match one
			},
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"name": "zgz", "addr": "bj"}),
			},
			result: [2]bool{
				false,
				true,
			},
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj"}),
				New(map[string]string{"name": "zgz", "addr": "cn"}),
			},
			result: [2]bool{
				false,
				false,
			},
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "addr": "bj", "age": "99"}),
				New(map[string]string{"age": "99", "prod": "true"}),
			},
			result: [2]bool{
				false,
				true,
			},
		},
		{
			data: [2]Labels{
				New(map[string]string{"name": "bbk", "prod": "true", "age": "99"}),
				New(map[string]string{"age": "99", "prod": "true"}),
			},
			result: [2]bool{
				true,
				true,
			},
		},
	}

	for _, data := range testData {
		var (
			lbs1    = data.data[0]
			filter  = data.data[1]
			result1 = data.result[0]
			result2 = data.result[1]
		)
		c.Assert(lbs1.MatchAll(filter), check.Equals, result1)
		c.Assert(lbs1.MatchOne(filter), check.Equals, result2)
	}
}
