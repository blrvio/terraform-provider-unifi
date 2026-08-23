package officialro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &switchLAGDataSource{}
	_ datasource.DataSourceWithConfigure = &switchLAGDataSource{}
	_ base.Resource                      = &switchLAGDataSource{}
)

type switchLAGDataSource struct {
	officialBase
}

type switchLAGModel struct {
	Site           types.String `tfsdk:"site"`
	ID             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	MembersJSON    types.String `tfsdk:"members_json"`
	JSON           types.String `tfsdk:"json"`
}

// NewSwitchLAGDataSource returns the unifi_switch_lag data source, which fetches a
// single switch link-aggregation group by id via the Official API's read-only
// LAG endpoint.
func NewSwitchLAGDataSource() datasource.DataSource {
	return &switchLAGDataSource{}
}

func (d *switchLAGDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_lag"
}

func (d *switchLAGDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a switch link-aggregation group (LAG) via the UniFi **Official API** " +
			"(`integration/v1`). LAGs are read-only on the Official API. Requires a controller running version " +
			"10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the LAG to look up.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The LAG type (e.g. `LOCAL`, `MC_LAG`, `SWITCH_STACK`).",
				Computed:            true,
			},
			"metadata_origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the LAG entity (e.g. `USER_DEFINED`).",
				Computed:            true,
			},
			"members_json": schema.StringAttribute{
				MarkdownDescription: "The LAG members (device id and aggregated port indexes) serialized as a JSON string.",
				Computed:            true,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full LAG details, including type-specific fields, serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *switchLAGDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state switchLAGModel
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
			"Invalid LAG id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.ID.ValueString(), err),
		)
		return
	}

	lag, err := d.client.Official().Switching().GetLag(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Switch LAG not found",
				fmt.Sprintf("No LAG with id %q was found on site %q.", state.ID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading switch LAG", err)...)
		return
	}

	raw, err := json.Marshal(lag)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing switch LAG", err.Error())
		return
	}
	members, err := json.Marshal(lag.Members)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing switch LAG members", err.Error())
		return
	}

	state.Site = types.StringValue(siteName)
	state.ID = types.StringValue(lag.Id.String())
	state.Type = types.StringValue(lag.Type)
	state.MetadataOrigin = types.StringValue(lag.Metadata.Origin)
	state.MembersJSON = types.StringValue(string(members))
	state.JSON = types.StringValue(string(raw))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
