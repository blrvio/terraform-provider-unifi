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
	_ datasource.DataSource              = &dpiApplicationsDataSource{}
	_ datasource.DataSourceWithConfigure = &dpiApplicationsDataSource{}
	_ base.Resource                      = &dpiApplicationsDataSource{}
)

type dpiApplicationsDataSource struct {
	officialBase
}

type dpiApplicationsModel struct {
	Filter       types.String      `tfsdk:"filter"`
	Applications []dpiCatalogEntry `tfsdk:"applications"`
}

// dpiCatalogEntry is the shared {id, name} model for the DPI application and
// category catalogs, which have identical shapes.
type dpiCatalogEntry struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// mapDPIApplication converts an Official-API DPI application into its model.
func mapDPIApplication(a official.DPIApplication) dpiCatalogEntry {
	return dpiCatalogEntry{
		ID:   types.Int64Value(int64(a.Id)),
		Name: types.StringValue(a.Name),
	}
}

// NewDPIApplicationsDataSource returns the unifi_dpi_applications data source,
// which lists the DPI application catalog via the Official API. Site-independent.
func NewDPIApplicationsDataSource() datasource.DataSource {
	return &dpiApplicationsDataSource{}
}

func (d *dpiApplicationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dpi_applications"
}

func dpiCatalogAttributes(kind string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "The numeric id of the DPI " + kind + ".",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The name of the DPI " + kind + ".",
			Computed:            true,
		},
	}
}

func (d *dpiApplicationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the DPI (deep packet inspection) application catalog via the UniFi **Official API** " +
			"(`integration/v1`). Read-only and site-independent. Requires a controller running version 10.1.78 or later " +
			"with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"filter": filterAttribute("DPI applications"),
			"applications": schema.ListNestedAttribute{
				MarkdownDescription: "The DPI applications.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: dpiCatalogAttributes("application"),
				},
			},
		},
	}
}

func (d *dpiApplicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dpiApplicationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(d.prepareVersionOnly(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apps, err := base.CollectAll(d.client.Official().Supporting().ListDpiApplicationsAll(ctx, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing DPI applications", err)...)
		return
	}

	state.Applications = make([]dpiCatalogEntry, 0, len(apps))
	for _, a := range apps {
		state.Applications = append(state.Applications, mapDPIApplication(a))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
