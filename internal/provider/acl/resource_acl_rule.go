package acl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// ACL rule action and type discriminators, matching the go-unifi Official API.
const (
	aclRuleActionAllow = "ALLOW"
	aclRuleActionBlock = "BLOCK"

	aclRuleTypeIPv4 = "IPV4"
	aclRuleTypeMAC  = "MAC"
)

var (
	_ resource.Resource                = &aclRuleResource{}
	_ resource.ResourceWithConfigure   = &aclRuleResource{}
	_ resource.ResourceWithImportState = &aclRuleResource{}
	_ resource.ResourceWithModifyPlan  = &aclRuleResource{}
	_ base.ResourceModel               = &aclRuleModel{}
	_ base.Resource                    = &aclRuleResource{}
)

type aclRuleModel struct {
	base.Model
	Name                  types.String `tfsdk:"name"`
	Action                types.String `tfsdk:"action"`
	Type                  types.String `tfsdk:"type"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Description           types.String `tfsdk:"description"`
	SourceFilter          types.String `tfsdk:"source_filter"`
	DestinationFilter     types.String `tfsdk:"destination_filter"`
	EnforcingDeviceFilter types.String `tfsdk:"enforcing_device_filter"`
	Index                 types.Int32  `tfsdk:"index"`
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

func (m *aclRuleModel) enabledValue() bool {
	if m.Enabled.IsNull() || m.Enabled.IsUnknown() {
		return true
	}
	return m.Enabled.ValueBool()
}

// filterFromJSON decodes an Optional JSON-string attribute into a free-form
// value for the polymorphic source/destination filters. An unset or empty
// string yields a nil filter (attribute omitted from the request).
func filterFromJSON(v types.String) (interface{}, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	s := strings.TrimSpace(v.ValueString())
	if s == "" {
		return nil, nil
	}
	var out interface{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// deviceFilterFromJSON decodes the Optional enforcing_device_filter JSON string
// into the concrete Official-API union type.
func deviceFilterFromJSON(v types.String) (*official.ACLRuleDeviceFilter, error) {
	if v.IsNull() || v.IsUnknown() {
		return nil, nil
	}
	s := strings.TrimSpace(v.ValueString())
	if s == "" {
		return nil, nil
	}
	var out official.ACLRuleDeviceFilter
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// filterToJSON marshals a returned polymorphic filter back to a compact JSON
// string, or null when the controller returned no filter.
func filterToJSON(v interface{}) (types.String, error) {
	if v == nil {
		return types.StringNull(), nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return types.StringNull(), err
	}
	return types.StringValue(string(raw)), nil
}

// AsUnifiModel builds the Official-API create/update body. ACLRuleUpdate uses a
// union with a MarshalJSON that overlays every top-level field, so a plain
// struct literal (union left nil) round-trips correctly; Enabled in particular
// must be set explicitly or it would always serialize as false. Index is
// DEPRECATED and left nil — ordering is managed by unifi_acl_rule_order.
func (m *aclRuleModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := &official.ACLRuleUpdate{
		Action:      official.ACLRuleAction(m.Action.ValueString()),
		Name:        m.Name.ValueString(),
		Type:        m.Type.ValueString(),
		Enabled:     m.enabledValue(),
		Description: strPtr(m.Description),
	}

	src, err := filterFromJSON(m.SourceFilter)
	if err != nil {
		diags.AddError("Invalid source_filter", fmt.Sprintf("source_filter must be valid JSON: %s", err))
		return nil, diags
	}
	body.SourceFilter = src

	dst, err := filterFromJSON(m.DestinationFilter)
	if err != nil {
		diags.AddError("Invalid destination_filter", fmt.Sprintf("destination_filter must be valid JSON: %s", err))
		return nil, diags
	}
	body.DestinationFilter = dst

	dev, err := deviceFilterFromJSON(m.EnforcingDeviceFilter)
	if err != nil {
		diags.AddError("Invalid enforcing_device_filter", fmt.Sprintf("enforcing_device_filter must be valid JSON: %s", err))
		return nil, diags
	}
	body.EnforcingDeviceFilter = dev

	return body, diags
}

// asACLRuleBody performs the checked type assertion from the AsUnifiModel
// interface{} return to the concrete Official-API body.
func asACLRuleBody(body interface{}) (*official.ACLRuleUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.ACLRuleUpdate)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.ACLRuleUpdate, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API ACLRule read response.
func (m *aclRuleModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	rule, ok := other.(*official.ACLRule)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.ACLRule, got %T", other))
		return diags
	}

	m.SetID(rule.Id.String())
	m.Name = types.StringValue(rule.Name)
	m.Action = types.StringValue(string(rule.Action))
	m.Type = types.StringValue(rule.Type)
	m.Enabled = types.BoolValue(rule.Enabled)
	m.Description = strFromPtr(rule.Description)
	m.Index = types.Int32Value(rule.Index)

	src, err := filterToJSON(rule.SourceFilter)
	if err != nil {
		diags.AddError("Failed to encode source_filter", err.Error())
		return diags
	}
	m.SourceFilter = src

	dst, err := filterToJSON(rule.DestinationFilter)
	if err != nil {
		diags.AddError("Failed to encode destination_filter", err.Error())
		return diags
	}
	m.DestinationFilter = dst

	if rule.EnforcingDeviceFilter != nil {
		dev, err := filterToJSON(rule.EnforcingDeviceFilter)
		if err != nil {
			diags.AddError("Failed to encode enforcing_device_filter", err.Error())
			return diags
		}
		m.EnforcingDeviceFilter = dev
	} else {
		m.EnforcingDeviceFilter = types.StringNull()
	}

	return diags
}

type aclRuleResource struct {
	*base.GenericResource[*aclRuleModel]
}

// NewACLRuleResource creates the unifi_acl_rule resource. It embeds a
// GenericResource purely for Configure/version-validator wiring; CRUD is custom
// because the Official API is keyed by site/entity UUID.
func NewACLRuleResource() resource.Resource {
	r := &aclRuleResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_acl_rule",
		func() *aclRuleModel { return &aclRuleModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *aclRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an ACL (access control list) rule via the UniFi **Official API** " +
			"(`integration/v1`). Requires a controller running version 10.1.78 or later with API-key authentication. " +
			"Rule priority/ordering is not managed here — use the `unifi_acl_rule_order` resource.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The ACL rule name.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The ACL rule action. One of `ALLOW` or `BLOCK`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(aclRuleActionAllow, aclRuleActionBlock),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The ACL rule type, selecting the traffic-matching family. One of `IPV4` " +
					"(IP-based) or `MAC` (MAC-based).",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(aclRuleTypeIPv4, aclRuleTypeMAC),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the ACL rule is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "An optional description for the ACL rule.",
				Optional:            true,
			},
			"source_filter": schema.StringAttribute{
				MarkdownDescription: "Traffic source filter, encoded as a JSON string. The nested shape is " +
					"polymorphic and depends on `type`; supply it verbatim (e.g. with `jsonencode(...)`). Must be " +
					"valid JSON. The controller normalizes the value, so use a compact JSON encoding to avoid drift.",
				Optional: true,
			},
			"destination_filter": schema.StringAttribute{
				MarkdownDescription: "Traffic destination filter, encoded as a JSON string. The nested shape is " +
					"polymorphic and depends on `type`; supply it verbatim (e.g. with `jsonencode(...)`). Must be " +
					"valid JSON. The controller normalizes the value, so use a compact JSON encoding to avoid drift.",
				Optional: true,
			},
			"enforcing_device_filter": schema.StringAttribute{
				MarkdownDescription: "Filter selecting the devices that enforce this rule, encoded as a JSON " +
					"string (e.g. with `jsonencode(...)`). Must be valid JSON. The controller normalizes the value, " +
					"so use a compact JSON encoding to avoid drift.",
				Optional: true,
			},
			"index": schema.Int32Attribute{
				MarkdownDescription: "The rule's position in the ACL ordering (lower index has higher priority). " +
					"Read-only here; manage ordering with the `unifi_acl_rule_order` resource.",
				Computed: true,
			},
		},
	}
}

func (r *aclRuleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
}

func (r *aclRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan aclRuleModel
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
	createBody, diags := asACLRuleBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := client.Official().ACLs().CreateRule(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating ACL rule", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aclRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state aclRuleModel
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
	rule, err := client.Official().ACLs().GetRule(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading ACL rule", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, rule)...)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *aclRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state aclRuleModel
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
	updateBody, diags := asACLRuleBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := client.Official().ACLs().UpdateRule(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating ACL rule", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *aclRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state aclRuleModel
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
	if err := client.Official().ACLs().DeleteRule(ctx, siteID, id); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting ACL rule", err)...)
	}
}

func (r *aclRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
