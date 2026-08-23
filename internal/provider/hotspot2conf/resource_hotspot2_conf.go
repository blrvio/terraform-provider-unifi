package hotspot2conf

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// ---------------------------------------------------------------------------
// Nested object models
// ---------------------------------------------------------------------------

// capabModel is one entry in the ANQP IP-connectivity capability list.
type capabModel struct {
	Port     types.Int32  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
	Status   types.String `tfsdk:"status"`
}

func (m capabModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"port":     types.Int32Type,
		"protocol": types.StringType,
		"status":   types.StringType,
	}
}

func capabObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: capabModel{}.AttributeTypes()}
}

// cellularNetworkModel is one entry in the 3GPP cellular network list.
type cellularNetworkModel struct {
	Mcc  types.Int32  `tfsdk:"mcc"`
	Mnc  types.Int32  `tfsdk:"mnc"`
	Name types.String `tfsdk:"name"`
}

func (m cellularNetworkModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mcc":  types.Int32Type,
		"mnc":  types.Int32Type,
		"name": types.StringType,
	}
}

func cellularNetworkObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: cellularNetworkModel{}.AttributeTypes()}
}

// languageTextModel is a shared {language, text} pair used by the top-level
// friendly_name list and by the osu description / friendly_name sub-lists.
type languageTextModel struct {
	Language types.String `tfsdk:"language"`
	Text     types.String `tfsdk:"text"`
}

func (m languageTextModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"language": types.StringType,
		"text":     types.StringType,
	}
}

func languageTextObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: languageTextModel{}.AttributeTypes()}
}

// iconsModel is one entry in the top-level icons list.
type iconsModel struct {
	Data     types.String `tfsdk:"data"`
	Filename types.String `tfsdk:"filename"`
	Height   types.Int32  `tfsdk:"height"`
	Language types.String `tfsdk:"language"`
	Media    types.String `tfsdk:"media"`
	Name     types.String `tfsdk:"name"`
	Size     types.Int32  `tfsdk:"size"`
	Width    types.Int32  `tfsdk:"width"`
}

func (m iconsModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"data":     types.StringType,
		"filename": types.StringType,
		"height":   types.Int32Type,
		"language": types.StringType,
		"media":    types.StringType,
		"name":     types.StringType,
		"size":     types.Int32Type,
		"width":    types.Int32Type,
	}
}

func iconsObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: iconsModel{}.AttributeTypes()}
}

// naiRealmModel is one entry in the NAI realm list.
type naiRealmModel struct {
	AuthIDs   types.String `tfsdk:"auth_ids"`
	AuthVals  types.String `tfsdk:"auth_vals"`
	EapMethod types.Int32  `tfsdk:"eap_method"`
	Encoding  types.Int32  `tfsdk:"encoding"`
	Name      types.String `tfsdk:"name"`
	Status    types.Bool   `tfsdk:"status"`
}

func (m naiRealmModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auth_ids":   types.StringType,
		"auth_vals":  types.StringType,
		"eap_method": types.Int32Type,
		"encoding":   types.Int32Type,
		"name":       types.StringType,
		"status":     types.BoolType,
	}
}

func naiRealmObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: naiRealmModel{}.AttributeTypes()}
}

// qosMapDcspModel is one entry in the QoS map DSCP list.
type qosMapDcspModel struct {
	High types.Int32 `tfsdk:"high"`
	Low  types.Int32 `tfsdk:"low"`
}

func (m qosMapDcspModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"high": types.Int32Type,
		"low":  types.Int32Type,
	}
}

func qosMapDcspObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: qosMapDcspModel{}.AttributeTypes()}
}

// qosMapExceptionModel is one entry in the QoS map exceptions list.
type qosMapExceptionModel struct {
	Dcsp types.Int32 `tfsdk:"dcsp"`
	Up   types.Int32 `tfsdk:"up"`
}

func (m qosMapExceptionModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"dcsp": types.Int32Type,
		"up":   types.Int32Type,
	}
}

func qosMapExceptionObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: qosMapExceptionModel{}.AttributeTypes()}
}

// roamingConsortiumModel is one entry in the roaming consortium list.
type roamingConsortiumModel struct {
	Name types.String `tfsdk:"name"`
	Oid  types.String `tfsdk:"oid"`
}

func (m roamingConsortiumModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
		"oid":  types.StringType,
	}
}

func roamingConsortiumObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: roamingConsortiumModel{}.AttributeTypes()}
}

// venueNameModel is one entry in the venue name list.
type venueNameModel struct {
	Language types.String `tfsdk:"language"`
	Name     types.String `tfsdk:"name"`
	URL      types.String `tfsdk:"url"`
}

