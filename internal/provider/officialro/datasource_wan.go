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
	_ datasource.DataSource              = &wanDataSource{}
	_ datasource.DataSourceWithConfigure = &wanDataSource{}
	_ base.Resource                      = &wanDataSource{}
)

type wanDataSource struct {
	officialBase
}

type wanModel struct {
	Site types.String `tfsdk:"site"`
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// NewWANDataSource returns the unifi_wan data source, which looks up a single WAN
// interface by id from the Official API's read-only WAN list.
func NewWANDataSource() datasource.DataSource {
	return &wanDataSource{}
}

func (d *wanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wan"
}

func (d *wanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a WAN interface via the UniFi **Official API** (`integration/v1`). " +
			"WAN interfaces are read-only on the Official API. Requires a controller running version 10.1.78 " +
			"or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the WAN interface to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the WAN interface.",
				Computed:            true,
			},
		},
	}
}

func (d *wanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state wanModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	wans, err := base.CollectAll(d.client.Official().Supporting().ListWansAll(ctx, siteID, ""))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing WAN interfaces", err)...)
		return
	}

	id := state.ID.ValueString()
	wan, found := findByID(wans, id, func(w official.WANOverview) string { return w.Id.String() })
	if !found {
		resp.Diagnostics.AddError(
			"WAN interface not found",
			fmt.Sprintf("No WAN interface with id %q was found on site %q.", id, siteName),
		)
		return
	}

	state.Site = types.StringValue(siteName)
	state.ID = types.StringValue(wan.Id.String())
	state.Name = types.StringValue(wan.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
