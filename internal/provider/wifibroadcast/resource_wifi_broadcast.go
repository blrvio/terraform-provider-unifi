package wifibroadcast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// WiFi broadcast type discriminators, matching the go-unifi Official API union.
const (
	wifiBroadcastTypeStandard = "STANDARD"
	wifiBroadcastTypeIoT      = "IOT_OPTIMIZED"
)

var (
	_ resource.Resource                = &wifiBroadcastResource{}
	_ resource.ResourceWithConfigure   = &wifiBroadcastResource{}
	_ resource.ResourceWithImportState = &wifiBroadcastResource{}
	_ resource.ResourceWithModifyPlan  = &wifiBroadcastResource{}
	_ base.ResourceModel               = &wifiBroadcastModel{}
	_ base.Resource                    = &wifiBroadcastResource{}
)

type wifiBroadcastModel struct {
	base.Model
	Name                                types.String `tfsdk:"name"`
	Type                                types.String `tfsdk:"type"`
	Enabled                             types.Bool   `tfsdk:"enabled"`
	HideName                            types.Bool   `tfsdk:"hide_name"`
	ClientIsolationEnabled              types.Bool   `tfsdk:"client_isolation_enabled"`
	UapsdEnabled                        types.Bool   `tfsdk:"uapsd_enabled"`
	MulticastToUnicastConversionEnabled types.Bool   `tfsdk:"multicast_to_unicast_conversion_enabled"`
	Channel2gLockedTo6                  types.Bool   `tfsdk:"channel_2g_locked_to_6"`
	DtimPeriod2gLockedTo3               types.Bool   `tfsdk:"dtim_period_2g_locked_to_3"`

	// The following attributes carry raw Official-API JSON object shapes. They
	// are surfaced as JSON strings because the underlying types are complex
	// (unions/nested objects) that don't map cleanly to a flat schema.
	SecurityConfiguration    types.String `tfsdk:"security_configuration"`
	Network                  types.String `tfsdk:"network"`
	BasicDataRate            types.String `tfsdk:"basic_data_rate"`
	BlackoutSchedule         types.String `tfsdk:"blackout_schedule"`
	BroadcastingDeviceFilter types.String `tfsdk:"broadcasting_device_filter"`
	ClientFilteringPolicy    types.String `tfsdk:"client_filtering_policy"`
	MdnsProxy                types.String `tfsdk:"mdns_proxy"`
	MulticastFilteringPolicy types.String `tfsdk:"multicast_filtering_policy"`
}

// boolWithDefault returns the configured bool, falling back to def when the
// value is null/unknown (belt-and-braces around the Computed defaults).
func boolWithDefault(v types.Bool, def bool) bool {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	return v.ValueBool()
}

// decodeJSONPtr unmarshals an optional JSON-string attribute into a pointer
// field on the Official-API body, leaving it nil when the attribute is empty.
func decodeJSONPtr[T any](v types.String, dst **T, attr string) diag.Diagnostics {
	var diags diag.Diagnostics
	if ut.IsEmptyString(v) {
		*dst = nil
		return diags
	}
	var out T
	if err := json.Unmarshal([]byte(v.ValueString()), &out); err != nil {
		diags.AddError("Invalid "+attr+" JSON", err.Error())
		return diags
	}
	*dst = &out
	return diags
}

// marshalJSONPtr renders an optional pointer field back to a compact JSON
// string, or null when the pointer is nil.
func marshalJSONPtr[T any](p *T, attr string) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	if p == nil {
		return types.StringNull(), diags
	}
	raw, err := json.Marshal(p)
	if err != nil {
		diags.AddError("Failed to encode "+attr, err.Error())
		return types.StringNull(), diags
	}
	return types.StringValue(string(raw)), diags
}

