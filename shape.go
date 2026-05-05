package main

import "regexp"

var reBlockedReason = regexp.MustCompile(`^Blocked: (.+)$`)

type classification struct {
	shape            string
	savedRunPath     *string
	insufficientKind *string
	blockedReason    *string
	warnings         []string
}

func classifyBlock(b rawBlock) classification {
	var c classification
	if m := reTerminatorSaved.FindStringSubmatch(b.terminator); m != nil {
		c.shape = "success"
		v := m[1]
		c.savedRunPath = &v
		return c
	}
	if m := reTerminatorInsufficient.FindStringSubmatch(b.terminator); m != nil {
		c.shape = "partial"
		v := m[1]
		c.insufficientKind = &v
		return c
	}
	if reTerminatorFailedConn.MatchString(b.terminator) {
		c.shape = "failure"
		for _, line := range b.postTerminator {
			if m := reBlockedReason.FindStringSubmatch(line); m != nil {
				v := m[1]
				c.blockedReason = &v
				break
			}
		}
		return c
	}
	c.shape = "unknown"
	return c
}
