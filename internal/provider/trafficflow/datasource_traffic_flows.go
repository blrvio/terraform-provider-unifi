package trafficflow

import (
	"context"
	"encoding/json"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &trafficFlowsDataSource{}
	_ datasource.DataSourceWithConfigure = &trafficFlowsDataSource{}
	_ base.Resource                      = &trafficFlowsDataSource{}
)

type trafficFlowsDataSource struct {
	trafficFlowBase
}

type trafficFlowsModel struct {
	Site types.String `tfsdk:"site"`

	// Filters (all optional) mirror unifi.TrafficFlowsRequest.
	Risk                 []string     `tfsdk:"risk"`
	Action               []string     `tfsdk:"action"`
	Direction            []string     `tfsdk:"direction"`
	Protocol             []string     `tfsdk:"protocol"`
	Policy               []string     `tfsdk:"policy"`
	PolicyType           []string     `tfsdk:"policy_type"`
	Service              []string     `tfsdk:"service"`
	SourceHost           []string     `tfsdk:"source_host"`
	SourceMAC            []string     `tfsdk:"source_mac"`
	SourceIP             []string     `tfsdk:"source_ip"`
	SourcePort           []int64      `tfsdk:"source_port"`
	SourceNetworkID      []string     `tfsdk:"source_network_id"`
	SourceDomain         []string     `tfsdk:"source_domain"`
	SourceZoneID         []string     `tfsdk:"source_zone_id"`
	SourceRegion         []string     `tfsdk:"source_region"`
	DestinationHost      []string     `tfsdk:"destination_host"`
	DestinationMAC       []string     `tfsdk:"destination_mac"`
	DestinationIP        []string     `tfsdk:"destination_ip"`
	DestinationPort      []int64      `tfsdk:"destination_port"`
	DestinationNetworkID []string     `tfsdk:"destination_network_id"`
	DestinationDomain    []string     `tfsdk:"destination_domain"`
	DestinationZoneID    []string     `tfsdk:"destination_zone_id"`
	DestinationRegion    []string     `tfsdk:"destination_region"`
	InNetworkID          []string     `tfsdk:"in_network_id"`
	OutNetworkID         []string     `tfsdk:"out_network_id"`
	NextAiQuery          []string     `tfsdk:"next_ai_query"`
	ExceptFor            []string     `tfsdk:"except_for"`
	TimestampFrom        types.Int64  `tfsdk:"timestamp_from"`
	TimestampTo          types.Int64  `tfsdk:"timestamp_to"`
	PageNumber           types.Int64  `tfsdk:"page_number"`
	PageSize             types.Int64  `tfsdk:"page_size"`
	SearchText           types.String `tfsdk:"search_text"`
	SkipCount            types.Bool   `tfsdk:"skip_count"`

	// Results (computed).
	Flows             []trafficFlowModel `tfsdk:"flows"`
	HasNext           types.Bool         `tfsdk:"has_next"`
	OrMore            types.Bool         `tfsdk:"or_more"`
	ResultPageNumber  types.Int64        `tfsdk:"result_page_number"`
	TotalElementCount types.Int64        `tfsdk:"total_element_count"`
	TotalPageCount    types.Int64        `tfsdk:"total_page_count"`
}

type trafficFlowModel struct {
	ID          types.String           `tfsdk:"id"`
	Action      types.String           `tfsdk:"action"`
	Direction   types.String           `tfsdk:"direction"`
	Protocol    types.String           `tfsdk:"protocol"`
	Risk        types.String           `tfsdk:"risk"`
	Service     types.String           `tfsdk:"service"`
	Count       types.Int64            `tfsdk:"count"`
	Time        types.Int64            `tfsdk:"time"`
	BytesRx     types.Int64            `tfsdk:"bytes_rx"`
	PacketsRx   types.Int64            `tfsdk:"packets_rx"`
	Source      trafficFlowTargetModel `tfsdk:"source"`
	Destination trafficFlowTargetModel `tfsdk:"destination"`
	JSON        types.String           `tfsdk:"json"`
}

