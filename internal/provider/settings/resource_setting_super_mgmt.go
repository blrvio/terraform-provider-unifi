package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// superMgmtModel represents the super (controller-wide) management settings for a
// UniFi controller, covering auto-upgrade, backups, contact info, data retention,
// and related administrative options.
type superMgmtModel struct {
	base.Model
	AnalyticsDisapprovedFor                  types.String `tfsdk:"analytics_disapproved_for"`
	AutoUpgrade                              types.Bool   `tfsdk:"auto_upgrade"`
	AutobackupCronExpr                       types.String `tfsdk:"autobackup_cron_expr"`
	AutobackupDays                           types.Int32  `tfsdk:"autobackup_days"`
	AutobackupEnabled                        types.Bool   `tfsdk:"autobackup_enabled"`
	AutobackupGcsBucket                      types.String `tfsdk:"autobackup_gcs_bucket"`
	AutobackupGcsCertificatePath             types.String `tfsdk:"autobackup_gcs_certificate_path"`
	AutobackupLocalPath                      types.String `tfsdk:"autobackup_local_path"`
	AutobackupMaxFiles                       types.Int32  `tfsdk:"autobackup_max_files"`
	AutobackupPostActions                    types.List   `tfsdk:"autobackup_post_actions"`
	AutobackupTimezone                       types.String `tfsdk:"autobackup_timezone"`
	BackupToCloudEnabled                     types.Bool   `tfsdk:"backup_to_cloud_enabled"`
	ContactInfoCity                          types.String `tfsdk:"contact_info_city"`
	ContactInfoCompanyName                   types.String `tfsdk:"contact_info_company_name"`
	ContactInfoCountry                       types.String `tfsdk:"contact_info_country"`
	ContactInfoFullName                      types.String `tfsdk:"contact_info_full_name"`
	ContactInfoPhoneNumber                   types.String `tfsdk:"contact_info_phone_number"`
	ContactInfoShippingAddress1              types.String `tfsdk:"contact_info_shipping_address_1"`
	ContactInfoShippingAddress2              types.String `tfsdk:"contact_info_shipping_address_2"`
	ContactInfoState                         types.String `tfsdk:"contact_info_state"`
	ContactInfoZip                           types.String `tfsdk:"contact_info_zip"`
	DataRetentionSettingPreference           types.String `tfsdk:"data_retention_setting_preference"`
	DataRetentionTimeInHoursFor5MinutesScale types.Int32  `tfsdk:"data_retention_time_in_hours_for_5minutes_scale"`
	DataRetentionTimeInHoursForDailyScale    types.Int32  `tfsdk:"data_retention_time_in_hours_for_daily_scale"`
	DataRetentionTimeInHoursForHourlyScale   types.Int32  `tfsdk:"data_retention_time_in_hours_for_hourly_scale"`
	DataRetentionTimeInHoursForMonthlyScale  types.Int32  `tfsdk:"data_retention_time_in_hours_for_monthly_scale"`
	DataRetentionTimeInHoursForOthers        types.Int32  `tfsdk:"data_retention_time_in_hours_for_others"`
	DefaultSiteDeviceAuthPasswordAlert       types.String `tfsdk:"default_site_device_auth_password_alert"`
	Discoverable                             types.Bool   `tfsdk:"discoverable"`
	EnableAnalytics                          types.Bool   `tfsdk:"enable_analytics"`
	GoogleMapsAPIKey                         types.String `tfsdk:"google_maps_api_key"`
	ImageMapsUseGoogleEngine                 types.Bool   `tfsdk:"image_maps_use_google_engine"`
	LedEnabled                               types.Bool   `tfsdk:"led_enabled"`
	LiveChat                                 types.String `tfsdk:"live_chat"`
	LiveUpdates                              types.String `tfsdk:"live_updates"`
	MinimumUsableHdSpace                     types.Int32  `tfsdk:"minimum_usable_hd_space"`
	MinimumUsableSdSpace                     types.Int32  `tfsdk:"minimum_usable_sd_space"`
	MultipleSitesEnabled                     types.Bool   `tfsdk:"multiple_sites_enabled"`
	OverrideInformHost                       types.Bool   `tfsdk:"override_inform_host"`
	OverrideInformHostLocation               types.String `tfsdk:"override_inform_host_location"`
	StoreEnabled                             types.String `tfsdk:"store_enabled"`
	TimeSeriesPerClientStatsEnabled          types.Bool   `tfsdk:"time_series_per_client_stats_enabled"`
	XSshPassword                             types.String `tfsdk:"x_ssh_password"`
	XSshUsername                             types.String `tfsdk:"x_ssh_username"`
}

