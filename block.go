package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reTerminatorSaved        = regexp.MustCompile(`^Saved: (.+)$`)
	reTerminatorInsufficient = regexp.MustCompile(`^Error: insufficient (.+) data to compute metrics$`)
	reTerminatorFailedConn   = regexp.MustCompile(`(?i)^(Error|ERROR): failed to confirm connection`)
)

func isTerminator(line string) bool {
	return reTerminatorSaved.MatchString(line) ||
		reTerminatorInsufficient.MatchString(line) ||
		reTerminatorFailedConn.MatchString(line)
}

type rawBlock struct {
	opener         string
	body           []string
	terminator     string
	postTerminator []string
	openLine       int
}

type blockScan struct {
	blocks   []rawBlock
	end      *string
	warnings []string
}

func splitBlocks(lines []string, lineOffset int) blockScan {
	var bs blockScan
	i := 0
	for i < len(lines) {
		line := lines[i]
		if line == "" {
			i++
			continue
		}
		if strings.HasPrefix(line, "End:") {
			v := strings.TrimPrefix(line, "End:")
			v = strings.TrimPrefix(v, " ")
			bs.end = &v
			i++
			continue
		}
		if strings.HasPrefix(line, "=== ") {
			b, consumed := readOneBlock(lines[i:], lineOffset+i+1)
			bs.blocks = append(bs.blocks, b)
			i += consumed
			continue
		}
		bs.warnings = append(bs.warnings,
			fmt.Sprintf("line %d: unexpected line outside block: %q", lineOffset+i+1, line))
		i++
	}
	return bs
}

func readOneBlock(lines []string, openerLineNumber int) (rawBlock, int) {
	b := rawBlock{opener: lines[0], openLine: openerLineNumber}
	end := 1
	for end < len(lines) {
		l := lines[end]
		if strings.HasPrefix(l, "=== ") || strings.HasPrefix(l, "End:") {
			break
		}
		end++
	}
	termIdx := -1
	for j := 1; j < end; j++ {
		if isTerminator(lines[j]) {
			termIdx = j
			break
		}
	}
	if termIdx == -1 {
		b.body = trimTrailingBlanks(lines[1:end])
		return b, end
	}
	b.terminator = lines[termIdx]
	b.body = append([]string(nil), lines[1:termIdx]...)
	b.postTerminator = trimBlanks(lines[termIdx+1 : end])
	return b, end
}

func trimTrailingBlanks(s []string) []string {
	n := len(s)
	for n > 0 && s[n-1] == "" {
		n--
	}
	return append([]string(nil), s[:n]...)
}

func trimBlanks(s []string) []string {
	start := 0
	for start < len(s) && s[start] == "" {
		start++
	}
	end := len(s)
	for end > start && s[end-1] == "" {
		end--
	}
	if start == end {
		return nil
	}
	return append([]string(nil), s[start:end]...)
}

func parseBlockHeader(opener string) (relay, blockDate string, ok bool) {
	m := reBlockOpener.FindStringSubmatch(opener)
	if m == nil {
		return "", "", false
	}
	return strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), true
}
