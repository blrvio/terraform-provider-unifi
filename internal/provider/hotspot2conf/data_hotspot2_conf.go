package hotspot2conf

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

// Hotspot2ConfDatasourceModel represents the data model for a Hotspot 2.0
// configuration data source. It mirrors Hotspot2ConfModel field-for-field.
type Hotspot2ConfDatasourceModel struct {
	base.Model

	// ints
	AnqpDomainID        types.Int32 `tfsdk:"anqp_domain_id"`
	DeauthReqTimeout    types.Int32 `tfsdk:"deauth_req_timeout"`
	GasComebackDelay    types.Int32 `tfsdk:"gas_comeback_delay"`
	GasFragLimit        types.Int32 `tfsdk:"gas_frag_limit"`
	IPaddrTypeAvailV4   types.Int32 `tfsdk:"ipaddr_type_avail_v4"`
	IPaddrTypeAvailV6   types.Int32 `tfsdk:"ipaddr_type_avail_v6"`
	MetricsDownlinkLoad types.Int32 `tfsdk:"metrics_downlink_load"`
	MetricsDownlinkSpd  types.Int32 `tfsdk:"metrics_downlink_speed"`
	MetricsMeasurement  types.Int32 `tfsdk:"metrics_measurement"`
	MetricsUplinkLoad   types.Int32 `tfsdk:"metrics_uplink_load"`
	MetricsUplinkSpeed  types.Int32 `tfsdk:"metrics_uplink_speed"`
	NetworkAuthType     types.Int32 `tfsdk:"network_auth_type"`
	NetworkType         types.Int32 `tfsdk:"network_type"`
	TCTimestamp         types.Int32 `tfsdk:"t_c_timestamp"`
	VenueGroup          types.Int32 `tfsdk:"venue_group"`
	VenueType           types.Int32 `tfsdk:"venue_type"`

	// bools
	DisableDgaf            types.Bool `tfsdk:"disable_dgaf"`
	GasAdvanced            types.Bool `tfsdk:"gas_advanced"`
	HessidUsed             types.Bool `tfsdk:"hessid_used"`
	MetricsDownlinkLoadSet types.Bool `tfsdk:"metrics_downlink_load_set"`
	MetricsDownlinkSpdSet  types.Bool `tfsdk:"metrics_downlink_speed_set"`
	MetricsInfoAtCapacity  types.Bool `tfsdk:"metrics_info_at_capacity"`
	MetricsInfoSymmetric   types.Bool `tfsdk:"metrics_info_symmetric"`
	MetricsMeasurementSet  types.Bool `tfsdk:"metrics_measurement_set"`
	MetricsStatus          types.Bool `tfsdk:"metrics_status"`
	MetricsUplinkLoadSet   types.Bool `tfsdk:"metrics_uplink_load_set"`
	MetricsUplinkSpeedSet  types.Bool `tfsdk:"metrics_uplink_speed_set"`
	NetworkAccessAsra      types.Bool `tfsdk:"network_access_asra"`
	NetworkAccessEsr       types.Bool `tfsdk:"network_access_esr"`
	NetworkAccessInternet  types.Bool `tfsdk:"network_access_internet"`
	NetworkAccessUesa      types.Bool `tfsdk:"network_access_uesa"`
	QOSMapStatus           types.Bool `tfsdk:"qos_map_status"`

	// strings
	Hessid                types.String `tfsdk:"hessid"`
	MetricsInfoLinkStatus types.String `tfsdk:"metrics_info_link_status"`
	Name                  types.String `tfsdk:"name"`
	NetworkAuthURL        types.String `tfsdk:"network_auth_url"`
	OsuSSID               types.String `tfsdk:"osu_ssid"`
	SaveTimestamp         types.String `tfsdk:"save_timestamp"`
	TCFilename            types.String `tfsdk:"t_c_filename"`

	// []string
	DomainNameList types.List `tfsdk:"domain_name_list"`

	// []struct lists
	Capab                 types.List `tfsdk:"capab"`
	CellularNetworkList   types.List `tfsdk:"cellular_network_list"`
	FriendlyName          types.List `tfsdk:"friendly_name"`
	Icons                 types.List `tfsdk:"icons"`
	NaiRealmList          types.List `tfsdk:"nai_realm_list"`
	QOSMapDcsp            types.List `tfsdk:"qos_map_dcsp"`
	QOSMapExceptions      types.List `tfsdk:"qos_map_exceptions"`
	RoamingConsortiumList types.List `tfsdk:"roaming_consortium_list"`
	VenueName             types.List `tfsdk:"venue_name"`
	Osu                   types.List `tfsdk:"osu"`
}

