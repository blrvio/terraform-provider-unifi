package officialro

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &dpiApplicationCategoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &dpiApplicationCategoriesDataSource{}
	_ base.Resource                      = &dpiApplicationCategoriesDataSource{}
)

type dpiApplicationCategoriesDataSource struct {
	officialBase
}

type dpiApplicationCategoriesModel struct {
	Filter     types.String      `tfsdk:"filter"`
	Categories []dpiCatalogEntry `tfsdk:"categories"`
}

// mapDPICategory converts an Official-API DPI category into its model.
func mapDPICategory(c official.DPICategory) dpiCatalogEntry {
	return dpiCatalogEntry{
		ID:   types.Int64Value(int64(c.Id)),
		Name: types.StringValue(c.Name),
	}
}

// NewDPIApplicationCategoriesDataSource returns the unifi_dpi_application_categories
// data source, which lists the DPI category catalog via the Official API.
// Site-independent.
func NewDPIApplicationCategoriesDataSource() datasource.DataSource {
	return &dpiApplicationCategoriesDataSource{}
}

func (d *dpiApplicationCategoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dpi_application_categories"
}

func (d *dpiApplicationCategoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the DPI (deep packet inspection) category catalog via the UniFi **Official API** " +
			"(`integration/v1`). Read-only and site-independent. Requires a controller running version 10.1.78 or later " +
			"with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"filter": filterAttribute("DPI categories"),
			"categories": schema.ListNestedAttribute{
				MarkdownDescription: "The DPI categories.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dpiCatalogAttributes("category"),
				},
			},
		},
	}
}

func (d *dpiApplicationCategoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dpiApplicationCategoriesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(d.prepareVersionOnly(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cats, err := base.CollectAll(d.client.Official().Supporting().ListDpiApplicationCategoriesAll(ctx, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing DPI categories", err)...)
		return
	}

	state.Categories = make([]dpiCatalogEntry, 0, len(cats))
	for _, c := range cats {
		state.Categories = append(state.Categories, mapDPICategory(c))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
