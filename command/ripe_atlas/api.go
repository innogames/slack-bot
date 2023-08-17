package ripe_atlas

import (
	"encoding/json"
	"fmt"
	"time"
)

type CreditsResponse struct {
	CurrentBalance            int    `json:"current_balance"`
	CreditChecked             bool   `json:"credit_checked"`
	MaxDailyCredits           int    `json:"max_daily_credits"`
	EstimatedDailyIncome      int    `json:"estimated_daily_income"`
	EstimatedDailyExpenditure int    `json:"estimated_daily_expenditure"`
	EstimatedDailyBalance     int    `json:"estimated_daily_balance"`
	CalculationTime           string `json:"calculation_time"`
	EstimatedRunoutSeconds    any    `json:"estimated_runout_seconds"`
	PastDayMeasurementResults int    `json:"past_day_measurement_results"`
	PastDayCreditsSpent       int    `json:"past_day_credits_spent"`
	LastDateDebited           string `json:"last_date_debited"`
	LastDateCredited          string `json:"last_date_credited"`
	IncomeItems               string `json:"income_items"`
	ExpenseItems              string `json:"expense_items"`
	Transactions              string `json:"transactions"`
}

type MeasurementsResponse struct {
	Count        int           `json:"count,omitempty"`
	Next         string        `json:"next,omitempty"`
	Previous     string        `json:"previous,omitempty"`
	Measurements []Measurement `json:"results"`
}

type MeasurementResult struct {
	Measurements []int `json:"measurements"`
}

type Measurement struct {
	AddressFamily          int               `json:"af"`
	CreationTime           int               `json:"creation_time"`
	CreditsPerResult       int               `json:"credits_per_result"`
	Description            string            `json:"description"`
	EstimatedResultsPerDay int               `json:"estimated_results_per_day"`
	Group                  string            `json:"group"`
	GroupID                int64             `json:"group_id"`
	ID                     int               `json:"id"`
	InWifiGroup            bool              `json:"in_wifi_group"`
	IncludeProbeID         bool              `json:"include_probe_id"`
	Interval               int               `json:"interval"`
	IsAllScheduled         bool              `json:"is_all_scheduled"`
	IsOneoff               bool              `json:"is_oneoff"`
	IsPublic               bool              `json:"is_public"`
	PacketInterval         int64             `json:"packet_interval"`
	Packets                int               `json:"packets"`
	ParticipantCount       int64             `json:"participant_count"`
	ProbesRequested        int64             `json:"probes_requested"`
	ProbesScheduled        int64             `json:"probes_scheduled"`
	ResolveOnProbe         bool              `json:"resolve_on_probe"`
	ResolvedIps            string            `json:"resolved_ips"`
	Result                 string            `json:"result"`
	Size                   int64             `json:"size"`
	Spread                 int64             `json:"spread"`
	StartTime              int               `json:"start_time"`
	Status                 MeasurementStatus `json:"status"`
	StopTime               int               `json:"stop_time"`
	Tags                   []string          `json:"tags"`
	Target                 string            `json:"target"`
	TargetAsn              int64             `json:"target_asn"`
	TargetIP               string            `json:"target_ip"`
	TargetPrefix           string            `json:"target_prefix"`
	Type                   string            `json:"type"`
}
type MeasurementStatus struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	When any    `json:"when"`
}

