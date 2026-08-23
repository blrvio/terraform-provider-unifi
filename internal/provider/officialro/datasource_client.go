package officialro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &clientDataSource{}
	_ datasource.DataSourceWithConfigure = &clientDataSource{}
	_ base.Resource                      = &clientDataSource{}
)

type clientDataSource struct {
	officialBase
}

type clientModel struct {
	Site        types.String `tfsdk:"site"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	IPAddress   types.String `tfsdk:"ip_address"`
	ConnectedAt types.String `tfsdk:"connected_at"`
	JSON        types.String `tfsdk:"json"`
}

// mapClientDetails converts an Official-API connected-client detail record into
// its Framework model, serializing the full record (including the polymorphic
// `access` payload) into the json attribute.
func mapClientDetails(c official.ClientDetails) (clientModel, error) {
	raw, err := json.Marshal(c)
	if err != nil {
		return clientModel{}, err
	}
	return clientModel{
		ID:          types.StringValue(c.Id.String()),
		Name:        types.StringValue(c.Name),
		Type:        types.StringValue(c.Type),
		IPAddress:   stringPtrValue(c.IpAddress),
		ConnectedAt: timePtrValue(c.ConnectedAt),
		JSON:        types.StringValue(string(raw)),
	}, nil
}

// NewClientDataSource returns the unifi_client data source, which fetches a single
// connected client by id via the Official API.
func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single connected client by id via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. This surface has no MAC field, so lookup is by client UUID only. Requires a controller running " +
			"version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the connected client to look up.",
				Required:            true,
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
		},
	}
}

func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clientModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid client id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.ID.ValueString(), err),
		)
		return
	}

	client, err := d.client.Official().Clients().GetConnected(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Client not found",
				fmt.Sprintf("No connected client with id %q was found on site %q.", state.ID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading connected client", err)...)
		return
	}

	model, err := mapClientDetails(*client)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing connected client", err.Error())
		return
	}
	model.Site = types.StringValue(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
