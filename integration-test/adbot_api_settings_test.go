package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/pkg/label"
	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/types"
)

func (s *ApiSuite) TestSettingShow(c *check.C) {
	startAt := time.Now()

	settings, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(settings, check.NotNil)

	costPrintln("TestSettingShow() passed", startAt)
}

func (s *ApiSuite) TestSettingUpdate(c *check.C) {
	startAt := time.Now()

	// save current
	saved, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(saved, check.NotNil)

	// test invalid updates
	var datas = map[*types.UpdateSettingsReq]string{
		&types.UpdateSettingsReq{LogLevel: ptype.String("")}:    "400 - .*not a valid logrus Level.*",
		&types.UpdateSettingsReq{LogLevel: ptype.String("xxx")}: "400 - .*not a valid logrus Level.*",
		&types.UpdateSettingsReq{LogLevel: ptype.String("")}:    "400 - .*not a valid logrus Level.*",
	}
	for new, errmsg := range datas {
		_, err = s.client.UpdateSettings(new)
		c.Assert(err, check.NotNil)
		c.Assert(err, check.ErrorMatches, errmsg)
	}

	// test normal update
	req := &types.UpdateSettingsReq{
		LogLevel: ptype.String("info"),
	}
	ret, err := s.client.UpdateSettings(req)
	c.Assert(err, check.IsNil)
	c.Assert(ret, check.NotNil)
	c.Assert(ret.LogLevel, check.Equals, ptype.StringV(req.LogLevel))

	current, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(current, check.NotNil)
	c.Assert(current.LogLevel, check.Equals, ptype.StringV(req.LogLevel))

	// set back
	ret, err = s.client.UpdateSettings(&types.UpdateSettingsReq{
		LogLevel: ptype.String(saved.LogLevel),
	})
	c.Assert(err, check.IsNil)
	c.Assert(ret, check.NotNil)

	costPrintln("TestSettingUpdate() passed", startAt)
}

func (s *ApiSuite) TestSettingReset(c *check.C) {
	startAt := time.Now()

	// save current
	saved, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(saved, check.NotNil)

	// reset
	err = s.client.ResetSettings()
	c.Assert(err, check.IsNil)

	settings, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(settings, check.NotNil)
	c.Assert(settings.LogLevel, check.Equals, "info")
	c.Assert(settings.UpdatedAt.IsZero(), check.Equals, true)
	c.Assert(settings.Initial, check.Equals, true)

	// set back
	ret, err := s.client.UpdateSettings(&types.UpdateSettingsReq{
		LogLevel: ptype.String(saved.LogLevel),
	})
	c.Assert(err, check.IsNil)
	c.Assert(ret, check.NotNil)

	costPrintln("TestSettingShow() passed", startAt)
}

func (s *ApiSuite) TestSettingSetAttrs(c *check.C) {
	startAt := time.Now()

	settings, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(settings, check.NotNil)

	attr1 := label.New(map[string]string{"zone": "us", "gpu": "true"})
	curr, err := s.client.SetGlobalAttrs(attr1)
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(settings.GlobalAttrs.Merge(attr1)), check.Equals, true)

	new, err := s.client.GetSettings() // re-inspect current settings
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(new.GlobalAttrs), check.Equals, true)

	attr2 := label.New(map[string]string{"zone": "us", "gpu": "false"})
	curr, err = s.client.SetGlobalAttrs(attr2)
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(settings.GlobalAttrs.Merge(attr1).Merge(attr2)), check.Equals, true)

	new, err = s.client.GetSettings() // re-inspect current settings
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(new.GlobalAttrs), check.Equals, true)

	costPrintln("TestSettingSetAttrs() passed", startAt)
}

