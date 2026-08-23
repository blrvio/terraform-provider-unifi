package officialro

import (
	"context"
	"encoding/json"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

type lagMemberModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	PortIdxs []int64      `tfsdk:"port_idxs"`
}

// mapLagMembers converts Official-API LAG members into their Framework models.
func mapLagMembers(members []official.LagMember) []lagMemberModel {
	out := make([]lagMemberModel, 0, len(members))
	for _, m := range members {
		idxs := make([]int64, 0, len(m.PortIdxs))
		for _, idx := range m.PortIdxs {
			idxs = append(idxs, int64(idx))
		}
		out = append(out, lagMemberModel{
			DeviceID: types.StringValue(m.DeviceId.String()),
			PortIdxs: idxs,
		})
	}
	return out
}

// mapLagDetails converts an Official-API LAG record into its Framework model,
// serializing the full record into the json attribute.
func mapLagDetails(l official.LAGDetails) (lagDetailsModel, error) {
	raw, err := json.Marshal(l)
	if err != nil {
		return lagDetailsModel{}, err
	}
	return lagDetailsModel{
		ID:             types.StringValue(l.Id.String()),
		Type:           types.StringValue(l.Type),
		MetadataOrigin: types.StringValue(l.Metadata.Origin),
		Members:        mapLagMembers(l.Members),
		JSON:           types.StringValue(string(raw)),
	}, nil
}

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
							MarkdownDescription: "The LAG type (e.g. `LOCAL`, `MC_LAG`, `SWITCH_STACK`).",
							Computed:            true,
						},
						"metadata_origin": schema.StringAttribute{
							MarkdownDescription: "The origin of the LAG entity (e.g. `USER_DEFINED`).",
							Computed:            true,
						},
						"members": schema.ListNestedAttribute{
							MarkdownDescription: "The LAG members.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"device_id": schema.StringAttribute{
										MarkdownDescription: "The UUID of the member device.",
										Computed:            true,
									},
									"port_idxs": schema.ListAttribute{
										MarkdownDescription: "The aggregated port indexes on the member device.",
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
		model, err := mapLagDetails(l)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing switch LAG", err.Error())
			return
		}
		state.LAGs = append(state.LAGs, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
