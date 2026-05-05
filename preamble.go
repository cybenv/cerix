package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	reMullvadVersion            = regexp.MustCompile(`^mullvad version: (.+)$`)
	reCloudflareSpeedCLIVersion = regexp.MustCompile(`^cloudflare-speed-cli version: (.+)$`)
	reTotalServers              = regexp.MustCompile(`^Total servers: +([0-9]+)$`)
	reStart                     = regexp.MustCompile(`^Start: (.+)$`)
	reBlockOpener               = regexp.MustCompile(`^=== (.+?) \| (.+) ===$`)
)

type preamble struct {
	mullvadVersion            *string
	cloudflareSpeedCLIVersion *string
	totalServersDeclared      *int
	start                     *string
	warnings                  []string
}

func parsePreamble(lines []string) (preamble, int) {
	var p preamble
	i := 0
	for ; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "=== ") {
			return p, i
		}
		if line == "" {
			continue
		}
		if m := reMullvadVersion.FindStringSubmatch(line); m != nil {
			v := m[1]
			p.mullvadVersion = &v
			continue
		}
		if m := reCloudflareSpeedCLIVersion.FindStringSubmatch(line); m != nil {
			v := m[1]
			p.cloudflareSpeedCLIVersion = &v
			continue
		}
		if m := reTotalServers.FindStringSubmatch(line); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				p.warnings = append(p.warnings, "preamble: invalid Total servers value: "+line)
				continue
			}
			p.totalServersDeclared = &n
			continue
		}
		if m := reStart.FindStringSubmatch(line); m != nil {
			v := m[1]
			p.start = &v
			continue
		}
		p.warnings = append(p.warnings, "preamble: unknown line: "+line)
	}
	return p, i
}
