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
	_ datasource.DataSource              = &radiusProfilesDataSource{}
	_ datasource.DataSourceWithConfigure = &radiusProfilesDataSource{}
	_ base.Resource                      = &radiusProfilesDataSource{}
)

type radiusProfilesDataSource struct {
	officialBase
}

type radiusProfilesModel struct {
	Site     types.String                 `tfsdk:"site"`
	Filter   types.String                 `tfsdk:"filter"`
	Profiles []radiusProfileOverviewModel `tfsdk:"profiles"`
}

type radiusProfileOverviewModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
}

// mapRadiusProfile converts an Official-API RADIUS profile overview into its model.
func mapRadiusProfile(p official.RadiusProfileOverview) radiusProfileOverviewModel {
	return radiusProfileOverviewModel{
		ID:             types.StringValue(p.Id.String()),
		Name:           types.StringValue(p.Name),
		MetadataOrigin: types.StringValue(p.Metadata.Origin),
	}
}

// NewRadiusProfilesDataSource returns the unifi_radius_profiles data source, which
// lists RADIUS profiles on a site via the Official API. Named distinctly from the
// internal unifi_radius_profile resource.
func NewRadiusProfilesDataSource() datasource.DataSource {
	return &radiusProfilesDataSource{}
}

func (d *radiusProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radius_profiles"
}

func (d *radiusProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists RADIUS profiles on a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Named distinctly from the internal `unifi_radius_profile` resource. Requires a controller " +
			"running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("RADIUS profiles"),
			"profiles": schema.ListNestedAttribute{
				MarkdownDescription: "The RADIUS profiles defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The UUID of the RADIUS profile.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The RADIUS profile name.",
							Computed:            true,
						},
						"metadata_origin": schema.StringAttribute{
							MarkdownDescription: "The origin of the RADIUS profile entity (e.g. `USER_DEFINED`, `DERIVED`).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *radiusProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state radiusProfilesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	profiles, err := base.CollectAll(d.client.Official().Supporting().ListRadiusProfilesAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing RADIUS profiles", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.Profiles = make([]radiusProfileOverviewModel, 0, len(profiles))
	for _, p := range profiles {
		state.Profiles = append(state.Profiles, mapRadiusProfile(p))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