func (d *superMgmtModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperMgmt{
		ID: d.ID.ValueString(),
	}

	if !ut.IsEmptyString(d.AnalyticsDisapprovedFor) {
		model.AnalyticsDisapprovedFor = d.AnalyticsDisapprovedFor.ValueString()
	}
	model.AutoUpgrade = d.AutoUpgrade.ValueBool()
	if !ut.IsEmptyString(d.AutobackupCronExpr) {
		model.AutobackupCronExpr = d.AutobackupCronExpr.ValueString()
	}
	if !d.AutobackupDays.IsNull() {
		model.AutobackupDays = int(d.AutobackupDays.ValueInt32())
	}
	model.AutobackupEnabled = d.AutobackupEnabled.ValueBool()
	if !ut.IsEmptyString(d.AutobackupGcsBucket) {
		model.AutobackupGcsBucket = d.AutobackupGcsBucket.ValueString()
	}
	if !ut.IsEmptyString(d.AutobackupGcsCertificatePath) {
		model.AutobackupGcsCertificatePath = d.AutobackupGcsCertificatePath.ValueString()
	}
	if !ut.IsEmptyString(d.AutobackupLocalPath) {
		model.AutobackupLocalPath = d.AutobackupLocalPath.ValueString()
	}
	if !d.AutobackupMaxFiles.IsNull() {
		model.AutobackupMaxFiles = int(d.AutobackupMaxFiles.ValueInt32())
	}
	if !d.AutobackupPostActions.IsNull() {
		var postActions []string
		diags.Append(ut.ListElementsAs(ctx, d.AutobackupPostActions, &postActions)...)
		if diags.HasError() {
			return nil, diags
		}
		model.AutobackupPostActions = postActions
	}
	if !ut.IsEmptyString(d.AutobackupTimezone) {
		model.AutobackupTimezone = d.AutobackupTimezone.ValueString()
	}
	model.BackupToCloudEnabled = d.BackupToCloudEnabled.ValueBool()
	if !ut.IsEmptyString(d.ContactInfoCity) {
		model.ContactInfoCity = d.ContactInfoCity.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoCompanyName) {
		model.ContactInfoCompanyName = d.ContactInfoCompanyName.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoCountry) {
		model.ContactInfoCountry = d.ContactInfoCountry.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoFullName) {
		model.ContactInfoFullName = d.ContactInfoFullName.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoPhoneNumber) {
		model.ContactInfoPhoneNumber = d.ContactInfoPhoneNumber.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoShippingAddress1) {
		model.ContactInfoShippingAddress1 = d.ContactInfoShippingAddress1.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoShippingAddress2) {
		model.ContactInfoShippingAddress2 = d.ContactInfoShippingAddress2.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoState) {
		model.ContactInfoState = d.ContactInfoState.ValueString()
	}
	if !ut.IsEmptyString(d.ContactInfoZip) {
		model.ContactInfoZip = d.ContactInfoZip.ValueString()
	}
	if !ut.IsEmptyString(d.DataRetentionSettingPreference) {
		model.DataRetentionSettingPreference = d.DataRetentionSettingPreference.ValueString()
	}
	if !d.DataRetentionTimeInHoursFor5MinutesScale.IsNull() {
		model.DataRetentionTimeInHoursFor5MinutesScale = int(d.DataRetentionTimeInHoursFor5MinutesScale.ValueInt32())
	}
	if !d.DataRetentionTimeInHoursForDailyScale.IsNull() {
		model.DataRetentionTimeInHoursForDailyScale = int(d.DataRetentionTimeInHoursForDailyScale.ValueInt32())
	}
	if !d.DataRetentionTimeInHoursForHourlyScale.IsNull() {
		model.DataRetentionTimeInHoursForHourlyScale = int(d.DataRetentionTimeInHoursForHourlyScale.ValueInt32())
	}
	if !d.DataRetentionTimeInHoursForMonthlyScale.IsNull() {
		model.DataRetentionTimeInHoursForMonthlyScale = int(d.DataRetentionTimeInHoursForMonthlyScale.ValueInt32())
	}
	if !d.DataRetentionTimeInHoursForOthers.IsNull() {
		model.DataRetentionTimeInHoursForOthers = int(d.DataRetentionTimeInHoursForOthers.ValueInt32())
	}
	if !ut.IsEmptyString(d.DefaultSiteDeviceAuthPasswordAlert) {
		model.DefaultSiteDeviceAuthPasswordAlert = d.DefaultSiteDeviceAuthPasswordAlert.ValueString()
	}
	model.Discoverable = d.Discoverable.ValueBool()
	model.EnableAnalytics = d.EnableAnalytics.ValueBool()
	if !ut.IsEmptyString(d.GoogleMapsAPIKey) {
		model.GoogleMapsApiKey = d.GoogleMapsAPIKey.ValueString()
	}
	model.ImageMapsUseGoogleEngine = d.ImageMapsUseGoogleEngine.ValueBool()
	model.LedEnabled = d.LedEnabled.ValueBool()
	if !ut.IsEmptyString(d.LiveChat) {
		model.LiveChat = d.LiveChat.ValueString()
	}
	if !ut.IsEmptyString(d.LiveUpdates) {
		model.LiveUpdates = d.LiveUpdates.ValueString()
	}
	if !d.MinimumUsableHdSpace.IsNull() {
		model.MinimumUsableHdSpace = int(d.MinimumUsableHdSpace.ValueInt32())
	}
	if !d.MinimumUsableSdSpace.IsNull() {
		model.MinimumUsableSdSpace = int(d.MinimumUsableSdSpace.ValueInt32())
	}
	model.MultipleSitesEnabled = d.MultipleSitesEnabled.ValueBool()
	model.OverrideInformHost = d.OverrideInformHost.ValueBool()
	if !ut.IsEmptyString(d.OverrideInformHostLocation) {
		model.OverrideInformHostLocation = d.OverrideInformHostLocation.ValueString()
	}
	if !ut.IsEmptyString(d.StoreEnabled) {
		model.StoreEnabled = d.StoreEnabled.ValueString()
	}
	model.TimeSeriesPerClientStatsEnabled = d.TimeSeriesPerClientStatsEnabled.ValueBool()
	if !ut.IsEmptyString(d.XSshPassword) {
		model.XSshPassword = d.XSshPassword.ValueString()
	}
	if !ut.IsEmptyString(d.XSshUsername) {
		model.XSshUsername = d.XSshUsername.ValueString()
	}

	return model, diags
}

