package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/acl"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/apgroup"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/broadcastgroup"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/channelplan"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/contentfiltering"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/dashboard"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/dhcpoption"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/dns"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/feature"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/firewall"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/heatmap"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/heatmappoint"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/hotspot"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/hotspot2conf"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/hotspotop"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/hotspotpackage"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/mediafile"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/officialfw"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/officialro"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/portal"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/qos"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/scheduletask"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/settings"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/spatialrecord"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/tag"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/trafficflow"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/trafficlist"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/unifimap"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/virtualdevice"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/wifibroadcast"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/wlangroup"
)

func NewV2(version string) func() provider.Provider {
	return func() provider.Provider {
		return &unifiProvider{
			version: version,
		}
	}
}

var _ provider.Provider = &unifiProvider{}

type unifiProvider struct {
	version string
}

type unifiProviderModel struct {
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
	APIKey     types.String `tfsdk:"api_key"`
	APIUrl     types.String `tfsdk:"api_url"`
	Site       types.String `tfsdk:"site"`
	Insecure   types.Bool   `tfsdk:"allow_insecure"`
	MaxRetries types.Int64  `tfsdk:"http_max_retries"`
}

func (p *unifiProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unifi"
	resp.Version = p.version
}

func (p *unifiProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// The SDKv2 provider (provider.go) appends the `Deprecated` marker to the
			// rendered description via SchemaDescriptionBuilder. The muxed providers must
			// expose byte-identical schemas, so mirror that suffix here explicitly.
			"username": schema.StringAttribute{
				MarkdownDescription: ProviderUsernameDescription + " " + ProviderUserPassDeprecated,
				Optional:            true,
				DeprecationMessage:  ProviderUserPassDeprecated,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: ProviderPasswordDescription + " " + ProviderUserPassDeprecated,
				Optional:            true,
				Sensitive:           true,
				DeprecationMessage:  ProviderUserPassDeprecated,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: ProviderAPIKeyDescription,
				Optional:            true,
				Sensitive:           true,
			},
			"api_url": schema.StringAttribute{
				MarkdownDescription: ProviderAPIURLDescription,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1), // workaround for `required: true`, because it fails on doc generation due to incorrectly detected difference between v1 and v2
					validators.HTTPSUrl(),
				},
				Optional: true,
			},
			"site": schema.StringAttribute{
				MarkdownDescription: ProviderSiteDescription,
				Optional:            true,
			},
			"allow_insecure": schema.BoolAttribute{
				MarkdownDescription: ProviderAllowInsecureDescription,
				Optional:            true,
			},
			"http_max_retries": schema.Int64Attribute{
				MarkdownDescription: ProviderMaxRetriesDescription,
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (p *unifiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Unifi provider...")
	// Retrieve provider data from the configuration
	var cfg unifiProviderModel
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if cfg.APIUrl.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_url"),
			"Unknown UniFi Controller API URL",
			"The provider cannot create the UniFi Controller API client as there is an unknown configuration value "+
				"for the API endpoint. Either target apply the source of the value first, set the value statically in "+
				"the configuration, or use the UNIFI_API environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	// Check environment variables
	username := utils.GetAnyStringEnv("UNIFI_USERNAME")
	password := utils.GetAnyStringEnv("UNIFI_PASSWORD")
	apiKey := utils.GetAnyStringEnv("UNIFI_API_KEY")
	apiURL := utils.GetAnyStringEnv("UNIFI_API")
	site := utils.GetAnyStringEnv("UNIFI_SITE")
	insecure := utils.GetAnyBoolEnv("UNIFI_INSECURE")
	maxRetries := utils.GetAnyIntEnv("UNIFI_MAX_RETRIES")

	if !cfg.Username.IsNull() {
		username = cfg.Username.ValueString()
	}
	if !cfg.Password.IsNull() {
		password = cfg.Password.ValueString()
	}
	if !cfg.APIKey.IsNull() {
		apiKey = cfg.APIKey.ValueString()
	}
	if !cfg.APIUrl.IsNull() {
		apiURL = cfg.APIUrl.ValueString()
	}
	if !cfg.Site.IsNull() {
		site = cfg.Site.ValueString()
	}
	if !cfg.Insecure.IsNull() {
		insecure = cfg.Insecure.ValueBool()
	}
	if !cfg.MaxRetries.IsNull() {
		maxRetries = int(cfg.MaxRetries.ValueInt64())
	}
	if username != "" || password != "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Username/password authentication is no longer supported",
			"Username/password authentication was removed in go-unifi v10. Unset `username`/`password` (and UNIFI_USERNAME/UNIFI_PASSWORD) and configure `api_key` (or UNIFI_API_KEY) instead.")
	} else if apiKey == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Missing UniFi API key", "The `api_key` attribute (or UNIFI_API_KEY) must be set.")
	}
	if apiURL == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_url"), "Missing UniFi API URL", "The `api_url` attribute must be set")
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if site == "" {
		site = "default" // set default site if not provided
	}
	c, err := base.NewClient(&base.ClientConfig{
		APIKey:     apiKey,
		URL:        apiURL,
		Site:       site,
		Insecure:   insecure,
		MaxRetries: maxRetries,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create UniFi client", err.Error())
		return
	}
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *unifiProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		acl.NewACLRuleResource,
		acl.NewACLRuleOrderResource,
		apgroup.NewAPGroupResource,
		broadcastgroup.NewBroadcastGroupResource,
		channelplan.NewChannelPlanResource,
		contentfiltering.NewContentFilteringResource,
		dhcpoption.NewDHCPOptionResource,
		scheduletask.NewScheduleTaskResource,
		tag.NewTagResource,
		heatmap.NewHeatMapResource,
		heatmappoint.NewHeatMapPointResource,
		unifimap.NewMapResource,
		spatialrecord.NewSpatialRecordResource,
		mediafile.NewMediaFileResource,
		dashboard.NewDashboardResource,
		hotspot2conf.NewHotspot2ConfResource,
		hotspotop.NewHotspotOperatorResource,
		hotspotpackage.NewHotspotPackageResource,
		virtualdevice.NewVirtualDeviceResource,
		wlangroup.NewWLANGroupResource,
		dns.NewDNSRecordResource,
		dns.NewDNSPolicyResource,
		hotspot.NewHotspotVoucherResource,
		trafficlist.NewTrafficMatchingListResource,
		wifibroadcast.NewWifiBroadcastResource,
		officialfw.NewOfficialFirewallZoneResource,
		officialfw.NewOfficialFirewallPolicyResource,
		officialfw.NewOfficialFirewallPolicyOrderResource,
		firewall.NewFirewallZoneResource,
		firewall.NewFirewallZonePolicyResource,
		firewall.NewFirewallZonePolicyOrderResource,
		portal.NewPortalFileResource,
		qos.NewQOSRuleResource,
		settings.NewAutoSpeedtestResource,
		settings.NewConnectivityResource,
		settings.NewCountryResource,
		settings.NewDpiResource,
		settings.NewEtherLightingResource,
		settings.NewGuestAccessResource,
		settings.NewIpsResource,
		settings.NewLcmResource,
		settings.NewLocaleResource,
		settings.NewMagicSiteToSiteVpnResource,
		settings.NewNetworkOptimizationResource,
		settings.NewNtpResource,
		settings.NewRsyslogdResource,
		settings.NewSslInspectionResource,
		settings.NewTeleportResource,
		settings.NewMgmtResource,
		settings.NewUsgResource,
		settings.NewUsgGeoResource,
		settings.NewUswResource,
		settings.NewGlobalSwitchResource,
		settings.NewSnmpResource,
		settings.NewMdnsResource,
		settings.NewDohResource,
		settings.NewGlobalNatResource,
		settings.NewGlobalApResource,
		settings.NewNetflowResource,
		settings.NewRoamingAssistantResource,
		settings.NewTrafficFlowResource,
		settings.NewBroadcastResource,
		settings.NewDashboardResource,
		settings.NewElementAdoptResource,
		settings.NewEvaluationScoreResource,
		settings.NewBaresipResource,
		settings.NewPortaResource,
		settings.NewRadioAiResource,
		settings.NewSuperCloudaccessResource,
		settings.NewSuperEventsResource,
		settings.NewSuperFwupdateResource,
		settings.NewSuperIdentityResource,
		settings.NewSuperMailResource,
		settings.NewSuperMgmtResource,
		settings.NewSuperSdnResource,
		settings.NewSuperSMTPResource,
	}
}