func (s *ApiSuite) TestSettingRemoveAttrs(c *check.C) {
	startAt := time.Now()

	settings, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(settings, check.NotNil)

	attr := label.New(map[string]string{"A": "B", "C": "D", "E": "F"})
	curr, err := s.client.SetGlobalAttrs(attr)
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(settings.GlobalAttrs.Merge(attr)), check.Equals, true)

	new, err := s.client.GetSettings() // re-inspect current settings
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(new.GlobalAttrs), check.Equals, true)

	rmAttr := []string{"xxx", "A", "C"}
	curr, err = s.client.RemoveGlobalAttrs(false, rmAttr)
	c.Assert(err, check.IsNil)
	for _, key := range rmAttr {
		attr.Del(key)
	}
	c.Assert(curr.EqualsTo(settings.GlobalAttrs.Merge(attr)), check.Equals, true)

	new, err = s.client.GetSettings() // re-inspect current settings
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(new.GlobalAttrs), check.Equals, true)

	curr, err = s.client.RemoveGlobalAttrs(true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(label.New(nil)), check.Equals, true)

	new, err = s.client.GetSettings() // re-inspect current settings
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(new.GlobalAttrs), check.Equals, true)

	costPrintln("TestSettingRemoveAttrs() passed", startAt)
}

func (s *ApiSuite) TestSettingSetAttrsConcurrency(c *check.C) {
	startAt := time.Now()

	// defer remove all
	defer s.client.RemoveGlobalAttrs(true, nil)

	// remove all attrs firstly
	curr, err := s.client.RemoveGlobalAttrs(true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(label.New(nil)), check.Equals, true)

	var (
		count       = 500
		concurrency = 10
	)

	var (
		wg           sync.WaitGroup
		tokens       = make(chan struct{}, concurrency)
		attrExpected = label.New(nil)
	)

	wg.Add(count)
	for idx := 1; idx <= count; idx++ {
		tokens <- struct{}{}

		attr := label.New(map[string]string{fmt.Sprintf("attr_%d", idx): strconv.Itoa(idx)})
		attrExpected = attrExpected.Merge(attr)

		go func(attr label.Labels) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			_, err := s.client.SetGlobalAttrs(attr)
			c.Assert(err, check.IsNil)
		}(attr)
	}
	wg.Wait()

	// ensure all attrs are set correctly
	new, err := s.client.GetSettings()
	c.Assert(err, check.IsNil)
	c.Assert(new.GlobalAttrs.Len(), check.Equals, count)
	c.Assert(new.GlobalAttrs.EqualsTo(attrExpected), check.Equals, true)

	costPrintln("TestSettingSetAttrsConcurrency() passed", startAt)
}

func (s *ApiSuite) TestSettingRemoveAttrsConcurrency(c *check.C) {
	startAt := time.Now()

	// defer remove all
	defer s.client.RemoveGlobalAttrs(true, nil)

	// remove all attrs firstly
	curr, err := s.client.RemoveGlobalAttrs(true, nil)
	c.Assert(err, check.IsNil)
	c.Assert(curr.EqualsTo(label.New(nil)), check.Equals, true)

	var (
		count       = 500
		concurrency = 10
	)
	var attrAll = label.New(nil)
	for idx := 1; idx <= count; idx++ {
		attr := label.New(map[string]string{fmt.Sprintf("lbs_%d", idx): strconv.Itoa(idx)})
		attrAll = attrAll.Merge(attr)
	}
	current, err := s.client.SetGlobalAttrs(attrAll)
	c.Assert(err, check.IsNil)
	c.Assert(current.Len(), check.Equals, count)
	c.Assert(current.EqualsTo(attrAll), check.Equals, true)

	// remove node label concurrency
	var (
		wg     sync.WaitGroup
		tokens = make(chan struct{}, concurrency)
	)

	wg.Add(count)
	for idx := 1; idx <= count; idx++ {
		tokens <- struct{}{}

		key := []string{fmt.Sprintf("lbs_%d", idx)}

		go func(key []string) {
			defer func() {
				wg.Done()
				<-tokens
			}()

			_, err := s.client.RemoveGlobalAttrs(false, key)
			c.Assert(err, check.IsNil)
		}(key)
	}
	wg.Wait()

	// ensure all attrs are removed correctly
	new, err := s.client.GetSettings()
	c.Assert(new.GlobalAttrs.Len(), check.Equals, 0)
	c.Assert(new.GlobalAttrs.EqualsTo(label.New(nil)), check.Equals, true)

	costPrintln("TestSettingRemoveAttrsConcurrency() passed", startAt)
}
