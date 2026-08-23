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
	_ datasource.DataSource              = &countriesDataSource{}
	_ datasource.DataSourceWithConfigure = &countriesDataSource{}
	_ base.Resource                      = &countriesDataSource{}
)

type countriesDataSource struct {
	officialBase
}

type countriesModel struct {
	Filter    types.String   `tfsdk:"filter"`
	Countries []countryEntry `tfsdk:"countries"`
}

type countryEntry struct {
	Code types.String `tfsdk:"code"`
	Name types.String `tfsdk:"name"`
}

// mapCountry converts an Official-API country definition into its model.
func mapCountry(c official.CountryDefinition) countryEntry {
	return countryEntry{
		Code: types.StringValue(c.Code),
		Name: types.StringValue(c.Name),
	}
}

// NewCountriesDataSource returns the unifi_countries data source, which lists the
// country/code definitions known to the controller via the Official API.
// Site-independent.
func NewCountriesDataSource() datasource.DataSource {
	return &countriesDataSource{}
}

func (d *countriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_countries"
}

func (d *countriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the country/code definitions known to the controller via the UniFi **Official API** " +
			"(`integration/v1`). Read-only and site-independent. Requires a controller running version 10.1.78 or later " +
			"with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"filter": filterAttribute("countries"),
			"countries": schema.ListNestedAttribute{
				MarkdownDescription: "The country definitions.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							MarkdownDescription: "The ISO 3166-1 alpha-2 country code.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The country name.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *countriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state countriesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(d.prepareVersionOnly(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	countries, err := base.CollectAll(d.client.Official().Supporting().ListCountriesAll(ctx, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing countries", err)...)
		return
	}

	state.Countries = make([]countryEntry, 0, len(countries))
	for _, c := range countries {
		state.Countries = append(state.Countries, mapCountry(c))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
