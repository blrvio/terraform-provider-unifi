package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// roamingAssistantModel represents the Roaming Assistant settings for a UniFi
// site, which nudges wireless clients to roam to a stronger access point once
// their signal drops below a configured RSSI threshold.
type roamingAssistantModel struct {
	base.Model
	Enabled types.Bool  `tfsdk:"enabled"`
	Rssi    types.Int32 `tfsdk:"rssi"`
}

func (d *roamingAssistantModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingRoamingAssistant{
		ID:      d.ID.ValueString(),
		Enabled: d.Enabled.ValueBool(),
	}

	// Only set optional fields if roaming assistant is enabled
	if d.Enabled.ValueBool() {
		if !d.Rssi.IsNull() {
			model.Rssi = int(d.Rssi.ValueInt32())
		}
	}

	return model, diags
}

func (d *roamingAssistantModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingRoamingAssistant)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingRoamingAssistant")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// Only set optional fields if roaming assistant is enabled
	if model.Enabled {
		d.Rssi = types.Int32Value(int32(model.Rssi))
	} else {
		d.Rssi = types.Int32Null()
	}

	return diags
}

var (
	_ base.ResourceModel                    = &roamingAssistantModel{}
	_ resource.Resource                     = &roamingAssistantResource{}
	_ resource.ResourceWithConfigure        = &roamingAssistantResource{}
	_ resource.ResourceWithImportState      = &roamingAssistantResource{}
	_ resource.ResourceWithConfigValidators = &roamingAssistantResource{}
)

type roamingAssistantResource struct {
	*base.GenericResource[*roamingAssistantModel]
}

func (r *roamingAssistantResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false), path.MatchRoot("rssi")),
	}
}

func (r *roamingAssistantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_roaming_assistant` resource manages the Roaming Assistant settings for a UniFi site. " +
			"When enabled, the controller encourages wireless clients to roam to a stronger access point once their signal " +
			"drops below the configured RSSI threshold.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the Roaming Assistant is enabled.",
				Required:            true,
			},
			"rssi": schema.Int32Attribute{
				MarkdownDescription: "The RSSI (signal strength) threshold, in dBm, below which clients are encouraged to roam. " +
					"Valid values are `-80` to `-60`. Only applicable when `enabled` is `true`.",
				Optional: true,
				Computed: true,
				Validators: []validator.Int32{
					int32validator.Between(-80, -60),
				},
			},
		},
	}
}

// NewRoamingAssistantResource creates a new instance of the Roaming Assistant setting resource.
func NewRoamingAssistantResource() resource.Resource {
	r := &roamingAssistantResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_roaming_assistant",
		func() *roamingAssistantModel { return &roamingAssistantModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingRoamingAssistant(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingRoamingAssistant)
			return client.UpdateSettingRoamingAssistant(ctx, site, b)
		},
	)
	return r
}
