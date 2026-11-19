package nat

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
)

var filterTypes = []string{"NONE", "ADDRESS_AND_PORT", "NETWORK_CONF", "FIREWALL_GROUPS"}

var (
	_ resource.Resource                = &natRuleResource{}
	_ resource.ResourceWithConfigure   = &natRuleResource{}
	_ resource.ResourceWithImportState = &natRuleResource{}
	_ base.ResourceModel               = &natRuleModel{}
)

type natFilterModel struct {
	FilterType       types.String `tfsdk:"filter_type"`
	Address          types.String `tfsdk:"address"`
	Port             types.String `tfsdk:"port"`
	InvertAddress    types.Bool   `tfsdk:"invert_address"`
	InvertPort       types.Bool   `tfsdk:"invert_port"`
	NetworkConfID    types.String `tfsdk:"network_conf_id"`
	FirewallGroupIDs types.Set    `tfsdk:"firewall_group_ids"`
}

func (natFilterModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"filter_type":        types.StringType,
		"address":            types.StringType,
		"port":               types.StringType,
		"invert_address":     types.BoolType,
		"invert_port":        types.BoolType,
		"network_conf_id":    types.StringType,
		"firewall_group_ids": types.SetType{ElemType: types.StringType},
	}
}

type natRuleModel struct {
	base.Model
	Description           types.String `tfsdk:"description"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	Type                  types.String `tfsdk:"type"`
	IPVersion             types.String `tfsdk:"ip_version"`
	InInterface           types.String `tfsdk:"in_interface"`
	OutInterface          types.String `tfsdk:"out_interface"`
	Protocol              types.String `tfsdk:"protocol"`
	Port                  types.String `tfsdk:"port"`
	IPAddress             types.String `tfsdk:"ip_address"`
	Logging               types.Bool   `tfsdk:"logging"`
	Exclude               types.Bool   `tfsdk:"exclude"`
	PppoeUseBaseInterface types.Bool   `tfsdk:"pppoe_use_base_interface"`
	RuleIndex             types.Int64  `tfsdk:"rule_index"`
	SourceFilter          types.Object `tfsdk:"source_filter"`
	DestinationFilter     types.Object `tfsdk:"destination_filter"`
}

type filterValues struct {
	FilterType, Address, Port, NetworkConfID string
	InvertAddress, InvertPort                bool
	FirewallGroupIDs                         []string
}

func filterFromObject(ctx context.Context, obj types.Object) (filterValues, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !ut.IsDefined(obj) {
		return filterValues{FilterType: "NONE"}, diags
	}
	var f natFilterModel
	diags.Append(obj.As(ctx, &f, basetypes.ObjectAsOptions{})...)
	var groups []string
	if ut.IsDefined(f.FirewallGroupIDs) {
		diags.Append(f.FirewallGroupIDs.ElementsAs(ctx, &groups, false)...)
	}
	return filterValues{
		FilterType: f.FilterType.ValueString(), Address: f.Address.ValueString(), Port: f.Port.ValueString(),
		NetworkConfID: f.NetworkConfID.ValueString(), InvertAddress: f.InvertAddress.ValueBool(),
		InvertPort: f.InvertPort.ValueBool(), FirewallGroupIDs: groups,
	}, diags
}

func filterToObject(ctx context.Context, v filterValues) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	var groups types.Set
	if len(v.FirewallGroupIDs) == 0 {
		groups = types.SetNull(types.StringType)
	} else {
		var d diag.Diagnostics
		groups, d = types.SetValueFrom(ctx, types.StringType, v.FirewallGroupIDs)
		diags.Append(d...)
	}
	f := natFilterModel{
		FilterType: types.StringValue(utils.OrDefault(v.FilterType, "NONE")), Address: ut.StringOrNull(v.Address),
		Port: ut.StringOrNull(v.Port), InvertAddress: types.BoolValue(v.InvertAddress), InvertPort: types.BoolValue(v.InvertPort),
		NetworkConfID: ut.StringOrNull(v.NetworkConfID), FirewallGroupIDs: groups,
	}
	obj, d := types.ObjectValueFrom(ctx, f.AttributeTypes(), f)
	diags.Append(d...)
	return obj, diags
}

func (m *natRuleModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	src, d := filterFromObject(ctx, m.SourceFilter)
	diags.Append(d...)
	dst, d := filterFromObject(ctx, m.DestinationFilter)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &unifi.NatRule{
		ID: m.ID.ValueString(), Description: m.Description.ValueString(), Enabled: m.Enabled.ValueBool(),
		Type: m.Type.ValueString(), IPVersion: m.IPVersion.ValueString(), InInterface: m.InInterface.ValueString(),
		OutInterface: m.OutInterface.ValueString(), Protocol: m.Protocol.ValueString(), Port: m.Port.ValueString(),
		IPAddress: m.IPAddress.ValueString(), Logging: m.Logging.ValueBool(), Exclude: m.Exclude.ValueBool(),
		PppoeUseBaseInterface: m.PppoeUseBaseInterface.ValueBool(), RuleIndex: int(m.RuleIndex.ValueInt64()),
		SettingPreference: "manual",
		SourceFilter: unifi.NatRuleSourceFilter{
			FilterType: src.FilterType, Address: src.Address, Port: src.Port, InvertAddress: src.InvertAddress,
			InvertPort: src.InvertPort, NetworkConfID: src.NetworkConfID, FirewallGroupIDs: src.FirewallGroupIDs,
		},
		DestinationFilter: unifi.NatRuleDestinationFilter{
			FilterType: dst.FilterType, Address: dst.Address, Port: dst.Port, InvertAddress: dst.InvertAddress,
			InvertPort: dst.InvertPort, NetworkConfID: dst.NetworkConfID, FirewallGroupIDs: dst.FirewallGroupIDs,
		},
	}, diags
}

func (m *natRuleModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	r, ok := other.(*unifi.NatRule)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.NatRule, got %T", other))
		return diags
	}
	m.ID = types.StringValue(r.ID)
	m.Description = types.StringValue(r.Description)
	m.Enabled = types.BoolValue(r.Enabled)
	m.Type = types.StringValue(r.Type)
	m.IPVersion = types.StringValue(utils.OrDefault(r.IPVersion, "IPV4"))
	m.InInterface = ut.StringOrNull(r.InInterface)
	m.OutInterface = ut.StringOrNull(r.OutInterface)
	m.Protocol = types.StringValue(utils.OrDefault(r.Protocol, "all"))
	m.Port = ut.StringOrNull(r.Port)
	m.IPAddress = ut.StringOrNull(r.IPAddress)
	m.Logging = types.BoolValue(r.Logging)
	m.Exclude = types.BoolValue(r.Exclude)
	m.PppoeUseBaseInterface = types.BoolValue(r.PppoeUseBaseInterface)
	m.RuleIndex = types.Int64Value(int64(r.RuleIndex))
	var d diag.Diagnostics
	s, ds := r.SourceFilter, r.DestinationFilter
	m.SourceFilter, d = filterToObject(ctx, filterValues{s.FilterType, s.Address, s.Port, s.NetworkConfID, s.InvertAddress, s.InvertPort, s.FirewallGroupIDs})
	diags.Append(d...)
	m.DestinationFilter, d = filterToObject(ctx, filterValues{ds.FilterType, ds.Address, ds.Port, ds.NetworkConfID, ds.InvertAddress, ds.InvertPort, ds.FirewallGroupIDs})
	diags.Append(d...)
	return diags
}

type natRuleResource struct {
	*base.GenericResource[*natRuleModel]
}

// NewNatRuleResource creates the unifi_nat_rule resource (v2 nat, user-defined rules).
func NewNatRuleResource() resource.Resource {
	return &natRuleResource{
		GenericResource: base.NewGenericResource(
			"unifi_nat_rule",
			func() *natRuleModel { return &natRuleModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetNatRule(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					b, _ := body.(*unifi.NatRule)
					return client.CreateNatRule(ctx, site, b)
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					b, _ := body.(*unifi.NatRule)
					return client.UpdateNatRule(ctx, site, b)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteNatRule(ctx, site, id)
				},
				ReadAfterWrite: true,
			},
		),
	}
}

func filterAttribute(desc string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: desc + " Defaults to matching anything (`filter_type = \"NONE\"`).",
		Optional:            true,
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"filter_type": schema.StringAttribute{
				MarkdownDescription: "`NONE`, `ADDRESS_AND_PORT` (`address` and/or `port`), `NETWORK_CONF` (`network_conf_id`) or `FIREWALL_GROUPS` (`firewall_group_ids`).",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf(filterTypes...)},
			},
			"address":            schema.StringAttribute{MarkdownDescription: "IP address, range or CIDR.", Optional: true},
			"port":               schema.StringAttribute{MarkdownDescription: "Port or port range.", Optional: true},
			"invert_address":     schema.BoolAttribute{MarkdownDescription: "Match everything except `address`.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"invert_port":        schema.BoolAttribute{MarkdownDescription: "Match everything except `port`.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"network_conf_id":    schema.StringAttribute{MarkdownDescription: "Network (networkconf) ID.", Optional: true},
			"firewall_group_ids": schema.SetAttribute{MarkdownDescription: "Firewall group (port/address object) IDs.", Optional: true, ElementType: types.StringType, Validators: []validator.Set{setvalidator.SizeAtLeast(1)}},
		},
	}
}