func (p *unifiProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		apgroup.NewAPGroupDatasource,
		tag.NewTagDatasource,
		heatmap.NewHeatMapDatasource,
		heatmap.NewHeatMapsDatasource,
		heatmappoint.NewHeatMapPointDatasource,
		unifimap.NewMapDatasource,
		unifimap.NewMapsDatasource,
		spatialrecord.NewSpatialRecordDatasource,
		mediafile.NewMediaFileDatasource,
		dashboard.NewDashboardDatasource,
		dashboard.NewDashboardsDatasource,
		hotspot2conf.NewHotspot2ConfDatasource,
		hotspotop.NewHotspotOperatorDatasource,
		hotspotpackage.NewHotspotPackageDatasource,
		hotspotpackage.NewHotspotPackagesDatasource,
		feature.NewFeatureDatasource,
		feature.NewFeaturesDatasource,
		feature.NewSystemInformationDatasource,
		wlangroup.NewWLANGroupDatasource,
		dns.NewDNSRecordsDatasource,
		dns.NewDNSRecordDatasource,
		firewall.NewFirewallZoneDatasource,
		officialro.NewWANDataSource,
		officialro.NewVPNServerDataSource,
		officialro.NewSiteToSiteTunnelDataSource,
		officialro.NewSwitchLAGDataSource,
		officialro.NewDeviceTagDataSource,
		officialro.NewClientsDataSource,
		officialro.NewClientDataSource,
		officialro.NewOfficialDevicesDataSource,
		officialro.NewOfficialDeviceDataSource,
		officialro.NewDeviceStatisticsDataSource,
		officialro.NewPendingDevicesDataSource,
		officialro.NewNetworksDataSource,
		officialro.NewNetworkReferencesDataSource,
		officialro.NewDPIApplicationsDataSource,
		officialro.NewDPIApplicationCategoriesDataSource,
		officialro.NewCountriesDataSource,
		officialro.NewRadiusProfilesDataSource,
		officialro.NewSwitchLAGsDataSource,
		officialro.NewMcLagDomainsDataSource,
		officialro.NewMcLagDomainDataSource,
		officialro.NewSwitchStacksDataSource,
		officialro.NewSwitchStackDataSource,
		officialro.NewControllerInfoDataSource,
		trafficflow.NewTrafficFlowsDataSource,
	}
}
