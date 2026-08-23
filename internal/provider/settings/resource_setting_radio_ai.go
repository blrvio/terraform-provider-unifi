package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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

// radioAiChannelsBlacklistModel represents a channel/width/radio combination
// that AI channel optimization must avoid.
type radioAiChannelsBlacklistModel struct {
	Channel      types.Int32  `tfsdk:"channel"`
	ChannelWidth types.Int32  `tfsdk:"channel_width"`
	Radio        types.String `tfsdk:"radio"`
}

func (m *radioAiChannelsBlacklistModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"channel":       types.Int32Type,
		"channel_width": types.Int32Type,
		"radio":         types.StringType,
	}
}

// radioAiRadiosConfigurationModel represents per-radio AI optimization
// parameters.
type radioAiRadiosConfigurationModel struct {
	ChannelWidth types.Int32  `tfsdk:"channel_width"`
	Dfs          types.Bool   `tfsdk:"dfs"`
	Radio        types.String `tfsdk:"radio"`
}

func (m *radioAiRadiosConfigurationModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"channel_width": types.Int32Type,
		"dfs":           types.BoolType,
		"radio":         types.StringType,
	}
}

// radioAiModel represents the AI-driven radio (channel and power) optimization
// settings for a UniFi site.
type radioAiModel struct {
	base.Model
	AutoAdjustChannelsToCountry types.Bool   `tfsdk:"auto_adjust_channels_to_country"`
	AutoChannelPresetsType      types.String `tfsdk:"auto_channel_presets_type"`
	Channels6E                  types.List   `tfsdk:"channels_6e"`
	ChannelsBlacklist           types.List   `tfsdk:"channels_blacklist"`
	ChannelsNa                  types.List   `tfsdk:"channels_na"`
	ChannelsNg                  types.List   `tfsdk:"channels_ng"`
	CronExpr                    types.String `tfsdk:"cron_expr"`
	Default                     types.Bool   `tfsdk:"default"`
	Enabled                     types.Bool   `tfsdk:"enabled"`
	ExcludeDevices              types.List   `tfsdk:"exclude_devices"`
	HighPriorityDevices         types.List   `tfsdk:"high_priority_devices"`
	HtModesNa                   types.List   `tfsdk:"ht_modes_na"`
	HtModesNg                   types.List   `tfsdk:"ht_modes_ng"`
	Optimize                    types.List   `tfsdk:"optimize"`
	Radios                      types.List   `tfsdk:"radios"`
	RadiosConfiguration         types.List   `tfsdk:"radios_configuration"`
	SettingPreference           types.String `tfsdk:"setting_preference"`
	UseXy                       types.Bool   `tfsdk:"use_xy"`
}