func (d *superMgmtModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperMgmt)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperMgmt")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.AnalyticsDisapprovedFor = ut.StringOrNull(model.AnalyticsDisapprovedFor)
	d.AutoUpgrade = types.BoolValue(model.AutoUpgrade)
	d.AutobackupCronExpr = ut.StringOrNull(model.AutobackupCronExpr)
	d.AutobackupDays = ut.Int32OrNull(model.AutobackupDays)
	d.AutobackupEnabled = types.BoolValue(model.AutobackupEnabled)
	d.AutobackupGcsBucket = ut.StringOrNull(model.AutobackupGcsBucket)
	d.AutobackupGcsCertificatePath = ut.StringOrNull(model.AutobackupGcsCertificatePath)
	d.AutobackupLocalPath = ut.StringOrNull(model.AutobackupLocalPath)
	d.AutobackupMaxFiles = ut.Int32OrNull(model.AutobackupMaxFiles)
	d.AutobackupTimezone = ut.StringOrNull(model.AutobackupTimezone)
	d.BackupToCloudEnabled = types.BoolValue(model.BackupToCloudEnabled)
	d.ContactInfoCity = ut.StringOrNull(model.ContactInfoCity)
	d.ContactInfoCompanyName = ut.StringOrNull(model.ContactInfoCompanyName)
	d.ContactInfoCountry = ut.StringOrNull(model.ContactInfoCountry)
	d.ContactInfoFullName = ut.StringOrNull(model.ContactInfoFullName)
	d.ContactInfoPhoneNumber = ut.StringOrNull(model.ContactInfoPhoneNumber)
	d.ContactInfoShippingAddress1 = ut.StringOrNull(model.ContactInfoShippingAddress1)
	d.ContactInfoShippingAddress2 = ut.StringOrNull(model.ContactInfoShippingAddress2)
	d.ContactInfoState = ut.StringOrNull(model.ContactInfoState)
	d.ContactInfoZip = ut.StringOrNull(model.ContactInfoZip)
	d.DataRetentionSettingPreference = ut.StringOrNull(model.DataRetentionSettingPreference)
	d.DataRetentionTimeInHoursFor5MinutesScale = ut.Int32OrNull(model.DataRetentionTimeInHoursFor5MinutesScale)
	d.DataRetentionTimeInHoursForDailyScale = ut.Int32OrNull(model.DataRetentionTimeInHoursForDailyScale)
	d.DataRetentionTimeInHoursForHourlyScale = ut.Int32OrNull(model.DataRetentionTimeInHoursForHourlyScale)
	d.DataRetentionTimeInHoursForMonthlyScale = ut.Int32OrNull(model.DataRetentionTimeInHoursForMonthlyScale)
	d.DataRetentionTimeInHoursForOthers = ut.Int32OrNull(model.DataRetentionTimeInHoursForOthers)
	d.DefaultSiteDeviceAuthPasswordAlert = ut.StringOrNull(model.DefaultSiteDeviceAuthPasswordAlert)
	d.Discoverable = types.BoolValue(model.Discoverable)
	d.EnableAnalytics = types.BoolValue(model.EnableAnalytics)
	d.GoogleMapsAPIKey = ut.StringOrNull(model.GoogleMapsApiKey)
	d.ImageMapsUseGoogleEngine = types.BoolValue(model.ImageMapsUseGoogleEngine)
	d.LedEnabled = types.BoolValue(model.LedEnabled)
	d.LiveChat = ut.StringOrNull(model.LiveChat)
	d.LiveUpdates = ut.StringOrNull(model.LiveUpdates)
	d.MinimumUsableHdSpace = ut.Int32OrNull(model.MinimumUsableHdSpace)
	d.MinimumUsableSdSpace = ut.Int32OrNull(model.MinimumUsableSdSpace)
	d.MultipleSitesEnabled = types.BoolValue(model.MultipleSitesEnabled)
	d.OverrideInformHost = types.BoolValue(model.OverrideInformHost)
	d.OverrideInformHostLocation = ut.StringOrNull(model.OverrideInformHostLocation)
	d.StoreEnabled = ut.StringOrNull(model.StoreEnabled)
	d.TimeSeriesPerClientStatsEnabled = types.BoolValue(model.TimeSeriesPerClientStatsEnabled)
	d.XSshPassword = ut.StringOrNull(model.XSshPassword)
	d.XSshUsername = ut.StringOrNull(model.XSshUsername)

	if len(model.AutobackupPostActions) > 0 {
		list, ld := types.ListValueFrom(ctx, types.StringType, model.AutobackupPostActions)
		diags.Append(ld...)
		if diags.HasError() {
			return diags
		}
		d.AutobackupPostActions = list
	} else {
		d.AutobackupPostActions = ut.EmptyList(types.StringType)
	}

	return diags
}

