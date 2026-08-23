package officialro

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &deviceTagDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceTagDataSource{}
	_ base.Resource                      = &deviceTagDataSource{}
)

type deviceTagDataSource struct {
	officialBase
}

type deviceTagModel struct {
	Site           types.String `tfsdk:"site"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	DeviceIDs      types.List   `tfsdk:"device_ids"`
}

// NewDeviceTagDataSource returns the unifi_device_tag data source, which looks up
// a single device tag by id from the Official API's read-only device-tag list.
func NewDeviceTagDataSource() datasource.DataSource {
	return &deviceTagDataSource{}
}

func (d *deviceTagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_tag"
}

func (d *deviceTagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a device tag via the UniFi **Official API** (`integration/v1`). " +
			"Device tags are read-only on the Official API. Requires a controller running version 10.1.78 " +
			"or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the device tag to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the device tag.",
				Computed:            true,
			},
			"metadata_origin": schema.StringAttribute{
				MarkdownDescription: "The origin of the device tag entity (e.g. `USER_DEFINED`, `ORCHESTRATED`).",
				Computed:            true,
			},
			"device_ids": schema.ListAttribute{
				MarkdownDescription: "The UUIDs of the devices assigned to this tag.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *deviceTagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceTagModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := base.CollectAll(d.client.Official().Supporting().ListDeviceTagsAll(ctx, siteID, ""))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing device tags", err)...)
		return
	}

	id := state.ID.ValueString()
	tag, found := findByID(tags, id, func(t official.DeviceTag) string { return t.Id.String() })
	if !found {
		resp.Diagnostics.AddError(
			"Device tag not found",
			fmt.Sprintf("No device tag with id %q was found on site %q.", id, siteName),
		)
		return
	}

	deviceIDs, listDiags := types.ListValueFrom(ctx, types.StringType, uuidsToStrings(tag.DeviceIds))
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Site = types.StringValue(siteName)
	state.ID = types.StringValue(tag.Id.String())
	state.Name = types.StringValue(tag.Name)
	state.MetadataOrigin = types.StringValue(tag.Metadata.Origin)
	state.DeviceIDs = deviceIDs
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
