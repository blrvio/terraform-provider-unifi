package officialro

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &vpnServersDataSource{}
	_ datasource.DataSourceWithConfigure = &vpnServersDataSource{}
	_ base.Resource                      = &vpnServersDataSource{}
)

type vpnServersDataSource struct {
	officialBase
}

type vpnServerItem struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
}

type vpnServersModel struct {
	Site    types.String    `tfsdk:"site"`
	Filter  types.String    `tfsdk:"filter"`
	Servers []vpnServerItem `tfsdk:"vpn_servers"`
}

// NewVPNServersDataSource returns the unifi_vpn_servers data source, which lists
// the VPN servers of a site via the Official API.
func NewVPNServersDataSource() datasource.DataSource {
	return &vpnServersDataSource{}
}

func (d *vpnServersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_servers"
}

func (d *vpnServersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the VPN servers of a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication. " +
			"The `id` is the Official API UUID (used e.g. as a member of the Vpn firewall zone); for a server " +
			"managed as `unifi_network`, the same value is its `official_id`.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("VPN servers"),
			"vpn_servers": schema.ListNestedAttribute{
				MarkdownDescription: "The VPN servers defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{MarkdownDescription: "The Official API UUID of the VPN server.", Computed: true},
						"name":            schema.StringAttribute{MarkdownDescription: "The name of the VPN server.", Computed: true},
						"type":            schema.StringAttribute{MarkdownDescription: "The VPN server type (e.g. `WIREGUARD`, `OPENVPN`).", Computed: true},
						"enabled":         schema.BoolAttribute{MarkdownDescription: "Whether the VPN server is enabled.", Computed: true},
						"metadata_origin": schema.StringAttribute{MarkdownDescription: "The origin of the entity (`USER_DEFINED` or `DERIVED`).", Computed: true},
					},
				},
			},
		},
	}
}

func (d *vpnServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vpnServersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	servers, err := base.CollectAll(d.client.Official().Supporting().ListVpnServersAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing VPN servers", err)...)
		return
	}
	state.Site = types.StringValue(siteName)
	state.Servers = make([]vpnServerItem, 0, len(servers))
	for _, s := range servers {
		state.Servers = append(state.Servers, vpnServerItem{
			ID:             types.StringValue(s.Id.String()),
			Name:           types.StringValue(s.Name),
			Type:           types.StringValue(s.Type),
			Enabled:        types.BoolValue(s.Enabled),
			MetadataOrigin: types.StringValue(s.Metadata.Origin),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