func (d *radioAiModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingRadioAi{
		ID:                          d.ID.ValueString(),
		AutoAdjustChannelsToCountry: d.AutoAdjustChannelsToCountry.ValueBool(),
		AutoChannelPresetsType:      d.AutoChannelPresetsType.ValueString(),
		CronExpr:                    d.CronExpr.ValueString(),
		Default:                     d.Default.ValueBool(),
		Enabled:                     d.Enabled.ValueBool(),
		SettingPreference:           d.SettingPreference.ValueString(),
		UseXy:                       d.UseXy.ValueBool(),
	}

	channels6E, d1 := int32ListToIntSlice(ctx, d.Channels6E)
	if d1.HasError() {
		diags.Append(d1...)
		return nil, diags
	}
	model.Channels6E = channels6E

	channelsNa, d2 := int32ListToIntSlice(ctx, d.ChannelsNa)
	if d2.HasError() {
		diags.Append(d2...)
		return nil, diags
	}
	model.ChannelsNa = channelsNa

	channelsNg, d3 := int32ListToIntSlice(ctx, d.ChannelsNg)
	if d3.HasError() {
		diags.Append(d3...)
		return nil, diags
	}
	model.ChannelsNg = channelsNg

	htModesNa, d4 := int32ListToIntSlice(ctx, d.HtModesNa)
	if d4.HasError() {
		diags.Append(d4...)
		return nil, diags
	}
	model.HtModesNa = htModesNa

	htModesNg, d5 := int32ListToIntSlice(ctx, d.HtModesNg)
	if d5.HasError() {
		diags.Append(d5...)
		return nil, diags
	}
	model.HtModesNg = htModesNg

	if ut.IsDefined(d.ExcludeDevices) {
		var xs []string
		diags.Append(ut.ListElementsAs(ctx, d.ExcludeDevices, &xs)...)
		if diags.HasError() {
			return nil, diags
		}
		model.ExcludeDevices = xs
	}

	if ut.IsDefined(d.HighPriorityDevices) {
		var xs []string
		diags.Append(ut.ListElementsAs(ctx, d.HighPriorityDevices, &xs)...)
		if diags.HasError() {
			return nil, diags
		}
		model.HighPriorityDevices = xs
	}

	if ut.IsDefined(d.Optimize) {
		var xs []string
		diags.Append(ut.ListElementsAs(ctx, d.Optimize, &xs)...)
		if diags.HasError() {
			return nil, diags
		}
		model.Optimize = xs
	}

	if ut.IsDefined(d.Radios) {
		var xs []string
		diags.Append(ut.ListElementsAs(ctx, d.Radios, &xs)...)
		if diags.HasError() {
			return nil, diags
		}
		model.Radios = xs
	}

	if ut.IsDefined(d.ChannelsBlacklist) {
		var items []radioAiChannelsBlacklistModel
		diags.Append(d.ChannelsBlacklist.ElementsAs(ctx, &items, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.ChannelsBlacklist = make([]unifi.SettingRadioAiChannelsBlacklist, 0, len(items))
		for _, it := range items {
			model.ChannelsBlacklist = append(model.ChannelsBlacklist, unifi.SettingRadioAiChannelsBlacklist{
				Channel:      int(it.Channel.ValueInt32()),
				ChannelWidth: int(it.ChannelWidth.ValueInt32()),
				Radio:        it.Radio.ValueString(),
			})
		}
	}

	if ut.IsDefined(d.RadiosConfiguration) {
		var items []radioAiRadiosConfigurationModel
		diags.Append(d.RadiosConfiguration.ElementsAs(ctx, &items, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.RadiosConfiguration = make([]unifi.SettingRadioAiRadiosConfiguration, 0, len(items))
		for _, it := range items {
			model.RadiosConfiguration = append(model.RadiosConfiguration, unifi.SettingRadioAiRadiosConfiguration{
				ChannelWidth: int(it.ChannelWidth.ValueInt32()),
				Dfs:          it.Dfs.ValueBool(),
				Radio:        it.Radio.ValueString(),
			})
		}
	}

	return model, diags
}

func (d *radioAiModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingRadioAi)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingRadioAi")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.AutoAdjustChannelsToCountry = types.BoolValue(model.AutoAdjustChannelsToCountry)
	d.AutoChannelPresetsType = ut.StringOrNull(model.AutoChannelPresetsType)
	d.CronExpr = ut.StringOrNull(model.CronExpr)
	d.Default = types.BoolValue(model.Default)
	d.Enabled = types.BoolValue(model.Enabled)
	d.SettingPreference = ut.StringOrNull(model.SettingPreference)
	d.UseXy = types.BoolValue(model.UseXy)

	channels6E, d1 := intSliceToInt32List(ctx, model.Channels6E)
	if d1.HasError() {
		diags.Append(d1...)
		return diags
	}
	d.Channels6E = channels6E

	channelsNa, d2 := intSliceToInt32List(ctx, model.ChannelsNa)
	if d2.HasError() {
		diags.Append(d2...)
		return diags
	}
	d.ChannelsNa = channelsNa

	channelsNg, d3 := intSliceToInt32List(ctx, model.ChannelsNg)
	if d3.HasError() {
		diags.Append(d3...)
		return diags
	}
	d.ChannelsNg = channelsNg

	htModesNa, d4 := intSliceToInt32List(ctx, model.HtModesNa)
	if d4.HasError() {
		diags.Append(d4...)
		return diags
	}
	d.HtModesNa = htModesNa

	htModesNg, d5 := intSliceToInt32List(ctx, model.HtModesNg)
	if d5.HasError() {
		diags.Append(d5...)
		return diags
	}
	d.HtModesNg = htModesNg

	excludeDevices, excludeDiags := types.ListValueFrom(ctx, types.StringType, model.ExcludeDevices)
	diags.Append(excludeDiags...)
	if diags.HasError() {
		return diags
	}
	d.ExcludeDevices = excludeDevices

	highPriorityDevices, highPriorityDiags := types.ListValueFrom(ctx, types.StringType, model.HighPriorityDevices)
	diags.Append(highPriorityDiags...)
	if diags.HasError() {
		return diags
	}
	d.HighPriorityDevices = highPriorityDevices

	optimize, optimizeDiags := types.ListValueFrom(ctx, types.StringType, model.Optimize)
	diags.Append(optimizeDiags...)
	if diags.HasError() {
		return diags
	}
	d.Optimize = optimize

	radios, radiosDiags := types.ListValueFrom(ctx, types.StringType, model.Radios)
	diags.Append(radiosDiags...)
	if diags.HasError() {
		return diags
	}
	d.Radios = radios

	blacklistItems := make([]radioAiChannelsBlacklistModel, 0, len(model.ChannelsBlacklist))
	for _, it := range model.ChannelsBlacklist {
		blacklistItems = append(blacklistItems, radioAiChannelsBlacklistModel{
			Channel:      ut.Int32OrNull(it.Channel),
			ChannelWidth: ut.Int32OrNull(it.ChannelWidth),
			Radio:        ut.StringOrNull(it.Radio),
		})
	}
	blacklistList, blacklistDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&radioAiChannelsBlacklistModel{}).AttributeTypes()}, blacklistItems)
	diags.Append(blacklistDiags...)
	if diags.HasError() {
		return diags
	}
	d.ChannelsBlacklist = blacklistList

	radiosConfigItems := make([]radioAiRadiosConfigurationModel, 0, len(model.RadiosConfiguration))
	for _, it := range model.RadiosConfiguration {
		radiosConfigItems = append(radiosConfigItems, radioAiRadiosConfigurationModel{
			ChannelWidth: ut.Int32OrNull(it.ChannelWidth),
			Dfs:          types.BoolValue(it.Dfs),
			Radio:        ut.StringOrNull(it.Radio),
		})
	}
	radiosConfigList, radiosConfigDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&radioAiRadiosConfigurationModel{}).AttributeTypes()}, radiosConfigItems)
	diags.Append(radiosConfigDiags...)
	if diags.HasError() {
		return diags
	}
	d.RadiosConfiguration = radiosConfigList

	return diags
}

