package hotspotpackage

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// HotspotPackageDatasourceModel represents the data model for a UniFi hotspot package data source.
type HotspotPackageDatasourceModel struct {
	base.Model
	Amount                         types.Float64 `tfsdk:"amount"`
	TrialReset                     types.Float64 `tfsdk:"trial_reset"`
	Hours                          types.Int32   `tfsdk:"hours"`
	Index                          types.Int32   `tfsdk:"index"`
	LimitDown                      types.Int32   `tfsdk:"limit_down"`
	LimitQuota                     types.Int32   `tfsdk:"limit_quota"`
	LimitUp                        types.Int32   `tfsdk:"limit_up"`
	TrialDurationMinutes           types.Int32   `tfsdk:"trial_duration_minutes"`
	ChargedAs                      types.String  `tfsdk:"charged_as"`
	Currency                       types.String  `tfsdk:"currency"`
	Name                           types.String  `tfsdk:"name"`
	CustomPaymentFieldsEnabled     types.Bool    `tfsdk:"custom_payment_fields_enabled"`
	LimitOverwrite                 types.Bool    `tfsdk:"limit_overwrite"`
	PaymentFieldsAddressEnabled    types.Bool    `tfsdk:"payment_fields_address_enabled"`
	PaymentFieldsAddressRequired   types.Bool    `tfsdk:"payment_fields_address_required"`
	PaymentFieldsCityEnabled       types.Bool    `tfsdk:"payment_fields_city_enabled"`
	PaymentFieldsCityRequired      types.Bool    `tfsdk:"payment_fields_city_required"`
	PaymentFieldsCountryEnabled    types.Bool    `tfsdk:"payment_fields_country_enabled"`
	PaymentFieldsCountryRequired   types.Bool    `tfsdk:"payment_fields_country_required"`
	PaymentFieldsEmailEnabled      types.Bool    `tfsdk:"payment_fields_email_enabled"`
	PaymentFieldsEmailRequired     types.Bool    `tfsdk:"payment_fields_email_required"`
	PaymentFieldsFirstNameEnabled  types.Bool    `tfsdk:"payment_fields_first_name_enabled"`
	PaymentFieldsFirstNameRequired types.Bool    `tfsdk:"payment_fields_first_name_required"`
	PaymentFieldsLastNameEnabled   types.Bool    `tfsdk:"payment_fields_last_name_enabled"`
	PaymentFieldsLastNameRequired  types.Bool    `tfsdk:"payment_fields_last_name_required"`
	PaymentFieldsStateEnabled      types.Bool    `tfsdk:"payment_fields_state_enabled"`
	PaymentFieldsStateRequired     types.Bool    `tfsdk:"payment_fields_state_required"`
	PaymentFieldsZipEnabled        types.Bool    `tfsdk:"payment_fields_zip_enabled"`
	PaymentFieldsZipRequired       types.Bool    `tfsdk:"payment_fields_zip_required"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HotspotPackageDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HotspotPackageDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.HotspotPackage)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.HotspotPackage, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Amount = ut.Float64OrNull(model.Amount)
	m.TrialReset = ut.Float64OrNull(model.TrialReset)
	m.Hours = ut.Int32OrNull(model.Hours)
	m.Index = ut.Int32OrNull(model.Index)
	m.LimitDown = ut.Int32OrNull(model.LimitDown)
	m.LimitQuota = ut.Int32OrNull(model.LimitQuota)
	m.LimitUp = ut.Int32OrNull(model.LimitUp)
	m.TrialDurationMinutes = ut.Int32OrNull(model.TrialDurationMinutes)
	m.ChargedAs = ut.StringOrNull(model.ChargedAs)
	m.Currency = ut.StringOrNull(model.Currency)
	m.Name = ut.StringOrNull(model.Name)
	m.CustomPaymentFieldsEnabled = types.BoolValue(model.CustomPaymentFieldsEnabled)
	m.LimitOverwrite = types.BoolValue(model.LimitOverwrite)
	m.PaymentFieldsAddressEnabled = types.BoolValue(model.PaymentFieldsAddressEnabled)
	m.PaymentFieldsAddressRequired = types.BoolValue(model.PaymentFieldsAddressRequired)
	m.PaymentFieldsCityEnabled = types.BoolValue(model.PaymentFieldsCityEnabled)
	m.PaymentFieldsCityRequired = types.BoolValue(model.PaymentFieldsCityRequired)
	m.PaymentFieldsCountryEnabled = types.BoolValue(model.PaymentFieldsCountryEnabled)
	m.PaymentFieldsCountryRequired = types.BoolValue(model.PaymentFieldsCountryRequired)
	m.PaymentFieldsEmailEnabled = types.BoolValue(model.PaymentFieldsEmailEnabled)
	m.PaymentFieldsEmailRequired = types.BoolValue(model.PaymentFieldsEmailRequired)
	m.PaymentFieldsFirstNameEnabled = types.BoolValue(model.PaymentFieldsFirstNameEnabled)
	m.PaymentFieldsFirstNameRequired = types.BoolValue(model.PaymentFieldsFirstNameRequired)
	m.PaymentFieldsLastNameEnabled = types.BoolValue(model.PaymentFieldsLastNameEnabled)
	m.PaymentFieldsLastNameRequired = types.BoolValue(model.PaymentFieldsLastNameRequired)
	m.PaymentFieldsStateEnabled = types.BoolValue(model.PaymentFieldsStateEnabled)
	m.PaymentFieldsStateRequired = types.BoolValue(model.PaymentFieldsStateRequired)
	m.PaymentFieldsZipEnabled = types.BoolValue(model.PaymentFieldsZipEnabled)
	m.PaymentFieldsZipRequired = types.BoolValue(model.PaymentFieldsZipRequired)
	return diags
}

var (
	_ datasource.DataSource              = &hotspotPackageDatasource{}
	_ datasource.DataSourceWithConfigure = &hotspotPackageDatasource{}
	_ base.Resource                      = &hotspotPackageDatasource{}
)

type hotspotPackageDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewHotspotPackageDatasource() datasource.DataSource {
	return &hotspotPackageDatasource{}
}

func (d *hotspotPackageDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *hotspotPackageDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *hotspotPackageDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *hotspotPackageDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *hotspotPackageDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_hotspot_package"
}

func (d *hotspotPackageDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot_package` data source retrieves an existing hotspot package by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"amount": schema.Float64Attribute{
				MarkdownDescription: "The price charged for the package, expressed in the package `currency`.",
				Computed:            true,
			},
			"trial_reset": schema.Float64Attribute{
				MarkdownDescription: "The interval, in hours, after which a guest's free trial eligibility resets.",
				Computed:            true,
			},
			"hours": schema.Int32Attribute{
				MarkdownDescription: "The duration of the package, in hours, that access remains valid after purchase.",
				Computed:            true,
			},
			"index": schema.Int32Attribute{
				MarkdownDescription: "The ordering index of the package within the hotspot package list.",
				Computed:            true,
			},
			"limit_down": schema.Int32Attribute{
				MarkdownDescription: "The download bandwidth limit, in kbps, applied to guests on this package.",
				Computed:            true,
			},
			"limit_quota": schema.Int32Attribute{
				MarkdownDescription: "The total data transfer quota, in megabytes, allowed on this package.",
				Computed:            true,
			},
			"limit_up": schema.Int32Attribute{
				MarkdownDescription: "The upload bandwidth limit, in kbps, applied to guests on this package.",
				Computed:            true,
			},
			"trial_duration_minutes": schema.Int32Attribute{
				MarkdownDescription: "The duration of the free trial, in minutes, offered by this package.",
				Computed:            true,
			},
			"charged_as": schema.StringAttribute{
				MarkdownDescription: "The label describing how the package is charged (for example the billing period shown to guests).",
				Computed:            true,
			},
			"currency": schema.StringAttribute{
				MarkdownDescription: "The ISO 4217 currency code (three uppercase letters, e.g. `USD`) used for the package price.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the hotspot package to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"custom_payment_fields_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether custom payment form fields are enabled for this package.",
				Computed:            true,
			},
			"limit_overwrite": schema.BoolAttribute{
				MarkdownDescription: "Whether this package's bandwidth and quota limits overwrite the global guest limits.",
				Computed:            true,
			},
			"payment_fields_address_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the address field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_address_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the address field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_city_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the city field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_city_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the city field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_country_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the country field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_country_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the country field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_email_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the email field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_email_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the email field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_first_name_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the first name field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_first_name_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the first name field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_last_name_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the last name field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_last_name_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the last name field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_state_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the state field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_state_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the state field is required on the payment form.",
				Computed:            true,
			},
			"payment_fields_zip_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the ZIP/postal code field is shown on the payment form.",
				Computed:            true,
			},
			"payment_fields_zip_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the ZIP/postal code field is required on the payment form.",
				Computed:            true,
			},
		},
	}
}

func (d *hotspotPackageDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HotspotPackageDatasourceModel
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

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a hotspot package.")
		return
	}

	var found *unifi.HotspotPackage
	for i := range hotspotPackages {
		h := hotspotPackages[i]
		if (id != "" && h.ID == id) || (name != "" && h.Name == name) {
			found = &h
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Hotspot package not found", fmt.Sprintf("No hotspot package matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