type trafficFlowTargetModel struct {
	ClientName  types.String `tfsdk:"client_name"`
	HostName    types.String `tfsdk:"host_name"`
	IP          types.String `tfsdk:"ip"`
	MAC         types.String `tfsdk:"mac"`
	Port        types.Int64  `tfsdk:"port"`
	NetworkID   types.String `tfsdk:"network_id"`
	NetworkName types.String `tfsdk:"network_name"`
	ZoneName    types.String `tfsdk:"zone_name"`
	Region      types.String `tfsdk:"region"`
}

func toIntSlice(in []int64) []int {
	if in == nil {
		return nil
	}
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

// buildTrafficFlowsRequest converts the data source model into the internal-SDK
// request payload. Unset (null) scalar values map to their zero value, which the
// controller treats as the default.
func buildTrafficFlowsRequest(state trafficFlowsModel) *unifi.TrafficFlowsRequest {
	return &unifi.TrafficFlowsRequest{
		Risk:                 state.Risk,
		Action:               state.Action,
		Direction:            state.Direction,
		Protocol:             state.Protocol,
		Policy:               state.Policy,
		PolicyType:           state.PolicyType,
		Service:              state.Service,
		SourceHost:           state.SourceHost,
		SourceMAC:            state.SourceMAC,
		SourceIP:             state.SourceIP,
		SourcePort:           toIntSlice(state.SourcePort),
		SourceNetworkID:      state.SourceNetworkID,
		SourceDomain:         state.SourceDomain,
		SourceZoneID:         state.SourceZoneID,
		SourceRegion:         state.SourceRegion,
		DestinationHost:      state.DestinationHost,
		DestinationMAC:       state.DestinationMAC,
		DestinationIP:        state.DestinationIP,
		DestinationPort:      toIntSlice(state.DestinationPort),
		DestinationNetworkID: state.DestinationNetworkID,
		DestinationDomain:    state.DestinationDomain,
		DestinationZoneID:    state.DestinationZoneID,
		DestinationRegion:    state.DestinationRegion,
		InNetworkID:          state.InNetworkID,
		OutNetworkID:         state.OutNetworkID,
		NextAiQuery:          state.NextAiQuery,
		ExceptFor:            state.ExceptFor,
		TimestampFrom:        state.TimestampFrom.ValueInt64(),
		TimestampTo:          state.TimestampTo.ValueInt64(),
		PageNumber:           int(state.PageNumber.ValueInt64()),
		SearchText:           state.SearchText.ValueString(),
		PageSize:             int(state.PageSize.ValueInt64()),
		SkipCount:            state.SkipCount.ValueBool(),
	}
}

// mapTrafficFlowTarget converts an internal-SDK flow endpoint into its model.
func mapTrafficFlowTarget(t unifi.TrafficFlowTarget) trafficFlowTargetModel {
	return trafficFlowTargetModel{
		ClientName:  types.StringValue(t.ClientName),
		HostName:    types.StringValue(t.HostName),
		IP:          types.StringValue(t.IP),
		MAC:         types.StringValue(t.MAC),
		Port:        types.Int64Value(int64(t.Port)),
		NetworkID:   types.StringValue(t.NetworkID),
		NetworkName: types.StringValue(t.NetworkName),
		ZoneName:    types.StringValue(t.ZoneName),
		Region:      types.StringValue(t.Region),
	}
}

// mapTrafficFlow converts an internal-SDK traffic flow into its model, serializing
// the full record (including the client fingerprint and applied policies) into the
// json attribute.
func mapTrafficFlow(f unifi.TrafficFlow) (trafficFlowModel, error) {
	raw, err := json.Marshal(f)
	if err != nil {
		return trafficFlowModel{}, err
	}
	return trafficFlowModel{
		ID:          types.StringValue(f.ID),
		Action:      types.StringValue(f.Action),
		Direction:   types.StringValue(f.Direction),
		Protocol:    types.StringValue(f.Protocol),
		Risk:        types.StringValue(f.Risk),
		Service:     types.StringValue(f.Service),
		Count:       types.Int64Value(int64(f.Count)),
		Time:        types.Int64Value(f.Time),
		BytesRx:     types.Int64Value(f.TrafficData.BytesRx),
		PacketsRx:   types.Int64Value(f.TrafficData.PacketsRx),
		Source:      mapTrafficFlowTarget(f.Source),
		Destination: mapTrafficFlowTarget(f.Destination),
		JSON:        types.StringValue(string(raw)),
	}, nil
}

// NewTrafficFlowsDataSource returns the unifi_traffic_flows data source, which
// queries observed traffic flows via the go-unifi internal API (API v2).
func NewTrafficFlowsDataSource() datasource.DataSource {
	return &trafficFlowsDataSource{}
}

func (d *trafficFlowsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_traffic_flows"
}

func stringListFilter(subject string) schema.ListAttribute {
	return schema.ListAttribute{
		MarkdownDescription: "Optional filter: restrict results to these " + subject + ".",
		Optional:            true,
		ElementType:         types.StringType,
	}
}

func trafficFlowTargetAttributes(role string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "The " + role + " endpoint of the flow.",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"client_name":  schema.StringAttribute{MarkdownDescription: "The client name, if known.", Computed: true},
			"host_name":    schema.StringAttribute{MarkdownDescription: "The host name, if known.", Computed: true},
			"ip":           schema.StringAttribute{MarkdownDescription: "The IP address.", Computed: true},
			"mac":          schema.StringAttribute{MarkdownDescription: "The MAC address.", Computed: true},
			"port":         schema.Int64Attribute{MarkdownDescription: "The port.", Computed: true},
			"network_id":   schema.StringAttribute{MarkdownDescription: "The network id.", Computed: true},
			"network_name": schema.StringAttribute{MarkdownDescription: "The network name.", Computed: true},
			"zone_name":    schema.StringAttribute{MarkdownDescription: "The firewall zone name.", Computed: true},
			"region":       schema.StringAttribute{MarkdownDescription: "The region.", Computed: true},
		},
	}
}

