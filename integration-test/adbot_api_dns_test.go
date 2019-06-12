package main

import (
	"net"
	"strings"
	"time"

	check "gopkg.in/check.v1"

	"github.com/bbklab/adbot/pkg/ptype"
	"github.com/bbklab/adbot/types"
	"github.com/miekg/dns"
)

// dns zones
//
func (s *ApiSuite) TestDNSZoneInspect(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)
	for _, data := range assertDNSZoneList() {
		zone, err := s.client.InspectDNSZone(data.origin)
		c.Assert(err, check.IsNil)
		c.Assert(zone.Origin, check.Equals, data.origin)
		c.Assert(zone.Desc, check.Equals, data.desc)
		c.Assert(zone.TTL, check.Equals, data.ttl)
		c.Assert(zone.Serial > 0, check.Equals, true)
		c.Assert(zone.Contact, check.Equals, data.contact)
		c.Assert(zone.RecordsCount, check.Equals, 8)
	}

	s.doClearAssertDNSZones(c)

	_, err := s.client.InspectDNSZone("dnszone.id.that.is.not.exists")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*not found.*")

	costPrintln("TestDNSZoneInspect() passed", startAt)
}

func (s *ApiSuite) TestDNSZoneList(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)
	zones, err := s.client.ListDNSZones()
	c.Assert(err, check.IsNil)
	c.Assert(len(zones) >= 3, check.Equals, true)
	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSZoneList() passed", startAt)
}

func (s *ApiSuite) TestDNSZoneCreate(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	zones, err := s.client.ListDNSZones()
	c.Assert(err, check.IsNil)
	c.Assert(len(zones) >= 3, check.Equals, true)

	// invalid dns zone create
	datas := map[*types.Zone]string{
		&types.Zone{Origin: "", Desc: "x"}:                                "400 - .*zone origin should be fqdn.*",
		&types.Zone{Origin: "a", Desc: "x"}:                               "400 - .*zone origin should be fqdn.*",
		&types.Zone{Origin: strings.Repeat("a", 64) + ".net.", Desc: "x"}: "400 - .*zone origin not recognized.*",
		&types.Zone{Origin: "xyz.net.", Desc: strings.Repeat("x", 1025)}:  "400 - .*zone desc length can't larger than 1024.*",
		&types.Zone{Origin: "xyz.net.", TTL: 0}:                           "400 - .*zone ttl 0 must be numberic between.*",
		&types.Zone{Origin: "xyz.net.", TTL: 3600*24 + 1}:                 "400 - .*zone ttl 86401 must be numberic between.*",
	}
	for data, errmsg := range datas {
		_, err := s.client.CreateDNSZone(data)
		c.Assert(err, check.NotNil)
		c.Assert(err, check.ErrorMatches, errmsg)
	}

	// dup name create
	_, err = s.client.CreateDNSZone(&types.Zone{Origin: zones[0].Origin, TTL: 10})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "409 - .*duplicate.*")

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSZoneCreate() passed", startAt)
}

func (s *ApiSuite) TestDNSZoneRemove(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	zones, err := s.client.ListDNSZones()
	c.Assert(err, check.IsNil)
	c.Assert(len(zones) >= 3, check.Equals, true)

	err = s.client.RemoveDNSZone(zones[0].Origin)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "423 - .*dns zone referenced by 8 dns records.*")

	records, _ := s.client.ListDNSRecords(zones[0].Origin)
	for _, record := range records {
		s.client.RemoveDNSRecord(zones[0].Origin, record.ID)
	}
	err = s.client.RemoveDNSZone(zones[0].Origin)
	c.Assert(err, check.IsNil)

	_, err = s.client.InspectDNSZone(zones[0].Origin)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, "404 - .*not found.*")

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSZoneRemove() passed", startAt)
}

// dns records
//
func (s *ApiSuite) TestDNSRecordList(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	zones, err := s.client.ListDNSZones()
	c.Assert(err, check.IsNil)
	c.Assert(len(zones) >= 3, check.Equals, true)

	records, _ := s.client.ListDNSRecords(zones[0].Origin)
	c.Assert(len(records), check.Equals, 8)
	for _, record := range records {
		c.Assert(record.Origin, check.Equals, zones[0].Origin)
	}

	costPrintln("TestDNSRecordList() passed", startAt)
}