func (m venueNameModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"language": types.StringType,
		"name":     types.StringType,
		"url":      types.StringType,
	}
}

func venueNameObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: venueNameModel{}.AttributeTypes()}
}

// osuIconModel is one entry in an osu provider's icon sub-list.
type osuIconModel struct {
	Name types.String `tfsdk:"name"`
}

func (m osuIconModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
	}
}

func osuIconObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: osuIconModel{}.AttributeTypes()}
}

// osuModel is one Online Sign-Up (OSU) provider. Each osu item itself contains
// three sub-lists (description, friendly_name, icon) modelled as nested lists.
type osuModel struct {
	Description      types.List   `tfsdk:"description"`
	FriendlyName     types.List   `tfsdk:"friendly_name"`
	Icon             types.List   `tfsdk:"icon"`
	MethodOmaDm      types.Bool   `tfsdk:"method_oma_dm"`
	MethodSoapXMLSpp types.Bool   `tfsdk:"method_soap_xml_spp"`
	Nai              types.String `tfsdk:"nai"`
	Nai2             types.String `tfsdk:"nai2"`
	OperatingClass   types.String `tfsdk:"operating_class"`
	ServerURI        types.String `tfsdk:"server_uri"`
}

func (m osuModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"description":         types.ListType{ElemType: languageTextObjectType()},
		"friendly_name":       types.ListType{ElemType: languageTextObjectType()},
		"icon":                types.ListType{ElemType: osuIconObjectType()},
		"method_oma_dm":       types.BoolType,
		"method_soap_xml_spp": types.BoolType,
		"nai":                 types.StringType,
		"nai2":                types.StringType,
		"operating_class":     types.StringType,
		"server_uri":          types.StringType,
	}
}

func osuObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: osuModel{}.AttributeTypes()}
}

// ---------------------------------------------------------------------------
// Top-level model
// ---------------------------------------------------------------------------

// Hotspot2ConfModel represents the data model for a UniFi Passpoint / Hotspot
// 2.0 configuration.
type Hotspot2ConfModel struct {
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

// ---------------------------------------------------------------------------
// Converters: Terraform list -> go-unifi slice
// ---------------------------------------------------------------------------

func capabAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfCapab, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []capabModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfCapab, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfCapab{
			Port:     int(it.Port.ValueInt32()),
			Protocol: it.Protocol.ValueString(),
			Status:   it.Status.ValueString(),
		})
	}
	return out, diags
}

func cellularNetworkAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfCellularNetworkList, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []cellularNetworkModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfCellularNetworkList, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfCellularNetworkList{
			Mcc:  int(it.Mcc.ValueInt32()),
			Mnc:  int(it.Mnc.ValueInt32()),
			Name: it.Name.ValueString(),
		})
	}
	return out, diags
}

func friendlyNameAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfFriendlyName, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []languageTextModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfFriendlyName, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfFriendlyName{
			Language: it.Language.ValueString(),
			Text:     it.Text.ValueString(),
		})
	}
	return out, diags
}

func descriptionAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfDescription, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []languageTextModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfDescription, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfDescription{
			Language: it.Language.ValueString(),
			Text:     it.Text.ValueString(),
		})
	}
	return out, diags
}

func osuIconAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfIcon, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []osuIconModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfIcon, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfIcon{Name: it.Name.ValueString()})
	}
	return out, diags
}

func iconsAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfIcons, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []iconsModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfIcons, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfIcons{
			Data:     it.Data.ValueString(),
			Filename: it.Filename.ValueString(),
			Height:   int(it.Height.ValueInt32()),
			Language: it.Language.ValueString(),
			Media:    it.Media.ValueString(),
			Name:     it.Name.ValueString(),
			Size:     int(it.Size.ValueInt32()),
			Width:    int(it.Width.ValueInt32()),
		})
	}
	return out, diags
}

func naiRealmAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfNaiRealmList, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []naiRealmModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfNaiRealmList, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfNaiRealmList{
			AuthIDs:   it.AuthIDs.ValueString(),
			AuthVals:  it.AuthVals.ValueString(),
			EapMethod: int(it.EapMethod.ValueInt32()),
			Encoding:  int(it.Encoding.ValueInt32()),
			Name:      it.Name.ValueString(),
			Status:    it.Status.ValueBool(),
		})
	}
	return out, diags
}

func qosMapDcspAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfQOSMapDcsp, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []qosMapDcspModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfQOSMapDcsp, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfQOSMapDcsp{
			High: int(it.High.ValueInt32()),
			Low:  int(it.Low.ValueInt32()),
		})
	}
	return out, diags
}

func qosMapExceptionAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfQOSMapExceptions, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []qosMapExceptionModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfQOSMapExceptions, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfQOSMapExceptions{
			Dcsp: int(it.Dcsp.ValueInt32()),
			Up:   int(it.Up.ValueInt32()),
		})
	}
	return out, diags
}

func roamingConsortiumAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfRoamingConsortiumList, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []roamingConsortiumModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfRoamingConsortiumList, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfRoamingConsortiumList{
			Name: it.Name.ValueString(),
			Oid:  it.Oid.ValueString(),
		})
	}
	return out, diags
}

func venueNameAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfVenueName, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []venueNameModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfVenueName, 0, len(items))
	for _, it := range items {
		out = append(out, unifi.Hotspot2ConfVenueName{
			Language: it.Language.ValueString(),
			Name:     it.Name.ValueString(),
			Url:      it.URL.ValueString(),
		})
	}
	return out, diags
}

func osuAsUnifi(ctx context.Context, list types.List) ([]unifi.Hotspot2ConfOsu, diag.Diagnostics) {
	var diags diag.Diagnostics
	var items []osuModel
	diags.Append(ut.ListElementsAs(ctx, list, &items)...)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]unifi.Hotspot2ConfOsu, 0, len(items))
	for _, it := range items {
		desc, d := descriptionAsUnifi(ctx, it.Description)
		diags.Append(d...)
		fn, d := friendlyNameAsUnifi(ctx, it.FriendlyName)
		diags.Append(d...)
		icon, d := osuIconAsUnifi(ctx, it.Icon)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out = append(out, unifi.Hotspot2ConfOsu{
			Description:      desc,
			FriendlyName:     fn,
			Icon:             icon,
			MethodOmaDm:      it.MethodOmaDm.ValueBool(),
			MethodSoapXmlSpp: it.MethodSoapXMLSpp.ValueBool(),
			Nai:              it.Nai.ValueString(),
			Nai2:             it.Nai2.ValueString(),
			OperatingClass:   it.OperatingClass.ValueString(),
			ServerUri:        it.ServerURI.ValueString(),
		})
	}
	return out, diags
}

// ---------------------------------------------------------------------------
// Converters: go-unifi slice -> Terraform list
// ---------------------------------------------------------------------------

func capabToList(ctx context.Context, in []unifi.Hotspot2ConfCapab) (types.List, diag.Diagnostics) {
	items := make([]capabModel, 0, len(in))
	for _, it := range in {
		items = append(items, capabModel{
			Port:     ut.Int32OrNull(it.Port),
			Protocol: ut.StringOrNull(it.Protocol),
			Status:   ut.StringOrNull(it.Status),
		})
	}
	return types.ListValueFrom(ctx, capabObjectType(), items)
}

func cellularNetworkToList(ctx context.Context, in []unifi.Hotspot2ConfCellularNetworkList) (types.List, diag.Diagnostics) {
	items := make([]cellularNetworkModel, 0, len(in))
	for _, it := range in {
		items = append(items, cellularNetworkModel{
			Mcc:  ut.Int32OrNull(it.Mcc),
			Mnc:  ut.Int32OrNull(it.Mnc),
			Name: ut.StringOrNull(it.Name),
		})
	}
	return types.ListValueFrom(ctx, cellularNetworkObjectType(), items)
}

func friendlyNameToList(ctx context.Context, in []unifi.Hotspot2ConfFriendlyName) (types.List, diag.Diagnostics) {
	items := make([]languageTextModel, 0, len(in))
	for _, it := range in {
		items = append(items, languageTextModel{
			Language: ut.StringOrNull(it.Language),
			Text:     ut.StringOrNull(it.Text),
		})
	}
	return types.ListValueFrom(ctx, languageTextObjectType(), items)
}

func descriptionToList(ctx context.Context, in []unifi.Hotspot2ConfDescription) (types.List, diag.Diagnostics) {
	items := make([]languageTextModel, 0, len(in))
	for _, it := range in {
		items = append(items, languageTextModel{
			Language: ut.StringOrNull(it.Language),
			Text:     ut.StringOrNull(it.Text),
		})
	}
	return types.ListValueFrom(ctx, languageTextObjectType(), items)
}

func osuIconToList(ctx context.Context, in []unifi.Hotspot2ConfIcon) (types.List, diag.Diagnostics) {
	items := make([]osuIconModel, 0, len(in))
	for _, it := range in {
		items = append(items, osuIconModel{Name: ut.StringOrNull(it.Name)})
	}
	return types.ListValueFrom(ctx, osuIconObjectType(), items)
}