// AsUnifiModel builds the Official-API create/update body as a PLAIN struct
// literal. WifiBroadcastCreateOrUpdate.MarshalJSON overlays every top-level
// field over its (nil) union, so setting Type plus the common fields directly
// is sufficient — no From* union helper is used and variant-specific extras
// (STANDARD/IOT_OPTIMIZED) are left to the server's defaults.
func (m *wifiBroadcastModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := &official.WifiBroadcastCreateOrUpdate{
		Name:                                m.Name.ValueString(),
		Type:                                m.Type.ValueString(),
		Enabled:                             boolWithDefault(m.Enabled, true),
		HideName:                            boolWithDefault(m.HideName, false),
		ClientIsolationEnabled:              boolWithDefault(m.ClientIsolationEnabled, false),
		UapsdEnabled:                        boolWithDefault(m.UapsdEnabled, false),
		MulticastToUnicastConversionEnabled: boolWithDefault(m.MulticastToUnicastConversionEnabled, false),
		Channel2gLockedTo6:                  boolWithDefault(m.Channel2gLockedTo6, false),
		DtimPeriod2gLockedTo3:               boolWithDefault(m.DtimPeriod2gLockedTo3, false),
	}

	// security_configuration is required (a complex union sent verbatim).
	if ut.IsEmptyString(m.SecurityConfiguration) {
		diags.AddError("Missing required attribute", "`security_configuration` must be set.")
		return nil, diags
	}
	if err := json.Unmarshal([]byte(m.SecurityConfiguration.ValueString()), &body.SecurityConfiguration); err != nil {
		diags.AddError("Invalid security_configuration JSON", err.Error())
		return nil, diags
	}

	diags.Append(decodeJSONPtr(m.Network, &body.Network, "network")...)
	diags.Append(decodeJSONPtr(m.BasicDataRate, &body.BasicDataRateKbpsByFrequencyGHz, "basic_data_rate")...)
	diags.Append(decodeJSONPtr(m.BlackoutSchedule, &body.BlackoutScheduleConfiguration, "blackout_schedule")...)
	diags.Append(decodeJSONPtr(m.BroadcastingDeviceFilter, &body.BroadcastingDeviceFilter, "broadcasting_device_filter")...)
	diags.Append(decodeJSONPtr(m.ClientFilteringPolicy, &body.ClientFilteringPolicy, "client_filtering_policy")...)
	diags.Append(decodeJSONPtr(m.MdnsProxy, &body.MdnsProxyConfiguration, "mdns_proxy")...)
	diags.Append(decodeJSONPtr(m.MulticastFilteringPolicy, &body.MulticastFilteringPolicy, "multicast_filtering_policy")...)
	if diags.HasError() {
		return nil, diags
	}
	return body, diags
}

// asWifiBroadcastBody performs the checked type assertion from the AsUnifiModel
// interface{} return to the concrete Official-API body.
func asWifiBroadcastBody(body interface{}) (*official.WifiBroadcastCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.WifiBroadcastCreateOrUpdate)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.WifiBroadcastCreateOrUpdate, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API WifiBroadcastDetails response.
func (m *wifiBroadcastModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	d, ok := other.(*official.WifiBroadcastDetails)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.WifiBroadcastDetails, got %T", other))
		return diags
	}

	m.SetID(d.Id.String())
	m.Name = types.StringValue(d.Name)
	m.Type = types.StringValue(d.Type)
	m.Enabled = types.BoolValue(d.Enabled)
	m.HideName = types.BoolValue(d.HideName)
	m.ClientIsolationEnabled = types.BoolValue(d.ClientIsolationEnabled)
	m.UapsdEnabled = types.BoolValue(d.UapsdEnabled)
	m.MulticastToUnicastConversionEnabled = types.BoolValue(d.MulticastToUnicastConversionEnabled)
	m.Channel2gLockedTo6 = types.BoolValue(d.Channel2gLockedTo6)
	m.DtimPeriod2gLockedTo3 = types.BoolValue(d.DtimPeriod2gLockedTo3)

	raw, err := json.Marshal(d.SecurityConfiguration)
	if err != nil {
		diags.AddError("Failed to encode security_configuration", err.Error())
		return diags
	}
	m.SecurityConfiguration = types.StringValue(string(raw))

	var sd diag.Diagnostics
	m.Network, sd = marshalJSONPtr(d.Network, "network")
	diags.Append(sd...)
	m.BasicDataRate, sd = marshalJSONPtr(d.BasicDataRateKbpsByFrequencyGHz, "basic_data_rate")
	diags.Append(sd...)
	m.BlackoutSchedule, sd = marshalJSONPtr(d.BlackoutScheduleConfiguration, "blackout_schedule")
	diags.Append(sd...)
	m.BroadcastingDeviceFilter, sd = marshalJSONPtr(d.BroadcastingDeviceFilter, "broadcasting_device_filter")
	diags.Append(sd...)
	m.ClientFilteringPolicy, sd = marshalJSONPtr(d.ClientFilteringPolicy, "client_filtering_policy")
	diags.Append(sd...)
	m.MdnsProxy, sd = marshalJSONPtr(d.MdnsProxyConfiguration, "mdns_proxy")
	diags.Append(sd...)
	m.MulticastFilteringPolicy, sd = marshalJSONPtr(d.MulticastFilteringPolicy, "multicast_filtering_policy")
	diags.Append(sd...)

	return diags
}

