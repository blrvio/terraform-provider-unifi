package appblock

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

var (
	_ resource.Resource                     = &firewallAppBlockResource{}
	_ resource.ResourceWithConfigure        = &firewallAppBlockResource{}
	_ resource.ResourceWithImportState      = &firewallAppBlockResource{}
	_ resource.ResourceWithConfigValidators = &firewallAppBlockResource{}
	_ base.Resource                         = &firewallAppBlockResource{}
	_ base.ResourceModel                    = &firewallAppBlockModel{}
)

type firewallAppBlockModel struct {
	base.Model
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	TargetType     types.String `tfsdk:"target_type"`
	AppCategoryIDs types.Set    `tfsdk:"app_category_ids"`
	AppIDs         types.Set    `tfsdk:"app_ids"`
	ClientMACs     types.Set    `tfsdk:"client_macs"`
	NetworkIDs     types.Set    `tfsdk:"network_ids"`
	Schedule       types.Object `tfsdk:"schedule"`
}

func int64s(ctx context.Context, set types.Set) ([]int, diag.Diagnostics) {
	var in []int64
	var diags diag.Diagnostics
	if ut.IsDefined(set) {
		diags.Append(set.ElementsAs(ctx, &in, false)...)
	}
	out := make([]int, 0, len(in))
	for _, v := range in {
		out = append(out, int(v))
	}
	return out, diags
}

func strs(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	out := []string{}
	var diags diag.Diagnostics
	if ut.IsDefined(set) {
		diags.Append(set.ElementsAs(ctx, &out, false)...)
	}
	return out, diags
}

func (m *firewallAppBlockModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	cats, d := int64s(ctx, m.AppCategoryIDs)
	diags.Append(d...)
	apps, d := int64s(ctx, m.AppIDs)
	diags.Append(d...)
	macs, d := strs(ctx, m.ClientMACs)
	diags.Append(d...)
	for i, mac := range macs {
		macs[i] = utils.CleanMAC(mac)
	}
	nets, d := strs(ctx, m.NetworkIDs)
	diags.Append(d...)
	sv, d := ut.ScheduleFromObject(ctx, m.Schedule)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &unifi.FirewallAppBlock{
		ID:             m.ID.ValueString(),
		Name:           m.Name.ValueString(),
		Type:           m.Type.ValueString(),
		TargetType:     m.TargetType.ValueString(),
		AppCategoryIDs: cats,
		AppIDs:         apps,
		ClientMACs:     macs,
		NetworkIDs:     nets,
		Schedule: unifi.FirewallAppBlockSchedule{
			Mode: sv.Mode, Date: sv.Date, DateStart: sv.DateStart, DateEnd: sv.DateEnd, RepeatOnDays: sv.RepeatOnDays,
			TimeAllDay: sv.TimeAllDay, TimeRangeStart: sv.TimeRangeStart, TimeRangeEnd: sv.TimeRangeEnd,
		},
	}, diags
}

func (m *firewallAppBlockModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	b, ok := other.(*unifi.FirewallAppBlock)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.FirewallAppBlock, got %T", other))
		return diags
	}
	m.ID = types.StringValue(b.ID)
	m.Name = types.StringValue(b.Name)
	m.Type = types.StringValue(b.Type)
	m.TargetType = types.StringValue(b.TargetType)
	toInt64 := func(in []int) []int64 {
		out := make([]int64, 0, len(in))
		for _, v := range in {
			out = append(out, int64(v))
		}
		return out
	}
	nonNil := func(in []string) []string {
		if in == nil {
			return []string{}
		}
		return in
	}
	var d diag.Diagnostics
	m.AppCategoryIDs, d = types.SetValueFrom(ctx, types.Int64Type, toInt64(b.AppCategoryIDs))
	diags.Append(d...)
	m.AppIDs, d = types.SetValueFrom(ctx, types.Int64Type, toInt64(b.AppIDs))
	diags.Append(d...)
	m.ClientMACs, d = types.SetValueFrom(ctx, ut.MACType{}, nonNil(b.ClientMACs))
	diags.Append(d...)
	m.NetworkIDs, d = types.SetValueFrom(ctx, types.StringType, nonNil(b.NetworkIDs))
	diags.Append(d...)
	s := b.Schedule
	m.Schedule, d = ut.ScheduleToObject(ctx, ut.ScheduleValues{
		Mode: s.Mode, Date: s.Date, DateStart: s.DateStart, DateEnd: s.DateEnd, RepeatOnDays: s.RepeatOnDays,
		TimeAllDay: s.TimeAllDay, TimeRangeStart: s.TimeRangeStart, TimeRangeEnd: s.TimeRangeEnd,
	})
	diags.Append(d...)
	return diags
}