func iconsToList(ctx context.Context, in []unifi.Hotspot2ConfIcons) (types.List, diag.Diagnostics) {
	items := make([]iconsModel, 0, len(in))
	for _, it := range in {
		items = append(items, iconsModel{
			Data:     ut.StringOrNull(it.Data),
			Filename: ut.StringOrNull(it.Filename),
			Height:   ut.Int32OrNull(it.Height),
			Language: ut.StringOrNull(it.Language),
			Media:    ut.StringOrNull(it.Media),
			Name:     ut.StringOrNull(it.Name),
			Size:     ut.Int32OrNull(it.Size),
			Width:    ut.Int32OrNull(it.Width),
		})
	}
	return types.ListValueFrom(ctx, iconsObjectType(), items)
}

func naiRealmToList(ctx context.Context, in []unifi.Hotspot2ConfNaiRealmList) (types.List, diag.Diagnostics) {
	items := make([]naiRealmModel, 0, len(in))
	for _, it := range in {
		items = append(items, naiRealmModel{
			AuthIDs:   ut.StringOrNull(it.AuthIDs),
			AuthVals:  ut.StringOrNull(it.AuthVals),
			EapMethod: ut.Int32OrNull(it.EapMethod),
			Encoding:  ut.Int32OrNull(it.Encoding),
			Name:      ut.StringOrNull(it.Name),
			Status:    types.BoolValue(it.Status),
		})
	}
	return types.ListValueFrom(ctx, naiRealmObjectType(), items)
}

func qosMapDcspToList(ctx context.Context, in []unifi.Hotspot2ConfQOSMapDcsp) (types.List, diag.Diagnostics) {
	items := make([]qosMapDcspModel, 0, len(in))
	for _, it := range in {
		items = append(items, qosMapDcspModel{
			High: ut.Int32OrNull(it.High),
			Low:  ut.Int32OrNull(it.Low),
		})
	}
	return types.ListValueFrom(ctx, qosMapDcspObjectType(), items)
}

func qosMapExceptionToList(ctx context.Context, in []unifi.Hotspot2ConfQOSMapExceptions) (types.List, diag.Diagnostics) {
	items := make([]qosMapExceptionModel, 0, len(in))
	for _, it := range in {
		items = append(items, qosMapExceptionModel{
			Dcsp: ut.Int32OrNull(it.Dcsp),
			Up:   ut.Int32OrNull(it.Up),
		})
	}
	return types.ListValueFrom(ctx, qosMapExceptionObjectType(), items)
}

func roamingConsortiumToList(ctx context.Context, in []unifi.Hotspot2ConfRoamingConsortiumList) (types.List, diag.Diagnostics) {
	items := make([]roamingConsortiumModel, 0, len(in))
	for _, it := range in {
		items = append(items, roamingConsortiumModel{
			Name: ut.StringOrNull(it.Name),
			Oid:  ut.StringOrNull(it.Oid),
		})
	}
	return types.ListValueFrom(ctx, roamingConsortiumObjectType(), items)
}

func venueNameToList(ctx context.Context, in []unifi.Hotspot2ConfVenueName) (types.List, diag.Diagnostics) {
	items := make([]venueNameModel, 0, len(in))
	for _, it := range in {
		items = append(items, venueNameModel{
			Language: ut.StringOrNull(it.Language),
			Name:     ut.StringOrNull(it.Name),
			URL:      ut.StringOrNull(it.Url),
		})
	}
	return types.ListValueFrom(ctx, venueNameObjectType(), items)
}

func osuToList(ctx context.Context, in []unifi.Hotspot2ConfOsu) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	items := make([]osuModel, 0, len(in))
	for _, it := range in {
		desc, d := descriptionToList(ctx, it.Description)
		diags.Append(d...)
		fn, d := friendlyNameToList(ctx, it.FriendlyName)
		diags.Append(d...)
		icon, d := osuIconToList(ctx, it.Icon)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(osuObjectType()), diags
		}
		items = append(items, osuModel{
			Description:      desc,
			FriendlyName:     fn,
			Icon:             icon,
			MethodOmaDm:      types.BoolValue(it.MethodOmaDm),
			MethodSoapXMLSpp: types.BoolValue(it.MethodSoapXmlSpp),
			Nai:              ut.StringOrNull(it.Nai),
			Nai2:             ut.StringOrNull(it.Nai2),
			OperatingClass:   ut.StringOrNull(it.OperatingClass),
			ServerURI:        ut.StringOrNull(it.ServerUri),
		})
	}
	list, d := types.ListValueFrom(ctx, osuObjectType(), items)
	diags.Append(d...)
	return list, diags
}