func (s *ApiSuite) TestDNSRecordCreate(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	zones, err := s.client.ListDNSZones()
	c.Assert(err, check.IsNil)
	c.Assert(len(zones) >= 3, check.Equals, true)

	zone := zones[0]

	datas := []struct {
		origin string
		typ    string
		ttl    *int
		data   interface{}
		errmsg string
	}{
		{"", "A", ptype.Int(30), nil, "record origin should be fqdn"},
		{"test.net", "A", ptype.Int(30), nil, "record origin should be fqdn"},
		{"test.net..", "A", ptype.Int(30), nil, "record origin not recognized"},

		{"test.net.", "A", ptype.Int(0), nil, "record ttl.*"},
		{"test.net.", "A", ptype.Int(3600*24 + 1), nil, "record ttl.*"},

		{"test.net.", "A", ptype.Int(30), nil, "A Data is required for type A"},
		{"test.net.", "A", ptype.Int(30), &types.AData{"", "1.1.1.1"}, "record Host.*"},
		{"test.net.", "A", ptype.Int(30), &types.AData{strings.Repeat("a", 257), "1.1.1.1"}, "record Host.*"},
		{"test.net.", "A", ptype.Int(30), &types.AData{"%$%", "1.1.1.1"}, "record Host.*"},
		{"test.net.", "A", ptype.Int(30), &types.AData{"web", "1.1.1.a"}, "invalid ip v4 address:.*"},
		{"test.net.", "A", ptype.Int(30), &types.AData{"web", "2401:3800:1002:12c::5ab2:de02"}, "invalid ip v4 address:.*"},
		// ok test
		{dns.Fqdn("test.net."), "A", ptype.Int(1), &types.AData{"web", "1.1.1.1"}, ""},
		{dns.Fqdn("test.net."), "A", ptype.Int(3600 * 24), &types.AData{"web", "1.1.1.1"}, ""},

		{"test.net.", "AAAA", ptype.Int(30), nil, "AAAA Data is required for type AAAA"},
		{"test.net.", "AAAA", ptype.Int(30), &types.AAAAData{"", "1.1.1.1"}, "record Host.*"},
		{"test.net.", "AAAA", ptype.Int(30), &types.AAAAData{strings.Repeat("a", 257), "1.1.1.1"}, "record Host.*"},
		{"test.net.", "AAAA", ptype.Int(30), &types.AAAAData{"%$%", "1.1.1.1"}, "record Host.*"},
		{"test.net.", "AAAA", ptype.Int(30), &types.AAAAData{"web", "1.1.1.a"}, "invalid ip v6 address:.*"},
		{"test.net.", "AAAA", ptype.Int(30), &types.AAAAData{"web", "abc:"}, "invalid ip v6 address:.*"},
		// ok test
		{dns.Fqdn("test.net."), "AAAA", ptype.Int(1), &types.AAAAData{"web", "2401:3800:1002:12c::5ab2:de02"}, ""},
		{dns.Fqdn("test.net."), "AAAA", ptype.Int(3600 * 24), &types.AAAAData{"web", "2401:3800:1002:12c::5ab2:de02"}, ""},

		{"test.net.", "MX", ptype.Int(30), nil, "Mx Data is required for type MX"},
		{"test.net.", "MX", ptype.Int(30), &types.MxData{0, "abc"}, "mx preference.*"},
		{"test.net.", "MX", ptype.Int(30), &types.MxData{101, "abc"}, "mx preference.*"},
		{"test.net.", "MX", ptype.Int(30), &types.MxData{100, ""}, "mx host.*"},
		{"test.net.", "MX", ptype.Int(30), &types.MxData{100, strings.Repeat("a", 257)}, "mx host.*"},
		{"test.net.", "MX", ptype.Int(30), &types.MxData{100, "$*&"}, "mx host.*"},
		// ok test
		{dns.Fqdn("test.net."), "MX", ptype.Int(1), &types.MxData{100, "gateway"}, ""},

		{"test.net.", "SRV", ptype.Int(30), nil, "Srv Data is required for type SRV"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{0, 0, 0, ""}, "srv priority.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{100001, 0, 0, ""}, "srv priority.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 0, 0, ""}, "srv weight.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 100001, 0, ""}, "srv weight.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 10000, 0, ""}, "srv port.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 10000, 65536, ""}, "srv port.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 10000, 65535, ""}, "srv target.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 10000, 65535, strings.Repeat("a", 257)}, "srv target.*"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 10000, 65535, "a..."}, "srv target not recognized as valid domain"},
		{"test.net.", "SRV", ptype.Int(30), &types.SrvData{10000, 10000, 65535, "-a.net"}, "srv target not recognized as valid domain"},
		// ok test
		{dns.Fqdn("test.net."), "SRV", ptype.Int(1), &types.SrvData{10000, 10000, 80, "http.test.net"}, ""},

		{"test.net.", "TXT", ptype.Int(30), nil, "Txt Data is required for type TXT/SPF"},
		{"test.net.", "TXT", ptype.Int(30), []string{}, "Txt Data is required for type TXT/SPF"},
		{"test.net.", "SPF", ptype.Int(30), nil, "Txt Data is required for type TXT/SPF"},
		{"test.net.", "SPF", ptype.Int(30), []string{}, "Txt Data is required for type TXT/SPF"},
		// ok test
		{dns.Fqdn("test.net."), "TXT", ptype.Int(1), []string{"", "v=spf1 -all"}, ""},
		{dns.Fqdn("test.net."), "SPF", ptype.Int(1), []string{"", "v=spf1 -all"}, ""},

		{"test.net.", "CNAME", ptype.Int(30), []string{}, "CName Data is required for type CNAME"},
		{"test.net.", "CNAME", ptype.Int(30), []string{"a"}, "CName Data must contains alias name & real name"},
		{"test.net.", "CNAME", ptype.Int(30), []string{"", "b"}, "cname alias.*"},
		{"test.net.", "CNAME", ptype.Int(30), []string{strings.Repeat("a", 257), "b"}, "cname alias.*"},
		{"test.net.", "CNAME", ptype.Int(30), []string{"a", ""}, "cname real.*"},
		{"test.net.", "CNAME", ptype.Int(30), []string{"a", strings.Repeat("a", 257)}, "cname real.*"},
		{"test.net.", "CNAME", ptype.Int(30), []string{"x", "x"}, "Alias can't be equals to Real for type CNAME"},
		// ok test
		{dns.Fqdn("test.net."), "CNAME", ptype.Int(1), []string{"www", "web"}, ""},

		{"test.net.", "NS", ptype.Int(30), nil, "Ns Data is required for type NS"},
		{"test.net.", "NS", ptype.Int(30), ptype.String(""), "ns data.*"},
		{"test.net.", "NS", ptype.Int(30), ptype.String(strings.Repeat("a", 257)), "ns data.*"},
		// ok test
		{dns.Fqdn("test.net."), "NS", ptype.Int(1), ptype.String("ns1"), ""},

		{"test.net.", "XXX", nil, nil, "dns record type not supported yet"},
	}

	for _, d := range datas {
		if d.errmsg == "" {
			record, err := types.NewRecord(d.origin, "desc text what ever ...", d.typ, d.ttl, d.data)
			c.Assert(err, check.IsNil)
			c.Assert(record, check.NotNil)
			_, err = s.client.CreateDNSRecord(zone.ID, record)
			c.Assert(err, check.IsNil)
		} else {
			record, err := types.NewRecord(d.origin, "desc text what ever ...", d.typ, d.ttl, d.data)
			c.Assert(err, check.NotNil)
			c.Assert(err, check.ErrorMatches, d.errmsg)
			c.Assert(record, check.IsNil)
		}
	}

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSRecordCreate() passed", startAt)
}