type firewallAppBlockResource struct {
	*base.GenericResource[*firewallAppBlockModel]
}

// NewFirewallAppBlockResource creates the unifi_firewall_app_block resource
// (CyberSecure "Simple App Blocking", v2 firewall-app-blocks).
func NewFirewallAppBlockResource() resource.Resource {
	return &firewallAppBlockResource{
		GenericResource: base.NewGenericResource(
			"unifi_firewall_app_block",
			func() *firewallAppBlockModel { return &firewallAppBlockModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetFirewallAppBlock(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					b, _ := body.(*unifi.FirewallAppBlock)
					return client.CreateFirewallAppBlock(ctx, site, b)
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					b, _ := body.(*unifi.FirewallAppBlock)
					return client.UpdateFirewallAppBlock(ctx, site, b)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteFirewallAppBlock(ctx, site, id)
				},
				// The v2 collection has no GET-by-id; the post-write state comes from the list.
				ReadAfterWrite: true,
			},
		),
	}
}

func (r *firewallAppBlockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyStrings := setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{}))
	emptyInts := setdefault.StaticValue(types.SetValueMust(types.Int64Type, []attr.Value{}))
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a CyberSecure **Simple App Blocking** rule (Settings > CyberSecure > Simple App Blocking): " +
			"block applications or application categories for a set of clients (`type = \"DEVICE\"`) or networks " +
			"(`type = \"NETWORK\"`), optionally on a schedule.\n\n" +
			"The controller derives the matching firewall policies from this rule (origin `simple_app_block`, one per source " +
			"zone). They appear as DERIVED policies in the policy table and change only through this resource.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the rule (shown in the policy table).",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "What the rule targets: `DEVICE` (the clients in `client_macs`) or `NETWORK` (the networks in `network_ids`). Defaults to `DEVICE`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("DEVICE"),
				Validators:          []validator.String{stringvalidator.OneOf("DEVICE", "NETWORK")},
			},
			"target_type": schema.StringAttribute{
				MarkdownDescription: "What is blocked: `APP_CATEGORY` (the categories in `app_category_ids`) or `APP` (the applications in `app_ids`).",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("APP", "APP_CATEGORY")},
			},
			"app_category_ids": schema.SetAttribute{
				MarkdownDescription: "DPI application category IDs to block (see the `unifi_dpi_application_categories` data source).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int64Type,
				Default:             emptyInts,
			},
			"app_ids": schema.SetAttribute{
				MarkdownDescription: "DPI application IDs to block (see the `unifi_dpi_applications` data source).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int64Type,
				Default:             emptyInts,
			},
			"client_macs": schema.SetAttribute{
				MarkdownDescription: "MAC addresses of the clients the rule applies to (`type = \"DEVICE\"`). Case-insensitive, `:` or `-` separators.",
				Optional:            true,
				Computed:            true,
				ElementType:         ut.MACType{},
				Default:             setdefault.StaticValue(types.SetValueMust(ut.MACType{}, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.Mac),
					validators.UniqueMACs(),
				},
			},
			"network_ids": schema.SetAttribute{
				MarkdownDescription: "IDs of the networks the rule applies to (`type = \"NETWORK\"`).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStrings,
			},
			"schedule": ut.ScheduleAttribute("When the rule is active. The UI's \"Always\" is `{ mode = \"ALWAYS\", repeat_on_days = [], time_all_day = false }`."),
		},
	}
}

func (r *firewallAppBlockResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("DEVICE"), path.MatchRoot("network_ids")),
		validators.RequiredNoneIf(path.MatchRoot("type"), types.StringValue("NETWORK"), path.MatchRoot("client_macs")),
		validators.RequiredNoneIf(path.MatchRoot("target_type"), types.StringValue("APP_CATEGORY"), path.MatchRoot("app_ids")),
		validators.RequiredNoneIf(path.MatchRoot("target_type"), types.StringValue("APP"), path.MatchRoot("app_category_ids")),
	}
}