// ---------------------------------------------------------------------------
// AsUnifiModel / Merge
// ---------------------------------------------------------------------------

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *Hotspot2ConfModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var domainNames []string
	diags.Append(ut.ListElementsAs(ctx, m.DomainNameList, &domainNames)...)
	if domainNames == nil {
		domainNames = []string{}
	}

	capab, d := capabAsUnifi(ctx, m.Capab)
	diags.Append(d...)
	cellular, d := cellularNetworkAsUnifi(ctx, m.CellularNetworkList)
	diags.Append(d...)
	friendly, d := friendlyNameAsUnifi(ctx, m.FriendlyName)
	diags.Append(d...)
	icons, d := iconsAsUnifi(ctx, m.Icons)
	diags.Append(d...)
	nai, d := naiRealmAsUnifi(ctx, m.NaiRealmList)
	diags.Append(d...)
	qosDcsp, d := qosMapDcspAsUnifi(ctx, m.QOSMapDcsp)
	diags.Append(d...)
	qosExc, d := qosMapExceptionAsUnifi(ctx, m.QOSMapExceptions)
	diags.Append(d...)
	roaming, d := roamingConsortiumAsUnifi(ctx, m.RoamingConsortiumList)
	diags.Append(d...)
	venue, d := venueNameAsUnifi(ctx, m.VenueName)
	diags.Append(d...)
	osu, d := osuAsUnifi(ctx, m.Osu)
	diags.Append(d...)

	if diags.HasError() {
		return nil, diags
	}

	return &unifi.Hotspot2Conf{
		ID:   m.ID.ValueString(),
		Name: m.Name.ValueString(),

		AnqpDomainID:         int(m.AnqpDomainID.ValueInt32()),
		DeauthReqTimeout:     int(m.DeauthReqTimeout.ValueInt32()),
		GasComebackDelay:     int(m.GasComebackDelay.ValueInt32()),
		GasFragLimit:         int(m.GasFragLimit.ValueInt32()),
		IPaddrTypeAvailV4:    int(m.IPaddrTypeAvailV4.ValueInt32()),
		IPaddrTypeAvailV6:    int(m.IPaddrTypeAvailV6.ValueInt32()),
		MetricsDownlinkLoad:  int(m.MetricsDownlinkLoad.ValueInt32()),
		MetricsDownlinkSpeed: int(m.MetricsDownlinkSpd.ValueInt32()),
		MetricsMeasurement:   int(m.MetricsMeasurement.ValueInt32()),
		MetricsUplinkLoad:    int(m.MetricsUplinkLoad.ValueInt32()),
		MetricsUplinkSpeed:   int(m.MetricsUplinkSpeed.ValueInt32()),
		NetworkAuthType:      int(m.NetworkAuthType.ValueInt32()),
		NetworkType:          int(m.NetworkType.ValueInt32()),
		TCTimestamp:          int(m.TCTimestamp.ValueInt32()),
		VenueGroup:           int(m.VenueGroup.ValueInt32()),
		VenueType:            int(m.VenueType.ValueInt32()),

		DisableDgaf:             m.DisableDgaf.ValueBool(),
		GasAdvanced:             m.GasAdvanced.ValueBool(),
		HessidUsed:              m.HessidUsed.ValueBool(),
		MetricsDownlinkLoadSet:  m.MetricsDownlinkLoadSet.ValueBool(),
		MetricsDownlinkSpeedSet: m.MetricsDownlinkSpdSet.ValueBool(),
		MetricsInfoAtCapacity:   m.MetricsInfoAtCapacity.ValueBool(),
		MetricsInfoSymmetric:    m.MetricsInfoSymmetric.ValueBool(),
		MetricsMeasurementSet:   m.MetricsMeasurementSet.ValueBool(),
		MetricsStatus:           m.MetricsStatus.ValueBool(),
		MetricsUplinkLoadSet:    m.MetricsUplinkLoadSet.ValueBool(),
		MetricsUplinkSpeedSet:   m.MetricsUplinkSpeedSet.ValueBool(),
		NetworkAccessAsra:       m.NetworkAccessAsra.ValueBool(),
		NetworkAccessEsr:        m.NetworkAccessEsr.ValueBool(),
		NetworkAccessInternet:   m.NetworkAccessInternet.ValueBool(),
		NetworkAccessUesa:       m.NetworkAccessUesa.ValueBool(),
		QOSMapStatus:            m.QOSMapStatus.ValueBool(),

		Hessid:                m.Hessid.ValueString(),
		MetricsInfoLinkStatus: m.MetricsInfoLinkStatus.ValueString(),
		NetworkAuthUrl:        m.NetworkAuthURL.ValueString(),
		OsuSSID:               m.OsuSSID.ValueString(),
		SaveTimestamp:         m.SaveTimestamp.ValueString(),
		TCFilename:            m.TCFilename.ValueString(),

		DomainNameList: domainNames,

		Capab:                 capab,
		CellularNetworkList:   cellular,
		FriendlyName:          friendly,
		Icons:                 icons,
		NaiRealmList:          nai,
		QOSMapDcsp:            qosDcsp,
		QOSMapExceptions:      qosExc,
		RoamingConsortiumList: roaming,
		VenueName:             venue,
		Osu:                   osu,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *Hotspot2ConfModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
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

	// []string — coalesce nil→empty so it round-trips without a perpetual diff.
	domainNames := model.DomainNameList
	if domainNames == nil {
		domainNames = []string{}
	}
	domainList, d := types.ListValueFrom(ctx, types.StringType, domainNames)
	diags.Append(d...)
	m.DomainNameList = domainList

	// []struct lists
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

// ---------------------------------------------------------------------------
// Resource
// ---------------------------------------------------------------------------

var (
	_ resource.Resource                = &hotspot2ConfResource{}
	_ resource.ResourceWithConfigure   = &hotspot2ConfResource{}
	_ resource.ResourceWithImportState = &hotspot2ConfResource{}
	_ base.Resource                    = &hotspot2ConfResource{}
	_ base.ResourceModel               = &Hotspot2ConfModel{}
)

type hotspot2ConfResource struct {
	*base.GenericResource[*Hotspot2ConfModel]
}

// NewHotspot2ConfResource creates a new instance of the Hotspot 2.0 config resource.
func NewHotspot2ConfResource() resource.Resource {
	return &hotspot2ConfResource{
		GenericResource: base.NewGenericResource(
			"unifi_hotspot2_conf",
			func() *Hotspot2ConfModel { return &Hotspot2ConfModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetHotspot2Conf(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					conf, _ := model.(*unifi.Hotspot2Conf)
					return client.CreateHotspot2Conf(ctx, site, conf)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					conf, _ := model.(*unifi.Hotspot2Conf)
					return client.UpdateHotspot2Conf(ctx, site, conf)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteHotspot2Conf(ctx, site, id)
				},
			},
		),
	}
}

// optionalComputedInt32 builds an Optional+Computed int32 attribute with optional validators.
func optionalComputedInt32(desc string, validators ...validator.Int32) schema.Int32Attribute {
	return schema.Int32Attribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		Validators:          validators,
	}
}

