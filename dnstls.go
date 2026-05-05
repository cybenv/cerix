package main

import (
	"regexp"
	"strconv"
)

var (
	reDNSSuccess = regexp.MustCompile(`^DNS: ([0-9.]+)ms$`)
	reDNSFailed  = regexp.MustCompile(`^DNS measurement failed: (.+)$`)
	reTLSSuccess = regexp.MustCompile(`^TLS: handshake ([0-9.]+)ms, (\S+) (.+)$`)
	reTLSFailed  = regexp.MustCompile(`^TLS measurement failed: (.+)$`)
)

type dnsTLS struct {
	dnsMs          *float64
	dnsFailed      *string
	tlsHandshakeMs *float64
	tlsProtocol    *string
	tlsCipher      *string
	tlsFailed      *string
	warnings       []string
}

func parseDNSTLS(body []string) dnsTLS {
	var r dnsTLS
	for _, line := range body {
		if m := reDNSSuccess.FindStringSubmatch(line); m != nil {
			v, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				r.warnings = append(r.warnings, "dns: invalid ms value: "+line)
				continue
			}
			r.dnsMs = &v
			continue
		}
		if m := reDNSFailed.FindStringSubmatch(line); m != nil {
			v := m[1]
			r.dnsFailed = &v
			continue
		}
		if m := reTLSSuccess.FindStringSubmatch(line); m != nil {
			v, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				r.warnings = append(r.warnings, "tls: invalid ms value: "+line)
				continue
			}
			r.tlsHandshakeMs = &v
			proto := m[2]
			cipher := m[3]
			r.tlsProtocol = &proto
			r.tlsCipher = &cipher
			continue
		}
		if m := reTLSFailed.FindStringSubmatch(line); m != nil {
			v := m[1]
			r.tlsFailed = &v
			continue
		}
	}
	return r
}
