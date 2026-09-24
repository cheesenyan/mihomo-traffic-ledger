package main

import (
	"testing"
	"time"
)

func TestPolicyGroupNameSeparatesSelectedNodeFromPolicyGroup(t *testing.T) {
	tests := []struct {
		name   string
		chains []string
		want   string
	}{
		{name: "nested policy", chains: []string{"日本家宽 专线4x", "国外流量"}, want: "国外流量"},
		{name: "multi hop", chains: []string{"局域网 HTTP 代理", "日本中转", "其他流量"}, want: "其他流量"},
		{name: "direct", chains: []string{"DIRECT"}, want: "DIRECT"},
		{name: "reject", chains: []string{"REJECT"}, want: "REJECT"},
		{name: "node only", chains: []string{"日本节点"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policyGroupName(tt.chains); got != tt.want {
				t.Fatalf("policyGroupName(%q)=%q want %q", tt.chains, got, tt.want)
			}
		})
	}
}

func TestOutboundNameKeepsOriginalNodeName(t *testing.T) {
	chains := []string{"🇯🇵日本家宽 专线4x", "国外流量"}
	if got := outboundName(chains); got != chains[0] {
		t.Fatalf("outboundName(%q)=%q want original node %q", chains, got, chains[0])
	}
}

func TestProviderChainsRemainOptionalAndPreserveNames(t *testing.T) {
	want := []string{"provider-ai", "日本供应商 原名"}
	raw := `["provider-ai",""," 日本供应商 原名 "]`
	got := parseOptionalNames(raw)
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("parseOptionalNames(%q)=%q want %q", raw, got, want)
	}
	if got := parseOptionalNames(""); len(got) != 0 {
		t.Fatalf("empty provider chains=%q want empty", got)
	}
}

func TestRouteAttributionRoundTripsThroughAggregateQueries(t *testing.T) {
	svc := newTestService(t)
	now := time.Date(2026, 9, 19, 20, 0, 0, 0, time.Local)
	log := trafficLog{
		SourceIP: "127.0.0.1", Host: "play.googleapis.com", DestinationIP: "142.250.1.1",
		Process: "agy.exe", RouteType: "PROXY", PolicyGroup: "国外流量",
		Outbound: "🇯🇵日本家宽 专线4x", Chains: []string{"🇯🇵日本家宽 专线4x", "国外流量"},
		ProviderChains: []string{"日本住宅 provider"}, Upload: 100, Download: 900,
	}
	if err := svc.addToAggregateBuffer([]trafficLog{log}, now.UnixMilli()); err != nil {
		t.Fatal(err)
	}

	groups, err := svc.queryAggregate("policyGroup", now.Add(-time.Minute).UnixMilli(), now.Add(time.Minute).UnixMilli(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 || groups[0].Label != "国外流量" || groups[0].Total != 1000 {
		t.Fatalf("policy groups=%+v", groups)
	}

	details, err := svc.queryConnectionDetails("policyGroup", "国外流量", "play.googleapis.com", now.Add(-time.Minute).UnixMilli(), now.Add(time.Minute).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(details) != 1 || details[0].Outbound != log.Outbound || details[0].PolicyGroup != log.PolicyGroup {
		t.Fatalf("details=%+v", details)
	}
	if len(details[0].ProviderChains) != 1 || details[0].ProviderChains[0] != log.ProviderChains[0] {
		t.Fatalf("provider chains=%q", details[0].ProviderChains)
	}
}