func (s *ApiSuite) TestDNSRecordRemove(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	zones, err := s.client.ListDNSZones()
	c.Assert(err, check.IsNil)
	c.Assert(len(zones) >= 3, check.Equals, true)

	records, _ := s.client.ListDNSRecords(zones[0].Origin)
	c.Assert(len(records), check.Equals, 8)

	err = s.client.RemoveDNSRecord(zones[0].Origin, records[0].ID)
	c.Assert(err, check.IsNil)
	records, _ = s.client.ListDNSRecords(zones[0].Origin)
	c.Assert(len(records), check.Equals, 7)

	err = s.client.RemoveDNSRecord(zones[0].Origin, records[0].ID)
	c.Assert(err, check.IsNil)
	records, _ = s.client.ListDNSRecords(zones[0].Origin)
	c.Assert(len(records), check.Equals, 6)

	err = s.client.RemoveDNSRecord(zones[0].Origin, "record.id.that.is.not.exists")
	c.Assert(err, check.IsNil) // remove not exists record won't complains error
	records, _ = s.client.ListDNSRecords(zones[0].Origin)
	c.Assert(len(records), check.Equals, 6)

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSRecordRemove() passed", startAt)
}

// dns node
//
/*
func (s *ApiSuite) TestDNSNodeEnable(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	// enable dns node
	nodeDNSServerAddr := s.enableAssertDNSNode(c, "0.0.0.0:5353")
	c.Assert(nodeDNSServerAddr, check.Not(check.Equals), "")

	// port checking
	host, port, _ := net.SplitHostPort(nodeDNSServerAddr)
	portN, _ := strconv.Atoi(port)
	_, err := netkits.TCPPortCheck(&netkits.PortCheckOption{host, portN, time.Second * 5})
	c.Assert(err, check.IsNil)

	// try query against dns server
	res, err := netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "web.integration-test.net.", "A", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err := dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rra, ok := rr.(*dns.A)
	c.Assert(ok, check.Equals, true)
	c.Assert(rra.A.To4().String(), check.Equals, "1.1.1.1")

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "ftp.integration-test.net.", "AAAA", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rraaaa, ok := rr.(*dns.AAAA)
	c.Assert(ok, check.Equals, true)
	c.Assert(rraaaa.AAAA.To16().String(), check.Equals, "2401:3800:1002:12c::5ab2:de02")

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "integration-test.net.", "MX", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rrmx, ok := rr.(*dns.MX)
	c.Assert(ok, check.Equals, true)
	c.Assert(rrmx.Mx, check.Equals, "mail-exchanger.integration-test.net.")
	c.Assert(rrmx.Preference, check.Equals, uint16(10))

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "integration-test.net.", "NS", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rrns, ok := rr.(*dns.NS)
	c.Assert(ok, check.Equals, true)
	c.Assert(rrns.Ns, check.Equals, "nameserver.integration-test.net.")

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "www.integration-test.net.", "CNAME", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rrcname, ok := rr.(*dns.CNAME)
	c.Assert(ok, check.Equals, true)
	c.Assert(rrcname.Target, check.Equals, "web.integration-test.net.")

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "integration-test.net.", "SPF", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rrspf, ok := rr.(*dns.TXT)
	c.Assert(ok, check.Equals, true)
	c.Assert(len(rrspf.Txt), check.Equals, 1)
	c.Assert(rrspf.Txt[0], check.Equals, "spf what ever here ...")

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "integration-test.net.", "TXT", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rrtxt, ok := rr.(*dns.TXT)
	c.Assert(ok, check.Equals, true)
	c.Assert(len(rrtxt.Txt), check.Equals, 1)
	c.Assert(rrtxt.Txt[0], check.Equals, "text what ever here ...")

	res, err = netkits.NsLookup(&netkits.LookupOption{nodeDNSServerAddr, "integration-test.net.", "SRV", nil})
	c.Assert(err, check.IsNil)
	c.Assert(len(res.Answer), check.Equals, 1)
	rr, err = dns.NewRR(res.Answer[0])
	c.Assert(err, check.IsNil)
	rrsrv, ok := rr.(*dns.SRV)
	c.Assert(ok, check.Equals, true)
	c.Assert(rrsrv.Priority, check.Equals, uint16(10))
	c.Assert(rrsrv.Weight, check.Equals, uint16(20))
	c.Assert(rrsrv.Port, check.Equals, uint16(80))
	c.Assert(rrsrv.Target, check.Equals, "www.sina.com.")

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSNodeEnable() passed", startAt)
}
*/

