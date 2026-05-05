package main

type SpeedStats struct {
	Avg *float64 `json:"avg"`
	Med *float64 `json:"med"`
	P25 *float64 `json:"p25"`
	P75 *float64 `json:"p75"`
}

type LatencyStats struct {
	AvgMs    *float64 `json:"avg_ms"`
	MedMs    *float64 `json:"med_ms"`
	P25Ms    *float64 `json:"p25_ms"`
	P75Ms    *float64 `json:"p75_ms"`
	LossPct  *float64 `json:"loss_pct"`
	JitterMs *float64 `json:"jitter_ms"`
}

type UDPQuality struct {
	Word       *string  `json:"word"`
	Mos        *float64 `json:"mos"`
	LossPct    *float64 `json:"loss_pct"`
	JitterMs   *float64 `json:"jitter_ms"`
	ReorderPct *float64 `json:"reorder_pct"`
	Rtt        *string  `json:"rtt"`
}

type Summary struct {
	IP                    *string       `json:"ip"`
	Colo                  *string       `json:"colo"`
	ASN                   *int          `json:"asn"`
	ASNOrg                *string       `json:"asn_org"`
	Download              *SpeedStats   `json:"download"`
	Upload                *SpeedStats   `json:"upload"`
	IdleLatency           *LatencyStats `json:"idle_latency"`
	LoadedLatencyDownload *LatencyStats `json:"loaded_latency_download"`
	LoadedLatencyUpload   *LatencyStats `json:"loaded_latency_upload"`
	UDPQuality            *UDPQuality   `json:"udp_quality"`
	SavedRunPath          *string       `json:"saved_run_path"`
}

type ServerRecord struct {
	Relay            string   `json:"relay"`
	BlockDate        *string  `json:"block_date"`
	Shape            string   `json:"shape"`
	Connected        bool     `json:"connected"`
	DNSms            *float64 `json:"dns_ms"`
	DNSFailed        *string  `json:"dns_failed"`
	TLSHandshakeMs   *float64 `json:"tls_handshake_ms"`
	TLSProtocol      *string  `json:"tls_protocol"`
	TLSCipher        *string  `json:"tls_cipher"`
	TLSFailed        *string  `json:"tls_failed"`
	Summary          Summary  `json:"summary"`
	InsufficientKind *string  `json:"insufficient_kind"`
	BlockedReason    *string  `json:"blocked_reason"`
	Warnings         []string `json:"warnings"`
}

type ToolVersions struct {
	Mullvad            *string `json:"mullvad"`
	CloudflareSpeedCLI *string `json:"cloudflare_speed_cli"`
}

type PreambleOut struct {
	TotalServersDeclared *int    `json:"total_servers_declared"`
	Start                *string `json:"start"`
}

type CountsByShape struct {
	Success int `json:"success"`
	Partial int `json:"partial"`
	Failure int `json:"failure"`
	Unknown int `json:"unknown"`
}

type Output struct {
	LogPath       string         `json:"log_path"`
	GeneratedAt   string         `json:"generated_at"`
	ToolVersions  ToolVersions   `json:"tool_versions"`
	Preamble      PreambleOut    `json:"preamble"`
	Servers       []ServerRecord `json:"servers"`
	End           *string        `json:"end"`
	CountsByShape CountsByShape  `json:"counts_by_shape"`
}