// optionalComputedString builds an Optional+Computed string attribute with optional validators.
func optionalComputedString(desc string, validators ...validator.String) schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
		Validators:          validators,
	}
}

// optionalComputedBool builds an Optional+Computed bool attribute.
func optionalComputedBool(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		Computed:            true,
	}
}

func languageTextResourceAttributes(langDesc, textDesc string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"language": optionalComputedString(langDesc),
		"text":     optionalComputedString(textDesc),
	}
}

// Schema defines the schema for the resource.
func (r *hotspot2ConfResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot2_conf` resource manages a Passpoint / Hotspot 2.0 " +
			"(Wi-Fi Alliance Online Sign-Up) configuration profile in the UniFi controller. The profile " +
			"carries the ANQP, venue, roaming, NAI-realm, OSU and QoS-map parameters advertised to " +
			"Passpoint-capable clients and can be attached to a WLAN.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),

			"name": optionalComputedString("The name of the Hotspot 2.0 configuration profile."),

			// ints
			"anqp_domain_id":     optionalComputedInt32("The ANQP domain identifier (0–65535)."),
			"deauth_req_timeout": optionalComputedInt32("The number of seconds a client should wait before re-associating after a deauthentication request."),
			"gas_comeback_delay": optionalComputedInt32("The GAS comeback delay, in TUs, advertised to clients."),
			"gas_frag_limit":     optionalComputedInt32("The GAS response fragmentation limit, in bytes."),
			"ipaddr_type_avail_v4": optionalComputedInt32(
				"The advertised IPv4 address-type availability. One of `0`–`7`.",
				int32validator.OneOf(0, 1, 2, 3, 4, 5, 6, 7),
			),
			"ipaddr_type_avail_v6": optionalComputedInt32(
				"The advertised IPv6 address-type availability. One of `0`, `1` or `2`.",
				int32validator.OneOf(0, 1, 2),
			),
			"metrics_downlink_load":  optionalComputedInt32("The advertised WAN downlink load, as a fraction of 255."),
			"metrics_downlink_speed": optionalComputedInt32("The advertised WAN downlink speed, in kbps."),
			"metrics_measurement":    optionalComputedInt32("The WAN metrics measurement duration, in tenths of a second."),
			"metrics_uplink_load":    optionalComputedInt32("The advertised WAN uplink load, as a fraction of 255."),
			"metrics_uplink_speed":   optionalComputedInt32("The advertised WAN uplink speed, in kbps."),
			"network_auth_type": optionalComputedInt32(
				"The network authentication type. One of `-1`, `0`, `1`, `2` or `3`.",
				int32validator.OneOf(-1, 0, 1, 2, 3),
			),
			"network_type": optionalComputedInt32(
				"The access-network type. One of `0`, `1`, `2`, `3`, `4`, `5`, `14` or `15`.",
				int32validator.OneOf(0, 1, 2, 3, 4, 5, 14, 15),
			),
			"t_c_timestamp": optionalComputedInt32("The terms-and-conditions file timestamp."),
			"venue_group": optionalComputedInt32(
				"The venue group. One of `0`–`11`.",
				int32validator.OneOf(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11),
			),
			"venue_type": optionalComputedInt32(
				"The venue type within the venue group. One of `0`–`15`.",
				int32validator.OneOf(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15),
			),

			// bools
			"disable_dgaf":               optionalComputedBool("Whether to disable Downstream Group-Addressed Forwarding (DGAF)."),
			"gas_advanced":               optionalComputedBool("Whether advanced GAS configuration is enabled."),
			"hessid_used":                optionalComputedBool("Whether the homogeneous ESS identifier (HESSID) is advertised."),
			"metrics_downlink_load_set":  optionalComputedBool("Whether the WAN downlink load value is set."),
			"metrics_downlink_speed_set": optionalComputedBool("Whether the WAN downlink speed value is set."),
			"metrics_info_at_capacity":   optionalComputedBool("Whether the WAN link is advertised as being at capacity."),
			"metrics_info_symmetric":     optionalComputedBool("Whether the WAN link is advertised as symmetric."),
			"metrics_measurement_set":    optionalComputedBool("Whether the WAN metrics measurement value is set."),
			"metrics_status":             optionalComputedBool("Whether WAN metrics advertisement is enabled."),
			"metrics_uplink_load_set":    optionalComputedBool("Whether the WAN uplink load value is set."),
			"metrics_uplink_speed_set":   optionalComputedBool("Whether the WAN uplink speed value is set."),
			"network_access_asra":        optionalComputedBool("Whether Additional Steps Required for Access (ASRA) is advertised."),
			"network_access_esr":         optionalComputedBool("Whether Emergency Services Reachable (ESR) is advertised."),
			"network_access_internet":    optionalComputedBool("Whether the network is advertised as providing internet access."),
			"network_access_uesa":        optionalComputedBool("Whether Unauthenticated Emergency Service Accessible (UESA) is advertised."),
			"qos_map_status":             optionalComputedBool("Whether the QoS map is advertised."),

			// strings
			"hessid": optionalComputedString("The homogeneous ESS identifier (HESSID), a MAC-formatted address, or empty."),
			"metrics_info_link_status": optionalComputedString(
				"The advertised WAN link status. One of `up`, `down` or `test`.",
				stringvalidator.OneOf("up", "down", "test"),
			),
			"network_auth_url": optionalComputedString("The network authentication redirect URL."),
			"osu_ssid":         optionalComputedString("The SSID used for Online Sign-Up (OSU)."),
			"save_timestamp":   optionalComputedString("The server-managed save timestamp of the profile."),
			"t_c_filename":     optionalComputedString("The filename of the uploaded terms-and-conditions document."),

			// []string
			"domain_name_list": schema.ListAttribute{
				MarkdownDescription: "The list of domain names advertised in the ANQP domain-name element.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},

			// []struct lists
			"capab": schema.ListNestedAttribute{
				MarkdownDescription: "The ANQP IP-connectivity capability list (protocol/port reachability).",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": optionalComputedInt32("The port number (0–65535)."),
						"protocol": optionalComputedString(
							"The protocol. One of `icmp`, `tcp_udp`, `tcp`, `udp` or `esp`.",
							stringvalidator.OneOf("icmp", "tcp_udp", "tcp", "udp", "esp"),
						),
						"status": optionalComputedString(
							"The reachability status. One of `closed`, `open` or `unknown`.",
							stringvalidator.OneOf("closed", "open", "unknown"),
						),
					},
				},
			},
			"cellular_network_list": schema.ListNestedAttribute{
				MarkdownDescription: "The 3GPP cellular network list advertised via ANQP.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mcc":  optionalComputedInt32("The Mobile Country Code (MCC)."),
						"mnc":  optionalComputedInt32("The Mobile Network Code (MNC)."),
						"name": optionalComputedString("The cellular network name."),
					},
				},
			},
			"friendly_name": schema.ListNestedAttribute{
				MarkdownDescription: "The operator friendly names, one per language.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: languageTextResourceAttributes(
						"The ISO-639 three-letter language code.",
						"The friendly name text (1–128 characters).",
					),
				},
			},
			"icons": schema.ListNestedAttribute{
				MarkdownDescription: "The operator icons advertised to clients.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"data":     optionalComputedString("The base64-encoded icon data."),
						"filename": optionalComputedString("The icon filename."),
						"height":   optionalComputedInt32("The icon height, in pixels."),
						"language": optionalComputedString("The ISO-639 three-letter language code."),
						"media":    optionalComputedString("The icon media (MIME) type."),
						"name":     optionalComputedString("The icon name."),
						"size":     optionalComputedInt32("The icon size, in bytes."),
						"width":    optionalComputedInt32("The icon width, in pixels."),
					},
				},
			},
			"nai_realm_list": schema.ListNestedAttribute{
				MarkdownDescription: "The NAI realm list advertised via ANQP.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"auth_ids":  optionalComputedString("The authentication parameter IDs."),
						"auth_vals": optionalComputedString("The authentication parameter values."),
						"eap_method": optionalComputedInt32(
							"The EAP method. One of `13`, `21`, `18`, `23` or `50`.",
							int32validator.OneOf(13, 21, 18, 23, 50),
						),
						"encoding": optionalComputedInt32(
							"The realm encoding. One of `0` or `1`.",
							int32validator.OneOf(0, 1),
						),
						"name":   optionalComputedString("The NAI realm name."),
						"status": optionalComputedBool("Whether this realm entry is enabled."),
					},
				},
			},
			"qos_map_dcsp": schema.ListNestedAttribute{
				MarkdownDescription: "The QoS map DSCP range entries.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"high": optionalComputedInt32("The high end of the DSCP range."),
						"low":  optionalComputedInt32("The low end of the DSCP range."),
					},
				},
			},
			"qos_map_exceptions": schema.ListNestedAttribute{
				MarkdownDescription: "The QoS map DSCP-to-user-priority exceptions.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dcsp": optionalComputedInt32("The DSCP value."),
						"up":   optionalComputedInt32("The user priority (0–7)."),
					},
				},
			},
			"roaming_consortium_list": schema.ListNestedAttribute{
				MarkdownDescription: "The roaming consortium organisation identifiers advertised via ANQP.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": optionalComputedString("The roaming consortium name."),
						"oid":  optionalComputedString("The roaming consortium organisation identifier (OID)."),
					},
				},
			},
			"venue_name": schema.ListNestedAttribute{
				MarkdownDescription: "The venue names, one per language.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"language": optionalComputedString("The ISO-639 three-letter language code."),
						"name":     optionalComputedString("The venue name text."),
						"url":      optionalComputedString("The venue URL."),
					},
				},
			},
			"osu": schema.ListNestedAttribute{
				MarkdownDescription: "The Online Sign-Up (OSU) provider list.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.ListNestedAttribute{
							MarkdownDescription: "The OSU provider descriptions, one per language.",
							Optional:            true,
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: languageTextResourceAttributes(
									"The ISO-639 three-letter language code.",
									"The description text (1–128 characters).",
								),
							},
						},
						"friendly_name": schema.ListNestedAttribute{
							MarkdownDescription: "The OSU provider friendly names, one per language.",
							Optional:            true,
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: languageTextResourceAttributes(
									"The ISO-639 three-letter language code.",
									"The friendly name text (1–128 characters).",
								),
							},
						},
						"icon": schema.ListNestedAttribute{
							MarkdownDescription: "The OSU provider icons (referencing entries in the top-level `icons` list by name).",
							Optional:            true,
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": optionalComputedString("The referenced icon name."),
								},
							},
						},
						"method_oma_dm":       optionalComputedBool("Whether the OMA-DM provisioning method is offered."),
						"method_soap_xml_spp": optionalComputedBool("Whether the SOAP-XML SPP provisioning method is offered."),
						"nai":                 optionalComputedString("The OSU Network Access Identifier."),
						"nai2":                optionalComputedString("The secondary OSU Network Access Identifier."),
						"operating_class":     optionalComputedString("The OSU operating class (hex-encoded)."),
						"server_uri":          optionalComputedString("The OSU server URI."),
					},
				},
			},
		},
	}
}
