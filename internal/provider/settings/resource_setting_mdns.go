package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
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

// mdnsCustomServiceModel represents a single user-defined mDNS service entry
// that the controller reflects across VLANs.
type mdnsCustomServiceModel struct {
	Address types.String `tfsdk:"address"`
	Name    types.String `tfsdk:"name"`
}

func (m *mdnsCustomServiceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
		"name":    types.StringType,
	}
}

// mdnsPredefinedServiceModel selects one of the controller's built-in mDNS
// service definitions by its code.
type mdnsPredefinedServiceModel struct {
	Code types.String `tfsdk:"code"`
}

func (m *mdnsPredefinedServiceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"code": types.StringType,
	}
}

// mdnsModel represents the multicast DNS (mDNS / Bonjour) reflector settings
// for a UniFi site.
type mdnsModel struct {
	base.Model
	Mode               types.String `tfsdk:"mode"`
	CustomServices     types.List   `tfsdk:"custom_services"`
	PredefinedServices types.List   `tfsdk:"predefined_services"`
}

func (d *mdnsModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingMdns{
		ID:   d.ID.ValueString(),
		Mode: d.Mode.ValueString(),
	}

	if ut.IsDefined(d.CustomServices) {
		var items []mdnsCustomServiceModel
		diags.Append(d.CustomServices.ElementsAs(ctx, &items, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.CustomServices = make([]unifi.SettingMdnsCustomServices, 0, len(items))
		for _, it := range items {
			model.CustomServices = append(model.CustomServices, unifi.SettingMdnsCustomServices{
				Address: it.Address.ValueString(),
				Name:    it.Name.ValueString(),
			})
		}
	}

	if ut.IsDefined(d.PredefinedServices) {
		var items []mdnsPredefinedServiceModel
		diags.Append(d.PredefinedServices.ElementsAs(ctx, &items, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.PredefinedServices = make([]unifi.SettingMdnsPredefinedServices, 0, len(items))
		for _, it := range items {
			model.PredefinedServices = append(model.PredefinedServices, unifi.SettingMdnsPredefinedServices{
				Code: it.Code.ValueString(),
			})
		}
	}

	return model, diags
}

func (d *mdnsModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingMdns)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingMdns")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Mode = ut.StringOrNull(model.Mode)

	customItems := make([]mdnsCustomServiceModel, 0, len(model.CustomServices))
	for _, it := range model.CustomServices {
		customItems = append(customItems, mdnsCustomServiceModel{
			Address: types.StringValue(it.Address),
			Name:    ut.StringOrNull(it.Name),
		})
	}
	customList, customDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&mdnsCustomServiceModel{}).AttributeTypes()}, customItems)
	diags.Append(customDiags...)
	if diags.HasError() {
		return diags
	}
	d.CustomServices = customList

	predefinedItems := make([]mdnsPredefinedServiceModel, 0, len(model.PredefinedServices))
	for _, it := range model.PredefinedServices {
		predefinedItems = append(predefinedItems, mdnsPredefinedServiceModel{
			Code: types.StringValue(it.Code),
		})
	}
	predefinedList, predefinedDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&mdnsPredefinedServiceModel{}).AttributeTypes()}, predefinedItems)
	diags.Append(predefinedDiags...)
	if diags.HasError() {
		return diags
	}
	d.PredefinedServices = predefinedList

	return diags
}

var (
	_ base.ResourceModel               = &mdnsModel{}
	_ resource.Resource                = &mdnsResource{}
	_ resource.ResourceWithConfigure   = &mdnsResource{}
	_ resource.ResourceWithImportState = &mdnsResource{}
)

type mdnsResource struct {
	*base.GenericResource[*mdnsModel]
}

func (r *mdnsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_mdns` resource manages the multicast DNS (mDNS / Bonjour) reflector for a " +
			"UniFi site. The reflector forwards service-discovery announcements (AirPlay, Chromecast, printers, etc.) " +
			"between VLANs so that clients on one network can find devices on another. `mode` selects which services are " +
			"reflected; `predefined_services` picks from the controller's built-in catalog, while `custom_services` adds " +
			"arbitrary DNS-SD service types.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"mode": schema.StringAttribute{
				MarkdownDescription: "Which mDNS services are reflected across networks. `all` reflects every discovered " +
					"service, `auto` lets the controller decide, and `custom` reflects only the services listed in " +
					"`predefined_services` and `custom_services`.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "auto", "custom"),
				},
			},
			"custom_services": schema.ListNestedAttribute{
				MarkdownDescription: "User-defined DNS-SD service types to reflect when `mode` is `custom`.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "DNS-SD service type to reflect, e.g. `_airplay._tcp.local` or `_ipp._tcp`.",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Human-friendly label for this custom service.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"predefined_services": schema.ListNestedAttribute{
				MarkdownDescription: "Built-in service definitions to reflect when `mode` is `custom`, selected by code.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							MarkdownDescription: "Identifier of a predefined mDNS service from the controller catalog.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"amazon_devices", "android_tv_remote", "apple_airDrop", "apple_airPlay",
									"apple_file_sharing", "apple_iChat", "apple_iTunes", "aqara", "bose",
									"dns_service_discovery", "ftp_servers", "google_chromecast", "homeKit",
									"matter_network", "philips_hue", "printers", "roku", "scanners", "sonos",
									"spotify_connect", "ssh_servers", "time_capsule", "web_servers",
									"windows_file_sharing_samba",
								),
							},
						},
					},
				},
			},
		},
	}
}

// NewMdnsResource creates a new instance of the mDNS setting resource.
func NewMdnsResource() resource.Resource {
	r := &mdnsResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_mdns",
		func() *mdnsModel { return &mdnsModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingMdns(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingMdns)
			return client.UpdateSettingMdns(ctx, site, b)
		},
	)
	return r
}
