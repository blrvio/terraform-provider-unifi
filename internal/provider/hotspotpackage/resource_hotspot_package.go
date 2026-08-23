package hotspotpackage

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// HotspotPackageModel represents the data model for a UniFi hotspot package.
type HotspotPackageModel struct {
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
func (m *HotspotPackageModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return &unifi.HotspotPackage{
		ID:                             m.ID.ValueString(),
		Amount:                         m.Amount.ValueFloat64(),
		TrialReset:                     m.TrialReset.ValueFloat64(),
		Hours:                          int(m.Hours.ValueInt32()),
		Index:                          int(m.Index.ValueInt32()),
		LimitDown:                      int(m.LimitDown.ValueInt32()),
		LimitQuota:                     int(m.LimitQuota.ValueInt32()),
		LimitUp:                        int(m.LimitUp.ValueInt32()),
		TrialDurationMinutes:           int(m.TrialDurationMinutes.ValueInt32()),
		ChargedAs:                      m.ChargedAs.ValueString(),
		Currency:                       m.Currency.ValueString(),
		Name:                           m.Name.ValueString(),
		CustomPaymentFieldsEnabled:     m.CustomPaymentFieldsEnabled.ValueBool(),
		LimitOverwrite:                 m.LimitOverwrite.ValueBool(),
		PaymentFieldsAddressEnabled:    m.PaymentFieldsAddressEnabled.ValueBool(),
		PaymentFieldsAddressRequired:   m.PaymentFieldsAddressRequired.ValueBool(),
		PaymentFieldsCityEnabled:       m.PaymentFieldsCityEnabled.ValueBool(),
		PaymentFieldsCityRequired:      m.PaymentFieldsCityRequired.ValueBool(),
		PaymentFieldsCountryEnabled:    m.PaymentFieldsCountryEnabled.ValueBool(),
		PaymentFieldsCountryRequired:   m.PaymentFieldsCountryRequired.ValueBool(),
		PaymentFieldsEmailEnabled:      m.PaymentFieldsEmailEnabled.ValueBool(),
		PaymentFieldsEmailRequired:     m.PaymentFieldsEmailRequired.ValueBool(),
		PaymentFieldsFirstNameEnabled:  m.PaymentFieldsFirstNameEnabled.ValueBool(),
		PaymentFieldsFirstNameRequired: m.PaymentFieldsFirstNameRequired.ValueBool(),
		PaymentFieldsLastNameEnabled:   m.PaymentFieldsLastNameEnabled.ValueBool(),
		PaymentFieldsLastNameRequired:  m.PaymentFieldsLastNameRequired.ValueBool(),
		PaymentFieldsStateEnabled:      m.PaymentFieldsStateEnabled.ValueBool(),
		PaymentFieldsStateRequired:     m.PaymentFieldsStateRequired.ValueBool(),
		PaymentFieldsZipEnabled:        m.PaymentFieldsZipEnabled.ValueBool(),
		PaymentFieldsZipRequired:       m.PaymentFieldsZipRequired.ValueBool(),
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HotspotPackageModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
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
	_ resource.Resource                = &hotspotPackageResource{}
	_ resource.ResourceWithConfigure   = &hotspotPackageResource{}
	_ resource.ResourceWithImportState = &hotspotPackageResource{}
	_ base.Resource                    = &hotspotPackageResource{}
	_ base.ResourceModel               = &HotspotPackageModel{}
)

type hotspotPackageResource struct {
	*base.GenericResource[*HotspotPackageModel]
}

// NewHotspotPackageResource creates a new instance of the hotspot package resource.
func NewHotspotPackageResource() resource.Resource {
	return &hotspotPackageResource{
		GenericResource: base.NewGenericResource(
			"unifi_hotspot_package",
			func() *HotspotPackageModel { return &HotspotPackageModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetHotspotPackage(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					hotspotPackage, _ := model.(*unifi.HotspotPackage)
					return client.CreateHotspotPackage(ctx, site, hotspotPackage)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					hotspotPackage, _ := model.(*unifi.HotspotPackage)
					return client.UpdateHotspotPackage(ctx, site, hotspotPackage)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteHotspotPackage(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *hotspotPackageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot_package` resource manages a hotspot (guest portal) payment " +
			"package in the UniFi controller. A package defines the price, duration, bandwidth limits and " +
			"the set of payment form fields collected from guests purchasing access.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"amount": schema.Float64Attribute{
				MarkdownDescription: "The price charged for the package, expressed in the package `currency`.",
				Optional:            true,
				Computed:            true,
			},
			"trial_reset": schema.Float64Attribute{
				MarkdownDescription: "The interval, in hours, after which a guest's free trial eligibility resets.",
				Optional:            true,
				Computed:            true,
			},
			"hours": schema.Int32Attribute{
				MarkdownDescription: "The duration of the package, in hours, that access remains valid after purchase.",
				Optional:            true,
				Computed:            true,
			},
			"index": schema.Int32Attribute{
				MarkdownDescription: "The ordering index of the package within the hotspot package list.",
				Optional:            true,
				Computed:            true,
			},
			"limit_down": schema.Int32Attribute{
				MarkdownDescription: "The download bandwidth limit, in kbps, applied to guests on this package.",
				Optional:            true,
				Computed:            true,
			},
			"limit_quota": schema.Int32Attribute{
				MarkdownDescription: "The total data transfer quota, in megabytes, allowed on this package.",
				Optional:            true,
				Computed:            true,
			},
			"limit_up": schema.Int32Attribute{
				MarkdownDescription: "The upload bandwidth limit, in kbps, applied to guests on this package.",
				Optional:            true,
				Computed:            true,
			},
			"trial_duration_minutes": schema.Int32Attribute{
				MarkdownDescription: "The duration of the free trial, in minutes, offered by this package.",
				Optional:            true,
				Computed:            true,
			},
			"charged_as": schema.StringAttribute{
				MarkdownDescription: "The label describing how the package is charged (for example the billing period shown to guests).",
				Optional:            true,
				Computed:            true,
			},
			"currency": schema.StringAttribute{
				MarkdownDescription: "The ISO 4217 currency code (three uppercase letters, e.g. `USD`) used for the package price.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the hotspot package.",
				Optional:            true,
				Computed:            true,
			},
			"custom_payment_fields_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether custom payment form fields are enabled for this package.",
				Optional:            true,
				Computed:            true,
			},
			"limit_overwrite": schema.BoolAttribute{
				MarkdownDescription: "Whether this package's bandwidth and quota limits overwrite the global guest limits.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_address_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the address field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_address_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the address field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_city_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the city field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_city_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the city field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_country_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the country field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_country_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the country field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_email_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the email field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_email_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the email field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_first_name_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the first name field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_first_name_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the first name field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_last_name_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the last name field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_last_name_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the last name field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_state_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the state field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_state_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the state field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_zip_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the ZIP/postal code field is shown on the payment form.",
				Optional:            true,
				Computed:            true,
			},
			"payment_fields_zip_required": schema.BoolAttribute{
				MarkdownDescription: "Whether the ZIP/postal code field is required on the payment form.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}
