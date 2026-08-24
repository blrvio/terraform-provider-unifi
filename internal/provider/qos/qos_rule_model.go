package qos

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var _ base.ResourceModel = &qosRuleModel{}

// qosRuleModel is the Terraform model for a UniFi QoS rule (v2 qos-rules).
type qosRuleModel struct {
	base.Model
	Name            types.String `tfsdk:"name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Index           types.Int64  `tfsdk:"index"`
	Objective       types.String `tfsdk:"objective"`
	DownloadBurst   types.String `tfsdk:"download_burst"`
	UploadBurst     types.String `tfsdk:"upload_burst"`
	WANOrVPNNetwork types.String `tfsdk:"wan_or_vpn_network"`
	Schedule        types.Object `tfsdk:"schedule"`
	Source          types.Object `tfsdk:"source"`
	Destination     types.Object `tfsdk:"destination"`
}

// qosScheduleModel is the nested `schedule` block.
type qosScheduleModel struct {
	Mode           types.String `tfsdk:"mode"`
	Date           types.String `tfsdk:"date"`
	DateStart      types.String `tfsdk:"date_start"`
	DateEnd        types.String `tfsdk:"date_end"`
	RepeatOnDays   types.Set    `tfsdk:"repeat_on_days"`
	TimeAllDay     types.Bool   `tfsdk:"time_all_day"`
	TimeRangeStart types.String `tfsdk:"time_range_start"`
	TimeRangeEnd   types.String `tfsdk:"time_range_end"`
}

func (qosScheduleModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":             types.StringType,
		"date":             types.StringType,
		"date_start":       types.StringType,
		"date_end":         types.StringType,
		"repeat_on_days":   types.SetType{ElemType: types.StringType},
		"time_all_day":     types.BoolType,
		"time_range_start": types.StringType,
		"time_range_end":   types.StringType,
	}
}

// qosSourceModel is the nested `source` matching block.
type qosSourceModel struct {
	MatchingTarget   types.String `tfsdk:"matching_target"`
	PortMatchingType types.String `tfsdk:"port_matching_type"`
	NetworkIDs       types.List   `tfsdk:"network_ids"`
	ClientMACs       types.List   `tfsdk:"client_macs"`
	IPs              types.List   `tfsdk:"ips"`
	Regions          types.List   `tfsdk:"regions"`
}

func (qosSourceModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"matching_target":    types.StringType,
		"port_matching_type": types.StringType,
		"network_ids":        types.ListType{ElemType: types.StringType},
		"client_macs":        types.ListType{ElemType: types.StringType},
		"ips":                types.ListType{ElemType: types.StringType},
		"regions":            types.ListType{ElemType: types.StringType},
	}
}

// qosDestinationModel is the nested `destination` matching block.
type qosDestinationModel struct {
	MatchingTarget   types.String `tfsdk:"matching_target"`
	PortMatchingType types.String `tfsdk:"port_matching_type"`
	AppIDs           types.List   `tfsdk:"app_ids"`
	AppCategoryIDs   types.List   `tfsdk:"app_category_ids"`
	NetworkIDs       types.List   `tfsdk:"network_ids"`
	IPs              types.List   `tfsdk:"ips"`
	Regions          types.List   `tfsdk:"regions"`
}

func (qosDestinationModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"matching_target":    types.StringType,
		"port_matching_type": types.StringType,
		"app_ids":            types.ListType{ElemType: types.Int64Type},
		"app_category_ids":   types.ListType{ElemType: types.Int64Type},
		"network_ids":        types.ListType{ElemType: types.StringType},
		"ips":                types.ListType{ElemType: types.StringType},
		"regions":            types.ListType{ElemType: types.StringType},
	}
}

func isZeroQOSSchedule(s unifi.QOSRuleSchedule) bool {
	return s.Mode == "" && s.Date == "" && s.DateStart == "" && s.DateEnd == "" &&
		s.TimeRangeStart == "" && s.TimeRangeEnd == "" && !s.TimeAllDay && len(s.RepeatOnDays) == 0
}

// AsUnifiModel converts the Terraform model to the go-unifi model.
func (m *qosRuleModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	rule := &unifi.QOSRule{
		ID:              m.ID.ValueString(),
		Name:            m.Name.ValueString(),
		Enabled:         m.Enabled.ValueBool(),
		Index:           int(m.Index.ValueInt64()),
		Objective:       m.Objective.ValueString(),
		DownloadBurst:   m.DownloadBurst.ValueString(),
		UploadBurst:     m.UploadBurst.ValueString(),
		WANOrVPNNetwork: m.WANOrVPNNetwork.ValueString(),
	}

	// schedule
	if ut.IsDefined(m.Schedule) {
		var sm qosScheduleModel
		diags.Append(m.Schedule.As(ctx, &sm, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		var repeatOnDays []string
		diags.Append(ut.ListElementsAs(ctx, listFromSet(ctx, sm.RepeatOnDays, &diags), &repeatOnDays)...)
		rule.Schedule = unifi.QOSRuleSchedule{
			Mode:           sm.Mode.ValueString(),
			Date:           sm.Date.ValueString(),
			DateStart:      sm.DateStart.ValueString(),
			DateEnd:        sm.DateEnd.ValueString(),
			RepeatOnDays:   repeatOnDays,
			TimeAllDay:     sm.TimeAllDay.ValueBool(),
			TimeRangeStart: sm.TimeRangeStart.ValueString(),
			TimeRangeEnd:   sm.TimeRangeEnd.ValueString(),
		}
	}

	// source
	if ut.IsDefined(m.Source) {
		var sm qosSourceModel
		diags.Append(m.Source.As(ctx, &sm, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		src := unifi.QOSRuleSource{
			MatchingTarget:   sm.MatchingTarget.ValueString(),
			PortMatchingType: sm.PortMatchingType.ValueString(),
		}
		diags.Append(ut.ListElementsAs(ctx, sm.NetworkIDs, &src.NetworkIDs)...)
		diags.Append(ut.ListElementsAs(ctx, sm.ClientMACs, &src.ClientMACs)...)
		diags.Append(ut.ListElementsAs(ctx, sm.IPs, &src.IPs)...)
		diags.Append(ut.ListElementsAs(ctx, sm.Regions, &src.Regions)...)
		rule.Source = src
	}

	// destination
	if ut.IsDefined(m.Destination) {
		var dm qosDestinationModel
		diags.Append(m.Destination.As(ctx, &dm, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		dst := unifi.QOSRuleDestination{
			MatchingTarget:   dm.MatchingTarget.ValueString(),
			PortMatchingType: dm.PortMatchingType.ValueString(),
			AppIDs:           int64ListToInts(ctx, &diags, dm.AppIDs),
			AppCategoryIDs:   int64ListToInts(ctx, &diags, dm.AppCategoryIDs),
		}
		diags.Append(ut.ListElementsAs(ctx, dm.NetworkIDs, &dst.NetworkIDs)...)
		diags.Append(ut.ListElementsAs(ctx, dm.IPs, &dst.IPs)...)
		diags.Append(ut.ListElementsAs(ctx, dm.Regions, &dst.Regions)...)
		rule.Destination = dst
	}

	return rule, diags
}

// Merge updates the Terraform model from the go-unifi model.
func (m *qosRuleModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	rule, ok := other.(*unifi.QOSRule)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.QOSRule, got %T", other))
		return diags
	}

	m.ID = types.StringValue(rule.ID)
	m.Name = types.StringValue(rule.Name)
	m.Enabled = types.BoolValue(rule.Enabled)
	m.Index = types.Int64Value(int64(rule.Index))
	m.Objective = ut.StringOrNull(rule.Objective)
	m.DownloadBurst = ut.StringOrNull(rule.DownloadBurst)
	m.UploadBurst = ut.StringOrNull(rule.UploadBurst)
	m.WANOrVPNNetwork = ut.StringOrNull(rule.WANOrVPNNetwork)

	var d diag.Diagnostics

	// schedule
	if isZeroQOSSchedule(rule.Schedule) {
		m.Schedule, d = ut.ObjectNull(&qosScheduleModel{})
		diags.Append(d...)
	} else {
		repeatOnDays, rd := types.SetValueFrom(ctx, types.StringType, orEmptyStrings(rule.Schedule.RepeatOnDays))
		diags.Append(rd...)
		sm := qosScheduleModel{
			Mode:           ut.StringOrNull(rule.Schedule.Mode),
			Date:           ut.StringOrNull(rule.Schedule.Date),
			DateStart:      ut.StringOrNull(rule.Schedule.DateStart),
			DateEnd:        ut.StringOrNull(rule.Schedule.DateEnd),
			RepeatOnDays:   repeatOnDays,
			TimeAllDay:     types.BoolValue(rule.Schedule.TimeAllDay),
			TimeRangeStart: ut.StringOrNull(rule.Schedule.TimeRangeStart),
			TimeRangeEnd:   ut.StringOrNull(rule.Schedule.TimeRangeEnd),
		}
		m.Schedule, d = types.ObjectValueFrom(ctx, sm.AttributeTypes(), sm)
		diags.Append(d...)
	}

	// source
	src := rule.Source
	sm := qosSourceModel{
		MatchingTarget:   ut.StringOrNull(src.MatchingTarget),
		PortMatchingType: ut.StringOrNull(src.PortMatchingType),
		NetworkIDs:       stringListOrNull(ctx, &diags, src.NetworkIDs),
		ClientMACs:       stringListOrNull(ctx, &diags, src.ClientMACs),
		IPs:              stringListOrNull(ctx, &diags, src.IPs),
		Regions:          stringListOrNull(ctx, &diags, src.Regions),
	}
	m.Source, d = types.ObjectValueFrom(ctx, sm.AttributeTypes(), sm)
	diags.Append(d...)

	// destination
	dst := rule.Destination
	dm := qosDestinationModel{
		MatchingTarget:   ut.StringOrNull(dst.MatchingTarget),
		PortMatchingType: ut.StringOrNull(dst.PortMatchingType),
		AppIDs:           int64ListOrNull(ctx, &diags, dst.AppIDs),
		AppCategoryIDs:   int64ListOrNull(ctx, &diags, dst.AppCategoryIDs),
		NetworkIDs:       stringListOrNull(ctx, &diags, dst.NetworkIDs),
		IPs:              stringListOrNull(ctx, &diags, dst.IPs),
		Regions:          stringListOrNull(ctx, &diags, dst.Regions),
	}
	m.Destination, d = types.ObjectValueFrom(ctx, dm.AttributeTypes(), dm)
	diags.Append(d...)

	return diags
}

// stringListOrNull builds a types.List of strings; empty/nil -> null list so
// unset matching fields don't produce a spurious empty list in state.
func stringListOrNull(ctx context.Context, diags *diag.Diagnostics, strs []string) types.List {
	if len(strs) == 0 {
		return types.ListNull(types.StringType)
	}
	l, d := types.ListValueFrom(ctx, types.StringType, strs)
	diags.Append(d...)
	return l
}

// int64ListOrNull builds a types.List of Int64 from an []int; empty/nil -> null.
func int64ListOrNull(ctx context.Context, diags *diag.Diagnostics, ints []int) types.List {
	if len(ints) == 0 {
		return types.ListNull(types.Int64Type)
	}
	vals := make([]int64, len(ints))
	for i, v := range ints {
		vals[i] = int64(v)
	}
	l, d := types.ListValueFrom(ctx, types.Int64Type, vals)
	diags.Append(d...)
	return l
}

// int64ListToInts reads an Int64 types.List back into an []int for the API model.
func int64ListToInts(ctx context.Context, diags *diag.Diagnostics, list types.List) []int {
	if !ut.IsDefined(list) || len(list.Elements()) == 0 {
		return nil
	}
	var vals []int64
	diags.Append(list.ElementsAs(ctx, &vals, false)...)
	out := make([]int, len(vals))
	for i, v := range vals {
		out[i] = int(v)
	}
	return out
}

func orEmptyStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// listFromSet converts a types.Set to a types.List for reuse of ListElementsAs.
func listFromSet(ctx context.Context, set types.Set, diags *diag.Diagnostics) types.List {
	if !ut.IsDefined(set) {
		return types.ListNull(types.StringType)
	}
	var vals []string
	diags.Append(set.ElementsAs(ctx, &vals, false)...)
	l, d := types.ListValueFrom(ctx, types.StringType, vals)
	diags.Append(d...)
	return l
}
