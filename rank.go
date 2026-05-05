package main

import (
	"fmt"
	"io"
	"sort"
)

func eligibleForRank(r *ServerRecord) bool {
	if r.Summary.Download == nil || r.Summary.Download.Med == nil {
		return false
	}
	if r.Summary.IdleLatency == nil {
		return false
	}
	if r.Summary.IdleLatency.MedMs == nil {
		return false
	}
	if r.Summary.IdleLatency.LossPct == nil {
		return false
	}
	return true
}

func topN(in []ServerRecord, n int) []ServerRecord {
	eligible := make([]ServerRecord, 0, len(in))
	for _, r := range in {
		if eligibleForRank(&r) {
			eligible = append(eligible, r)
		}
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		a, b := &eligible[i], &eligible[j]
		if *a.Summary.Download.Med != *b.Summary.Download.Med {
			return *a.Summary.Download.Med > *b.Summary.Download.Med
		}
		if *a.Summary.IdleLatency.MedMs != *b.Summary.IdleLatency.MedMs {
			return *a.Summary.IdleLatency.MedMs < *b.Summary.IdleLatency.MedMs
		}
		if *a.Summary.IdleLatency.LossPct != *b.Summary.IdleLatency.LossPct {
			return *a.Summary.IdleLatency.LossPct < *b.Summary.IdleLatency.LossPct
		}
		return a.Relay < b.Relay
	})
	if n > len(eligible) {
		n = len(eligible)
	}
	return eligible[:n]
}

func writeOverview(w io.Writer, out *Output) {
	total := len(out.Servers)
	fmt.Fprintf(w, "cerix: %s -> %d server records | success=%d partial=%d failure=%d unknown=%d\n",
		out.LogPath, total,
		out.CountsByShape.Success, out.CountsByShape.Partial,
		out.CountsByShape.Failure, out.CountsByShape.Unknown)
	if out.ToolVersions.Mullvad != nil || out.ToolVersions.CloudflareSpeedCLI != nil {
		var mull, cfs string
		if out.ToolVersions.Mullvad != nil {
			mull = *out.ToolVersions.Mullvad
		} else {
			mull = "(none)"
		}
		if out.ToolVersions.CloudflareSpeedCLI != nil {
			cfs = *out.ToolVersions.CloudflareSpeedCLI
		} else {
			cfs = "(none)"
		}
		fmt.Fprintf(w, "  tool versions: mullvad=%q cloudflare-speed-cli=%q\n", mull, cfs)
	}
	if out.Preamble.Start != nil {
		fmt.Fprintf(w, "  start: %s\n", *out.Preamble.Start)
	}
	if out.End != nil {
		fmt.Fprintf(w, "  end:   %s\n", *out.End)
	}
}

func writeTopN(w io.Writer, servers []ServerRecord, n int) {
	top := topN(servers, n)
	if len(top) == 0 {
		fmt.Fprintln(w, "Top-N: no eligible records (need download.med + idle.med_ms + idle.loss_pct)")
		return
	}
	fmt.Fprintf(w, "Top %d of %d eligible relays (sort: download.med DESC, tiebreak idle.med/loss/relay):\n",
		len(top), countEligible(servers))
	fmt.Fprintf(w, "  %-3s  %-22s  %10s  %10s  %8s\n", "#", "relay", "dl.med", "idle.med", "loss%")
	for i, r := range top {
		fmt.Fprintf(w, "  %-3d  %-22s  %10.2f  %8.1fms  %7.1f%%\n",
			i+1, r.Relay,
			*r.Summary.Download.Med,
			*r.Summary.IdleLatency.MedMs,
			*r.Summary.IdleLatency.LossPct)
	}
}

func countEligible(servers []ServerRecord) int {
	n := 0
	for i := range servers {
		if eligibleForRank(&servers[i]) {
			n++
		}
	}
	return n
}