type MeasurementRequest struct {
	Definitions []MeasurementDefinition `json:"definitions"`
	Probes      []Probes                `json:"probes"`
	IsOneOff    bool                    `json:"is_oneoff"`
}
type MeasurementDefinition struct {
	Target                string `json:"target,omitempty"`
	Af                    int    `json:"af,omitempty"`
	ResponseTimeout       int    `json:"response_timeout,omitempty"`
	Description           string `json:"description,omitempty"`
	Protocol              string `json:"protocol,omitempty"`
	ResolveOnProbe        bool   `json:"resolve_on_probe,omitempty"`
	Packets               int    `json:"packets,omitempty"`
	Size                  int    `json:"size,omitempty"`
	FirstHop              int    `json:"first_hop,omitempty"`
	MaxHops               int    `json:"max_hops,omitempty"`
	Paris                 int    `json:"paris,omitempty"`
	DestinationOptionSize int    `json:"destination_option_size,omitempty"`
	HopByHopOptionSize    int    `json:"hop_by_hop_option_size,omitempty"`
	DontFragment          bool   `json:"dont_fragment,omitempty"`
	SkipDNSCheck          bool   `json:"skip_dns_check,omitempty"`
	Type                  string `json:"type,omitempty"`
	IsPublic              bool   `json:"is_public"`
}
type Probes struct {
	Type      string `json:"type,omitempty"`
	Value     string `json:"value,omitempty"`
	Requested int    `json:"requested,omitempty"`
}

type StreamingResponse struct {
	Type    string                   `json:"type"`
	Payload StreamingResponsePayload `json:"payload"`
}

func (sr *StreamingResponse) UnmarshalJSON(b []byte) error {
	a := []interface{}{&sr.Type, &sr.Payload}
	return json.Unmarshal(b, &a)
}

type StreamingResponsePayload struct {
	Fw        int             `json:"fw,omitempty"`
	Mver      string          `json:"mver,omitempty"`
	Lts       int             `json:"lts,omitempty"`
	Endtime   int             `json:"endtime,omitempty"`
	DstName   string          `json:"dst_name,omitempty"`
	DstAddr   string          `json:"dst_addr,omitempty"`
	SrcAddr   string          `json:"src_addr,omitempty"`
	Proto     string          `json:"proto,omitempty"`
	Af        int             `json:"af,omitempty"`
	Size      int             `json:"size,omitempty"`
	ParisID   int             `json:"paris_id,omitempty"`
	Result    []PayloadResult `json:"result,omitempty"`
	MsmID     int             `json:"msm_id,omitempty"`
	PrbID     int             `json:"prb_id,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	MsmName   string          `json:"msm_name,omitempty"`
	From      string          `json:"from,omitempty"`
	Type      string          `json:"type,omitempty"`
	GroupID   int             `json:"group_id,omitempty"`
}

func (srp StreamingResponsePayload) String() string {
	//Start: 2023-08-03T14:01:07Z
	//HOST: 2a02:1811:c1c:7800:a62b:b0ff:fef1:5062 Loss%  Last
	//1  . AS0        172.20.0.1        0%   0.139
	//2  . AS0        172.26.4.1        0%   0.397
	//3  . AS0        192.168.144.1     0%   1.693
	var text string
	text += "```\n"
	text += fmt.Sprintf("Start: %s\n", time.Unix(srp.Timestamp, 0))
	text += fmt.Sprintf("HOST: %-40s Loss%%  RTT\n", srp.SrcAddr)

	for _, res := range srp.Result {
		var from string
		switch {
		case len(res.Result[0].From) > 0:
			from = res.Result[0].From
		case len(res.Result[1].From) > 0:
			from = res.Result[1].From
		case len(res.Result[2].From) > 0:
			from = res.Result[2].From
		default:
			from = "???"
		}

		text += fmt.Sprintf("%2d .  %-40s %4d%%  %7.3f %7.3f %7.3f\n", res.Hop, from, 0, res.Result[0].Rtt, res.Result[1].Rtt, res.Result[2].Rtt)
	}

	text += "```\n"
	return text
}

type HopResult struct {
	From string  `json:"from,omitempty"`
	TTL  int     `json:"ttl,omitempty"`
	Size int     `json:"size,omitempty"`
	Rtt  float64 `json:"rtt,omitempty"`
	Loss string  `json:"x,omitempty"`
}
type PayloadResult struct {
	Hop    int         `json:"hop,omitempty"`
	Result []HopResult `json:"result,omitempty"`
}
