package officialfw

import (
	"context"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                = &officialFirewallPredefinedPolicyResource{}
	_ resource.ResourceWithConfigure   = &officialFirewallPredefinedPolicyResource{}
	_ resource.ResourceWithImportState = &officialFirewallPredefinedPolicyResource{}
	_ resource.ResourceWithModifyPlan  = &officialFirewallPredefinedPolicyResource{}
	_ base.Resource                    = &officialFirewallPredefinedPolicyResource{}
	_ base.ResourceModel               = &officialFirewallPredefinedPolicyModel{}
)

type officialFirewallPredefinedPolicyModel struct {
	base.Model
	PolicyID          types.String `tfsdk:"policy_id"`
	Name              types.String `tfsdk:"name"`
	Origin            types.String `tfsdk:"origin"`
	LoggingEnabled    types.Bool   `tfsdk:"logging_enabled"`
	ReportFlowEnabled types.Bool   `tfsdk:"report_flow_enabled"`
}

func (m *officialFirewallPredefinedPolicyModel) merge(p *official.FirewallPolicy) {
	m.SetID(p.Id.String())
	m.PolicyID = types.StringValue(p.Id.String())
	m.Name = types.StringValue(p.Name)
	m.Origin = types.StringValue(p.Metadata.Origin)
	m.LoggingEnabled = types.BoolValue(p.LoggingEnabled)
	m.ReportFlowEnabled = types.BoolValue(p.ReportFlowEnabled)
}

// AsUnifiModel returns the PATCH body (base.ResourceModel).
func (m *officialFirewallPredefinedPolicyModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	b := m.patchBody()
	return &b, nil
}

// Merge populates the model from an Official-API firewall policy (base.ResourceModel).
func (m *officialFirewallPredefinedPolicyModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	p, ok := other.(*official.FirewallPolicy)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.FirewallPolicy, got %T", other))
		return diags
	}
	m.merge(p)
	return diags
}

// patchBody sends only the attributes the configuration sets.
func (m *officialFirewallPredefinedPolicyModel) patchBody() official.PatchFirewallPolicy {
	var body official.PatchFirewallPolicy
	if ut.IsDefined(m.LoggingEnabled) {
		v := m.LoggingEnabled.ValueBool()
		body.LoggingEnabled = &v
	}
	if ut.IsDefined(m.ReportFlowEnabled) {
		v := m.ReportFlowEnabled.ValueBool()
		body.ReportFlowEnabled = &v
	}
	return body
}

// officialFirewallPredefinedPolicyResource manages the two fields the controller lets
// a user change on a predefined (SYSTEM_DEFINED or DERIVED) firewall policy.
type officialFirewallPredefinedPolicyResource struct {
	*base.GenericResource[*officialFirewallPredefinedPolicyModel]
}

// NewOfficialFirewallPredefinedPolicyResource creates the
// unifi_official_firewall_predefined_policy resource.
func NewOfficialFirewallPredefinedPolicyResource() resource.Resource {
	return &officialFirewallPredefinedPolicyResource{
		GenericResource: base.NewGenericResource("unifi_official_firewall_predefined_policy",
			func() *officialFirewallPredefinedPolicyModel { return &officialFirewallPredefinedPolicyModel{} }, base.ResourceFunctions{}),
	}
}

func (r *officialFirewallPredefinedPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the logging settings of a **predefined** firewall policy (`SYSTEM_DEFINED`, or `DERIVED` " +
			"from another object) via the UniFi **Official API** (`integration/v1`) PATCH endpoint. Those are the only fields " +
			"the controller (and the UI's policy table) lets you change on a predefined policy; everything else is read-only. " +
			"Use `unifi_official_firewall_policy` for user-defined policies. Destroying the resource only removes it from state " +
			"and leaves the policy as it is. Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "The Official API UUID of the predefined firewall policy. Changing it forces a new resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The policy name (read-only).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"origin": schema.StringAttribute{
				MarkdownDescription: "`SYSTEM_DEFINED` or `DERIVED` (read-only).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"logging_enabled": schema.BoolAttribute{
				MarkdownDescription: "Send a syslog entry when the policy matches (UI: \"Syslog Logging\"). Left as is when unset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"report_flow_enabled": schema.BoolAttribute{
				MarkdownDescription: "Report matched traffic in Insights > Flows (UI: \"Flow Logging\"). The controller ignores it, " +
					"and always reports false, for the Allow Return Traffic, multicast and IPv6 neighbor discovery policies. " +
					"Left as is when unset. Requires UniFi Network 11.0.99 or later.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *officialFirewallPredefinedPolicyResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
	resp.Diagnostics.Append(r.RequireMinVersionForPath("11.0.99", path.Root("report_flow_enabled"), req.Config)...)
}

func (r *officialFirewallPredefinedPolicyResource) fetch(ctx context.Context, m *officialFirewallPredefinedPolicyModel) (*official.FirewallPolicy, diag.Diagnostics) {
	client := r.GetClient()
	siteName := client.ResolveSite(m)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, m.PolicyID.ValueString())
	if diags.HasError() {
		return nil, diags
	}
	m.SetSite(siteName)
	p, err := client.Official().Firewall().GetPolicy(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return nil, nil
		}
		diags.Append(base.OfficialAPIErrorDiagnostics("Error reading firewall policy", err)...)
	}
	return p, diags
}

func (r *officialFirewallPredefinedPolicyResource) patch(ctx context.Context, m *officialFirewallPredefinedPolicyModel) diag.Diagnostics {
	current, diags := r.fetch(ctx, m)
	if diags.HasError() {
		return diags
	}
	if current == nil {
		diags.AddAttributeError(path.Root("policy_id"), "Firewall policy not found", fmt.Sprintf("No firewall policy %q on site %q.", m.PolicyID.ValueString(), m.GetSite()))
		return diags
	}
	if current.Metadata.Origin == "USER_DEFINED" {
		diags.AddAttributeError(path.Root("policy_id"), "Not a predefined firewall policy",
			fmt.Sprintf("Policy %q is user-defined; manage it with unifi_official_firewall_policy.", current.Name))
		return diags
	}
	client := r.GetClient()
	siteID, id, d := client.ResolveSiteAndID(ctx, m.GetSite(), m.PolicyID.ValueString())
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	body := m.patchBody()
	if body.LoggingEnabled == nil && body.ReportFlowEnabled == nil {
		m.merge(current)
		return diags
	}
	updated, err := client.Official().Firewall().PatchPolicy(ctx, siteID, id, body)
	if err != nil {
		diags.Append(base.OfficialAPIErrorDiagnostics("Error updating firewall policy", err)...)
		return diags
	}
	m.merge(updated)
	return diags
}

func (r *officialFirewallPredefinedPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(base.CheckConfigured(r.GetClient())...)
	var plan officialFirewallPredefinedPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.patch(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *officialFirewallPredefinedPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(base.CheckConfigured(r.GetClient())...)
	var state officialFirewallPredefinedPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if state.PolicyID.IsNull() || state.PolicyID.ValueString() == "" {
		state.PolicyID = state.ID
	}
	p, diags := r.fetch(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if p == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.merge(p)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *officialFirewallPredefinedPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(base.CheckConfigured(r.GetClient())...)
	var plan officialFirewallPredefinedPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(r.patch(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete only removes the resource from state: a predefined policy cannot be deleted.
func (r *officialFirewallPredefinedPolicyResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *officialFirewallPredefinedPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithOptionalSite(req, resp, r.GetClient().Site)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