type wifiBroadcastResource struct {
	*base.GenericResource[*wifiBroadcastModel]
}

// NewWifiBroadcastResource creates the unifi_wifi_broadcast resource. It embeds
// a GenericResource purely for Configure/version-validator wiring; CRUD is
// custom because the Official API is keyed by site/entity UUID.
func NewWifiBroadcastResource() resource.Resource {
	r := &wifiBroadcastResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_wifi_broadcast",
		func() *wifiBroadcastModel { return &wifiBroadcastModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *wifiBroadcastResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	jsonNote := " Accepts a JSON string matching the corresponding UniFi **Official API** object shape."
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a WiFi broadcast (SSID) via the UniFi **Official API** (`integration/v1`). " +
			"Requires a controller running version 10.1.78 or later with API-key authentication.\n\n" +
			"**Limitation:** this resource manages only the fields common to both the `STANDARD` and " +
			"`IOT_OPTIMIZED` broadcast variants (plus the nested JSON objects below). Variant-specific " +
			"extras (e.g. band steering, broadcasting frequencies, hotspot/captive portal) are not " +
			"exposed here — the controller applies its defaults for those.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The SSID / name of the WiFi broadcast.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The broadcast type. One of `STANDARD`, `IOT_OPTIMIZED`. " +
					"Changing the type forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(wifiBroadcastTypeStandard, wifiBroadcastTypeIoT),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the WiFi broadcast is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"hide_name": schema.BoolAttribute{
				MarkdownDescription: "Whether the SSID is hidden from beacon frames. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"client_isolation_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether client isolation is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"uapsd_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Unscheduled Automatic Power Save Delivery (U-APSD) is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"multicast_to_unicast_conversion_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether multicast-to-unicast conversion is enabled. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"channel_2g_locked_to_6": schema.BoolAttribute{
				MarkdownDescription: "Locks the 2.4GHz radio channel to 6 on all broadcasting devices. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"dtim_period_2g_locked_to_3": schema.BoolAttribute{
				MarkdownDescription: "Locks the DTIM period to 3 for the 2.4GHz radio. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"security_configuration": schema.StringAttribute{
				MarkdownDescription: "The WiFi security configuration." + jsonNote +
					" This is a required complex union (e.g. `{\"type\":\"WPA2_PERSONAL\",\"passphrase\":\"...\"}`).",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"network": schema.StringAttribute{
				MarkdownDescription: "The network reference the broadcast is attached to." + jsonNote,
				Optional:            true,
			},
			"basic_data_rate": schema.StringAttribute{
				MarkdownDescription: "The basic data rate configuration by frequency (kbps)." + jsonNote,
				Optional:            true,
			},
			"blackout_schedule": schema.StringAttribute{
				MarkdownDescription: "The blackout schedule configuration." + jsonNote,
				Optional:            true,
			},
			"broadcasting_device_filter": schema.StringAttribute{
				MarkdownDescription: "The broadcasting device filter." + jsonNote,
				Optional:            true,
			},
			"client_filtering_policy": schema.StringAttribute{
				MarkdownDescription: "The client (MAC) filtering policy." + jsonNote,
				Optional:            true,
			},
			"mdns_proxy": schema.StringAttribute{
				MarkdownDescription: "The mDNS proxy configuration." + jsonNote,
				Optional:            true,
			},
			"multicast_filtering_policy": schema.StringAttribute{
				MarkdownDescription: "The multicast filtering policy." + jsonNote,
				Optional:            true,
			},
		},
	}
}

func (r *wifiBroadcastResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
}

func (r *wifiBroadcastResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan wifiBroadcastModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&plan)
	siteID, diags := client.ResolveSiteUUID(ctx, siteName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createBody, diags := asWifiBroadcastBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := client.Official().WifiBroadcasts().Create(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating WiFi broadcast", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wifiBroadcastResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state wifiBroadcastModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	broadcast, err := client.Official().WifiBroadcasts().Get(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading WiFi broadcast", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, broadcast)...)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *wifiBroadcastResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state wifiBroadcastModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&plan)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateBody, diags := asWifiBroadcastBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := client.Official().WifiBroadcasts().Update(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating WiFi broadcast", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *wifiBroadcastResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state wifiBroadcastModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := client.Official().WifiBroadcasts().Delete(ctx, siteID, id, nil); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting WiFi broadcast", err)...)
	}
}

func (r *wifiBroadcastResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
