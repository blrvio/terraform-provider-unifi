package tag

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// TagModel represents the data model for a UniFi tag.
type TagModel struct {
	base.Model
	Name        types.String `tfsdk:"name"`
	MemberTable types.Set    `tfsdk:"member_table"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *TagModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var members []string
	if ut.IsDefined(m.MemberTable) {
		diags.Append(m.MemberTable.ElementsAs(ctx, &members, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &unifi.Tag{
		ID:          m.ID.ValueString(),
		Name:        m.Name.ValueString(),
		MemberTable: members,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *TagModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.Tag)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Tag, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)

	// Coalesce a nil slice to an empty (non-null) set so an empty member list
	// round-trips to the schema default and does not produce a perpetual diff.
	members := model.MemberTable
	if members == nil {
		members = []string{}
	}
	set, diags := types.SetValueFrom(ctx, types.StringType, members)
	m.MemberTable = set
	return diags
}

var (
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithConfigure   = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
	_ base.Resource                    = &tagResource{}
	_ base.ResourceModel               = &TagModel{}
)

type tagResource struct {
	*base.GenericResource[*TagModel]
}

// NewTagResource creates a new instance of the tag resource.
func NewTagResource() resource.Resource {
	return &tagResource{
		GenericResource: base.NewGenericResource(
			"unifi_tag",
			func() *TagModel { return &TagModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetTag(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					tag, _ := model.(*unifi.Tag)
					return client.CreateTag(ctx, site, tag)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					tag, _ := model.(*unifi.Tag)
					return client.UpdateTag(ctx, site, tag)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteTag(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_tag` resource manages a tag in the UniFi controller.\n\n" +
			"Tags are named groups of members (typically device or client identifiers) that can be " +
			"referenced elsewhere in the controller configuration. This is the internal-SDK, fully " +
			"manageable tag resource; it is distinct from the read-only `unifi_device_tag` data source " +
			"exposed via the Official API.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the tag.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"member_table": schema.SetAttribute{
				MarkdownDescription: "The set of member identifiers (for example device MAC addresses or " +
					"object IDs) that belong to this tag.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}