/*
func (s *ApiSuite) TestDNSNodeDisable(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	// enable dns node
	nodeDNSServerAddr := s.enableAssertDNSNode(c, "0.0.0.0:5353")
	c.Assert(nodeDNSServerAddr, check.Not(check.Equals), "")

	// port checking
	host, port, _ := net.SplitHostPort(nodeDNSServerAddr)
	portN, _ := strconv.Atoi(port)
	_, err := netkits.TCPPortCheck(&netkits.PortCheckOption{host, portN, time.Second * 5})
	c.Assert(err, check.IsNil)

	// disable dns node
	node := s.getAssertOnlineNode(c)
	err = s.client.DisableDNSNode(node.ID)
	c.Assert(err, check.IsNil)
	_, err = netkits.TCPPortCheck(&netkits.PortCheckOption{host, portN, time.Second * 5})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, ".*connection refused.*")

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSNodeDisable() passed", startAt)
}
*/

/*
func (s *ApiSuite) TestDNSNodeCompare(c *check.C) {
	startAt := time.Now()

	s.doClearAssertDNSZones(c)
	s.doLoadAssertDNSZones(c)

	// enable dns node
	nodeDNSServerAddr := s.enableAssertDNSNode(c, "0.0.0.0:5353")
	c.Assert(nodeDNSServerAddr, check.Not(check.Equals), "")

	// port checking
	host, port, _ := net.SplitHostPort(nodeDNSServerAddr)
	portN, _ := strconv.Atoi(port)
	_, err := netkits.TCPPortCheck(&netkits.PortCheckOption{host, portN, time.Second * 5})
	c.Assert(err, check.IsNil)

	// ensure synced
	node := s.getAssertOnlineNode(c)
	_, err = s.client.CompareDNSNode(node.ID, true)
	c.Assert(err, check.IsNil)

	// change zones/records and compare again
	foobarZone := "foo.bar.zone."
	// add zone
	_, err = s.client.CreateDNSZone(&types.Zone{Origin: foobarZone, TTL: 100})
	defer func() {
		records, _ := s.client.ListDNSRecords(foobarZone)
		for _, record := range records {
			s.client.RemoveDNSRecord(foobarZone, record.ID)
		}
		s.client.RemoveDNSZone(foobarZone)
	}()
	c.Assert(err, check.IsNil)
	time.Sleep(time.Millisecond * 1000) // should synced in 1s
	_, err = s.client.CompareDNSNode(node.ID, true)
	c.Assert(err, check.IsNil)
	// add record
	foobarRecord, err := types.NewRecord(foobarZone, "desc text what ever ...", "A", nil, &types.AData{"foo", "7.7.7.7"})
	c.Assert(err, check.IsNil)
	created, err := s.client.CreateDNSRecord(foobarZone, foobarRecord)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Millisecond * 500)
	_, err = s.client.CompareDNSNode(node.ID, true)
	c.Assert(err, check.IsNil)
	// remove record
	err = s.client.RemoveDNSRecord(foobarZone, created.ID)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Millisecond * 500)
	_, err = s.client.CompareDNSNode(node.ID, true)
	c.Assert(err, check.IsNil)
	// remove zone
	err = s.client.RemoveDNSZone(foobarZone)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Millisecond * 500)
	_, err = s.client.CompareDNSNode(node.ID, true)
	c.Assert(err, check.IsNil)

	// disable dns node
	node = s.getAssertOnlineNode(c)
	err = s.client.DisableDNSNode(node.ID)
	c.Assert(err, check.IsNil)
	_, err = netkits.TCPPortCheck(&netkits.PortCheckOption{host, portN, time.Second * 5})
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, ".*connection refused.*")
	// it should be different on disabled nodes
	_, err = s.client.CompareDNSNode(node.ID, true)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, ".*not synced.*")

	s.doClearAssertDNSZones(c)

	costPrintln("TestDNSNodeCompare() passed", startAt)
}
*/

