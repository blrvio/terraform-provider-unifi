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

// dohCustomServerModel represents a single user-supplied DNS-over-HTTPS
// resolver entry.
type dohCustomServerModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	SdnsStamp  types.String `tfsdk:"sdns_stamp"`
	ServerName types.String `tfsdk:"server_name"`
}

func (m *dohCustomServerModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":     types.BoolType,
		"sdns_stamp":  types.StringType,
		"server_name": types.StringType,
	}
}

// dohModel represents the DNS-over-HTTPS (DoH) settings for a UniFi site.
type dohModel struct {
	base.Model
	State         types.String `tfsdk:"state"`
	ServerNames   types.List   `tfsdk:"server_names"`
	CustomServers types.List   `tfsdk:"custom_servers"`
}

func (d *dohModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingDoh{
		ID:    d.ID.ValueString(),
		State: d.State.ValueString(),
	}

	if ut.IsDefined(d.ServerNames) {
		var names []string
		diags.Append(ut.ListElementsAs(ctx, d.ServerNames, &names)...)
		if diags.HasError() {
			return nil, diags
		}
		model.ServerNames = names
	}

	if ut.IsDefined(d.CustomServers) {
		var items []dohCustomServerModel
		diags.Append(d.CustomServers.ElementsAs(ctx, &items, false)...)
		if diags.HasError() {
			return nil, diags
		}
		model.CustomServers = make([]unifi.SettingDohCustomServers, 0, len(items))
		for _, it := range items {
			model.CustomServers = append(model.CustomServers, unifi.SettingDohCustomServers{
				Enabled:    it.Enabled.ValueBool(),
				SdnsStamp:  it.SdnsStamp.ValueString(),
				ServerName: it.ServerName.ValueString(),
			})
		}
	}

	return model, diags
}

func (d *dohModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingDoh)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingDoh")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.State = ut.StringOrNull(model.State)

	serverNames, serverNamesDiags := types.ListValueFrom(ctx, types.StringType, model.ServerNames)
	diags.Append(serverNamesDiags...)
	if diags.HasError() {
		return diags
	}
	d.ServerNames = serverNames

	customItems := make([]dohCustomServerModel, 0, len(model.CustomServers))
	for _, it := range model.CustomServers {
		customItems = append(customItems, dohCustomServerModel{
			Enabled:    types.BoolValue(it.Enabled),
			SdnsStamp:  ut.StringOrNull(it.SdnsStamp),
			ServerName: ut.StringOrNull(it.ServerName),
		})
	}
	customList, customDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&dohCustomServerModel{}).AttributeTypes()}, customItems)
	diags.Append(customDiags...)
	if diags.HasError() {
		return diags
	}
	d.CustomServers = customList

	return diags
}

var (
	_ base.ResourceModel               = &dohModel{}
	_ resource.Resource                = &dohResource{}
	_ resource.ResourceWithConfigure   = &dohResource{}
	_ resource.ResourceWithImportState = &dohResource{}
)

type dohResource struct {
	*base.GenericResource[*dohModel]
}

func (r *dohResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_doh` resource manages DNS-over-HTTPS (DoH) for a UniFi site. When enabled, the " +
			"gateway encrypts outbound DNS by resolving through HTTPS resolvers instead of plaintext UDP/53. `state` chooses " +
			"between the controller's automatic resolver selection, a manually picked set of known providers " +
			"(`server_names`), or fully custom resolvers defined in `custom_servers`.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"state": schema.StringAttribute{
				MarkdownDescription: "DoH operating mode. `off` disables DoH, `auto` lets the controller select resolvers, " +
					"`manual` uses the providers listed in `server_names`, and `custom` uses the resolvers in `custom_servers`.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("off", "auto", "manual", "custom"),
				},
			},
			"server_names": schema.ListAttribute{
				MarkdownDescription: "Names of known DoH providers to use when `state` is `manual` (e.g. `cloudflare`, `google`).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"custom_servers": schema.ListNestedAttribute{
				MarkdownDescription: "User-defined DoH resolvers, used when `state` is `custom`.",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether this custom resolver is active.",
							Required:            true,
						},
						"sdns_stamp": schema.StringAttribute{
							MarkdownDescription: "DNS Stamp (`sdns://...`) describing the resolver's address, protocol, and public key.",
							Optional:            true,
							Computed:            true,
						},
						"server_name": schema.StringAttribute{
							MarkdownDescription: "Label for this custom resolver.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// NewDohResource creates a new instance of the DNS-over-HTTPS setting resource.
func NewDohResource() resource.Resource {
	r := &dohResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_doh",
		func() *dohModel { return &dohModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingDoh(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingDoh)
			return client.UpdateSettingDoh(ctx, site, b)
		},
	)
	return r
}
