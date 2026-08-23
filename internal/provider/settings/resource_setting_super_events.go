package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// superEventsModel represents the Super Events console setting for a UniFi
// site. The controller exposes only an opaque `_ignored` field for this
// setting.
type superEventsModel struct {
	base.Model
	Ignored types.String `tfsdk:"ignored"`
}

func (d *superEventsModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingSuperEvents{
		ID: d.ID.ValueString(),
	}
	if !ut.IsEmptyString(d.Ignored) {
		model.Ignored = d.Ignored.ValueString()
	}

	return model, diags
}

func (d *superEventsModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingSuperEvents)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingSuperEvents")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Ignored = ut.StringOrNull(model.Ignored)

	return diags
}

var (
	_ base.ResourceModel               = &superEventsModel{}
	_ resource.Resource                = &superEventsResource{}
	_ resource.ResourceWithConfigure   = &superEventsResource{}
	_ resource.ResourceWithImportState = &superEventsResource{}
)

type superEventsResource struct {
	*base.GenericResource[*superEventsModel]
}

func (r *superEventsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_super_events` resource manages the Super Events console setting for a UniFi site. " +
			"This console-level setting exposes only a single opaque `_ignored` field, surfaced here as the `ignored` attribute.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"ignored": schema.StringAttribute{
				MarkdownDescription: "Opaque value mapped to the controller's `_ignored` field for the Super Events setting.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// NewSuperEventsResource creates a new instance of the Super Events setting resource.
func NewSuperEventsResource() resource.Resource {
	r := &superEventsResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_super_events",
		func() *superEventsModel { return &superEventsModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingSuperEvents(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingSuperEvents)
			return client.UpdateSettingSuperEvents(ctx, site, b)
		},
	)
	return r
}