// AsUnifiModel is unused for a data source.
func (m *Hotspot2ConfDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
// It reuses the resource-side list converters, which produce the same element
// object types.
func (m *Hotspot2ConfDatasourceModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.Hotspot2Conf)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Hotspot2Conf, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = ut.StringOrNull(model.Name)

	// ints
	m.AnqpDomainID = ut.Int32OrNull(model.AnqpDomainID)
	m.DeauthReqTimeout = ut.Int32OrNull(model.DeauthReqTimeout)
	m.GasComebackDelay = ut.Int32OrNull(model.GasComebackDelay)
	m.GasFragLimit = ut.Int32OrNull(model.GasFragLimit)
	m.IPaddrTypeAvailV4 = ut.Int32OrNull(model.IPaddrTypeAvailV4)
	m.IPaddrTypeAvailV6 = ut.Int32OrNull(model.IPaddrTypeAvailV6)
	m.MetricsDownlinkLoad = ut.Int32OrNull(model.MetricsDownlinkLoad)
	m.MetricsDownlinkSpd = ut.Int32OrNull(model.MetricsDownlinkSpeed)
	m.MetricsMeasurement = ut.Int32OrNull(model.MetricsMeasurement)
	m.MetricsUplinkLoad = ut.Int32OrNull(model.MetricsUplinkLoad)
	m.MetricsUplinkSpeed = ut.Int32OrNull(model.MetricsUplinkSpeed)
	m.NetworkAuthType = ut.Int32OrNull(model.NetworkAuthType)
	m.NetworkType = ut.Int32OrNull(model.NetworkType)
	m.TCTimestamp = ut.Int32OrNull(model.TCTimestamp)
	m.VenueGroup = ut.Int32OrNull(model.VenueGroup)
	m.VenueType = ut.Int32OrNull(model.VenueType)

	// bools
	m.DisableDgaf = types.BoolValue(model.DisableDgaf)
	m.GasAdvanced = types.BoolValue(model.GasAdvanced)
	m.HessidUsed = types.BoolValue(model.HessidUsed)
	m.MetricsDownlinkLoadSet = types.BoolValue(model.MetricsDownlinkLoadSet)
	m.MetricsDownlinkSpdSet = types.BoolValue(model.MetricsDownlinkSpeedSet)
	m.MetricsInfoAtCapacity = types.BoolValue(model.MetricsInfoAtCapacity)
	m.MetricsInfoSymmetric = types.BoolValue(model.MetricsInfoSymmetric)
	m.MetricsMeasurementSet = types.BoolValue(model.MetricsMeasurementSet)
	m.MetricsStatus = types.BoolValue(model.MetricsStatus)
	m.MetricsUplinkLoadSet = types.BoolValue(model.MetricsUplinkLoadSet)
	m.MetricsUplinkSpeedSet = types.BoolValue(model.MetricsUplinkSpeedSet)
	m.NetworkAccessAsra = types.BoolValue(model.NetworkAccessAsra)
	m.NetworkAccessEsr = types.BoolValue(model.NetworkAccessEsr)
	m.NetworkAccessInternet = types.BoolValue(model.NetworkAccessInternet)
	m.NetworkAccessUesa = types.BoolValue(model.NetworkAccessUesa)
	m.QOSMapStatus = types.BoolValue(model.QOSMapStatus)

	// strings
	m.Hessid = ut.StringOrNull(model.Hessid)
	m.MetricsInfoLinkStatus = ut.StringOrNull(model.MetricsInfoLinkStatus)
	m.NetworkAuthURL = ut.StringOrNull(model.NetworkAuthUrl)
	m.OsuSSID = ut.StringOrNull(model.OsuSSID)
	m.SaveTimestamp = ut.StringOrNull(model.SaveTimestamp)
	m.TCFilename = ut.StringOrNull(model.TCFilename)

	// []string
	domainNames := model.DomainNameList
	if domainNames == nil {
		domainNames = []string{}
	}
	domainList, d := types.ListValueFrom(ctx, types.StringType, domainNames)
	diags.Append(d...)
	m.DomainNameList = domainList

	// []struct lists (reuse resource-side converters)
	capab, d := capabToList(ctx, model.Capab)
	diags.Append(d...)
	m.Capab = capab

	cellular, d := cellularNetworkToList(ctx, model.CellularNetworkList)
	diags.Append(d...)
	m.CellularNetworkList = cellular

	friendly, d := friendlyNameToList(ctx, model.FriendlyName)
	diags.Append(d...)
	m.FriendlyName = friendly

	icons, d := iconsToList(ctx, model.Icons)
	diags.Append(d...)
	m.Icons = icons

	nai, d := naiRealmToList(ctx, model.NaiRealmList)
	diags.Append(d...)
	m.NaiRealmList = nai

	qosDcsp, d := qosMapDcspToList(ctx, model.QOSMapDcsp)
	diags.Append(d...)
	m.QOSMapDcsp = qosDcsp

	qosExc, d := qosMapExceptionToList(ctx, model.QOSMapExceptions)
	diags.Append(d...)
	m.QOSMapExceptions = qosExc

	roaming, d := roamingConsortiumToList(ctx, model.RoamingConsortiumList)
	diags.Append(d...)
	m.RoamingConsortiumList = roaming

	venue, d := venueNameToList(ctx, model.VenueName)
	diags.Append(d...)
	m.VenueName = venue

	osu, d := osuToList(ctx, model.Osu)
	diags.Append(d...)
	m.Osu = osu

	return diags
}