// int32ListToIntSlice converts a Framework Int32 list into a []int for the
// go-unifi model, returning nil when the list is not set.
func int32ListToIntSlice(ctx context.Context, list types.List) ([]int, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !ut.IsDefined(list) {
		return nil, diags
	}
	var xs []int32
	diags.Append(ut.ListElementsAs(ctx, list, &xs)...)
	if diags.HasError() {
		return nil, diags
	}
	conv := make([]int, 0, len(xs))
	for _, x := range xs {
		conv = append(conv, int(x))
	}
	return conv, diags
}

// intSliceToInt32List converts a go-unifi []int into a Framework Int32 list,
// guarding against a nil source slice.
func intSliceToInt32List(ctx context.Context, xs []int) (types.List, diag.Diagnostics) {
	conv := make([]int32, 0, len(xs))
	for _, x := range xs {
		conv = append(conv, int32(x))
	}
	return types.ListValueFrom(ctx, types.Int32Type, conv)
}

var (
	_ base.ResourceModel               = &radioAiModel{}
	_ resource.Resource                = &radioAiResource{}
	_ resource.ResourceWithConfigure   = &radioAiResource{}
	_ resource.ResourceWithImportState = &radioAiResource{}
)

type radioAiResource struct {
	*base.GenericResource[*radioAiModel]
}

