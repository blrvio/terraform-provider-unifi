package hotspotop

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

// HotspotOperatorModel represents the data model for a UniFi hotspot operator.
type HotspotOperatorModel struct {
	base.Model
	Name      types.String `tfsdk:"name"`
	Note      types.String `tfsdk:"note"`
	XPassword types.String `tfsdk:"x_password"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HotspotOperatorModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return &unifi.HotspotOp{
		ID:        m.ID.ValueString(),
		Name:      m.Name.ValueString(),
		Note:      m.Note.ValueString(),
		XPassword: m.XPassword.ValueString(),
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HotspotOperatorModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.HotspotOp)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.HotspotOp, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Note = ut.StringOrNull(model.Note)
	m.XPassword = ut.StringOrNull(model.XPassword)
	return diags
}

var (
	_ resource.Resource                = &hotspotOperatorResource{}
	_ resource.ResourceWithConfigure   = &hotspotOperatorResource{}
	_ resource.ResourceWithImportState = &hotspotOperatorResource{}
	_ base.Resource                    = &hotspotOperatorResource{}
	_ base.ResourceModel               = &HotspotOperatorModel{}
)

type hotspotOperatorResource struct {
	*base.GenericResource[*HotspotOperatorModel]
}

// NewHotspotOperatorResource creates a new instance of the hotspot operator resource.
func NewHotspotOperatorResource() resource.Resource {
	return &hotspotOperatorResource{
		GenericResource: base.NewGenericResource(
			"unifi_hotspot_operator",
			func() *HotspotOperatorModel { return &HotspotOperatorModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetHotspotOp(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					hotspotOp, _ := model.(*unifi.HotspotOp)
					return client.CreateHotspotOp(ctx, site, hotspotOp)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					hotspotOp, _ := model.(*unifi.HotspotOp)
					return client.UpdateHotspotOp(ctx, site, hotspotOp)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteHotspotOp(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *hotspotOperatorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot_operator` resource manages a hotspot operator account in the " +
			"UniFi controller. Hotspot operators are the accounts used to authorise guests and manage " +
			"guest hotspot access.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the hotspot operator.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "A free-form note for the hotspot operator.",
				Optional:            true,
				Computed:            true,
			},
			"x_password": schema.StringAttribute{
				MarkdownDescription: "The password for the hotspot operator account.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
		},
	}
}
