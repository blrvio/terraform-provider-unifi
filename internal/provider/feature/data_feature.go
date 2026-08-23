package feature

import (
	"context"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// featureDatasourceModel represents a single described feature of the controller.
type featureDatasourceModel struct {
	Site          types.String `tfsdk:"site"`
	Name          types.String `tfsdk:"name"`
	FeatureExists types.Bool   `tfsdk:"feature_exists"`
}

func (m *featureDatasourceModel) GetSite() string          { return m.Site.ValueString() }
func (m *featureDatasourceModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *featureDatasourceModel) GetRawSite() types.String { return m.Site }

var (
	_ datasource.DataSource              = &featureDatasource{}
	_ datasource.DataSourceWithConfigure = &featureDatasource{}
	_ base.Resource                      = &featureDatasource{}
)

type featureDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewFeatureDatasource() datasource.DataSource {
	return &featureDatasource{}
}

func (d *featureDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *featureDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *featureDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *featureDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *featureDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_feature"
}

func (d *featureDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_feature` data source reports whether a named controller feature is " +
			"available on a site. Feature names are case-insensitive.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the feature to look up (case-insensitive).",
				Required:            true,
			},
			"feature_exists": schema.BoolAttribute{
				MarkdownDescription: "Whether the feature exists/is available on the controller.",
				Computed:            true,
			},
		},
	}
}

func (d *featureDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state featureDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	name := state.Name.ValueString()
	feature, err := d.client.GetFeature(ctx, site, name)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError("Feature not found", fmt.Sprintf("No feature named %q was found on site %q", name, site))
			return
		}
		resp.Diagnostics.AddError("Failed to get feature", err.Error())
		return
	}

	state.Name = types.StringValue(feature.Name)
	state.FeatureExists = types.BoolValue(feature.FeatureExists)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
