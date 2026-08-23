package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// broadcastModel represents the Broadcast (media broadcast) settings for a
// UniFi site, controlling the sounds played before and after a broadcast.
type broadcastModel struct {
	base.Model
	SoundAfterEnabled   types.Bool   `tfsdk:"sound_after_enabled"`
	SoundAfterResource  types.String `tfsdk:"sound_after_resource"`
	SoundAfterType      types.String `tfsdk:"sound_after_type"`
	SoundBeforeEnabled  types.Bool   `tfsdk:"sound_before_enabled"`
	SoundBeforeResource types.String `tfsdk:"sound_before_resource"`
	SoundBeforeType     types.String `tfsdk:"sound_before_type"`
}

func (d *broadcastModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingBroadcast{
		ID:                 d.ID.ValueString(),
		SoundAfterEnabled:  d.SoundAfterEnabled.ValueBool(),
		SoundBeforeEnabled: d.SoundBeforeEnabled.ValueBool(),
	}
	if !ut.IsEmptyString(d.SoundAfterResource) {
		model.SoundAfterResource = d.SoundAfterResource.ValueString()
	}
	if !ut.IsEmptyString(d.SoundAfterType) {
		model.SoundAfterType = d.SoundAfterType.ValueString()
	}
	if !ut.IsEmptyString(d.SoundBeforeResource) {
		model.SoundBeforeResource = d.SoundBeforeResource.ValueString()
	}
	if !ut.IsEmptyString(d.SoundBeforeType) {
		model.SoundBeforeType = d.SoundBeforeType.ValueString()
	}

	return model, diags
}

func (d *broadcastModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingBroadcast)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingBroadcast")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.SoundAfterEnabled = types.BoolValue(model.SoundAfterEnabled)
	d.SoundBeforeEnabled = types.BoolValue(model.SoundBeforeEnabled)
	d.SoundAfterResource = ut.StringOrNull(model.SoundAfterResource)
	d.SoundAfterType = ut.StringOrNull(model.SoundAfterType)
	d.SoundBeforeResource = ut.StringOrNull(model.SoundBeforeResource)
	d.SoundBeforeType = ut.StringOrNull(model.SoundBeforeType)

	return diags
}

var (
	_ base.ResourceModel               = &broadcastModel{}
	_ resource.Resource                = &broadcastResource{}
	_ resource.ResourceWithConfigure   = &broadcastResource{}
	_ resource.ResourceWithImportState = &broadcastResource{}
)

type broadcastResource struct {
	*base.GenericResource[*broadcastModel]
}

func (r *broadcastResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_broadcast` resource manages the media Broadcast settings for a UniFi site, controlling the " +
			"sounds played before and after a broadcast.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"sound_after_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether a sound is played after a broadcast.",
				Optional:            true,
				Computed:            true,
			},
			"sound_after_resource": schema.StringAttribute{
				MarkdownDescription: "The resource identifier of the sound played after a broadcast.",
				Optional:            true,
				Computed:            true,
			},
			"sound_after_type": schema.StringAttribute{
				MarkdownDescription: "The type of the sound played after a broadcast. Valid values are `sample` and `media`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("sample", "media"),
				},
			},
			"sound_before_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether a sound is played before a broadcast.",
				Optional:            true,
				Computed:            true,
			},
			"sound_before_resource": schema.StringAttribute{
				MarkdownDescription: "The resource identifier of the sound played before a broadcast.",
				Optional:            true,
				Computed:            true,
			},
			"sound_before_type": schema.StringAttribute{
				MarkdownDescription: "The type of the sound played before a broadcast. Valid values are `sample` and `media`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("sample", "media"),
				},
			},
		},
	}
}

// NewBroadcastResource creates a new instance of the Broadcast setting resource.
func NewBroadcastResource() resource.Resource {
	r := &broadcastResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_broadcast",
		func() *broadcastModel { return &broadcastModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingBroadcast(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingBroadcast)
			return client.UpdateSettingBroadcast(ctx, site, b)
		},
	)
	return r
}
