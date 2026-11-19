package officialfw

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                 = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithConfigure    = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithImportState  = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithModifyPlan   = &officialFirewallPolicyOrderResource{}
	_ resource.ResourceWithUpgradeState = &officialFirewallPolicyOrderResource{}
	_ base.ResourceModel                = &officialFirewallPolicyOrderModel{}
	_ base.Resource                     = &officialFirewallPolicyOrderResource{}
)

type officialFirewallPolicyOrderModel struct {
	base.Model
	SourceZoneID      types.String `tfsdk:"source_zone_id"`
	DestinationZoneID types.String `tfsdk:"destination_zone_id"`
	BeforeSystemIDs   types.List   `tfsdk:"before_system_defined_policy_ids"`
	AfterSystemIDs    types.List   `tfsdk:"after_system_defined_policy_ids"`
}

// orderedIDs parses one of the two ordered lists. A null/unknown list becomes an
// empty, non-nil slice: the Official API rejects (or misreads) a null list.
func orderedIDs(ctx context.Context, l types.List) ([]uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	var in []string
	if ut.IsDefined(l) {
		diags.Append(ut.ListElementsAs(ctx, l, &in)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	ids, convDiags := stringsToUUIDs(in)
	diags.Append(convDiags...)
	return ids, diags
}

// AsUnifiModel builds the Official-API ordering body from both configured lists.
// Source/destination zones travel as query parameters, not in the body.
func (m *officialFirewallPolicyOrderModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	before, d := orderedIDs(ctx, m.BeforeSystemIDs)
	diags.Append(d...)
	after, d := orderedIDs(ctx, m.AfterSystemIDs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &official.FirewallPolicyOrdering{OrderedFirewallPolicyIds: official.OrderedFirewallPolicyIDs{
		BeforeSystemDefined: before,
		AfterSystemDefined:  after,
	}}, diags
}

// Merge populates both ordered lists from an Official-API ordering response.
func (m *officialFirewallPolicyOrderModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	ordering, ok := other.(*official.FirewallPolicyOrdering)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.FirewallPolicyOrdering, got %T", other))
		return diags
	}
	o := ordering.OrderedFirewallPolicyIds
	before, d := types.ListValueFrom(ctx, types.StringType, uuidsToStrings(o.BeforeSystemDefined))
	diags.Append(d...)
	after, d := types.ListValueFrom(ctx, types.StringType, uuidsToStrings(o.AfterSystemDefined))
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	m.BeforeSystemIDs = before
	m.AfterSystemIDs = after
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
		MarkdownDescription: "Manages the evaluation order of the user-defined firewall policies of a " +
			"source/destination zone pair via the UniFi **Official API** (`integration/v1`). Requires a controller " +
			"running version 10.1.78 or later with API-key authentication.\n\n" +
			"The controller keeps two ordered lists per zone pair: user-defined policies evaluated **before** the " +
			"system-defined (predefined) policies, and those evaluated **after** them. This resource owns both lists " +
			"of the pair: every user-defined policy of the pair must appear in exactly one of them, and a policy " +
			"created outside Terraform shows up as drift. Destroying the resource only removes it from state; the " +
			"order on the controller is left as is.",
		Version: 1,
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
			"before_system_defined_policy_ids": schema.ListAttribute{
				MarkdownDescription: "Ordered UUIDs of the user-defined policies evaluated **before** the system-defined policies " +
					"(the UniFi UI default for new policies). Defaults to an empty list.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(ut.EmptyList(types.StringType)),
			},
			"after_system_defined_policy_ids": schema.ListAttribute{
				MarkdownDescription: "Ordered UUIDs of the user-defined policies evaluated **after** the system-defined policies. " +
					"Defaults to an empty list. Replaces the former `policy_ids` attribute.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(ut.EmptyList(types.StringType)),
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
	// The PUT echo is not merged back: the planned lists are what was applied, and the
	// next Read reports any controller-side difference as drift instead of an apply error.
	if _, err := client.Official().Firewall().UpdatePolicyOrdering(ctx, siteID, sourceZone, destZone, *orderingBody); err != nil {
		diags.Append(base.OfficialAPIErrorDiagnostics("Error updating firewall policy ordering", err)...)
		return
	}
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

// UpgradeState moves the v0 `policy_ids` (which the provider always sent as the
// after-system-defined list) into `after_system_defined_policy_ids`.
func (r *officialFirewallPolicyOrderResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id":                  schema.StringAttribute{Computed: true},
					"site":                schema.StringAttribute{Optional: true, Computed: true},
					"source_zone_id":      schema.StringAttribute{Required: true},
					"destination_zone_id": schema.StringAttribute{Required: true},
					"policy_ids":          schema.ListAttribute{Required: true, ElementType: types.StringType},
				},
			},
			StateUpgrader: upgradePolicyOrderV0,
		},
	}
}

type policyOrderModelV0 struct {
	ID                types.String `tfsdk:"id"`
	Site              types.String `tfsdk:"site"`
	SourceZoneID      types.String `tfsdk:"source_zone_id"`
	DestinationZoneID types.String `tfsdk:"destination_zone_id"`
	PolicyIDs         types.List   `tfsdk:"policy_ids"`
}

func upgradePolicyOrderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior policyOrderModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}
	after := prior.PolicyIDs
	if !ut.IsDefined(after) {
		after = ut.EmptyList(types.StringType)
	}
	upgraded := officialFirewallPolicyOrderModel{
		SourceZoneID:      prior.SourceZoneID,
		DestinationZoneID: prior.DestinationZoneID,
		BeforeSystemIDs:   ut.EmptyList(types.StringType),
		AfterSystemIDs:    after,
	}
	upgraded.ID = prior.ID
	upgraded.Site = prior.Site
	resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
}
