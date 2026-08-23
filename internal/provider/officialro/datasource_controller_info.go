package officialro

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &controllerInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &controllerInfoDataSource{}
	_ base.Resource                      = &controllerInfoDataSource{}
)

type controllerInfoDataSource struct {
	officialBase
}

type controllerInfoModel struct {
	ApplicationVersion types.String `tfsdk:"application_version"`
}

// NewControllerInfoDataSource returns the unifi_controller_info data source, which
// reports the controller application info via the Official API. Site-independent.
func NewControllerInfoDataSource() datasource.DataSource {
	return &controllerInfoDataSource{}
}

func (d *controllerInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_controller_info"
}

func (d *controllerInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reports the UniFi controller application info via the **Official API** (`integration/v1`). " +
			"Read-only and site-independent. Requires a controller running version 10.1.78 or later with API-key " +
			"authentication.",
		Attributes: map[string]schema.Attribute{
			"application_version": schema.StringAttribute{
				MarkdownDescription: "The controller application version (e.g. `10.5.67`).",
				Computed:            true,
			},
		},
	}
}

func (d *controllerInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	resp.Diagnostics.Append(d.prepareVersionOnly(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	info, err := d.client.Official().Info().Get(ctx)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading controller info", err)...)
		return
	}

	state := controllerInfoModel{
		ApplicationVersion: types.StringValue(info.ApplicationVersion),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