func (r *radioAiResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_radio_ai` resource manages AI-driven radio optimization (\"AI WiFi\") for a " +
			"UniFi site. When enabled, the controller periodically analyzes the RF environment and adjusts channel, channel " +
			"width, and transmit power on managed access points. This resource controls the scope of that optimization " +
			"(which radios and channels are eligible), the schedule (`cron_expr`), and per-device or per-radio overrides.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether AI radio optimization is enabled for the site.",
				Optional:            true,
				Computed:            true,
			},
			"auto_adjust_channels_to_country": schema.BoolAttribute{
				MarkdownDescription: "Whether the eligible channel set is automatically constrained to those permitted by the site's country/regulatory domain.",
				Optional:            true,
				Computed:            true,
			},
			"auto_channel_presets_type": schema.StringAttribute{
				MarkdownDescription: "Optimization preset governing how aggressively channels are chosen. One of `maximum_speed`, `conservative`, or `custom`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("maximum_speed", "conservative", "custom"),
				},
			},
			"channels_6e": schema.ListAttribute{
				MarkdownDescription: "Eligible channels for the 6 GHz (6E) radio during optimization.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int32Type,
			},
			"channels_na": schema.ListAttribute{
				MarkdownDescription: "Eligible channels for the 5 GHz (na) radio during optimization.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int32Type,
			},
			"channels_ng": schema.ListAttribute{
				MarkdownDescription: "Eligible channels for the 2.4 GHz (ng) radio during optimization.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int32Type,
			},
			"ht_modes_na": schema.ListAttribute{
				MarkdownDescription: "Eligible channel widths (MHz) for the 5 GHz (na) radio, e.g. `20`, `40`, `80`, `160`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int32Type,
			},
			"ht_modes_ng": schema.ListAttribute{
				MarkdownDescription: "Eligible channel widths (MHz) for the 2.4 GHz (ng) radio, e.g. `20`, `40`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int32Type,
			},
			"cron_expr": schema.StringAttribute{
				MarkdownDescription: "Cron expression defining when the optimization runs (e.g. `0 3 * * *` for 03:00 daily).",
				Optional:            true,
				Computed:            true,
			},
			"default": schema.BoolAttribute{
				MarkdownDescription: "Whether this configuration is the controller's default AI optimization profile.",
				Optional:            true,
				Computed:            true,
			},
			"exclude_devices": schema.ListAttribute{
				MarkdownDescription: "MAC addresses of access points excluded from AI optimization.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"high_priority_devices": schema.ListAttribute{
				MarkdownDescription: "MAC addresses of access points the optimizer should prioritize.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"optimize": schema.ListAttribute{
				MarkdownDescription: "Which radio properties the optimizer may adjust. Values: `channel`, `power`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf("channel", "power")),
				},
			},
			"radios": schema.ListAttribute{
				MarkdownDescription: "Which radio bands participate in optimization. Values: `na` (5 GHz), `ng` (2.4 GHz), `6e` (6 GHz).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf("na", "ng", "6e")),
				},
			},
			"setting_preference": schema.StringAttribute{
				MarkdownDescription: "Whether the settings are managed automatically by the controller (`auto`) or overridden manually (`manual`).",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "manual"),
				},
			},
			"use_xy": schema.BoolAttribute{
				MarkdownDescription: "Whether the newer (XY) optimization engine is used.",
				Optional:            true,
				Computed:            true,
			},
			"channels_blacklist": schema.ListNestedAttribute{
				MarkdownDescription: "Channel/width/radio combinations the optimizer must never select.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel": schema.Int32Attribute{
							MarkdownDescription: "Channel number to exclude.",
							Optional:            true,
							Computed:            true,
						},
						"channel_width": schema.Int32Attribute{
							MarkdownDescription: "Channel width (MHz) the exclusion applies to.",
							Optional:            true,
							Computed:            true,
						},
						"radio": schema.StringAttribute{
							MarkdownDescription: "Radio band the exclusion applies to. One of `na`, `ng`, `6e`.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("na", "ng", "6e"),
							},
						},
					},
				},
			},
			"radios_configuration": schema.ListNestedAttribute{
				MarkdownDescription: "Per-radio optimization parameters such as allowed channel width and DFS usage.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel_width": schema.Int32Attribute{
							MarkdownDescription: "Channel width (MHz) the optimizer may use for this radio.",
							Optional:            true,
							Computed:            true,
						},
						"dfs": schema.BoolAttribute{
							MarkdownDescription: "Whether DFS (radar-shared) channels may be selected for this radio.",
							Optional:            true,
							Computed:            true,
						},
						"radio": schema.StringAttribute{
							MarkdownDescription: "Radio band this configuration applies to. One of `na`, `ng`, `6e`.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("na", "ng", "6e"),
							},
						},
					},
				},
			},
		},
	}
}

// NewRadioAiResource creates a new instance of the AI radio optimization setting resource.
func NewRadioAiResource() resource.Resource {
	r := &radioAiResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_radio_ai",
		func() *radioAiModel { return &radioAiModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingRadioAi(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingRadioAi)
			return client.UpdateSettingRadioAi(ctx, site, b)
		},
	)
	return r
}