func (d *trafficFlowsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Queries observed traffic flows via the go-unifi **internal API** (API v2). Unlike the " +
			"`officialro` data sources this does not use the Official API and is not gated on the Official-API controller " +
			"version. All arguments are optional filters; results are returned in the `flows` list.",
		Attributes: map[string]schema.Attribute{
			"site": schema.StringAttribute{
				MarkdownDescription: "The name of the UniFi site to query. If not specified, the provider's default site is used.",
				Optional:            true,
				Computed:            true,
			},
			"risk":                   stringListFilter("risk levels"),
			"action":                 stringListFilter("actions"),
			"direction":              stringListFilter("directions"),
			"protocol":               stringListFilter("protocols"),
			"policy":                 stringListFilter("policy ids"),
			"policy_type":            stringListFilter("policy types"),
			"service":                stringListFilter("services"),
			"source_host":            stringListFilter("source hosts"),
			"source_mac":             stringListFilter("source MAC addresses"),
			"source_ip":              stringListFilter("source IP addresses"),
			"source_port":            portListFilter("source ports"),
			"source_network_id":      stringListFilter("source network ids"),
			"source_domain":          stringListFilter("source domains"),
			"source_zone_id":         stringListFilter("source zone ids"),
			"source_region":          stringListFilter("source regions"),
			"destination_host":       stringListFilter("destination hosts"),
			"destination_mac":        stringListFilter("destination MAC addresses"),
			"destination_ip":         stringListFilter("destination IP addresses"),
			"destination_port":       portListFilter("destination ports"),
			"destination_network_id": stringListFilter("destination network ids"),
			"destination_domain":     stringListFilter("destination domains"),
			"destination_zone_id":    stringListFilter("destination zone ids"),
			"destination_region":     stringListFilter("destination regions"),
			"in_network_id":          stringListFilter("ingress network ids"),
			"out_network_id":         stringListFilter("egress network ids"),
			"next_ai_query":          stringListFilter("Next AI query terms"),
			"except_for":             stringListFilter("exclusion terms"),
			"timestamp_from": schema.Int64Attribute{
				MarkdownDescription: "Optional lower bound (epoch milliseconds) for the flow time window.",
				Optional:            true,
			},
			"timestamp_to": schema.Int64Attribute{
				MarkdownDescription: "Optional upper bound (epoch milliseconds) for the flow time window.",
				Optional:            true,
			},
			"page_number": schema.Int64Attribute{
				MarkdownDescription: "Optional page number to request (default first page).",
				Optional:            true,
			},
			"page_size": schema.Int64Attribute{
				MarkdownDescription: "Optional page size (default controller page size).",
				Optional:            true,
			},
			"search_text": schema.StringAttribute{
				MarkdownDescription: "Optional free-text search applied server-side.",
				Optional:            true,
			},
			"skip_count": schema.BoolAttribute{
				MarkdownDescription: "When true, ask the controller to skip computing the total element count.",
				Optional:            true,
			},
			"flows": schema.ListNestedAttribute{
				MarkdownDescription: "The matching traffic flows.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{MarkdownDescription: "The flow id.", Computed: true},
						"action":      schema.StringAttribute{MarkdownDescription: "The flow action (e.g. `ALLOW`, `BLOCK`).", Computed: true},
						"direction":   schema.StringAttribute{MarkdownDescription: "The flow direction.", Computed: true},
						"protocol":    schema.StringAttribute{MarkdownDescription: "The transport protocol.", Computed: true},
						"risk":        schema.StringAttribute{MarkdownDescription: "The assessed risk level.", Computed: true},
						"service":     schema.StringAttribute{MarkdownDescription: "The identified service.", Computed: true},
						"count":       schema.Int64Attribute{MarkdownDescription: "The number of aggregated connections.", Computed: true},
						"time":        schema.Int64Attribute{MarkdownDescription: "The flow timestamp (epoch milliseconds).", Computed: true},
						"bytes_rx":    schema.Int64Attribute{MarkdownDescription: "Bytes received.", Computed: true},
						"packets_rx":  schema.Int64Attribute{MarkdownDescription: "Packets received.", Computed: true},
						"source":      trafficFlowTargetAttributes("source"),
						"destination": trafficFlowTargetAttributes("destination"),
						"json": schema.StringAttribute{
							MarkdownDescription: "The full flow record, including the client fingerprint and applied policies, serialized as a JSON string.",
							Computed:            true,
						},
					},
				},
			},
			"has_next": schema.BoolAttribute{
				MarkdownDescription: "Whether more pages are available after the returned page.",
				Computed:            true,
			},
			"or_more": schema.BoolAttribute{
				MarkdownDescription: "Whether the total element count is a lower bound (`or more`).",
				Computed:            true,
			},
			"result_page_number": schema.Int64Attribute{
				MarkdownDescription: "The page number of the returned results.",
				Computed:            true,
			},
			"total_element_count": schema.Int64Attribute{
				MarkdownDescription: "The total number of matching flows (unless skip_count was set).",
				Computed:            true,
			},
			"total_page_count": schema.Int64Attribute{
				MarkdownDescription: "The total number of pages of matching flows.",
				Computed:            true,
			},
		},
	}
}

