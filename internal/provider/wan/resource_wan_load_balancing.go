package wan

import (
	"context"
	"fmt"
	"sort"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

const wanLoadBalancingID = "wan_load_balancing"

var wanModes = []string{"FAILOVER_ONLY", "DISTRIBUTED"}

var (
	_ resource.Resource                = &wanLoadBalancingResource{}
	_ resource.ResourceWithConfigure   = &wanLoadBalancingResource{}
	_ resource.ResourceWithImportState = &wanLoadBalancingResource{}
	_ base.ResourceModel               = &wanLoadBalancingModel{}
)

type wanInterfaceModel struct {
	Name     types.String `tfsdk:"name"`
	Mode     types.String `tfsdk:"mode"`
	Priority types.Int64  `tfsdk:"priority"`
	Weight   types.Int64  `tfsdk:"weight"`
}

func (wanInterfaceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.StringType,
		"mode":     types.StringType,
		"priority": types.Int64Type,
		"weight":   types.Int64Type,
	}
}

type wanLoadBalancingModel struct {
	base.Model
	Mode          types.String `tfsdk:"mode"`
	WANInterfaces types.Map    `tfsdk:"wan_interfaces"`
}

func (m *wanLoadBalancingModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	ifaces := map[string]wanInterfaceModel{}
	if ut.IsDefined(m.WANInterfaces) {
		diags.Append(m.WANInterfaces.ElementsAs(ctx, &ifaces, false)...)
	}
	if diags.HasError() {
		return nil, diags
	}
	out := &unifi.WANLoadBalancing{Mode: m.Mode.ValueString(), WANInterfaces: make([]unifi.WANLoadBalancingInterface, 0, len(ifaces))}
	for group, i := range ifaces {
		out.WANInterfaces = append(out.WANInterfaces, unifi.WANLoadBalancingInterface{
			WANNetworkgroup: group,
			Name:            i.Name.ValueString(),
			Mode:            i.Mode.ValueString(),
			Priority:        int(i.Priority.ValueInt64()),
			Weight:          int(i.Weight.ValueInt64()),
		})
	}
	sort.Slice(out.WANInterfaces, func(a, b int) bool {
		return out.WANInterfaces[a].Priority < out.WANInterfaces[b].Priority ||
			(out.WANInterfaces[a].Priority == out.WANInterfaces[b].Priority && out.WANInterfaces[a].WANNetworkgroup < out.WANInterfaces[b].WANNetworkgroup)
	})
	return out, diags
}

func (m *wanLoadBalancingModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	lb, ok := other.(*unifi.WANLoadBalancing)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.WANLoadBalancing, got %T", other))
		return diags
	}
	m.ID = types.StringValue(wanLoadBalancingID)
	m.Mode = types.StringValue(lb.Mode)
	ifaces := make(map[string]wanInterfaceModel, len(lb.WANInterfaces))
	for _, i := range lb.WANInterfaces {
		ifaces[i.WANNetworkgroup] = wanInterfaceModel{
			Name:     types.StringValue(i.Name),
			Mode:     types.StringValue(i.Mode),
			Priority: types.Int64Value(int64(i.Priority)),
			Weight:   types.Int64Value(int64(i.Weight)),
		}
	}
	mv, d := types.MapValueFrom(ctx, types.ObjectType{AttrTypes: wanInterfaceModel{}.AttributeTypes()}, ifaces)
	diags.Append(d...)
	m.WANInterfaces = mv
	return diags
}

// overlayWANLoadBalancing applies the planned mode and per-WAN values onto the current
// configuration, keeping fields the schema does not model. Planned WANs unknown to the
// controller are an error: the WAN list itself comes from the WAN networks.
func overlayWANLoadBalancing(current, planned *unifi.WANLoadBalancing) (*unifi.WANLoadBalancing, error) {
	out := *current
	out.Mode = planned.Mode
	out.WANInterfaces = append([]unifi.WANLoadBalancingInterface(nil), current.WANInterfaces...)
	idx := map[string]int{}
	for i, w := range out.WANInterfaces {
		idx[w.WANNetworkgroup] = i
	}
	for _, p := range planned.WANInterfaces {
		i, ok := idx[p.WANNetworkgroup]
		if !ok {
			return nil, fmt.Errorf("WAN %q is not a WAN of this site; the WAN list follows the WAN networks", p.WANNetworkgroup)
		}
		out.WANInterfaces[i].Mode = p.Mode
		out.WANInterfaces[i].Priority = p.Priority
		out.WANInterfaces[i].Weight = p.Weight
	}
	return &out, nil
}

type wanLoadBalancingResource struct {
	*base.GenericResource[*wanLoadBalancingModel]
}

// NewWANLoadBalancingResource creates the unifi_wan_load_balancing singleton resource.
func NewWANLoadBalancingResource() resource.Resource {
	write := func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
		planned, _ := body.(*unifi.WANLoadBalancing)
		current, err := client.GetWANLoadBalancing(ctx, site)
		if err != nil {
			return nil, fmt.Errorf("reading the WAN load-balancing configuration before update: %w", err)
		}
		merged, err := overlayWANLoadBalancing(current, planned)
		if err != nil {
			return nil, err
		}
		return client.UpdateWANLoadBalancing(ctx, site, merged)
	}
	return &wanLoadBalancingResource{
		GenericResource: base.NewGenericResource(
			"unifi_wan_load_balancing",
			func() *wanLoadBalancingModel { return &wanLoadBalancingModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, _ string) (interface{}, error) {
					return client.GetWANLoadBalancing(ctx, site)
				},
				Create:                   write,
				Update:                   write,
				ReadAfterWrite:           true,
				ResolveUnknownsFromState: true,
			},
		),
	}
}

func (r *wanLoadBalancingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the site's WAN mode (Settings > Internet > WAN Mode): failover only, or load balancing " +
			"across the WANs, with each WAN's priority and weight. This is a singleton: it always exists, and destroying the " +
			"resource only removes it from state.\n\n" +
			"**Changing it can drop internet access** — apply over a connection that does not depend on the WAN being changed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `wan_load_balancing`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"site": ut.SiteAttribute(),
			"mode": schema.StringAttribute{
				MarkdownDescription: "The WAN mode: `FAILOVER_ONLY` (one WAN at a time, by priority) or `DISTRIBUTED` (load balancing).",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf(wanModes...)},
			},
			"wan_interfaces": schema.MapNestedAttribute{
				MarkdownDescription: "Per-WAN settings, keyed by the WAN network group (`WAN`, `WAN2`, ...). List every WAN of the site.",
				Required:            true,
				Validators:          []validator.Map{mapvalidator.SizeAtLeast(1)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The WAN name (read-only, from the WAN network).",
							Computed:            true,
							PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"mode": schema.StringAttribute{
							MarkdownDescription: "The WAN's role: `DISTRIBUTED` (takes part in load balancing / primary) or `FAILOVER_ONLY` (backup).",
							Required:            true,
							Validators:          []validator.String{stringvalidator.OneOf(wanModes...)},
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "Failover priority (1 = first).",
							Required:            true,
							Validators:          []validator.Int64{int64validator.AtLeast(1)},
						},
						"weight": schema.Int64Attribute{
							MarkdownDescription: "Load-balancing weight in percent.",
							Required:            true,
							Validators:          []validator.Int64{int64validator.Between(0, 100)},
						},
					},
				},
			},
		},
	}
}
