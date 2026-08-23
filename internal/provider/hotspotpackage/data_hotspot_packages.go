package hotspotpackage

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// HotspotPackagesDatasourceModel represents the data model for a list of UniFi hotspot packages.
type HotspotPackagesDatasourceModel struct {
	Site            types.String                    `tfsdk:"site"`
	HotspotPackages []HotspotPackageDatasourceModel `tfsdk:"hotspot_packages"`
}

func (m *HotspotPackagesDatasourceModel) GetSite() string          { return m.Site.ValueString() }
func (m *HotspotPackagesDatasourceModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *HotspotPackagesDatasourceModel) GetRawSite() types.String { return m.Site }

var (
	_ datasource.DataSource              = &hotspotPackagesDatasource{}
	_ datasource.DataSourceWithConfigure = &hotspotPackagesDatasource{}
	_ base.Resource                      = &hotspotPackagesDatasource{}
)

type hotspotPackagesDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewHotspotPackagesDatasource() datasource.DataSource {
	return &hotspotPackagesDatasource{}
}

func (d *hotspotPackagesDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *hotspotPackagesDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *hotspotPackagesDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *hotspotPackagesDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *hotspotPackagesDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_hotspot_packages"
}

func (d *hotspotPackagesDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot_packages` data source retrieves all hotspot packages configured on a site.",
		Attributes: map[string]schema.Attribute{
			"site": ut.SiteAttribute(),
			"hotspot_packages": schema.ListNestedAttribute{
				MarkdownDescription: "The list of hotspot packages.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                     ut.ID(),
						"site":                   ut.SiteAttribute(),
						"amount":                 schema.Float64Attribute{Computed: true, MarkdownDescription: "The price charged for the package, expressed in the package currency."},
						"trial_reset":            schema.Float64Attribute{Computed: true, MarkdownDescription: "The interval, in hours, after which a guest's free trial eligibility resets."},
						"hours":                  schema.Int32Attribute{Computed: true, MarkdownDescription: "The duration of the package, in hours, that access remains valid after purchase."},
						"index":                  schema.Int32Attribute{Computed: true, MarkdownDescription: "The ordering index of the package within the hotspot package list."},
						"limit_down":             schema.Int32Attribute{Computed: true, MarkdownDescription: "The download bandwidth limit, in kbps, applied to guests on this package."},
						"limit_quota":            schema.Int32Attribute{Computed: true, MarkdownDescription: "The total data transfer quota, in megabytes, allowed on this package."},
						"limit_up":               schema.Int32Attribute{Computed: true, MarkdownDescription: "The upload bandwidth limit, in kbps, applied to guests on this package."},
						"trial_duration_minutes": schema.Int32Attribute{Computed: true, MarkdownDescription: "The duration of the free trial, in minutes, offered by this package."},
						"charged_as":             schema.StringAttribute{Computed: true, MarkdownDescription: "The label describing how the package is charged."},
						"currency":               schema.StringAttribute{Computed: true, MarkdownDescription: "The ISO 4217 currency code used for the package price."},
						"name":                   schema.StringAttribute{Computed: true, MarkdownDescription: "The display name of the hotspot package."},

						"custom_payment_fields_enabled":      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether custom payment form fields are enabled for this package."},
						"limit_overwrite":                    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this package's limits overwrite the global guest limits."},
						"payment_fields_address_enabled":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the address field is shown on the payment form."},
						"payment_fields_address_required":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the address field is required on the payment form."},
						"payment_fields_city_enabled":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the city field is shown on the payment form."},
						"payment_fields_city_required":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the city field is required on the payment form."},
						"payment_fields_country_enabled":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the country field is shown on the payment form."},
						"payment_fields_country_required":    schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the country field is required on the payment form."},
						"payment_fields_email_enabled":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the email field is shown on the payment form."},
						"payment_fields_email_required":      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the email field is required on the payment form."},
						"payment_fields_first_name_enabled":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the first name field is shown on the payment form."},
						"payment_fields_first_name_required": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the first name field is required on the payment form."},
						"payment_fields_last_name_enabled":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the last name field is shown on the payment form."},
						"payment_fields_last_name_required":  schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the last name field is required on the payment form."},
						"payment_fields_state_enabled":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the state field is shown on the payment form."},
						"payment_fields_state_required":      schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the state field is required on the payment form."},
						"payment_fields_zip_enabled":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the ZIP/postal code field is shown on the payment form."},
						"payment_fields_zip_required":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the ZIP/postal code field is required on the payment form."},
					},
				},
			},
		},
	}
}

func (d *hotspotPackagesDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HotspotPackagesDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	hotspotPackages, err := d.client.ListHotspotPackage(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list hotspot packages", err.Error())
		return
	}

	state.HotspotPackages = make([]HotspotPackageDatasourceModel, 0, len(hotspotPackages))
	for i := range hotspotPackages {
		var item HotspotPackageDatasourceModel
		resp.Diagnostics.Append(item.Merge(ctx, &hotspotPackages[i])...)
		if resp.Diagnostics.HasError() {
			return
		}
		item.SetSite(site)
		state.HotspotPackages = append(state.HotspotPackages, item)
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
