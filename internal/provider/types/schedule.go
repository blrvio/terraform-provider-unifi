package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// ScheduleValues is the plain form of the v2 rule schedule shared by content
// filtering, app blocking and other v2 policies ({mode, date*, repeat_on_days,
// time_all_day, time_range_*}). Domains copy their SDK schedule struct to and from it.
type ScheduleValues struct {
	Mode           string
	Date           string
	DateStart      string
	DateEnd        string
	RepeatOnDays   []string
	TimeAllDay     bool
	TimeRangeStart string
	TimeRangeEnd   string
}

// IsZero reports whether the schedule carries no configuration.
func (s ScheduleValues) IsZero() bool {
	return s.Mode == "" && s.Date == "" && s.DateStart == "" && s.DateEnd == "" &&
		s.TimeRangeStart == "" && s.TimeRangeEnd == "" && !s.TimeAllDay && len(s.RepeatOnDays) == 0
}

// ScheduleModel is the Framework model of the single nested `schedule` attribute.
type ScheduleModel struct {
	Mode           types.String `tfsdk:"mode"`
	Date           types.String `tfsdk:"date"`
	DateStart      types.String `tfsdk:"date_start"`
	DateEnd        types.String `tfsdk:"date_end"`
	RepeatOnDays   types.Set    `tfsdk:"repeat_on_days"`
	TimeAllDay     types.Bool   `tfsdk:"time_all_day"`
	TimeRangeStart types.String `tfsdk:"time_range_start"`
	TimeRangeEnd   types.String `tfsdk:"time_range_end"`
}

func (m ScheduleModel) AttributeTypes() map[string]attr.Type {
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

// ScheduleFromObject converts the `schedule` object to ScheduleValues; a null or
// unknown object yields the zero schedule.
func ScheduleFromObject(ctx context.Context, obj types.Object) (ScheduleValues, diag.Diagnostics) {
	var diags diag.Diagnostics
	if !IsDefined(obj) {
		return ScheduleValues{}, diags
	}
	var sm ScheduleModel
	diags.Append(obj.As(ctx, &sm, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return ScheduleValues{}, diags
	}
	days := []string{}
	if IsDefined(sm.RepeatOnDays) {
		diags.Append(sm.RepeatOnDays.ElementsAs(ctx, &days, false)...)
	}
	return ScheduleValues{
		Mode:           sm.Mode.ValueString(),
		Date:           sm.Date.ValueString(),
		DateStart:      sm.DateStart.ValueString(),
		DateEnd:        sm.DateEnd.ValueString(),
		RepeatOnDays:   days,
		TimeAllDay:     sm.TimeAllDay.ValueBool(),
		TimeRangeStart: sm.TimeRangeStart.ValueString(),
		TimeRangeEnd:   sm.TimeRangeEnd.ValueString(),
	}, diags
}

// ScheduleToObject converts ScheduleValues to the `schedule` object; the zero
// schedule maps to a null object so an unset schedule round-trips.
func ScheduleToObject(ctx context.Context, s ScheduleValues) (types.Object, diag.Diagnostics) {
	if s.IsZero() {
		return ObjectNull(&ScheduleModel{})
	}
	days := s.RepeatOnDays
	if days == nil {
		days = []string{}
	}
	var diags diag.Diagnostics
	repeat, d := types.SetValueFrom(ctx, types.StringType, days)
	diags.Append(d...)
	sm := ScheduleModel{
		Mode:           StringOrNull(s.Mode),
		Date:           StringOrNull(s.Date),
		DateStart:      StringOrNull(s.DateStart),
		DateEnd:        StringOrNull(s.DateEnd),
		RepeatOnDays:   repeat,
		TimeAllDay:     types.BoolValue(s.TimeAllDay),
		TimeRangeStart: StringOrNull(s.TimeRangeStart),
		TimeRangeEnd:   StringOrNull(s.TimeRangeEnd),
	}
	obj, d := types.ObjectValueFrom(ctx, sm.AttributeTypes(), sm)
	diags.Append(d...)
	return obj, diags
}

// ScheduleAttribute returns the optional nested `schedule` attribute.
func ScheduleAttribute(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: description,
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				MarkdownDescription: "The schedule mode. One of `ALWAYS`, `EVERY_DAY`, `EVERY_WEEK`, `ONE_TIME_ONLY`, `CUSTOM`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALWAYS", "EVERY_DAY", "EVERY_WEEK", "ONE_TIME_ONLY", "CUSTOM"),
				},
			},
			"date": schema.StringAttribute{
				MarkdownDescription: "A single date (`YYYY-MM-DD`) for one-time schedules.",
				Optional:            true,
			},
			"date_start": schema.StringAttribute{
				MarkdownDescription: "The start date (`YYYY-MM-DD`) of the schedule window.",
				Optional:            true,
			},
			"date_end": schema.StringAttribute{
				MarkdownDescription: "The end date (`YYYY-MM-DD`) of the schedule window.",
				Optional:            true,
			},
			"repeat_on_days": schema.SetAttribute{
				MarkdownDescription: "Days of the week the schedule repeats on. Each of `mon`, `tue`, `wed`, `thu`, `fri`, `sat`, `sun`.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf("mon", "tue", "wed", "thu", "fri", "sat", "sun")),
				},
			},
			"time_all_day": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is active for the entire day.",
				Optional:            true,
			},
			"time_range_start": schema.StringAttribute{
				MarkdownDescription: "The start time (`HH:MM`) of the active window.",
				Optional:            true,
			},
			"time_range_end": schema.StringAttribute{
				MarkdownDescription: "The end time (`HH:MM`) of the active window.",
				Optional:            true,
			},
		},
	}
}
