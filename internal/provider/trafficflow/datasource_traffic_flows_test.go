package trafficflow

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildTrafficFlowsRequest(t *testing.T) {
	state := trafficFlowsModel{
		Risk:            []string{"high"},
		Action:          []string{"BLOCK"},
		SourceIP:        []string{"10.0.0.1"},
		SourcePort:      []int64{443, 8443},
		DestinationPort: []int64{53},
		SearchText:      types.StringValue("dns"),
		TimestampFrom:   types.Int64Value(1000),
		TimestampTo:     types.Int64Value(2000),
		PageNumber:      types.Int64Value(2),
		PageSize:        types.Int64Value(50),
		SkipCount:       types.BoolValue(true),
	}
	req := buildTrafficFlowsRequest(state)

	if len(req.Risk) != 1 || req.Risk[0] != "high" || req.Action[0] != "BLOCK" {
		t.Fatalf("string filters not mapped: %+v", req)
	}
	if len(req.SourcePort) != 2 || req.SourcePort[0] != 443 || req.SourcePort[1] != 8443 {
		t.Fatalf("source ports = %v", req.SourcePort)
	}
	if len(req.DestinationPort) != 1 || req.DestinationPort[0] != 53 {
		t.Fatalf("destination ports = %v", req.DestinationPort)
	}
	if req.SearchText != "dns" || req.TimestampFrom != 1000 || req.TimestampTo != 2000 {
		t.Fatalf("scalars not mapped: %+v", req)
	}
	if req.PageNumber != 2 || req.PageSize != 50 || !req.SkipCount {
		t.Fatalf("pagination not mapped: %+v", req)
	}
}

func TestBuildTrafficFlowsRequest_nullScalarsAreZero(t *testing.T) {
	req := buildTrafficFlowsRequest(trafficFlowsModel{})
	if req.TimestampFrom != 0 || req.PageNumber != 0 || req.SearchText != "" || req.SkipCount {
		t.Fatalf("expected zero-valued defaults, got %+v", req)
	}
	if req.Risk != nil || req.SourcePort != nil {
		t.Fatalf("expected nil slices for unset filters, got %+v", req)
	}
}

func TestMapTrafficFlow(t *testing.T) {
	flow := unifi.TrafficFlow{
		ID:        "flow-1",
		Action:    "ALLOW",
		Direction: "OUTBOUND",
		Protocol:  "tcp",
		Risk:      "low",
		Service:   "https",
		Count:     5,
		Time:      12345,
		Source:    unifi.TrafficFlowTarget{ClientName: "laptop", IP: "10.0.0.5", Port: 51000},
		Destination: unifi.TrafficFlowTarget{
			HostName: "example.com", IP: "93.184.216.34", Port: 443, NetworkName: "WAN",
		},
		TrafficData: unifi.TrafficFlowTrafficData{BytesRx: 900, PacketsRx: 10},
	}
	m, err := mapTrafficFlow(flow)
	if err != nil {
		t.Fatalf("mapTrafficFlow: %v", err)
	}
	if m.ID.ValueString() != "flow-1" || m.Action.ValueString() != "ALLOW" || m.Count.ValueInt64() != 5 {
		t.Fatalf("scalar mapping: %+v", m)
	}
	if m.BytesRx.ValueInt64() != 900 || m.PacketsRx.ValueInt64() != 10 {
		t.Fatalf("traffic data: %+v", m)
	}
	if m.Source.ClientName.ValueString() != "laptop" || m.Source.Port.ValueInt64() != 51000 {
		t.Fatalf("source: %+v", m.Source)
	}
	if m.Destination.HostName.ValueString() != "example.com" || m.Destination.NetworkName.ValueString() != "WAN" {
		t.Fatalf("destination: %+v", m.Destination)
	}
	if m.JSON.IsNull() || m.JSON.ValueString() == "" {
		t.Fatalf("expected json to be populated")
	}
}

func TestTrafficFlowsStateRoundTrip(t *testing.T) {
	ds := NewTrafficFlowsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}

	flow, _ := mapTrafficFlow(unifi.TrafficFlow{ID: "f1"})
	state := trafficFlowsModel{
		Site:              types.StringValue("default"),
		Risk:              []string{"high"},
		SourcePort:        []int64{443},
		Flows:             []trafficFlowModel{flow},
		HasNext:           types.BoolValue(false),
		OrMore:            types.BoolValue(false),
		ResultPageNumber:  types.Int64Value(1),
		TotalElementCount: types.Int64Value(1),
		TotalPageCount:    types.Int64Value(1),
	}

	s := tfsdk.State{Schema: resp.Schema}
	if diags := s.Set(context.Background(), &state); diags.HasError() {
		t.Fatalf("state round-trip diagnostics: %v", diags)
	}
}
