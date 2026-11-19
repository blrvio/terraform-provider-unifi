package officialro

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &wansDataSource{}
	_ datasource.DataSourceWithConfigure = &wansDataSource{}
	_ base.Resource                      = &wansDataSource{}
)

type wansDataSource struct {
	officialBase
}

type wanItem struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type wansModel struct {
	Site   types.String `tfsdk:"site"`
	Filter types.String `tfsdk:"filter"`
	WANs   []wanItem    `tfsdk:"wans"`
}

// NewWANsDataSource returns the unifi_wans data source, which lists the WAN
// interfaces of a site (Official API UUID and name) so a configuration can look
// one up by name instead of hard-coding its UUID.
func NewWANsDataSource() datasource.DataSource {
	return &wansDataSource{}
}

func (d *wansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wans"
}

func (d *wansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the WAN interfaces of a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication. " +
			"The `id` is the Official API UUID (used e.g. as a member of the External firewall zone); the same " +
			"value is exposed as `official_id` on `unifi_network`.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("WAN interfaces"),
			"wans": schema.ListNestedAttribute{
				MarkdownDescription: "The WAN interfaces defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{MarkdownDescription: "The Official API UUID of the WAN interface.", Computed: true},
						"name": schema.StringAttribute{MarkdownDescription: "The name of the WAN interface.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *wansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state wansModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	wans, err := base.CollectAll(d.client.Official().Supporting().ListWansAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing WAN interfaces", err)...)
		return
	}
	state.Site = types.StringValue(siteName)
	state.WANs = make([]wanItem, 0, len(wans))
	for _, w := range wans {
		state.WANs = append(state.WANs, wanItem{ID: types.StringValue(w.Id.String()), Name: types.StringValue(w.Name)})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
