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
	_ datasource.DataSource              = &siteToSiteTunnelDataSource{}
	_ datasource.DataSourceWithConfigure = &siteToSiteTunnelDataSource{}
	_ base.Resource                      = &siteToSiteTunnelDataSource{}
)

type siteToSiteTunnelDataSource struct {
	officialBase
}

type siteToSiteTunnelModel struct {
	Site           types.String `tfsdk:"site"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	JSON           types.String `tfsdk:"json"`
}

// NewSiteToSiteTunnelDataSource returns the unifi_site_to_site_tunnel data
// source, which looks up a single site-to-site VPN tunnel by id from the
// Official API's read-only tunnel list.
func NewSiteToSiteTunnelDataSource() datasource.DataSource {
	return &siteToSiteTunnelDataSource{}
}

func (d *siteToSiteTunnelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_to_site_tunnel"
}

func (d *siteToSiteTunnelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a site-to-site VPN tunnel via the UniFi **Official API** (`integration/v1`). " +
			"Site-to-site tunnels are read-only on the Official API. Requires a controller running version 10.1.78 " +
			"or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the site-to-site VPN tunnel to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the site-to-site VPN tunnel.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The tunnel type (e.g. `IPSEC`, `OPENVPN`, `WIREGUARD`).",
				Computed:            true,
			},
			"metadata_origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the tunnel entity (e.g. `USER_DEFINED`, `DERIVED`).",
				Computed:            true,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full tunnel overview, including type-specific fields, serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *siteToSiteTunnelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state siteToSiteTunnelModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tunnels, err := base.CollectAll(d.client.Official().Supporting().ListSiteToSiteVpnTunnelsAll(ctx, siteID, ""))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing site-to-site VPN tunnels", err)...)
		return
	}

	id := state.ID.ValueString()
	tunnel, found := findByID(tunnels, id, func(t official.SiteToSiteVPNTunnelOverview) string { return t.Id.String() })
	if !found {
		resp.Diagnostics.AddError(
			"Site-to-site VPN tunnel not found",
			fmt.Sprintf("No site-to-site VPN tunnel with id %q was found on site %q.", id, siteName),
		)
		return
	}

	raw, err := json.Marshal(tunnel)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing site-to-site VPN tunnel", err.Error())
		return
	}

	state.Site = types.StringValue(siteName)
	state.ID = types.StringValue(tunnel.Id.String())
	state.Name = types.StringValue(tunnel.Name)
	state.Type = types.StringValue(tunnel.Type)
	state.MetadataOrigin = types.StringValue(tunnel.Metadata.Origin)
	state.JSON = types.StringValue(string(raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
