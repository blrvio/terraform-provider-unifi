package feature

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// featuresDatasourceModel represents the full list of described features on a site.
type featuresDatasourceModel struct {
	Site     types.String             `tfsdk:"site"`
	Features []featureDatasourceModel `tfsdk:"features"`
}

func (m *featuresDatasourceModel) GetSite() string          { return m.Site.ValueString() }
func (m *featuresDatasourceModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *featuresDatasourceModel) GetRawSite() types.String { return m.Site }

var (
	_ datasource.DataSource              = &featuresDatasource{}
	_ datasource.DataSourceWithConfigure = &featuresDatasource{}
	_ base.Resource                      = &featuresDatasource{}
)

type featuresDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewFeaturesDatasource() datasource.DataSource {
	return &featuresDatasource{}
}

func (d *featuresDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *featuresDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *featuresDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *featuresDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *featuresDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_features"
}

func (d *featuresDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_features` data source lists all controller features and their " +
			"availability on a site.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"features": schema.ListNestedAttribute{
				MarkdownDescription: "The list of controller features.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"site":           ut.SiteAttribute(),
						"name":           schema.StringAttribute{Computed: true, MarkdownDescription: "The name of the feature."},
						"feature_exists": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the feature exists/is available."},
					},
				},
			},
		},
	}
}

func (d *featuresDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state featuresDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	features, err := d.client.ListFeatures(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list features", err.Error())
		return
	}

	state.Features = make([]featureDatasourceModel, 0, len(features))
	for i := range features {
		state.Features = append(state.Features, featureDatasourceModel{
			Site:          types.StringValue(site),
			Name:          types.StringValue(features[i].Name),
			FeatureExists: types.BoolValue(features[i].FeatureExists),
		})
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
