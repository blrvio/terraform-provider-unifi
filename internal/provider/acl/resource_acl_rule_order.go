package acl

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                = &aclRuleOrderResource{}
	_ resource.ResourceWithConfigure   = &aclRuleOrderResource{}
	_ resource.ResourceWithImportState = &aclRuleOrderResource{}
	_ resource.ResourceWithModifyPlan  = &aclRuleOrderResource{}
	_ base.Resource                    = &aclRuleOrderResource{}
)

type aclRuleOrderModel struct {
	base.Model
	RuleIDs types.List `tfsdk:"rule_ids"`
}

// AsUnifiModel and Merge exist only to satisfy base.ResourceModel; this
// resource implements fully custom CRUD and never calls them.
func (m *aclRuleOrderModel) AsUnifiModel(context.Context) (interface{}, diag.Diagnostics) {
	return nil, nil
}

func (m *aclRuleOrderModel) Merge(context.Context, interface{}) diag.Diagnostics {
	return nil
}

// ruleIDsToUUIDs parses the desired-order rule id strings into UUIDs.
func ruleIDsToUUIDs(ctx context.Context, list types.List) ([]uuid.UUID, diag.Diagnostics) {
	var ids []string
	diags := list.ElementsAs(ctx, &ids, false)
	if diags.HasError() {
		return nil, diags
	}
	out := make([]uuid.UUID, 0, len(ids))
	for _, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil {
			diags.AddError("Invalid ACL rule ID", fmt.Sprintf("Could not parse %q as a UUID: %s", s, err))
			return nil, diags
		}
		out = append(out, id)
	}
	return out, diags
}

// uuidsToRuleIDs converts the controller's ordered UUIDs into a Framework list
// of strings.
func uuidsToRuleIDs(ctx context.Context, ids []uuid.UUID) (types.List, diag.Diagnostics) {
	elems := make([]string, 0, len(ids))
	for _, id := range ids {
		elems = append(elems, id.String())
	}
	return types.ListValueFrom(ctx, types.StringType, elems)
}

type aclRuleOrderResource struct {
	*base.GenericResource[*aclRuleOrderModel]
}

// NewACLRuleOrderResource creates the unifi_acl_rule_order resource, a
// singleton-style ordering resource for the site's ACL rules.
func NewACLRuleOrderResource() resource.Resource {
	r := &aclRuleOrderResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_acl_rule_order",
		func() *aclRuleOrderModel { return &aclRuleOrderModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *aclRuleOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the ordering of ACL rules for a site via the UniFi **Official API** " +
			"(`integration/v1`). Requires a controller running version 10.1.78 or later with API-key authentication. " +
			"This is a singleton resource: exactly one per site. Deleting it removes it from state but does not " +
			"change the controller's ordering.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"rule_ids": schema.ListAttribute{
				MarkdownDescription: "The ACL rule UUIDs, in the desired order (first has the highest priority).",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *aclRuleOrderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
}

// apply pushes the desired ordering to the controller and writes state.
func (r *aclRuleOrderResource) apply(ctx context.Context, plan aclRuleOrderModel, setState func(context.Context, interface{}) diag.Diagnostics) diag.Diagnostics {
	client := r.GetClient()
	var diags diag.Diagnostics
	diags.Append(base.CheckConfigured(client)...)
	if diags.HasError() {
		return diags
	}
	siteName := client.ResolveSite(&plan)
	siteID, d := client.ResolveSiteUUID(ctx, siteName)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	ids, d := ruleIDsToUUIDs(ctx, plan.RuleIDs)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	updated, err := client.Official().ACLs().UpdateRuleOrdering(ctx, siteID, official.ACLRuleOrdering{OrderedAclRuleIds: ids})
	if err != nil {
		diags.Append(base.OfficialAPIErrorDiagnostics("Error updating ACL rule ordering", err)...)
		return diags
	}
	ruleIDs, d := uuidsToRuleIDs(ctx, updated.OrderedAclRuleIds)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	plan.RuleIDs = ruleIDs
	plan.SetSite(siteName)
	plan.SetID(orderID(siteName))
	diags.Append(setState(ctx, &plan)...)
	return diags
}

func (r *aclRuleOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aclRuleOrderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, plan, resp.State.Set)...)
}

func (r *aclRuleOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state aclRuleOrderModel
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
	ordering, err := client.Official().ACLs().GetRuleOrdering(ctx, siteID)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading ACL rule ordering", err)...)
		return
	}
	ruleIDs, d := uuidsToRuleIDs(ctx, ordering.OrderedAclRuleIds)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.RuleIDs = ruleIDs
	state.SetSite(siteName)
	state.SetID(orderID(siteName))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aclRuleOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan aclRuleOrderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.apply(ctx, plan, resp.State.Set)...)
}

// Delete is a no-op: the controller has no concept of deleting an ordering, so
// the resource is simply removed from state.
func (r *aclRuleOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *aclRuleOrderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	site := req.ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), orderID(site))...)
}

// orderID derives the singleton resource id from the site name.
func orderID(site string) string {
	if site == "" {
		return "ordering"
	}
	return site
}