// utils
//
func assertDNSZoneList() []*assertDNSZone {
	datas := []*assertDNSZone{
		{"integration-test.net.", "test.net desc", 10, "abc"},
		{"integration-test.com.", "test.com desc", 20, "lmn"},
		{"integration-test.org.", "test.org desc", 30, "xyz"},
	}
	return datas
}

func assertDNSRecordList() []*assertDNSRecord {
	datas := []*assertDNSRecord{
		{"", "A record", ptype.Int(99), "A", &types.AData{"web", "1.1.1.1"}},
		{"", "AAAA record", ptype.Int(999), "AAAA", &types.AAAAData{"ftp", "2401:3800:1002:12c::5ab2:de02"}},
		{"", "MX record", ptype.Int(3600), "MX", &types.MxData{10, "mail-exchanger"}},
		{"", "NS record", ptype.Int(120), "NS", ptype.String("nameserver")},
		{"", "CNAME record", ptype.Int(360), "CNAME", []string{"www", "web"}},
		{"", "TXT record", ptype.Int(720), "TXT", []string{"text what ever here ..."}},
		{"", "SPF record", ptype.Int(720), "SPF", []string{"spf what ever here ..."}},
		{"", "SRV record", ptype.Int(60), "SRV", &types.SrvData{10, 20, 80, "www.sina.com"}},
	}
	return datas
}

