package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

func parseLog(path string) (Output, error) {
	lines, err := readLines(path)
	if err != nil {
		return Output{}, err
	}

	pre, rest := parsePreamble(lines)
	bs := splitBlocks(lines[rest:], rest)

	out := Output{
		LogPath:     path,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		ToolVersions: ToolVersions{
			Mullvad:            pre.mullvadVersion,
			CloudflareSpeedCLI: pre.cloudflareSpeedCLIVersion,
		},
		Preamble: PreambleOut{
			TotalServersDeclared: pre.totalServersDeclared,
			Start:                pre.start,
		},
		End:     bs.end,
		Servers: make([]ServerRecord, 0, len(bs.blocks)),
	}

	for _, b := range bs.blocks {
		sr := buildServerRecord(b)
		out.Servers = append(out.Servers, sr)
		switch sr.Shape {
		case "success":
			out.CountsByShape.Success++
		case "partial":
			out.CountsByShape.Partial++
		case "failure":
			out.CountsByShape.Failure++
		default:
			out.CountsByShape.Unknown++
		}
	}
	return out, nil
}

func buildServerRecord(b rawBlock) ServerRecord {
	var warnings []string

	relay, blockDate, ok := parseBlockHeader(b.opener)
	if !ok {
		warnings = append(warnings, fmt.Sprintf("header: malformed opener at line %d: %q", b.openLine, b.opener))
	}
	var blockDatePtr *string
	if blockDate != "" {
		bd := blockDate
		blockDatePtr = &bd
	}

	class := classifyBlock(b)
	warnings = append(warnings, class.warnings...)

	dt := parseDNSTLS(b.body)
	warnings = append(warnings, dt.warnings...)

	summary, summaryWarn := parseSummary(b.body)
	warnings = append(warnings, summaryWarn...)

	summary.SavedRunPath = class.savedRunPath

	if warnings == nil {
		warnings = []string{}
	}

	return ServerRecord{
		Relay:            relay,
		BlockDate:        blockDatePtr,
		Shape:            class.shape,
		Connected:        hasConnected(b.body),
		DNSms:            dt.dnsMs,
		DNSFailed:        dt.dnsFailed,
		TLSHandshakeMs:   dt.tlsHandshakeMs,
		TLSProtocol:      dt.tlsProtocol,
		TLSCipher:        dt.tlsCipher,
		TLSFailed:        dt.tlsFailed,
		Summary:          summary,
		InsufficientKind: class.insufficientKind,
		BlockedReason:    class.blockedReason,
		Warnings:         warnings,
	}
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