func portListFilter(subject string) schema.ListAttribute {
	return schema.ListAttribute{
		MarkdownDescription: "Optional filter: restrict results to these " + subject + ".",
		Optional:            true,
		ElementType:         types.Int64Type,
	}
}

func (d *trafficFlowsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state trafficFlowsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(base.CheckConfigured(d.client)...)
	if resp.Diagnostics.HasError() {
		return
	}

	site := state.Site.ValueString()
	if site == "" {
		site = d.client.Site
	}

	flowReq := buildTrafficFlowsRequest(state)
	result, err := d.client.GetTrafficFlows(ctx, site, flowReq)
	if err != nil {
		resp.Diagnostics.AddError("Error reading traffic flows", err.Error())
		return
	}

	state.Site = types.StringValue(site)
	state.Flows = make([]trafficFlowModel, 0, len(result.Data))
	for _, f := range result.Data {
		model, err := mapTrafficFlow(f)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing traffic flow", err.Error())
			return
		}
		state.Flows = append(state.Flows, model)
	}
	state.HasNext = types.BoolValue(result.HasNext)
	state.OrMore = types.BoolValue(result.OrMore)
	state.ResultPageNumber = types.Int64Value(int64(result.PageNumber))
	state.TotalElementCount = types.Int64Value(int64(result.TotalElementCount))
	state.TotalPageCount = types.Int64Value(int64(result.TotalPageCount))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