func (r *natRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "**Experimental.** Manages a user-defined NAT rule (Settings > Policy Table > NAT): masquerade, " +
			"source NAT or destination NAT. The schema follows the UniFi Network 11.0.99 UI and has not been validated against a " +
			"rule imported from a console yet.\n\n" +
			"The automatic per-WAN \"NAT masquerade\" rows the policy table shows are not NAT rules: the UI builds them from the " +
			"network configuration, and they are controlled by `unifi_setting_global_nat` (`mode`, `excluded_network_ids`).",
		Attributes: map[string]schema.Attribute{
			"id":          ut.ID(),
			"site":        ut.SiteAttribute(),
			"description": schema.StringAttribute{MarkdownDescription: "The rule name shown in the policy table.", Required: true},
			"enabled":     schema.BoolAttribute{MarkdownDescription: "Whether the rule is active. Defaults to `true`.", Optional: true, Computed: true, Default: booldefault.StaticBool(true)},
			"type": schema.StringAttribute{
				MarkdownDescription: "`MASQUERADE`, `SNAT` or `DNAT`.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("MASQUERADE", "SNAT", "DNAT")},
			},
			"ip_version": schema.StringAttribute{
				MarkdownDescription: "`IPV4` (default) or `IPV6`.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("IPV4"),
				Validators: []validator.String{stringvalidator.OneOf("IPV4", "IPV6")},
			},
			"in_interface":  schema.StringAttribute{MarkdownDescription: "Inbound interface (network ID) for `DNAT` rules.", Optional: true},
			"out_interface": schema.StringAttribute{MarkdownDescription: "Outbound interface (network ID) for `MASQUERADE`/`SNAT` rules.", Optional: true},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "`all` (default), `tcp`, `udp` or `tcp_udp`.",
				Optional:            true, Computed: true, Default: stringdefault.StaticString("all"),
				Validators: []validator.String{stringvalidator.OneOf("all", "tcp", "udp", "tcp_udp")},
			},
			"port":       schema.StringAttribute{MarkdownDescription: "Translated port (`DNAT`).", Optional: true},
			"ip_address": schema.StringAttribute{MarkdownDescription: "Translated IP address (`SNAT`/`DNAT`).", Optional: true},
			"logging":    schema.BoolAttribute{MarkdownDescription: "Syslog logging. Defaults to `false`.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"exclude":    schema.BoolAttribute{MarkdownDescription: "Exclude matching traffic from NAT. Defaults to `false`.", Optional: true, Computed: true, Default: booldefault.StaticBool(false)},
			"pppoe_use_base_interface": schema.BoolAttribute{
				MarkdownDescription: "On a PPPoE WAN, use the underlying interface. Defaults to `false`.",
				Optional:            true, Computed: true, Default: booldefault.StaticBool(false),
			},
			"rule_index": schema.Int64Attribute{
				MarkdownDescription: "Evaluation order; assigned by the controller when unset.",
				Optional:            true, Computed: true,
			},
			"source_filter":      filterAttribute("Source match."),
			"destination_filter": filterAttribute("Destination match."),
		},
	}
}
