package broadcastgroup

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

// BroadcastGroupModel represents the data model for a UniFi broadcast group.
type BroadcastGroupModel struct {
	base.Model
	Name        types.String `tfsdk:"name"`
	MemberTable types.Set    `tfsdk:"member_table"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *BroadcastGroupModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var members []string
	if ut.IsDefined(m.MemberTable) {
		diags.Append(m.MemberTable.ElementsAs(ctx, &members, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	return &unifi.BroadcastGroup{
		ID:          m.ID.ValueString(),
		Name:        m.Name.ValueString(),
		MemberTable: members,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *BroadcastGroupModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.BroadcastGroup)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.BroadcastGroup, got %T", other))
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
	_ resource.Resource                = &broadcastGroupResource{}
	_ resource.ResourceWithConfigure   = &broadcastGroupResource{}
	_ resource.ResourceWithImportState = &broadcastGroupResource{}
	_ base.Resource                    = &broadcastGroupResource{}
	_ base.ResourceModel               = &BroadcastGroupModel{}
)

type broadcastGroupResource struct {
	*base.GenericResource[*BroadcastGroupModel]
}

// NewBroadcastGroupResource creates a new instance of the broadcast group resource.
func NewBroadcastGroupResource() resource.Resource {
	return &broadcastGroupResource{
		GenericResource: base.NewGenericResource(
			"unifi_broadcast_group",
			func() *BroadcastGroupModel { return &BroadcastGroupModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetBroadcastGroup(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					group, _ := model.(*unifi.BroadcastGroup)
					return client.CreateBroadcastGroup(ctx, site, group)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					group, _ := model.(*unifi.BroadcastGroup)
					return client.UpdateBroadcastGroup(ctx, site, group)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteBroadcastGroup(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *broadcastGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_broadcast_group` resource manages a broadcast group in the UniFi " +
			"controller.\n\n" +
			"A broadcast group is the internal-SDK grouping of members (typically access points) used by " +
			"WLANs to scope broadcast/multicast traffic. This is distinct from `unifi_wifi_broadcast`, " +
			"which manages IoT/Official-API WiFi broadcast announcements.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the broadcast group.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"member_table": schema.SetAttribute{
				MarkdownDescription: "The set of member identifiers (typically access point object IDs or MAC " +
					"addresses) that belong to this broadcast group.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}
