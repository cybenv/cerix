package main

import (
	"regexp"
	"strconv"
)

var (
	reIPColoASN     = regexp.MustCompile(`^IP/Colo/ASN: (\S+) / (\S+) / (\S+)(?: \((.+)\))?$`)
	reDownload      = regexp.MustCompile(`^Download: +avg (\S+) med (\S+) p25 (\S+) p75 (\S+)$`)
	reUpload        = regexp.MustCompile(`^Upload: +avg (\S+) med (\S+) p25 (\S+) p75 (\S+)$`)
	reIdleLatency   = regexp.MustCompile(`^Idle latency: avg (\S+) med (\S+) p25 (\S+) p75 (\S+) ms \(loss (\S+)%, jitter (\S+) ms\)$`)
	reLoadedDlLat   = regexp.MustCompile(`^Loaded latency \(download\): avg (\S+) med (\S+) p25 (\S+) p75 (\S+) ms \(loss (\S+)%, jitter (\S+) ms\)$`)
	reLoadedUlLat   = regexp.MustCompile(`^Loaded latency \(upload\): avg (\S+) med (\S+) p25 (\S+) p75 (\S+) ms \(loss (\S+)%, jitter (\S+) ms\)$`)
	reUDPQuality    = regexp.MustCompile(`^UDP quality: (.+?) \((.+?)\) \| loss (\S+)% jitter (\S+) reorder (\S+)% rtt (\S+)$`)
	reUDPQualityMOS = regexp.MustCompile(`^MOS (\S+)$`)
)

func parseSummary(body []string) (Summary, []string) {
	var s Summary
	var warnings []string

	for _, line := range body {
		if m := reIPColoASN.FindStringSubmatch(line); m != nil {
			s.IP = dashToNil(m[1])
			s.Colo = dashToNil(m[2])
			if m[3] != "-" {
				if n, err := strconv.Atoi(m[3]); err == nil {
					s.ASN = &n
				} else {
					warnings = append(warnings, "summary: invalid ASN: "+line)
				}
			}
			if len(m) >= 5 {
				s.ASNOrg = dashToNil(m[4])
			}
			continue
		}
		if m := reDownload.FindStringSubmatch(line); m != nil {
			s.Download = parseSpeedStats(m[1:5], "download", line, &warnings)
			continue
		}
		if m := reUpload.FindStringSubmatch(line); m != nil {
			s.Upload = parseSpeedStats(m[1:5], "upload", line, &warnings)
			continue
		}
		if m := reIdleLatency.FindStringSubmatch(line); m != nil {
			s.IdleLatency = parseLatencyStats(m[1:7], "idle_latency", line, &warnings)
			continue
		}
		if m := reLoadedDlLat.FindStringSubmatch(line); m != nil {
			s.LoadedLatencyDownload = parseLatencyStats(m[1:7], "loaded_latency_download", line, &warnings)
			continue
		}
		if m := reLoadedUlLat.FindStringSubmatch(line); m != nil {
			s.LoadedLatencyUpload = parseLatencyStats(m[1:7], "loaded_latency_upload", line, &warnings)
			continue
		}
		if m := reUDPQuality.FindStringSubmatch(line); m != nil {
			s.UDPQuality = parseUDPQualityFromMatch(m, &warnings)
			continue
		}
	}
	return s, warnings
}

func dashToNil(s string) *string {
	if s == "-" {
		return nil
	}
	v := s
	return &v
}

func parseFloatToken(tok, field, line string, warnings *[]string) *float64 {
	v, err := strconv.ParseFloat(tok, 64)
	if err != nil {
		*warnings = append(*warnings, "summary: "+field+": invalid numeric "+tok+" in "+line)
		return nil
	}
	return &v
}

func parseSpeedStats(toks []string, section, line string, warnings *[]string) *SpeedStats {
	return &SpeedStats{
		Avg: parseFloatToken(toks[0], section+".avg", line, warnings),
		Med: parseFloatToken(toks[1], section+".med", line, warnings),
		P25: parseFloatToken(toks[2], section+".p25", line, warnings),
		P75: parseFloatToken(toks[3], section+".p75", line, warnings),
	}
}

func parseLatencyStats(toks []string, section, line string, warnings *[]string) *LatencyStats {
	return &LatencyStats{
		AvgMs:    parseFloatToken(toks[0], section+".avg_ms", line, warnings),
		MedMs:    parseFloatToken(toks[1], section+".med_ms", line, warnings),
		P25Ms:    parseFloatToken(toks[2], section+".p25_ms", line, warnings),
		P75Ms:    parseFloatToken(toks[3], section+".p75_ms", line, warnings),
		LossPct:  parseFloatToken(toks[4], section+".loss_pct", line, warnings),
		JitterMs: parseFloatToken(toks[5], section+".jitter_ms", line, warnings),
	}
}

func parseUDPQualityFromMatch(m []string, warnings *[]string) *UDPQuality {
	q := &UDPQuality{}
	word := m[1]
	q.Word = &word

	if mosMatch := reUDPQualityMOS.FindStringSubmatch(m[2]); mosMatch != nil {
		if v, err := strconv.ParseFloat(mosMatch[1], 64); err == nil {
			q.Mos = &v
		} else {
			*warnings = append(*warnings, "summary: udp_quality.mos: invalid "+mosMatch[1])
		}
	}
	q.LossPct = parseFloatToken(m[3], "udp_quality.loss_pct", m[0], warnings)
	jitterTok := m[4]
	if jitterTok != "-" {
		stripped := jitterTok
		if l := len(stripped); l > 2 && stripped[l-2:] == "ms" {
			stripped = stripped[:l-2]
		}
		q.JitterMs = parseFloatToken(stripped, "udp_quality.jitter_ms", m[0], warnings)
	}
	q.ReorderPct = parseFloatToken(m[5], "udp_quality.reorder_pct", m[0], warnings)
	if m[6] != "-" {
		rtt := m[6]
		q.Rtt = &rtt
	}
	return q
}
