package dns

import (
	"context"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// DNS policy type discriminators, matching the go-unifi Official API union.
const (
	dnsPolicyTypeA             = "A_RECORD"
	dnsPolicyTypeAAAA          = "AAAA_RECORD"
	dnsPolicyTypeCNAME         = "CNAME_RECORD"
	dnsPolicyTypeMX            = "MX_RECORD"
	dnsPolicyTypeSRV           = "SRV_RECORD"
	dnsPolicyTypeTXT           = "TXT_RECORD"
	dnsPolicyTypeForwardDomain = "FORWARD_DOMAIN"
)

var (
	_ resource.Resource                = &dnsPolicyResource{}
	_ resource.ResourceWithConfigure   = &dnsPolicyResource{}
	_ resource.ResourceWithImportState = &dnsPolicyResource{}
	_ resource.ResourceWithModifyPlan  = &dnsPolicyResource{}
	_ base.ResourceModel               = &dnsPolicyModel{}
	_ base.Resource                    = &dnsPolicyResource{}
)

type dnsPolicyModel struct {
	base.Model
	Type             types.String `tfsdk:"type"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Domain           types.String `tfsdk:"domain"`
	TTLSeconds       types.Int32  `tfsdk:"ttl_seconds"`
	IPv4Address      types.String `tfsdk:"ipv4_address"`
	IPv6Address      types.String `tfsdk:"ipv6_address"`
	TargetDomain     types.String `tfsdk:"target_domain"`
	MailServerDomain types.String `tfsdk:"mail_server_domain"`
	Priority         types.Int32  `tfsdk:"priority"`
	Port             types.Int32  `tfsdk:"port"`
	Weight           types.Int32  `tfsdk:"weight"`
	Service          types.String `tfsdk:"service"`
	Protocol         types.String `tfsdk:"protocol"`
	ServerDomain     types.String `tfsdk:"server_domain"`
	Text             types.String `tfsdk:"text"`
	IPAddress        types.String `tfsdk:"ip_address"`
}

func strPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func int32Ptr(v types.Int32) *int32 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt32()
	return &i
}

func strFromPtr(p *string) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(*p)
}

func int32FromPtr(p *int32) types.Int32 {
	if p == nil {
		return types.Int32Null()
	}
	return types.Int32Value(*p)
}

func (m *dnsPolicyModel) enabledValue() bool {
	if m.Enabled.IsNull() || m.Enabled.IsUnknown() {
		return true
	}
	return m.Enabled.ValueBool()
}

// AsUnifiModel builds the Official-API create/update body for the configured DNS
// policy type. Note that DNSPolicyCreateOrUpdate.MarshalJSON overlays the
// top-level Enabled/Type over the union payload, so Enabled must be set on the
// body in addition to the variant.
func (m *dnsPolicyModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := &official.DNSPolicyCreateOrUpdate{}
	enabled := m.enabledValue()

	var err error
	switch m.Type.ValueString() {
	case dnsPolicyTypeA:
		err = body.FromDnsARecordCreateUpdate(official.DnsARecordCreateUpdate{
			Domain:      strPtr(m.Domain),
			Enabled:     enabled,
			Ipv4Address: strPtr(m.IPv4Address),
			TtlSeconds:  int32Ptr(m.TTLSeconds),
		})
	case dnsPolicyTypeAAAA:
		err = body.FromDnsAaaaRecordCreateUpdate(official.DnsAaaaRecordCreateUpdate{
			Domain:      strPtr(m.Domain),
			Enabled:     enabled,
			Ipv6Address: strPtr(m.IPv6Address),
			TtlSeconds:  int32Ptr(m.TTLSeconds),
		})
	case dnsPolicyTypeCNAME:
		err = body.FromDnsCnameRecordCreateUpdate(official.DnsCnameRecordCreateUpdate{
			Domain:       strPtr(m.Domain),
			Enabled:      enabled,
			TargetDomain: strPtr(m.TargetDomain),
			TtlSeconds:   int32Ptr(m.TTLSeconds),
		})
	case dnsPolicyTypeMX:
		err = body.FromDnsMxRecordCreateUpdate(official.DnsMxRecordCreateUpdate{
			Domain:           strPtr(m.Domain),
			Enabled:          enabled,
			MailServerDomain: strPtr(m.MailServerDomain),
			Priority:         int32Ptr(m.Priority),
		})
	case dnsPolicyTypeSRV:
		err = body.FromDnsSrvRecordCreateUpdate(official.DnsSrvRecordCreateUpdate{
			Domain:       strPtr(m.Domain),
			Enabled:      enabled,
			Port:         int32Ptr(m.Port),
			Priority:     int32Ptr(m.Priority),
			Protocol:     strPtr(m.Protocol),
			ServerDomain: strPtr(m.ServerDomain),
			Service:      strPtr(m.Service),
			Weight:       int32Ptr(m.Weight),
		})
	case dnsPolicyTypeTXT:
		err = body.FromDnsTxtRecordCreateUpdate(official.DnsTxtRecordCreateUpdate{
			Domain:  strPtr(m.Domain),
			Enabled: enabled,
			Text:    strPtr(m.Text),
		})
	case dnsPolicyTypeForwardDomain:
		err = body.FromDnsForwardDomainPolicyCreateUpdate(official.DnsForwardDomainPolicyCreateUpdate{
			Domain:    strPtr(m.Domain),
			Enabled:   enabled,
			IpAddress: strPtr(m.IPAddress),
		})
	default:
		diags.AddError("Unsupported DNS policy type", fmt.Sprintf("Unknown type %q", m.Type.ValueString()))
		return nil, diags
	}
	if err != nil {
		diags.AddError("Failed to build DNS policy request", err.Error())
		return nil, diags
	}
	// The union From* helpers set only Type; MarshalJSON overlays the top-level
	// Enabled, so it must be set explicitly or every request would send false.
	body.Enabled = enabled
	return body, diags
}

// asDNSPolicyBody performs the checked type assertion from the AsUnifiModel
// interface{} return to the concrete Official-API body.
func asDNSPolicyBody(body interface{}) (*official.DNSPolicyCreateOrUpdate, diag.Diagnostics) {
	var diags diag.Diagnostics
	b, ok := body.(*official.DNSPolicyCreateOrUpdate)
	if !ok {
		diags.AddError("Internal error", fmt.Sprintf("expected *official.DNSPolicyCreateOrUpdate, got %T", body))
	}
	return b, diags
}

// Merge populates the model from an Official-API DNSPolicy read response.
func (m *dnsPolicyModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	policy, ok := other.(*official.DNSPolicy)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *official.DNSPolicy, got %T", other))
		return diags
	}

	m.SetID(policy.Id.String())
	m.Type = types.StringValue(policy.Type)
	m.Enabled = types.BoolValue(policy.Enabled)

	switch policy.Type {
	case dnsPolicyTypeA:
		v, err := policy.AsDnsARecord()
		if err != nil {
			diags.AddError("Failed to decode A record DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.IPv4Address = strFromPtr(v.Ipv4Address)
		m.TTLSeconds = int32FromPtr(v.TtlSeconds)
	case dnsPolicyTypeAAAA:
		v, err := policy.AsDnsAaaaRecord()
		if err != nil {
			diags.AddError("Failed to decode AAAA record DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.IPv6Address = strFromPtr(v.Ipv6Address)
		m.TTLSeconds = int32FromPtr(v.TtlSeconds)
	case dnsPolicyTypeCNAME:
		v, err := policy.AsDnsCnameRecord()
		if err != nil {
			diags.AddError("Failed to decode CNAME record DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.TargetDomain = strFromPtr(v.TargetDomain)
		m.TTLSeconds = int32FromPtr(v.TtlSeconds)
	case dnsPolicyTypeMX:
		v, err := policy.AsDnsMxRecord()
		if err != nil {
			diags.AddError("Failed to decode MX record DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.MailServerDomain = strFromPtr(v.MailServerDomain)
		m.Priority = int32FromPtr(v.Priority)
	case dnsPolicyTypeSRV:
		v, err := policy.AsDnsSrvRecord()
		if err != nil {
			diags.AddError("Failed to decode SRV record DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.Port = int32FromPtr(v.Port)
		m.Priority = int32FromPtr(v.Priority)
		m.Protocol = strFromPtr(v.Protocol)
		m.ServerDomain = strFromPtr(v.ServerDomain)
		m.Service = strFromPtr(v.Service)
		m.Weight = int32FromPtr(v.Weight)
	case dnsPolicyTypeTXT:
		v, err := policy.AsDnsTxtRecord()
		if err != nil {
			diags.AddError("Failed to decode TXT record DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.Text = strFromPtr(v.Text)
	case dnsPolicyTypeForwardDomain:
		v, err := policy.AsDnsForwardDomainPolicy()
		if err != nil {
			diags.AddError("Failed to decode forward-domain DNS policy", err.Error())
			return diags
		}
		m.Domain = strFromPtr(v.Domain)
		m.IPAddress = strFromPtr(v.IpAddress)
	default:
		diags.AddError("Unsupported DNS policy type", fmt.Sprintf("Controller returned unknown type %q", policy.Type))
	}
	return diags
}

type dnsPolicyResource struct {
	*base.GenericResource[*dnsPolicyModel]
}

// NewDNSPolicyResource creates the unifi_dns_policy resource. It embeds a
// GenericResource purely for Configure/version-validator wiring; CRUD is custom
// because the Official API is keyed by site/entity UUID.
func NewDNSPolicyResource() resource.Resource {
	r := &dnsPolicyResource{}
	r.GenericResource = base.NewGenericResource(
		"unifi_dns_policy",
		func() *dnsPolicyModel { return &dnsPolicyModel{} },
		base.ResourceFunctions{},
	)
	return r
}

func (r *dnsPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a DNS policy (record or forwarding rule) via the UniFi **Official API** " +
			"(`integration/v1`). Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"type": schema.StringAttribute{
				MarkdownDescription: "The DNS policy type. One of `A_RECORD`, `AAAA_RECORD`, `CNAME_RECORD`, " +
					"`MX_RECORD`, `SRV_RECORD`, `TXT_RECORD`, `FORWARD_DOMAIN`. Changing the type forces a new resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						dnsPolicyTypeA, dnsPolicyTypeAAAA, dnsPolicyTypeCNAME, dnsPolicyTypeMX,
						dnsPolicyTypeSRV, dnsPolicyTypeTXT, dnsPolicyTypeForwardDomain,
					),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the DNS policy is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain the policy applies to. Required for all record types and for `FORWARD_DOMAIN`.",
				Optional:            true,
				Computed:            true,
			},
			"ttl_seconds": schema.Int32Attribute{
				MarkdownDescription: "Time to live in seconds. Applies to `A_RECORD`, `AAAA_RECORD` and `CNAME_RECORD`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.Int32{int32validator.AtLeast(0)},
			},
			"ipv4_address": schema.StringAttribute{
				MarkdownDescription: "IPv4 address for an `A_RECORD`.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ipv6_address": schema.StringAttribute{
				MarkdownDescription: "IPv6 address for an `AAAA_RECORD`.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"target_domain": schema.StringAttribute{
				MarkdownDescription: "Target domain for a `CNAME_RECORD`.",
				Optional:            true,
			},
			"mail_server_domain": schema.StringAttribute{
				MarkdownDescription: "Mail server domain for an `MX_RECORD`.",
				Optional:            true,
			},
			"priority": schema.Int32Attribute{
				MarkdownDescription: "Priority (lower is preferred). Applies to `MX_RECORD` and `SRV_RECORD`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.Int32{int32validator.Between(0, 65535)},
			},
			"port": schema.Int32Attribute{
				MarkdownDescription: "Service port for an `SRV_RECORD`.",
				Optional:            true,
				Validators:          []validator.Int32{int32validator.Between(0, 65535)},
			},
			"weight": schema.Int32Attribute{
				MarkdownDescription: "Relative weight for records with the same priority (`SRV_RECORD`).",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.Int32{int32validator.Between(0, 65535)},
			},
			"service": schema.StringAttribute{
				MarkdownDescription: "Service name for an `SRV_RECORD` (e.g. `_sip`).",
				Optional:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol used by the service for an `SRV_RECORD` (e.g. `_tcp`).",
				Optional:            true,
			},
			"server_domain": schema.StringAttribute{
				MarkdownDescription: "Domain of the server running the service for an `SRV_RECORD`.",
				Optional:            true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "Text value for a `TXT_RECORD`.",
				Optional:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "IP address of the DNS server queries are forwarded to (`FORWARD_DOMAIN`).",
				Optional:            true,
			},
		},
	}
}

// requiredFor maps each DNS policy type to the attributes it requires.
var dnsPolicyRequiredAttrs = map[string][]string{
	dnsPolicyTypeA:             {"domain", "ipv4_address"},
	dnsPolicyTypeAAAA:          {"domain", "ipv6_address"},
	dnsPolicyTypeCNAME:         {"domain", "target_domain"},
	dnsPolicyTypeMX:            {"domain", "mail_server_domain"},
	dnsPolicyTypeSRV:           {"domain", "service", "protocol"},
	dnsPolicyTypeTXT:           {"domain", "text"},
	dnsPolicyTypeForwardDomain: {"domain", "ip_address"},
}

func (r *dnsPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans carry a null plan; nothing to validate.
	if req.Plan.Raw.IsNull() {
		return
	}
	resp.Diagnostics.Append(r.GetClient().RequireOfficialAPI()...)
	resp.Diagnostics.Append(r.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)

	var plan dnsPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyType := plan.Type.ValueString()
	if policyType == "" {
		return
	}
	for _, attrName := range dnsPolicyRequiredAttrs[policyType] {
		if attrIsUnset(&plan, attrName) {
			resp.Diagnostics.AddAttributeError(
				path.Root(attrName),
				"Missing required attribute",
				fmt.Sprintf("`%s` is required when `type` is %q.", attrName, policyType),
			)
		}
	}
}

func attrIsUnset(m *dnsPolicyModel, attrName string) bool {
	switch attrName {
	case "domain":
		return ut.IsEmptyString(m.Domain)
	case "ipv4_address":
		return ut.IsEmptyString(m.IPv4Address)
	case "ipv6_address":
		return ut.IsEmptyString(m.IPv6Address)
	case "target_domain":
		return ut.IsEmptyString(m.TargetDomain)
	case "mail_server_domain":
		return ut.IsEmptyString(m.MailServerDomain)
	case "service":
		return ut.IsEmptyString(m.Service)
	case "protocol":
		return ut.IsEmptyString(m.Protocol)
	case "text":
		return ut.IsEmptyString(m.Text)
	case "ip_address":
		return ut.IsEmptyString(m.IPAddress)
	default:
		return false
	}
}

func (r *dnsPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan dnsPolicyModel
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
	createBody, diags := asDNSPolicyBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := client.Official().DNSPolicies().Create(ctx, siteID, *createBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error creating DNS policy", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, created)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dnsPolicyModel
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
	policy, err := client.Official().DNSPolicies().Get(ctx, siteID, id)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading DNS policy", err)...)
		return
	}
	resp.Diagnostics.Append(state.Merge(ctx, policy)...)
	state.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dnsPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan, state dnsPolicyModel
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
	updateBody, diags := asDNSPolicyBody(body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updated, err := client.Official().DNSPolicies().Update(ctx, siteID, id, *updateBody)
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error updating DNS policy", err)...)
		return
	}
	resp.Diagnostics.Append(plan.Merge(ctx, updated)...)
	plan.SetSite(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dnsPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	client := r.GetClient()
	resp.Diagnostics.Append(base.CheckConfigured(client)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state dnsPolicyModel
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
	if err := client.Official().DNSPolicies().Delete(ctx, siteID, id); err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error deleting DNS policy", err)...)
	}
}

func (r *dnsPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, site := base.ImportIDWithSite(req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site"), site)...)
}
