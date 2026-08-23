package officialfw

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithConfigure   = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithImportState = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithModifyPlan  = &officialFirewallPolicyOrderResource{}
	_ base.ResourceModel               = &officialFirewallPolicyOrderModel{}
	_ base.Resource                    = &officialFirewallPolicyOrderResource{}
)

type officialFirewallPolicyOrderModel struct {
	base.Model
	SourceZoneID      types.String `tfsdk:"source_zone_id"`
	DestinationZoneID types.String `tfsdk:"destination_zone_id"`
	PolicyIDs         types.List   `tfsdk:"policy_ids"`
}

// orderingFromStrings converts an ordered list of policy-ID strings into the
// Official-API ordering body. The managed order is applied to the user-defined
// policies evaluated after the system-defined ones (AfterSystemDefined).
func orderingFromStrings(in []string) (official.OrderedFirewallPolicyIDs, diag.Diagnostics) {
	ids, diags := stringsToUUIDs(in)
	if diags.HasError() {
		return official.OrderedFirewallPolicyIDs{}, diags
	}
	return official.OrderedFirewallPolicyIDs{AfterSystemDefined: ids}, diags
}

// stringsFromOrdering renders the AfterSystemDefined policy IDs of an ordering
// as their canonical string form.
func stringsFromOrdering(o official.OrderedFirewallPolicyIDs) []string {
	return uuidsToStrings(o.AfterSystemDefined)
}

// AsUnifiModel builds the Official-API ordering body from the configured policy
// IDs. Source/destination zones travel as query parameters, not in the body.
func (m *officialFirewallPolicyOrderModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	var policyStrings []string
	diags.Append(ut.ListElementsAs(ctx, m.PolicyIDs, &policyStrings)...)
	if diags.HasError() {
		return nil, diags
	}
	ordered, convDiags := orderingFromStrings(policyStrings)
	diags.Append(convDiags...)
	if diags.HasError() {
		return nil, diags
	}
	return &official.FirewallPolicyOrdering{OrderedFirewallPolicyIds: ordered}, diags
}

// Merge populates the model's policy_ids from an Official-API ordering response.
func (m *officialFirewallPolicyOrderModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	ordering, ok := other.(*official.FirewallPolicyOrdering)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.FirewallPolicyOrdering, got %T", other))
		return diags
	}
	list, listDiags := types.ListValueFrom(ctx, types.StringType, stringsFromOrdering(ordering.OrderedFirewallPolicyIds))
	diags.Append(listDiags...)
	if diags.HasError() {
		return diags
	}
	m.PolicyIDs = list
	return diags
}

// asFirewallPolicyOrderingBody performs the checked type assertion from the
// AsUnifiModel interface{} return to the concrete Official-API body.
func asFirewallPolicyOrderingBody(body interface{}) (*official.FirewallPolicyOrdering, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.FirewallPolicyOrdering)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.FirewallPolicyOrdering, got %T", body))
	}
	return b, diags
}

type officialFirewallPolicyOrderResource struct {
	*base.GenericResource[*officialFirewallPolicyOrderModel]
}

// NewOfficialFirewallPolicyOrderResource creates the
// unifi_official_firewall_policy_order resource. It manages the evaluation order
// of firewall policies for a source/destination zone pair via the Official API.
func NewOfficialFirewallPolicyOrderResource() resource.Resource {
	r := &officialFirewallPolicyOrderResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_official_firewall_policy_order",
		func() *officialFirewallPolicyOrderModel { return &officialFirewallPolicyOrderModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *officialFirewallPolicyOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the evaluation order of firewall policies for a source/destination zone pair " +
			"via the UniFi **Official API** (`integration/v1`). Requires a controller running version 10.1.78 or later " +
			"with API-key authentication. The order applies to user-defined policies evaluated after the " +
			"system-defined policies.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID("The identifier of this ordering, in the form `<source_zone_id>:<destination_zone_id>`."),
			"site": ut.SiteAttribute(),
			"source_zone_id": schema.StringAttribute{
				MarkdownDescription: "ID of the firewall zone from which the matched traffic originates. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_zone_id": schema.StringAttribute{
				MarkdownDescription: "ID of the firewall zone to which the matched traffic is destined. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_ids": schema.ListAttribute{
				MarkdownDescription: "Ordered list of firewall policy UUIDs for the zone pair.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *officialFirewallPolicyOrderResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
}

// orderingID builds the synthetic resource ID for a zone pair.
func orderingID(sourceZone, destZone string) string {
	return sourceZone + ":" + destZone
}

func (r *officialFirewallPolicyOrderResource) upsert(ctx context.Context, plan *officialFirewallPolicyOrderModel, diags *diag.Diagnostics) {
	client := r.GetClient()
	siteName := client.ResolveSite(plan)
	siteID, siteDiags := client.ResolveSiteUUID(ctx, siteName)
	diags.Append(siteDiags...)
	if diags.HasError() {
		return
	}
	body, bodyDiags := plan.AsUnifiModel(ctx)
	diags.Append(bodyDiags...)
	if diags.HasError() {
		return
	}
	orderingBody, castDiags := asFirewallPolicyOrderingBody(body)
	diags.Append(castDiags...)
	if diags.HasError() {
		return
	}
	sourceZone := plan.SourceZoneID.ValueString()
	destZone := plan.DestinationZoneID.ValueString()
	updated, err := client.Official().Firewall().UpdatePolicyOrdering(ctx, siteID, sourceZone, destZone, *orderingBody)
	if err != nil {
		diags.Append(base.OfficialAPIErrorDiagnostics("Error updating firewall policy ordering", err)...)
		return
	}
	diags.Append(plan.Merge(ctx, updated)...)
	plan.SetID(orderingID(sourceZone, destZone))
	plan.SetSite(siteName)
}

func (r *officialFirewallPolicyOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan officialFirewallPolicyOrderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *officialFirewallPolicyOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state officialFirewallPolicyOrderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, diags := client.ResolveSiteUUID(ctx, siteName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	sourceZone := state.SourceZoneID.ValueString()
	destZone := state.DestinationZoneID.ValueString()
	ordering, err := client.Official().Firewall().GetPolicyOrdering(ctx, siteID, sourceZone, destZone)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading firewall policy ordering", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, ordering)...)
	state.SetID(orderingID(sourceZone, destZone))
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *officialFirewallPolicyOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan officialFirewallPolicyOrderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.upsert(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is a no-op: policy ordering is a controller-managed property of a zone
// pair, not a standalone entity, so there is nothing to remove.
func (r *officialFirewallPolicyOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *officialFirewallPolicyOrderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: site:sourceZoneId:destinationZoneId
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the form `site:sourceZoneId:destinationZoneId`, got %q.", req.ID),
		)
		return
	}
	site, sourceZone, destZone := parts[0], parts[1], parts[2]
	if _, err := uuid.Parse(sourceZone); err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Could not parse source zone %q as a UUID: %s", sourceZone, err))
		return
	}
	if _, err := uuid.Parse(destZone); err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Could not parse destination zone %q as a UUID: %s", destZone, err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), orderingID(sourceZone, destZone))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source_zone_id"), sourceZone)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("destination_zone_id"), destZone)...)
}
