package officialro

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &vpnServerDataSource{}
	_ datasource.DataSourceWithConfigure = &vpnServerDataSource{}
	_ base.Resource                      = &vpnServerDataSource{}
)

type vpnServerDataSource struct {
	officialBase
}

type vpnServerModel struct {
	Site           types.String `tfsdk:"site"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	JSON           types.String `tfsdk:"json"`
}

// NewVPNServerDataSource returns the unifi_vpn_server data source, which looks up
// a single VPN server by id from the Official API's read-only VPN server list.
func NewVPNServerDataSource() datasource.DataSource {
	return &vpnServerDataSource{}
}

func (d *vpnServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_server"
}

func (d *vpnServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a VPN server via the UniFi **Official API** (`integration/v1`). " +
			"VPN servers are read-only on the Official API. Requires a controller running version 10.1.78 " +
			"or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the VPN server to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the VPN server.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The VPN server type (e.g. `WIREGUARD_SERVER`, `OPENVPN_SERVER`, `L2TP_SERVER`, `PPTP_SERVER`).",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the VPN server is enabled.",
				Computed:            true,
			},
			"metadata_origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the VPN server entity (e.g. `USER_DEFINED`, `DERIVED`).",
				Computed:            true,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full VPN server overview, including type-specific fields, serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *vpnServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vpnServerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	servers, err := base.CollectAll(d.client.Official().Supporting().ListVpnServersAll(ctx, siteID, ""))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing VPN servers", err)...)
		return
	}

	id := state.ID.ValueString()
	server, found := findByID(servers, id, func(s official.VPNServerOverview) string { return s.Id.String() })
	if !found {
		resp.Diagnostics.AddError(
			"VPN server not found",
			fmt.Sprintf("No VPN server with id %q was found on site %q.", id, siteName),
		)
		return
	}

	raw, err := json.Marshal(server)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing VPN server", err.Error())
		return
	}

	state.Site = types.StringValue(siteName)
	state.ID = types.StringValue(server.Id.String())
	state.Name = types.StringValue(server.Name)
	state.Type = types.StringValue(server.Type)
	state.Enabled = types.BoolValue(server.Enabled)
	state.MetadataOrigin = types.StringValue(server.Metadata.Origin)
	state.JSON = types.StringValue(string(raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
