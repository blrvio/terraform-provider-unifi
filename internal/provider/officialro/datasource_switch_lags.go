package officialro

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &switchLAGsDataSource{}
	_ datasource.DataSourceWithConfigure = &switchLAGsDataSource{}
	_ base.Resource                      = &switchLAGsDataSource{}
)

type switchLAGsDataSource struct {
	officialBase
}

type switchLAGsModel struct {
	Site   types.String      `tfsdk:"site"`
	Filter types.String      `tfsdk:"filter"`
	LAGs   []lagDetailsModel `tfsdk:"lags"`
}

type lagDetailsModel struct {
	ID             types.String     `tfsdk:"id"`
	Type           types.String     `tfsdk:"type"`
	MetadataOrigin types.String     `tfsdk:"metadata_origin"`
	Members        []lagMemberModel `tfsdk:"members"`
	JSON           types.String     `tfsdk:"json"`
}

// lagMemberModel is one member of a LAG. Device-based LAGs (LOCAL,
// MULTI_CHASSIS) identify the member by device_id; switch-stack LAGs identify
// it by stack unit (unit_id, unit_mac_address). The fields that do not apply to
// the LAG's type are null.
type lagMemberModel struct {
	DeviceID       types.String `tfsdk:"device_id"`
	UnitID         types.Int64  `tfsdk:"unit_id"`
	UnitMacAddress types.String `tfsdk:"unit_mac_address"`
	PortIdxs       []int64      `tfsdk:"port_idxs"`
}

// LAG type discriminators of the Official-API LAGDetails union.
const (
	lagTypeLocal        = "LOCAL"
	lagTypeMultiChassis = "MULTI_CHASSIS"
	lagTypeSwitchStack  = "SWITCH_STACK"
)

func portIdxsToInt64(idxs []int32) []int64 {
	out := make([]int64, 0, len(idxs))
	for _, idx := range idxs {
		out = append(out, int64(idx))
	}
	return out
}

// sliceOrEmpty dereferences an optional slice, mapping an absent or null one
// to an empty (non-nil) slice so it serializes as [] rather than null.
func sliceOrEmpty[T any](p *[]T) []T {
	if p == nil || *p == nil {
		return []T{}
	}
	return *p
}

// mapDeviceLagMembers converts device-based LAG members (LOCAL, MULTI_CHASSIS).
func mapDeviceLagMembers(members []official.LagMember) []lagMemberModel {
	out := make([]lagMemberModel, 0, len(members))
	for _, m := range members {
		out = append(out, lagMemberModel{
			DeviceID:       types.StringValue(m.DeviceId.String()),
			UnitID:         types.Int64Null(),
			UnitMacAddress: types.StringNull(),
			PortIdxs:       portIdxsToInt64(m.PortIdxs),
		})
	}
	return out
}

// mapSwitchStackLagMembers converts switch-stack LAG members, which reference a
// stack unit instead of a device.
func mapSwitchStackLagMembers(members []official.SwitchStackLagMember) []lagMemberModel {
	out := make([]lagMemberModel, 0, len(members))
	for _, m := range members {
		out = append(out, lagMemberModel{
			DeviceID:       types.StringNull(),
			UnitID:         types.Int64Value(int64(m.UnitId)),
			UnitMacAddress: types.StringValue(m.UnitMacAddress),
			PortIdxs:       portIdxsToInt64(m.PortIdxs),
		})
	}
	return out
}

// lagMembers reads a LAG's members from its type-specific variant (selected by
// the type discriminator) and returns their models plus their JSON
// serialization for members_json. An unrecognized LAG type does not fail the
// read: it yields no members and a warning, and the full record stays
// available in the json attribute.
func lagMembers(l official.LAGDetails) ([]lagMemberModel, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var (
		models []lagMemberModel
		raw    any
	)
	switch l.Type {
	case lagTypeLocal:
		v, err := l.AsLocalLagGlobal()
		if err != nil {
			diags.AddError("Error decoding switch LAG", fmt.Sprintf("LAG %s (%s): %s", l.Id, l.Type, err))
			return nil, "", diags
		}
		members := sliceOrEmpty(v.Members)
		models, raw = mapDeviceLagMembers(members), members
	case lagTypeMultiChassis:
		v, err := l.AsMcLagGlobal()
		if err != nil {
			diags.AddError("Error decoding switch LAG", fmt.Sprintf("LAG %s (%s): %s", l.Id, l.Type, err))
			return nil, "", diags
		}
		members := sliceOrEmpty(v.Members)
		models, raw = mapDeviceLagMembers(members), members
	case lagTypeSwitchStack:
		v, err := l.AsSwitchStackLagGlobal()
		if err != nil {
			diags.AddError("Error decoding switch LAG", fmt.Sprintf("LAG %s (%s): %s", l.Id, l.Type, err))
			return nil, "", diags
		}
		members := sliceOrEmpty(v.Members)
		models, raw = mapSwitchStackLagMembers(members), members
	default:
		diags.AddWarning(
			"Unsupported switch LAG type",
			fmt.Sprintf("LAG %s has type %q, which this provider version does not recognize. Its members are "+
				"reported as empty; the full record is still available in the json attribute.", l.Id, l.Type),
		)
		models, raw = []lagMemberModel{}, []struct{}{}
	}
	js, err := json.Marshal(raw)
	if err != nil {
		diags.AddError("Error serializing switch LAG members", err.Error())
		return nil, "", diags
	}
	return models, string(js), diags
}