var (
	_ base.ResourceModel               = &superMgmtModel{}
	_ resource.Resource                = &superMgmtResource{}
	_ resource.ResourceWithConfigure   = &superMgmtResource{}
	_ resource.ResourceWithImportState = &superMgmtResource{}
)

type superMgmtResource struct {
	*base.GenericResource[*superMgmtModel]
}

func (r *superMgmtResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_mgmt` resource manages controller-wide (super) management settings for a " +
			"UniFi controller, including auto-upgrade, automatic backups, contact information, data retention, and related " +
			"administrative options.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"analytics_disapproved_for": schema.StringAttribute{
				MarkdownDescription: "Identifier of the analytics collection the user has disapproved.",
				Optional:            true,
				Computed:            true,
			},
			"auto_upgrade": schema.BoolAttribute{
				MarkdownDescription: "Whether devices are automatically upgraded to new firmware.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_cron_expr": schema.StringAttribute{
				MarkdownDescription: "Cron expression defining the automatic backup schedule.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_days": schema.Int32Attribute{
				MarkdownDescription: "Number of days of data to include in automatic backups.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether automatic backups are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_gcs_bucket": schema.StringAttribute{
				MarkdownDescription: "Google Cloud Storage bucket name used for automatic backups.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_gcs_certificate_path": schema.StringAttribute{
				MarkdownDescription: "Path to the Google Cloud Storage service account certificate used for backups.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_local_path": schema.StringAttribute{
				MarkdownDescription: "Local filesystem path where automatic backups are stored.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_max_files": schema.Int32Attribute{
				MarkdownDescription: "Maximum number of automatic backup files to retain.",
				Optional:            true,
				Computed:            true,
			},
			"autobackup_post_actions": schema.ListAttribute{
				MarkdownDescription: "Actions performed after an automatic backup completes. Valid values: `copy_local`, `copy_gcs`, `copy_cloud`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf("copy_local", "copy_gcs", "copy_cloud")),
				},
			},
			"autobackup_timezone": schema.StringAttribute{
				MarkdownDescription: "Timezone used when scheduling automatic backups.",
				Optional:            true,
				Computed:            true,
			},
			"backup_to_cloud_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether backups are copied to the UniFi cloud.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_city": schema.StringAttribute{
				MarkdownDescription: "Contact information: city.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_company_name": schema.StringAttribute{
				MarkdownDescription: "Contact information: company name.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_country": schema.StringAttribute{
				MarkdownDescription: "Contact information: country.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_full_name": schema.StringAttribute{
				MarkdownDescription: "Contact information: full name.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_phone_number": schema.StringAttribute{
				MarkdownDescription: "Contact information: phone number.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_shipping_address_1": schema.StringAttribute{
				MarkdownDescription: "Contact information: first line of the shipping address.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_shipping_address_2": schema.StringAttribute{
				MarkdownDescription: "Contact information: second line of the shipping address.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_state": schema.StringAttribute{
				MarkdownDescription: "Contact information: state or province.",
				Optional:            true,
				Computed:            true,
			},
			"contact_info_zip": schema.StringAttribute{
				MarkdownDescription: "Contact information: postal / ZIP code.",
				Optional:            true,
				Computed:            true,
			},
			"data_retention_setting_preference": schema.StringAttribute{
				MarkdownDescription: "Data retention preference. Valid values: `auto`, `manual`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},
			"data_retention_time_in_hours_for_5minutes_scale": schema.Int32Attribute{
				MarkdownDescription: "Retention time (hours) for statistics at the 5-minute scale.",
				Optional:            true,
				Computed:            true,
			},
			"data_retention_time_in_hours_for_daily_scale": schema.Int32Attribute{
				MarkdownDescription: "Retention time (hours) for statistics at the daily scale.",
				Optional:            true,
				Computed:            true,
			},
			"data_retention_time_in_hours_for_hourly_scale": schema.Int32Attribute{
				MarkdownDescription: "Retention time (hours) for statistics at the hourly scale.",
				Optional:            true,
				Computed:            true,
			},
			"data_retention_time_in_hours_for_monthly_scale": schema.Int32Attribute{
				MarkdownDescription: "Retention time (hours) for statistics at the monthly scale.",
				Optional:            true,
				Computed:            true,
			},
			"data_retention_time_in_hours_for_others": schema.Int32Attribute{
				MarkdownDescription: "Retention time (hours) for other statistics categories.",
				Optional:            true,
				Computed:            true,
			},
			"default_site_device_auth_password_alert": schema.StringAttribute{
				MarkdownDescription: "Default site device authentication password alert setting.",
				Optional:            true,
				Computed:            true,
			},
			"discoverable": schema.BoolAttribute{
				MarkdownDescription: "Whether the controller is discoverable on the network.",
				Optional:            true,
				Computed:            true,
			},
			"enable_analytics": schema.BoolAttribute{
				MarkdownDescription: "Whether usage analytics collection is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"google_maps_api_key": schema.StringAttribute{
				MarkdownDescription: "Google Maps API key used for map rendering.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"image_maps_use_google_engine": schema.BoolAttribute{
				MarkdownDescription: "Whether the Google Maps engine is used for image maps.",
				Optional:            true,
				Computed:            true,
			},
			"led_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether device status LEDs are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"live_chat": schema.StringAttribute{
				MarkdownDescription: "Live chat availability. Valid values: `disabled`, `super-only`, `everyone`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "super-only", "everyone"),
				},
			},
			"live_updates": schema.StringAttribute{
				MarkdownDescription: "Live updates mode. Valid values: `disabled`, `live`, `auto`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "live", "auto"),
				},
			},
			"minimum_usable_hd_space": schema.Int32Attribute{
				MarkdownDescription: "Minimum usable hard-disk space (in MB) before alerts are raised.",
				Optional:            true,
				Computed:            true,
			},
			"minimum_usable_sd_space": schema.Int32Attribute{
				MarkdownDescription: "Minimum usable SD-card space (in MB) before alerts are raised.",
				Optional:            true,
				Computed:            true,
			},
			"multiple_sites_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether multiple sites are enabled on the controller.",
				Optional:            true,
				Computed:            true,
			},
			"override_inform_host": schema.BoolAttribute{
				MarkdownDescription: "Whether the inform host is overridden for device communication.",
				Optional:            true,
				Computed:            true,
			},
			"override_inform_host_location": schema.StringAttribute{
				MarkdownDescription: "Override inform host location (hostname or IP) used by devices to reach the controller.",
				Optional:            true,
				Computed:            true,
			},
			"store_enabled": schema.StringAttribute{
				MarkdownDescription: "UniFi store availability. Valid values: `disabled`, `super-only`, `everyone`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "super-only", "everyone"),
				},
			},
			"time_series_per_client_stats_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether per-client time-series statistics are collected.",
				Optional:            true,
				Computed:            true,
			},
			"x_ssh_password": schema.StringAttribute{
				MarkdownDescription: "SSH password used for device management.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"x_ssh_username": schema.StringAttribute{
				MarkdownDescription: "SSH username used for device management.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

// NewSuperMgmtResource creates a new instance of the super management setting resource.
func NewSuperMgmtResource() resource.Resource {
	r := &superMgmtResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_mgmt",
		func() *superMgmtModel { return &superMgmtModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperMgmt(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperMgmt)
			return client.UpdateSettingSuperMgmt(ctx, site, b)
		},
	)
	return r
}
