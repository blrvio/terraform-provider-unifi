package officialfw

import (
	"context"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// Firewall policy action discriminators, matching the go-unifi Official API union.
const (
	firewallPolicyActionAllow  = "ALLOW"
	firewallPolicyActionBlock  = "BLOCK"
	firewallPolicyActionReject = "REJECT"
)

// IPsec filter values, matching the FirewallPolicyIpsecFilter enum.
const (
	firewallPolicyIpsecMatchEncrypted    = "MATCH_ENCRYPTED"
	firewallPolicyIpsecMatchNotEncrypted = "MATCH_NOT_ENCRYPTED"
)

// Connection-state filter values, matching the FirewallPolicyConnectionStateFilter enum.
const (
	firewallPolicyConnStateNew         = "NEW"
	firewallPolicyConnStateInvalid     = "INVALID"
	firewallPolicyConnStateEstablished = "ESTABLISHED"
	firewallPolicyConnStateRelated     = "RELATED"
)

var (
	_ resource.Resource                     = &officialFirewallPolicyResource{}
	_ resource.ResourceWithConfigure        = &officialFirewallPolicyResource{}
	_ resource.ResourceWithImportState      = &officialFirewallPolicyResource{}
	_ resource.ResourceWithModifyPlan       = &officialFirewallPolicyResource{}
	_ resource.ResourceWithConfigValidators = &officialFirewallPolicyResource{}
	_ base.ResourceModel                    = &officialFirewallPolicyModel{}
	_ base.Resource                         = &officialFirewallPolicyResource{}
)

type officialFirewallPolicyModel struct {
	base.Model
	Name                  types.String `tfsdk:"name"`
	Action                types.String `tfsdk:"action"`
	AllowReturnTraffic    types.Bool   `tfsdk:"allow_return_traffic"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	LoggingEnabled        types.Bool   `tfsdk:"logging_enabled"`
	ReportFlowEnabled     types.Bool   `tfsdk:"report_flow_enabled"`
	Description           types.String `tfsdk:"description"`
	IpsecFilter           types.String `tfsdk:"ipsec_filter"`
	ConnectionStateFilter types.List   `tfsdk:"connection_state_filter"`
	Index                 types.Int32  `tfsdk:"index"`
	Source                ut.JSONValue `tfsdk:"source"`
	Destination           ut.JSONValue `tfsdk:"destination"`
	IPProtocolScope       ut.JSONValue `tfsdk:"ip_protocol_scope"`
	Schedule              ut.JSONValue `tfsdk:"schedule"`
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func strFromPtr(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

// AsUnifiModel builds the Official-API create/update body for a firewall policy.
// The complex source/destination/ipProtocolScope/schedule attributes are carried
// as JSON strings and decoded into their Official-API types here.
func (m *officialFirewallPolicyModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	action, err := m.actionBody()
	if err != nil {
		diags.AddError("Failed to encode firewall policy action", err.Error())
		return nil, diags
	}
	body := &official.FirewallPolicyCreateOrUpdate{
		Name:           m.Name.ValueString(),
		Action:         action,
		Enabled:        m.enabledValue(),
		LoggingEnabled: m.loggingEnabledValue(),
		Description:    strPtr(m.Description),
	}

	// The PUT is a full replace and the controller resets an omitted
	// reportFlowEnabled to true, so send it whenever the plan knows the value
	// (configured, or carried over from state by UseStateForUnknown). Only an
	// unknown value (create without the attribute) is omitted, letting the
	// controller apply its default, which is then read back.
	if ut.IsDefined(m.ReportFlowEnabled) {
		v := m.ReportFlowEnabled.ValueBool()
		body.ReportFlowEnabled = &v
	}

	if v := strPtr(m.IpsecFilter); v != nil {
		f := official.FirewallPolicyIpsecFilter(*v)
		body.IpsecFilter = &f
	}

	if csf, csfDiags := m.connectionStateFilter(ctx); csfDiags.HasError() {
		diags.Append(csfDiags...)
		return nil, diags
	} else if csf != nil {
		body.ConnectionStateFilter = csf
	}

	for _, f := range []struct {
		name   string
		value  ut.JSONValue
		target any
	}{
		{"source", m.Source, &body.Source},
		{"destination", m.Destination, &body.Destination},
		{"ip_protocol_scope", m.IPProtocolScope, &body.IpProtocolScope},
	} {
		if _, d := ut.UnmarshalJSONValue(f.name, f.value, f.target); d.HasError() {
			diags.Append(d...)
			return nil, diags
		}
	}
	var schedule official.FirewallSchedule
	defined, d := ut.UnmarshalJSONValue("schedule", m.Schedule, &schedule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	if defined {
		body.Schedule = &schedule
	}

	return body, diags
}

// actionBody builds the action union. The controller rejects an ALLOW without allowReturnTraffic, so an
// unknown value (create without the attribute) is sent as false.
func (m *officialFirewallPolicyModel) actionBody() (official.FirewallPolicyAction, error) {
	action := official.FirewallPolicyAction{Type: m.Action.ValueString()}
	if action.Type != firewallPolicyActionAllow {
		return action, nil
	}
	allowReturn := false
	if ut.IsDefined(m.AllowReturnTraffic) {
		allowReturn = m.AllowReturnTraffic.ValueBool()
	}
	err := action.FromFirewallPolicyActionAllow(official.FirewallPolicyActionAllow{
		Type:               firewallPolicyActionAllow,
		AllowReturnTraffic: &allowReturn,
	})
	return action, err
}

func (m *officialFirewallPolicyModel) enabledValue() bool {
	if m.Enabled.IsNull() || m.Enabled.IsUnknown() {
		return true
	}
	return m.Enabled.ValueBool()
}

func (m *officialFirewallPolicyModel) loggingEnabledValue() bool {
	if m.LoggingEnabled.IsNull() || m.LoggingEnabled.IsUnknown() {
		return false
	}
	return m.LoggingEnabled.ValueBool()
}

// connectionStateFilter converts the list attribute to the pointer-to-slice the
// Official API expects, returning nil when the list is unset or empty.
func (m *officialFirewallPolicyModel) connectionStateFilter(ctx context.Context) (*[]official.FirewallPolicyConnectionStateFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !ut.IsDefined(m.ConnectionStateFilter) {
		return nil, diags
	}
	var states []string
	diags.Append(ut.ListElementsAs(ctx, m.ConnectionStateFilter, &states)...)
	if diags.HasError() || len(states) == 0 {
		return nil, diags
	}
	out := make([]official.FirewallPolicyConnectionStateFilter, 0, len(states))
	for _, s := range states {
		out = append(out, official.FirewallPolicyConnectionStateFilter(s))
	}
	return &out, diags
}

// asFirewallPolicyBody performs the checked type assertion from the AsUnifiModel
// interface{} return to the concrete Official-API body.
func asFirewallPolicyBody(body interface{}) (*official.FirewallPolicyCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.FirewallPolicyCreateOrUpdate)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.FirewallPolicyCreateOrUpdate, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API FirewallPolicy read response.
func (m *officialFirewallPolicyModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	policy, ok := other.(*official.FirewallPolicy)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.FirewallPolicy, got %T", other))
		return diags
	}

	m.SetID(policy.Id.String())
	m.Name = types.StringValue(policy.Name)
	m.Action = types.StringValue(policy.Action.Type)
	m.AllowReturnTraffic = types.BoolNull()
	if policy.Action.Type == firewallPolicyActionAllow {
		allow, err := policy.Action.AsFirewallPolicyActionAllow()
		if err != nil {
			diags.AddError("Failed to decode firewall policy action", err.Error())
			return diags
		}
		if allow.AllowReturnTraffic != nil {
			m.AllowReturnTraffic = types.BoolValue(*allow.AllowReturnTraffic)
		}
	}
	m.Enabled = types.BoolValue(policy.Enabled)
	m.LoggingEnabled = types.BoolValue(policy.LoggingEnabled)
	m.ReportFlowEnabled = types.BoolValue(policy.ReportFlowEnabled)
	m.Index = types.Int32Value(policy.Index)
	m.Description = strFromPtr(policy.Description)

	if policy.IpsecFilter != nil {
		m.IpsecFilter = types.StringValue(string(*policy.IpsecFilter))
	} else {
		m.IpsecFilter = types.StringNull()
	}

	if policy.ConnectionStateFilter != nil {
		states := make([]string, 0, len(*policy.ConnectionStateFilter))
		for _, s := range *policy.ConnectionStateFilter {
			states = append(states, string(s))
		}
		list, listDiags := types.ListValueFrom(ctx, types.StringType, states)
		diags.Append(listDiags...)
		if diags.HasError() {
			return diags
		}
		m.ConnectionStateFilter = list
	} else {
		m.ConnectionStateFilter = types.ListNull(types.StringType)
	}

	source, srcDiags := ut.MarshalJSONValue("source", policy.Source)
	diags.Append(srcDiags...)
	m.Source = source

	destination, dstDiags := ut.MarshalJSONValue("destination", policy.Destination)
	diags.Append(dstDiags...)
	m.Destination = destination

	scope, scopeDiags := ut.MarshalJSONValue("ip_protocol_scope", policy.IpProtocolScope)
	diags.Append(scopeDiags...)
	m.IPProtocolScope = scope

	schedule, schedDiags := ut.MarshalJSONValue("schedule", policy.Schedule)
	diags.Append(schedDiags...)
	m.Schedule = schedule

	return diags
}

type officialFirewallPolicyResource struct {
	*base.GenericResource[*officialFirewallPolicyModel]
}

// NewOfficialFirewallPolicyResource creates the unifi_official_firewall_policy
// resource. CRUD is custom because the Official API is keyed by site/policy UUID.
func NewOfficialFirewallPolicyResource() resource.Resource {
	r := &officialFirewallPolicyResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_official_firewall_policy",
		func() *officialFirewallPolicyModel { return &officialFirewallPolicyModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *officialFirewallPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a firewall policy (zone-based firewall rule) via the UniFi **Official API** " +
			"(`integration/v1`). Requires a controller running version 10.1.78 or later with API-key authentication. " +
			"The `source`, `destination`, `ip_protocol_scope` and `schedule` attributes are JSON-encoded strings " +
			"matching the Official-API object shapes.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the firewall policy.",
				Required:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Action applied to matched traffic. One of `ALLOW`, `BLOCK`, `REJECT`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(firewallPolicyActionAllow, firewallPolicyActionBlock, firewallPolicyActionReject),
				},
			},
			"allow_return_traffic": schema.BoolAttribute{
				MarkdownDescription: "Only for `action = \"ALLOW\"`: have the controller create a derived policy for the " +
					"mirrored zone pair that allows the return traffic (shown as \"<name> (Return)\"). The controller " +
					"requires this on every ALLOW policy. When omitted, `false` is sent on create and the value read from " +
					"the controller is kept afterwards, so an existing policy is not changed. Must be unset for `BLOCK` " +
					"and `REJECT`, where it is always null.",
				Optional: true,
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the firewall policy is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"logging_enabled": schema.BoolAttribute{
				MarkdownDescription: "Generate syslog entries when traffic is matched. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"report_flow_enabled": schema.BoolAttribute{
				MarkdownDescription: "Report traffic matched by this policy in Insights Flows. Requires a controller " +
					"running 11.0.99 or later; on older controllers it cannot be set and is always null. When omitted " +
					"on create the controller default (`true`) is used and read back; from then on the value in state " +
					"is sent on every update, so a policy with reporting turned off is not switched back on. Removing " +
					"the attribute from the configuration keeps the last known value.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Optional description of the firewall policy.",
				Optional:            true,
			},
			"ipsec_filter": schema.StringAttribute{
				MarkdownDescription: "Match on traffic encrypted, or not encrypted by IPsec. One of " +
					"`MATCH_ENCRYPTED`, `MATCH_NOT_ENCRYPTED`. If unset, matches all traffic.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.OneOf(firewallPolicyIpsecMatchEncrypted, firewallPolicyIpsecMatchNotEncrypted),
				},
			},
			"connection_state_filter": schema.ListAttribute{
				MarkdownDescription: "Match on firewall connection state. Any of `NEW`, `INVALID`, `ESTABLISHED`, " +
					"`RELATED`. If unset, matches all connection states.",
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(
						firewallPolicyConnStateNew, firewallPolicyConnStateInvalid,
						firewallPolicyConnStateEstablished, firewallPolicyConnStateRelated,
					)),
				},
			},
			"index": schema.Int32Attribute{
				MarkdownDescription: "The evaluation index of the policy, assigned by the controller.",
				Computed:            true,
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded firewall policy source object (must include at least `zoneId`). Compared semantically: key order, whitespace and null-valued keys do not cause a diff, so `jsonencode()` works.",
				Required:            true,
				CustomType:          ut.JSONType{},
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded firewall policy destination object (must include at least `zoneId`). Compared semantically: key order, whitespace and null-valued keys do not cause a diff, so `jsonencode()` works.",
				Required:            true,
				CustomType:          ut.JSONType{},
			},
			"ip_protocol_scope": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded IP protocol scope object (must include at least `ipVersion`). Compared semantically: key order, whitespace and null-valued keys do not cause a diff, so `jsonencode()` works.",
				Required:            true,
				CustomType:          ut.JSONType{},
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded firewall schedule object. If unset, the policy is always active. Compared semantically: key order, whitespace and null-valued keys do not cause a diff, so `jsonencode()` works.",
				Optional:            true,
				CustomType:          ut.JSONType{},
			},
		},
	}
}

