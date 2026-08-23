package wlangroup

import (
	"context"
	"fmt"

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

// WLANGroupModel represents the data model for a UniFi WLAN group.
type WLANGroupModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *WLANGroupModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.WLANGroup{
		ID:   m.ID.ValueString(),
		Name: m.Name.ValueString(),
	}, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *WLANGroupModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.WLANGroup)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.WLANGroup, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	return diag.Diagnostics{}
}

var (
	_ resource.Resource                = &wlanGroupResource{}
	_ resource.ResourceWithConfigure   = &wlanGroupResource{}
	_ resource.ResourceWithImportState = &wlanGroupResource{}
	_ base.Resource                    = &wlanGroupResource{}
	_ base.ResourceModel               = &WLANGroupModel{}
)

type wlanGroupResource struct {
	*base.GenericResource[*WLANGroupModel]
}

// NewWLANGroupResource creates a new instance of the WLAN group resource.
func NewWLANGroupResource() resource.Resource {
	return &wlanGroupResource{
		GenericResource: base.NewGenericResource(
			"unifi_wlan_group",
			func() *WLANGroupModel { return &WLANGroupModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetWLANGroup(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					group, _ := model.(*unifi.WLANGroup)
					return client.CreateWLANGroup(ctx, site, group)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					group, _ := model.(*unifi.WLANGroup)
					return client.UpdateWLANGroup(ctx, site, group)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteWLANGroup(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *wlanGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_wlan_group` resource manages a WLAN group in the UniFi controller.\n\n" +
			"WLAN groups organize wireless networks so they can be assigned together to access points. " +
			"WLANs reference a WLAN group by id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the WLAN group (1-128 characters).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
		},
	}
}
