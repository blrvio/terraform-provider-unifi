package officialro

import (
	"context"
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
	_ datasource.DataSource              = &switchStackDataSource{}
	_ datasource.DataSourceWithConfigure = &switchStackDataSource{}
	_ base.Resource                      = &switchStackDataSource{}
)

type switchStackDataSource struct {
	officialBase
}

type switchStackModel struct {
	Site            types.String `tfsdk:"site"`
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	MetadataOrigin  types.String `tfsdk:"metadata_origin"`
	MemberDeviceIDs []string     `tfsdk:"member_device_ids"`
	JSON            types.String `tfsdk:"json"`
}

// NewSwitchStackDataSource returns the unifi_switch_stack data source, which
// fetches a single switch stack by id via the Official API.
func NewSwitchStackDataSource() datasource.DataSource {
	return &switchStackDataSource{}
}

func (d *switchStackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_stack"
}

func (d *switchStackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single switch stack by id via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the switch stack to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The switch stack name.",
				Computed:            true,
			},
			"metadata_origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the switch stack entity (e.g. `USER_DEFINED`).",
				Computed:            true,
			},
			"member_device_ids": schema.ListAttribute{
				MarkdownDescription: "The UUIDs of the devices that are members of the stack.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full switch stack record, including nested lags, serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *switchStackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state switchStackModel
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
			"Invalid switch stack id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.ID.ValueString(), err),
		)
		return
	}

	stack, err := d.client.Official().Switching().GetSwitchStack(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Switch stack not found",
				fmt.Sprintf("No switch stack with id %q was found on site %q.", state.ID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading switch stack", err)...)
		return
	}

	item, err := mapSwitchStack(*stack)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing switch stack", err.Error())
		return
	}
	state.Site = types.StringValue(siteName)
	state.ID = item.ID
	state.Name = item.Name
	state.MetadataOrigin = item.MetadataOrigin
	state.MemberDeviceIDs = item.MemberDeviceIDs
	state.JSON = item.JSON
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