func (r *officialFirewallPolicyResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("action"), types.StringValue(firewallPolicyActionBlock), path.MatchRoot("allow_return_traffic")),
		validators.RequiredNoneIf(path.MatchRoot("action"), types.StringValue(firewallPolicyActionReject), path.MatchRoot("allow_return_traffic")),
	}
}

func (r *officialFirewallPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
	resp.Diagnostics.Append(r.RequireMinVersionForPath(
		base.ControllerVersionFirewallPolicyReportFlow.String(), path.Root("report_flow_enabled"), req.Config,
	)...)

	var action, stateAction types.String
	var configured, stateValue types.Bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("action"), &action)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("allow_return_traffic"), &configured)...)
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("action"), &stateAction)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("allow_return_traffic"), &stateValue)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("allow_return_traffic"),
		planAllowReturnTraffic(action, configured, stateAction, stateValue))...)
}

// planAllowReturnTraffic plans allow_return_traffic: null off ALLOW, the configured value, else the value
// already on an ALLOW policy, else unknown (create sends false and reads it back).
func planAllowReturnTraffic(action types.String, configured types.Bool, stateAction types.String, stateValue types.Bool) types.Bool {
	switch {
	case action.IsUnknown():
		return types.BoolUnknown()
	case action.ValueString() != firewallPolicyActionAllow:
		return types.BoolNull()
	case !configured.IsNull():
		return configured
	case stateAction.ValueString() == firewallPolicyActionAllow && ut.IsDefined(stateValue):
		return stateValue
	default:
		return types.BoolUnknown()
	}
}