type assertDNSZone struct {
	origin  string
	desc    string
	ttl     int
	contact string
}

type assertDNSRecord struct {
	origin string
	desc   string
	ttl    *int
	typ    string
	data   interface{}
}

func (s *ApiSuite) doLoadAssertDNSZones(c *check.C) {
	for _, data := range assertDNSZoneList() {
		origin, desc, ttl, contact := data.origin, data.desc, data.ttl, data.contact
		created, err := s.client.CreateDNSZone(&types.Zone{
			Origin:  origin,
			Desc:    desc,
			TTL:     ttl,
			Contact: contact,
		})
		c.Assert(err, check.IsNil)
		c.Assert(created.Origin, check.Equals, origin)
		c.Assert(created.Desc, check.Equals, desc)
		c.Assert(created.TTL, check.Equals, ttl)
		c.Assert(created.Contact, check.Equals, contact)

		origin, zoneID := created.Origin, created.ID

		for _, rdata := range assertDNSRecordList() {
			desc, ttl, typ, data := rdata.desc, rdata.ttl, rdata.typ, rdata.data
			record, err := types.NewRecord(origin, desc, typ, ttl, data)
			c.Assert(err, check.IsNil)
			c.Assert(record, check.NotNil)
			rcreated, err := s.client.CreateDNSRecord(zoneID, record)
			c.Assert(err, check.IsNil)
			c.Assert(rcreated, check.NotNil)
			c.Assert(rcreated.Origin, check.Equals, origin)
			c.Assert(rcreated.Desc, check.Equals, desc)
			c.Assert(rcreated.Type, check.Equals, typ)
		}
	}
}

func (s *ApiSuite) doClearAssertDNSZones(c *check.C) {
	for _, data := range assertDNSZoneList() {
		records, _ := s.client.ListDNSRecords(data.origin)
		for _, record := range records {
			err := s.client.RemoveDNSRecord(data.origin, record.ID)
			c.Assert(err, check.IsNil)
		}
		s.client.RemoveDNSZone(data.origin)
	}
}

func (s *ApiSuite) enableAssertDNSNode(c *check.C, listen string) (dnsServingAddr string) {
	node := s.getAssertOnlineNode(c)

	err := s.client.DisableDNSNode(node.ID)
	c.Assert(err, check.IsNil)
	err = s.client.EnableDNSNode(node.ID, listen)
	c.Assert(err, check.IsNil)
	time.Sleep(time.Millisecond * 500)
	_, err = s.client.CompareDNSNode(node.ID, true) // ensure synced
	c.Assert(err, check.IsNil)

	host, _, err := net.SplitHostPort(node.RemoteAddr)
	c.Assert(err, check.IsNil)
	_, port, err := net.SplitHostPort(listen)
	c.Assert(err, check.IsNil)

	dnsServingAddr = net.JoinHostPort(host, port)
	return
}