var (
	_ datasource.DataSource              = &hotspot2ConfDatasource{}
	_ datasource.DataSourceWithConfigure = &hotspot2ConfDatasource{}
	_ base.Resource                      = &hotspot2ConfDatasource{}
)

type hotspot2ConfDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

// NewHotspot2ConfDatasource creates a new instance of the Hotspot 2.0 config data source.
func NewHotspot2ConfDatasource() datasource.DataSource {
	return &hotspot2ConfDatasource{}
}

func (d *hotspot2ConfDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *hotspot2ConfDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *hotspot2ConfDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *hotspot2ConfDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *hotspot2ConfDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_hotspot2_conf"
}

// computedInt32 builds a Computed-only int32 datasource attribute.
func computedInt32(desc string) schema.Int32Attribute {
	return schema.Int32Attribute{MarkdownDescription: desc, Computed: true}
}

// computedString builds a Computed-only string datasource attribute.
func computedString(desc string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

// computedBool builds a Computed-only bool datasource attribute.
func computedBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{MarkdownDescription: desc, Computed: true}
}

func languageTextDatasourceAttributes(langDesc, textDesc string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"language": computedString(langDesc),
		"text":     computedString(textDesc),
	}
}

func (d *hotspot2ConfDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot2_conf` data source retrieves an existing Passpoint / " +
			"Hotspot 2.0 configuration profile by `name` or `id`.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),

			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the profile to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},

			// ints
			"anqp_domain_id":         computedInt32("The ANQP domain identifier."),
			"deauth_req_timeout":     computedInt32("The deauthentication request timeout, in seconds."),
			"gas_comeback_delay":     computedInt32("The GAS comeback delay, in TUs."),
			"gas_frag_limit":         computedInt32("The GAS response fragmentation limit, in bytes."),
			"ipaddr_type_avail_v4":   computedInt32("The advertised IPv4 address-type availability."),
			"ipaddr_type_avail_v6":   computedInt32("The advertised IPv6 address-type availability."),
			"metrics_downlink_load":  computedInt32("The advertised WAN downlink load."),
			"metrics_downlink_speed": computedInt32("The advertised WAN downlink speed, in kbps."),
			"metrics_measurement":    computedInt32("The WAN metrics measurement duration."),
			"metrics_uplink_load":    computedInt32("The advertised WAN uplink load."),
			"metrics_uplink_speed":   computedInt32("The advertised WAN uplink speed, in kbps."),
			"network_auth_type":      computedInt32("The network authentication type."),
			"network_type":           computedInt32("The access-network type."),
			"t_c_timestamp":          computedInt32("The terms-and-conditions file timestamp."),
			"venue_group":            computedInt32("The venue group."),
			"venue_type":             computedInt32("The venue type within the venue group."),

			// bools
			"disable_dgaf":               computedBool("Whether DGAF is disabled."),
			"gas_advanced":               computedBool("Whether advanced GAS configuration is enabled."),
			"hessid_used":                computedBool("Whether the HESSID is advertised."),
			"metrics_downlink_load_set":  computedBool("Whether the WAN downlink load value is set."),
			"metrics_downlink_speed_set": computedBool("Whether the WAN downlink speed value is set."),
			"metrics_info_at_capacity":   computedBool("Whether the WAN link is advertised as at capacity."),
			"metrics_info_symmetric":     computedBool("Whether the WAN link is advertised as symmetric."),
			"metrics_measurement_set":    computedBool("Whether the WAN metrics measurement value is set."),
			"metrics_status":             computedBool("Whether WAN metrics advertisement is enabled."),
			"metrics_uplink_load_set":    computedBool("Whether the WAN uplink load value is set."),
			"metrics_uplink_speed_set":   computedBool("Whether the WAN uplink speed value is set."),
			"network_access_asra":        computedBool("Whether ASRA is advertised."),
			"network_access_esr":         computedBool("Whether ESR is advertised."),
			"network_access_internet":    computedBool("Whether internet access is advertised."),
			"network_access_uesa":        computedBool("Whether UESA is advertised."),
			"qos_map_status":             computedBool("Whether the QoS map is advertised."),

			// strings
			"hessid":                   computedString("The homogeneous ESS identifier (HESSID)."),
			"metrics_info_link_status": computedString("The advertised WAN link status."),
			"network_auth_url":         computedString("The network authentication redirect URL."),
			"osu_ssid":                 computedString("The SSID used for Online Sign-Up (OSU)."),
			"save_timestamp":           computedString("The server-managed save timestamp of the profile."),
			"t_c_filename":             computedString("The filename of the uploaded terms-and-conditions document."),

			// []string
			"domain_name_list": schema.ListAttribute{
				MarkdownDescription: "The list of domain names advertised in the ANQP domain-name element.",
				Computed:            true,
				ElementType:         types.StringType,
			},

			// []struct lists
			"capab": schema.ListNestedAttribute{
				MarkdownDescription: "The ANQP IP-connectivity capability list.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port":     computedInt32("The port number."),
						"protocol": computedString("The protocol."),
						"status":   computedString("The reachability status."),
					},
				},
			},
			"cellular_network_list": schema.ListNestedAttribute{
				MarkdownDescription: "The 3GPP cellular network list.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mcc":  computedInt32("The Mobile Country Code (MCC)."),
						"mnc":  computedInt32("The Mobile Network Code (MNC)."),
						"name": computedString("The cellular network name."),
					},
				},
			},
			"friendly_name": schema.ListNestedAttribute{
				MarkdownDescription: "The operator friendly names, one per language.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: languageTextDatasourceAttributes(
						"The ISO-639 three-letter language code.",
						"The friendly name text.",
					),
				},
			},
			"icons": schema.ListNestedAttribute{
				MarkdownDescription: "The operator icons.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"data":     computedString("The base64-encoded icon data."),
						"filename": computedString("The icon filename."),
						"height":   computedInt32("The icon height, in pixels."),
						"language": computedString("The ISO-639 three-letter language code."),
						"media":    computedString("The icon media (MIME) type."),
						"name":     computedString("The icon name."),
						"size":     computedInt32("The icon size, in bytes."),
						"width":    computedInt32("The icon width, in pixels."),
					},
				},
			},
			"nai_realm_list": schema.ListNestedAttribute{
				MarkdownDescription: "The NAI realm list.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"auth_ids":   computedString("The authentication parameter IDs."),
						"auth_vals":  computedString("The authentication parameter values."),
						"eap_method": computedInt32("The EAP method."),
						"encoding":   computedInt32("The realm encoding."),
						"name":       computedString("The NAI realm name."),
						"status":     computedBool("Whether this realm entry is enabled."),
					},
				},
			},
			"qos_map_dcsp": schema.ListNestedAttribute{
				MarkdownDescription: "The QoS map DSCP range entries.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"high": computedInt32("The high end of the DSCP range."),
						"low":  computedInt32("The low end of the DSCP range."),
					},
				},
			},
			"qos_map_exceptions": schema.ListNestedAttribute{
				MarkdownDescription: "The QoS map DSCP-to-user-priority exceptions.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dcsp": computedInt32("The DSCP value."),
						"up":   computedInt32("The user priority."),
					},
				},
			},
			"roaming_consortium_list": schema.ListNestedAttribute{
				MarkdownDescription: "The roaming consortium organisation identifiers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": computedString("The roaming consortium name."),
						"oid":  computedString("The roaming consortium organisation identifier (OID)."),
					},
				},
			},
			"venue_name": schema.ListNestedAttribute{
				MarkdownDescription: "The venue names, one per language.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"language": computedString("The ISO-639 three-letter language code."),
						"name":     computedString("The venue name text."),
						"url":      computedString("The venue URL."),
					},
				},
			},
			"osu": schema.ListNestedAttribute{
				MarkdownDescription: "The Online Sign-Up (OSU) provider list.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.ListNestedAttribute{
							MarkdownDescription: "The OSU provider descriptions, one per language.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: languageTextDatasourceAttributes(
									"The ISO-639 three-letter language code.",
									"The description text.",
								),
							},
						},
						"friendly_name": schema.ListNestedAttribute{
							MarkdownDescription: "The OSU provider friendly names, one per language.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: languageTextDatasourceAttributes(
									"The ISO-639 three-letter language code.",
									"The friendly name text.",
								),
							},
						},
						"icon": schema.ListNestedAttribute{
							MarkdownDescription: "The OSU provider icons.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": computedString("The referenced icon name."),
								},
							},
						},
						"method_oma_dm":       computedBool("Whether the OMA-DM provisioning method is offered."),
						"method_soap_xml_spp": computedBool("Whether the SOAP-XML SPP provisioning method is offered."),
						"nai":                 computedString("The OSU Network Access Identifier."),
						"nai2":                computedString("The secondary OSU Network Access Identifier."),
						"operating_class":     computedString("The OSU operating class (hex-encoded)."),
						"server_uri":          computedString("The OSU server URI."),
					},
				},
			},
		},
	}
}

func (d *hotspot2ConfDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state Hotspot2ConfDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	confs, err := d.client.ListHotspot2Conf(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Hotspot 2.0 configurations", err.Error())
		return
	}

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a Hotspot 2.0 configuration.")
		return
	}

	var found *unifi.Hotspot2Conf
	for i := range confs {
		conf := confs[i]
		if (id != "" && conf.ID == id) || (name != "" && conf.Name == name) {
			found = &conf
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Hotspot 2.0 configuration not found", fmt.Sprintf("No Hotspot 2.0 configuration matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