// reportFlowSupported reports whether the controller models reportFlowEnabled.
// Older controllers neither return nor accept it, so the provider never sends it
// to them and keeps report_flow_enabled null in state.
func (r *officialFirewallPolicyResource) reportFlowSupported() bool {
	return r.GetClient().SupportsFirewallPolicyReportFlow()
}

// dropUnsupportedReportFlow strips reportFlowEnabled from the request body when
// the controller predates it.
func (r *officialFirewallPolicyResource) dropUnsupportedReportFlow(body *official.FirewallPolicyCreateOrUpdate) {
	if !r.reportFlowSupported() {
		body.ReportFlowEnabled = nil
	}
}

// nullUnsupportedReportFlow clears report_flow_enabled after a merge when the
// controller predates it (its zero-valued false would be meaningless).
func (r *officialFirewallPolicyResource) nullUnsupportedReportFlow(m *officialFirewallPolicyModel) {
	if !r.reportFlowSupported() {
		m.ReportFlowEnabled = types.BoolNull()
	}
}

func (r *officialFirewallPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan officialFirewallPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&plan)
	siteID, diags := client.ResolveSiteUUID(ctx, siteName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	createBody, diags := asFirewallPolicyBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.dropUnsupportedReportFlow(createBody)
	created, err := client.Official().Firewall().CreatePolicy(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating firewall policy", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
	r.nullUnsupportedReportFlow(&plan)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *officialFirewallPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state officialFirewallPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policy, err := client.Official().Firewall().GetPolicy(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading firewall policy", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, policy)...)
	r.nullUnsupportedReportFlow(&state)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *officialFirewallPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state officialFirewallPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&plan)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := plan.AsUnifiModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updateBody, diags := asFirewallPolicyBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.dropUnsupportedReportFlow(updateBody)
	updated, err := client.Official().Firewall().UpdatePolicy(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating firewall policy", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	r.nullUnsupportedReportFlow(&plan)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *officialFirewallPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state officialFirewallPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	siteName := client.ResolveSite(&state)
	siteID, id, diags := client.ResolveSiteAndID(ctx, siteName, state.GetID())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := client.Official().Firewall().DeletePolicy(ctx, siteID, id); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting firewall policy", err)...)
	}
}

func (r *officialFirewallPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
