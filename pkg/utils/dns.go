package utils

import (
	"github.com/miekg/dns"
)

// LookupHostA same as `host -t a {host} {server}`
func LookupHostA(host, server string) ([]string, error) {
	var (
		addrs = make([]string, 0)
		msg   = new(dns.Msg)
	)

	msg.SetQuestion(dns.Fqdn(host), dns.TypeA)
	rmsg, err := dns.Exchange(msg, server)
	if err != nil {
		return addrs, err
	}

	for _, answer := range rmsg.Answer {
		if a, ok := answer.(*dns.A); ok {
			addrs = append(addrs, a.A.String())
		}
	}
	return addrs, nil
}

// LookupHostTXT same as `host -t txt {host} {server}`
func LookupHostTXT(host, server string) ([]string, error) {
	var (
		addrs = make([]string, 0)
		msg   = new(dns.Msg)
	)

	msg.SetQuestion(dns.Fqdn(host), dns.TypeTXT)
	rmsg, err := dns.Exchange(msg, server)
	if err != nil {
		return addrs, err
	}

	for _, answer := range rmsg.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			addrs = append(addrs, txt.Txt...)
		}
	}
	return addrs, nil
}
