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
	_ datasource.DataSource              = &networksDataSource{}
	_ datasource.DataSourceWithConfigure = &networksDataSource{}
	_ base.Resource                      = &networksDataSource{}
)

type networksDataSource struct {
	officialBase
}

type networksModel struct {
	Site     types.String           `tfsdk:"site"`
	Filter   types.String           `tfsdk:"filter"`
	Networks []networkOverviewModel `tfsdk:"networks"`
}

type networkOverviewModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Default        types.Bool   `tfsdk:"default"`
	Management     types.String `tfsdk:"management"`
	VlanID         types.Int64  `tfsdk:"vlan_id"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	JSON           types.String `tfsdk:"json"`
}

// mapNetworkOverview converts an Official-API network overview into its Framework
// model, serializing the full record into the json attribute.
func mapNetworkOverview(n official.NetworkOverview) (networkOverviewModel, error) {
	raw, err := json.Marshal(n)
	if err != nil {
		return networkOverviewModel{}, err
	}
	return networkOverviewModel{
		ID:             types.StringValue(n.Id.String()),
		Name:           types.StringValue(n.Name),
		Enabled:        types.BoolValue(n.Enabled),
		Default:        types.BoolValue(n.Default),
		Management:     types.StringValue(n.Management),
		VlanID:         types.Int64Value(int64(n.VlanId)),
		MetadataOrigin: types.StringValue(n.Metadata.Origin),
		JSON:           types.StringValue(string(raw)),
	}, nil
}

// NewNetworksDataSource returns the unifi_networks data source, which lists the
// networks on a site via the Official API (read-only). Writable network
// management remains the internal unifi_network resource.
func NewNetworksDataSource() datasource.DataSource {
	return &networksDataSource{}
}

func (d *networksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networks"
}

func (d *networksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the networks on a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only — writable network management is the internal `unifi_network` resource. Requires a controller " +
			"running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("networks"),
			"networks": schema.ListNestedAttribute{
				MarkdownDescription: "The networks defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The UUID of the network.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The network name.",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the network is enabled.",
							Computed:            true,
						},
						"default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default network.",
							Computed:            true,
						},
						"management": schema.StringAttribute{
							MarkdownDescription: "The network management mode.",
							Computed:            true,
						},
						"vlan_id": schema.Int64Attribute{
							MarkdownDescription: "The VLAN ID.",
							Computed:            true,
						},
						"metadata_origin": schema.StringAttribute{
							MarkdownDescription: "The origin of the network entity (e.g. `USER_DEFINED`, `SYSTEM_DEFINED`).",
							Computed:            true,
						},
						"json": schema.StringAttribute{
							MarkdownDescription: "The full network overview serialized as a JSON string.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *networksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networksModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	networks, err := base.CollectAll(d.client.Official().Networks().ListAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing networks", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.Networks = make([]networkOverviewModel, 0, len(networks))
	for _, n := range networks {
		model, err := mapNetworkOverview(n)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing network", err.Error())
			return
		}
		state.Networks = append(state.Networks, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
