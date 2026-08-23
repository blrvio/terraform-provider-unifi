package officialfw

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
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
	_ resource.Resource                = &officialFirewallPolicyResource{}
	_ resource.ResourceWithConfigure   = &officialFirewallPolicyResource{}
	_ resource.ResourceWithImportState = &officialFirewallPolicyResource{}
	_ resource.ResourceWithModifyPlan  = &officialFirewallPolicyResource{}
	_ base.ResourceModel               = &officialFirewallPolicyModel{}
	_ base.Resource                    = &officialFirewallPolicyResource{}
)

type officialFirewallPolicyModel struct {
	base.Model
	Name                  types.String `tfsdk:"name"`
	Action                types.String `tfsdk:"action"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	LoggingEnabled        types.Bool   `tfsdk:"logging_enabled"`
	Description           types.String `tfsdk:"description"`
	IpsecFilter           types.String `tfsdk:"ipsec_filter"`
	ConnectionStateFilter types.List   `tfsdk:"connection_state_filter"`
	Index                 types.Int32  `tfsdk:"index"`
	Source                types.String `tfsdk:"source"`
	Destination           types.String `tfsdk:"destination"`
	IPProtocolScope       types.String `tfsdk:"ip_protocol_scope"`
	Schedule              types.String `tfsdk:"schedule"`
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
	body := &official.FirewallPolicyCreateOrUpdate{
		Name:           m.Name.ValueString(),
		Action:         official.FirewallPolicyAction{Type: m.Action.ValueString()},
		Enabled:        m.enabledValue(),
		LoggingEnabled: m.loggingEnabledValue(),
		Description:    strPtr(m.Description),
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

	if err := json.Unmarshal([]byte(m.Source.ValueString()), &body.Source); err != nil {
		diags.AddError("Invalid source", fmt.Sprintf("`source` is not valid JSON for a firewall policy source: %s", err))
		return nil, diags
	}
	if err := json.Unmarshal([]byte(m.Destination.ValueString()), &body.Destination); err != nil {
		diags.AddError("Invalid destination", fmt.Sprintf("`destination` is not valid JSON for a firewall policy destination: %s", err))
		return nil, diags
	}
	if err := json.Unmarshal([]byte(m.IPProtocolScope.ValueString()), &body.IpProtocolScope); err != nil {
		diags.AddError("Invalid ip_protocol_scope", fmt.Sprintf("`ip_protocol_scope` is not valid JSON for a firewall policy IP protocol scope: %s", err))
		return nil, diags
	}
	if !m.Schedule.IsNull() && !m.Schedule.IsUnknown() {
		var schedule official.FirewallSchedule
		if err := json.Unmarshal([]byte(m.Schedule.ValueString()), &schedule); err != nil {
			diags.AddError("Invalid schedule", fmt.Sprintf("`schedule` is not valid JSON for a firewall schedule: %s", err))
			return nil, diags
		}
		body.Schedule = &schedule
	}

	return body, diags
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
	m.Enabled = types.BoolValue(policy.Enabled)
	m.LoggingEnabled = types.BoolValue(policy.LoggingEnabled)
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

	source, srcDiags := marshalJSONString("source", policy.Source)
	diags.Append(srcDiags...)
	m.Source = source

	destination, dstDiags := marshalJSONString("destination", policy.Destination)
	diags.Append(dstDiags...)
	m.Destination = destination

	scope, scopeDiags := marshalJSONString("ip_protocol_scope", policy.IpProtocolScope)
	diags.Append(scopeDiags...)
	m.IPProtocolScope = scope

	if policy.Schedule != nil {
		schedule, schedDiags := marshalJSONString("schedule", policy.Schedule)
		diags.Append(schedDiags...)
		m.Schedule = schedule
	} else {
		m.Schedule = types.StringNull()
	}

	return diags
}

// marshalJSONString marshals an Official-API sub-object into a compact JSON
// string value, reporting failures as a diagnostic tied to the attribute name.
func marshalJSONString(attr string, v interface{}) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	raw, err := json.Marshal(v)
	if err != nil {
		diags.AddError("Failed to encode "+attr, err.Error())
		return types.StringNull(), diags
	}
	return types.StringValue(string(raw)), diags
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
				MarkdownDescription: "JSON-encoded firewall policy source object (must include at least `zoneId`).",
				Required:            true,
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded firewall policy destination object (must include at least `zoneId`).",
				Required:            true,
			},
			"ip_protocol_scope": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded IP protocol scope object (must include at least `ipVersion`).",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				MarkdownDescription: "JSON-encoded firewall schedule object. If unset, the policy is always active.",
				Optional:            true,
			},
		},
	}
}

func (r *officialFirewallPolicyResource) ModifyPlan(_ context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
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
	created, err := client.Official().Firewall().CreatePolicy(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating firewall policy", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
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
	updated, err := client.Official().Firewall().UpdatePolicy(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating firewall policy", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
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
