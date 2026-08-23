package dhcpoption

import (
	"context"
	"fmt"
	"regexp"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var dhcpOptionNameRegex = regexp.MustCompile(`^[A-Za-z0-9-_]{1,25}$`)

// DHCPOptionModel represents the data model for a UniFi custom DHCP option.
type DHCPOptionModel struct {
	base.Model
	Code   types.String `tfsdk:"code"`
	Name   types.String `tfsdk:"name"`
	Signed types.Bool   `tfsdk:"signed"`
	Type   types.String `tfsdk:"type"`
	Width  types.Int32  `tfsdk:"width"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *DHCPOptionModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.DHCPOption{
		ID:     m.ID.ValueString(),
		Code:   m.Code.ValueString(),
		Name:   m.Name.ValueString(),
		Signed: m.Signed.ValueBool(),
		Type:   m.Type.ValueString(),
		Width:  int(m.Width.ValueInt32()),
	}, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *DHCPOptionModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.DHCPOption)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.DHCPOption, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Code = types.StringValue(model.Code)
	m.Name = types.StringValue(model.Name)
	m.Signed = types.BoolValue(model.Signed)
	m.Type = types.StringValue(model.Type)
	m.Width = ut.Int32OrNull(model.Width)
	return diag.Diagnostics{}
}

var (
	_ resource.Resource                = &dhcpOptionResource{}
	_ resource.ResourceWithConfigure   = &dhcpOptionResource{}
	_ resource.ResourceWithImportState = &dhcpOptionResource{}
	_ base.Resource                    = &dhcpOptionResource{}
	_ base.ResourceModel               = &DHCPOptionModel{}
)

type dhcpOptionResource struct {
	*base.GenericResource[*DHCPOptionModel]
}

// NewDHCPOptionResource creates a new instance of the DHCP option resource.
func NewDHCPOptionResource() resource.Resource {
	return &dhcpOptionResource{
		GenericResource: base.NewGenericResource(
			"unifi_dhcp_option",
			func() *DHCPOptionModel { return &DHCPOptionModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetDHCPOption(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					opt, _ := model.(*unifi.DHCPOption)
					return client.CreateDHCPOption(ctx, site, opt)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					opt, _ := model.(*unifi.DHCPOption)
					return client.UpdateDHCPOption(ctx, site, opt)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteDHCPOption(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *dhcpOptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_dhcp_option` resource manages a custom DHCP option definition in the " +
			"UniFi controller.\n\n" +
			"A custom DHCP option registers an option code, its wire `type` and `width`, so it can then be " +
			"served to clients by a network's DHCP server.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"code": schema.StringAttribute{
				MarkdownDescription: "The DHCP option code (option number), as a string. Reserved codes " +
					"already handled natively by the controller (for example `15`, `43`, `66`, `67`, `252`) " +
					"are rejected by the controller.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the DHCP option. Must match `^[A-Za-z0-9-_]{1,25}$`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(dhcpOptionNameRegex,
						"must be 1-25 characters of letters, digits, `-` or `_`"),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The value type of the option. One of `boolean`, `hexarray`, `integer`, " +
					"`ipaddress`, `macaddress`, `text`.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("boolean", "hexarray", "integer", "ipaddress", "macaddress", "text"),
				},
			},
			"width": schema.Int32Attribute{
				MarkdownDescription: "The bit width of the value for numeric option types. One of `8`, `16`, `32`.",
				Optional:            true,
				Validators: []validator.Int32{
					int32validator.OneOf(8, 16, 32),
				},
			},
			"signed": schema.BoolAttribute{
				MarkdownDescription: "Whether the numeric value is interpreted as signed. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}
