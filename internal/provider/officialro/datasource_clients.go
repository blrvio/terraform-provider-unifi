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
	_ datasource.DataSource              = &clientsDataSource{}
	_ datasource.DataSourceWithConfigure = &clientsDataSource{}
	_ base.Resource                      = &clientsDataSource{}
)

type clientsDataSource struct {
	officialBase
}

type clientsModel struct {
	Site    types.String          `tfsdk:"site"`
	Filter  types.String          `tfsdk:"filter"`
	Clients []clientOverviewModel `tfsdk:"clients"`
}

type clientOverviewModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	IPAddress   types.String `tfsdk:"ip_address"`
	ConnectedAt types.String `tfsdk:"connected_at"`
	JSON        types.String `tfsdk:"json"`
}

// mapClientOverview converts an Official-API connected-client overview into its
// Framework model, serializing the full record (including the polymorphic
// `access` payload) into the json attribute.
func mapClientOverview(c official.ClientOverview) (clientOverviewModel, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return clientOverviewModel{}, err
	}
	return clientOverviewModel{
		ID:          types.StringValue(c.Id.String()),
		Name:        types.StringValue(c.Name),
		Type:        types.StringValue(c.Type),
		IPAddress:   stringPtrValue(c.IpAddress),
		ConnectedAt: timePtrValue(c.ConnectedAt),
		JSON:        types.StringValue(string(raw)),
	}, nil
}

// NewClientsDataSource returns the unifi_clients data source, which lists the
// clients currently connected to a site (wired, wireless, VPN and teleport) via
// the Official API.
func NewClientsDataSource() datasource.DataSource {
	return &clientsDataSource{}
}

func (d *clientsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clients"
}

func clientAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The UUID of the connected client.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The client name.",
			Computed:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "The connection type (e.g. `WIRED`, `WIRELESS`, `VPN`, `TELEPORT`).",
			Computed:            true,
		},
		"ip_address": schema.StringAttribute{
			MarkdownDescription: "The client IP address, if known.",
			Computed:            true,
		},
		"connected_at": schema.StringAttribute{
			MarkdownDescription: "When the client connected, in RFC 3339 format.",
			Computed:            true,
		},
		"json": schema.StringAttribute{
			MarkdownDescription: "The full client record, including the type-specific `access` payload, serialized as a JSON string.",
			Computed:            true,
		},
	}
}

func (d *clientsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the clients currently connected to a site (wired, wireless, VPN and teleport) via the " +
			"UniFi **Official API** (`integration/v1`). Read-only. Requires a controller running version 10.1.78 or " +
			"later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("connected clients"),
			"clients": schema.ListNestedAttribute{
				MarkdownDescription: "The connected clients.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: clientAttributes(),
				},
			},
		},
	}
}

func (d *clientsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clientsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clients, err := base.CollectAll(d.client.Official().Clients().ListConnectedAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing connected clients", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.Clients = make([]clientOverviewModel, 0, len(clients))
	for _, c := range clients {
		model, err := mapClientOverview(c)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing connected client", err.Error())
			return
		}
		state.Clients = append(state.Clients, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