// mapLagDetails converts an Official-API LAG record into its Framework model,
// reading the members from the LAG's variant and serializing the full record
// into the json attribute.
func mapLagDetails(l official.LAGDetails) (lagDetailsModel, diag.Diagnostics) {
	members, _, diags := lagMembers(l)
	if diags.HasError() {
		return lagDetailsModel{}, diags
	}
	raw, err := json.Marshal(l)
	if err != nil {
		diags.AddError("Error serializing switch LAG", err.Error())
		return lagDetailsModel{}, diags
	}
	return lagDetailsModel{
		ID:             types.StringValue(l.Id.String()),
		Type:           types.StringValue(l.Type),
		MetadataOrigin: types.StringValue(l.Metadata.Origin),
		Members:        members,
		JSON:           types.StringValue(string(raw)),
	}, diags
}

// lagTypeDescription documents the LAG type attribute of both LAG data sources.
const lagTypeDescription = "The LAG type: `LOCAL` (ports on one switch), `MULTI_CHASSIS` (MC-LAG across two " +
	"switches) or `SWITCH_STACK` (ports across units of a switch stack)."

// NewSwitchLAGsDataSource returns the unifi_switch_lags data source, which lists
// switch link-aggregation groups on a site via the Official API. The singular
// unifi_switch_lag data source fetches one by id.
func NewSwitchLAGsDataSource() datasource.DataSource {
	return &switchLAGsDataSource{}
}

func (d *switchLAGsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_lags"
}

func (d *switchLAGsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists switch link-aggregation groups (LAGs) on a site via the UniFi **Official API** " +
			"(`integration/v1`). Read-only. Requires a controller running version 10.1.78 or later with API-key " +
			"authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("switch LAGs"),
			"lags": schema.ListNestedAttribute{
				MarkdownDescription: "The switch LAGs defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The UUID of the LAG.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: lagTypeDescription,
							Computed:            true,
						},
						"metadata_origin": schema.StringAttribute{
							MarkdownDescription: "The origin of the LAG entity (e.g. `USER_DEFINED`).",
							Computed:            true,
						},
						"members": schema.ListNestedAttribute{
							MarkdownDescription: "The LAG members, read from the type-specific LAG variant. Empty when " +
								"the controller reports none or the LAG type is not recognized.",
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"device_id": schema.StringAttribute{
										MarkdownDescription: "The UUID of the member device, for `LOCAL` and `MULTI_CHASSIS` " +
											"LAGs. Null for `SWITCH_STACK` LAGs.",
										Computed: true,
									},
									"unit_id": schema.Int64Attribute{
										MarkdownDescription: "The stack unit id of the member (see `units` on " +
											"`unifi_switch_stack`), for `SWITCH_STACK` LAGs. Null otherwise.",
										Computed: true,
									},
									"unit_mac_address": schema.StringAttribute{
										MarkdownDescription: "The MAC address of the member stack unit, for `SWITCH_STACK` LAGs. Null otherwise.",
										Computed:            true,
									},
									"port_idxs": schema.ListAttribute{
										MarkdownDescription: "The aggregated port indexes on the member device or unit.",
										Computed:            true,
										ElementType:         types.Int64Type,
									},
								},
							},
						},
						"json": schema.StringAttribute{
							MarkdownDescription: "The full LAG details serialized as a JSON string.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *switchLAGsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state switchLAGsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	lags, err := base.CollectAll(d.client.Official().Switching().ListLagsAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing switch LAGs", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.LAGs = make([]lagDetailsModel, 0, len(lags))
	for _, l := range lags {
		model, diags := mapLagDetails(l)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}
		state.LAGs = append(state.LAGs, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
