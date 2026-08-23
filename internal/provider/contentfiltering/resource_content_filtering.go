package contentfiltering

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// scheduleModel is the single nested `schedule` block of a content filtering rule.
type scheduleModel struct {
	Mode           types.String `tfsdk:"mode"`
	Date           types.String `tfsdk:"date"`
	DateStart      types.String `tfsdk:"date_start"`
	DateEnd        types.String `tfsdk:"date_end"`
	RepeatOnDays   types.Set    `tfsdk:"repeat_on_days"`
	TimeAllDay     types.Bool   `tfsdk:"time_all_day"`
	TimeRangeStart types.String `tfsdk:"time_range_start"`
	TimeRangeEnd   types.String `tfsdk:"time_range_end"`
}

func (m scheduleModel) AttributeTypes() map[string]attr.Type {
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

// ContentFilteringModel represents the data model for a UniFi content filtering rule.
type ContentFilteringModel struct {
	base.Model
	Name       types.String `tfsdk:"name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	AllowList  types.Set    `tfsdk:"allow_list"`
	BlockList  types.Set    `tfsdk:"block_list"`
	Categories types.Set    `tfsdk:"categories"`
	ClientMACs types.Set    `tfsdk:"client_macs"`
	NetworkIDs types.Set    `tfsdk:"network_ids"`
	SafeSearch types.Set    `tfsdk:"safe_search"`
	Schedule   types.Object `tfsdk:"schedule"`
}

func stringSlice(ctx context.Context, set types.Set) ([]string, diag.Diagnostics) {
	var out []string
	diags := diag.Diagnostics{}
	if ut.IsDefined(set) {
		diags.Append(set.ElementsAs(ctx, &out, false)...)
	}
	if out == nil {
		out = []string{}
	}
	return out, diags
}

func stringSetOrEmpty(ctx context.Context, elemType attr.Type, values []string) (types.Set, diag.Diagnostics) {
	if values == nil {
		values = []string{}
	}
	return types.SetValueFrom(ctx, elemType, values)
}

// isZeroSchedule reports whether the API schedule carries no configuration, so an
// unset schedule round-trips to a null object rather than an empty-but-present one.
func isZeroSchedule(s unifi.ContentFilteringSchedule) bool {
	return s.Mode == "" && s.Date == "" && s.DateStart == "" && s.DateEnd == "" &&
		s.TimeRangeStart == "" && s.TimeRangeEnd == "" && !s.TimeAllDay && len(s.RepeatOnDays) == 0
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *ContentFilteringModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	allowList, d := stringSlice(ctx, m.AllowList)
	diags.Append(d...)
	blockList, d := stringSlice(ctx, m.BlockList)
	diags.Append(d...)
	categories, d := stringSlice(ctx, m.Categories)
	diags.Append(d...)
	clientMACs, d := stringSlice(ctx, m.ClientMACs)
	diags.Append(d...)
	// Normalize to the controller's canonical MAC form defensively; MACType
	// already provides semantic equality, this just keeps the persisted value tidy.
	for i, mac := range clientMACs {
		clientMACs[i] = utils.CleanMAC(mac)
	}
	networkIDs, d := stringSlice(ctx, m.NetworkIDs)
	diags.Append(d...)
	safeSearch, d := stringSlice(ctx, m.SafeSearch)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	var schedule unifi.ContentFilteringSchedule
	if ut.IsDefined(m.Schedule) {
		var sm scheduleModel
		diags.Append(m.Schedule.As(ctx, &sm, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}
		repeatOnDays, rd := stringSlice(ctx, sm.RepeatOnDays)
		diags.Append(rd...)
		if diags.HasError() {
			return nil, diags
		}
		schedule = unifi.ContentFilteringSchedule{
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

	return &unifi.ContentFiltering{
		ID:         m.ID.ValueString(),
		Name:       m.Name.ValueString(),
		Enabled:    m.Enabled.ValueBool(),
		AllowList:  allowList,
		BlockList:  blockList,
		Categories: categories,
		ClientMACs: clientMACs,
		NetworkIDs: networkIDs,
		SafeSearch: safeSearch,
		Schedule:   schedule,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *ContentFilteringModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.ContentFiltering)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.ContentFiltering, got %T", other))
		return diags
	}

	diags := diag.Diagnostics{}
	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Enabled = types.BoolValue(model.Enabled)

	var d diag.Diagnostics
	m.AllowList, d = stringSetOrEmpty(ctx, types.StringType, model.AllowList)
	diags.Append(d...)
	m.BlockList, d = stringSetOrEmpty(ctx, types.StringType, model.BlockList)
	diags.Append(d...)
	m.Categories, d = stringSetOrEmpty(ctx, types.StringType, model.Categories)
	diags.Append(d...)
	m.ClientMACs, d = stringSetOrEmpty(ctx, ut.MACType{}, model.ClientMACs)
	diags.Append(d...)
	m.NetworkIDs, d = stringSetOrEmpty(ctx, types.StringType, model.NetworkIDs)
	diags.Append(d...)
	m.SafeSearch, d = stringSetOrEmpty(ctx, types.StringType, model.SafeSearch)
	diags.Append(d...)

	if isZeroSchedule(model.Schedule) {
		m.Schedule, d = ut.ObjectNull(&scheduleModel{})
		diags.Append(d...)
	} else {
		repeatOnDays, rd := stringSetOrEmpty(ctx, types.StringType, model.Schedule.RepeatOnDays)
		diags.Append(rd...)
		sm := scheduleModel{
			Mode:           ut.StringOrNull(model.Schedule.Mode),
			Date:           ut.StringOrNull(model.Schedule.Date),
			DateStart:      ut.StringOrNull(model.Schedule.DateStart),
			DateEnd:        ut.StringOrNull(model.Schedule.DateEnd),
			RepeatOnDays:   repeatOnDays,
			TimeAllDay:     types.BoolValue(model.Schedule.TimeAllDay),
			TimeRangeStart: ut.StringOrNull(model.Schedule.TimeRangeStart),
			TimeRangeEnd:   ut.StringOrNull(model.Schedule.TimeRangeEnd),
		}
		m.Schedule, d = types.ObjectValueFrom(ctx, sm.AttributeTypes(), sm)
		diags.Append(d...)
	}

	return diags
}

var (
	_ resource.Resource                = &contentFilteringResource{}
	_ resource.ResourceWithConfigure   = &contentFilteringResource{}
	_ resource.ResourceWithImportState = &contentFilteringResource{}
	_ base.Resource                    = &contentFilteringResource{}
	_ base.ResourceModel               = &ContentFilteringModel{}
)

type contentFilteringResource struct {
	*base.GenericResource[*ContentFilteringModel]
}

// NewContentFilteringResource creates a new instance of the content filtering resource.
func NewContentFilteringResource() resource.Resource {
	return &contentFilteringResource{
		GenericResource: base.NewGenericResource(
			"unifi_content_filtering",
			func() *ContentFilteringModel { return &ContentFilteringModel{} },
			base.ResourceFunctions{
				// The SDK does not expose a public GetContentFiltering; read by
				// listing and filtering on the id, reporting ErrNotFound so the
				// GenericResource surfaces a "not found".
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					list, err := client.ListContentFiltering(ctx, site)
					if err != nil {
						return nil, err
					}
					for i := range list {
						if list[i].ID == id {
							return &list[i], nil
						}
					}
					return nil, unifi.ErrNotFound
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					cf, _ := model.(*unifi.ContentFiltering)
					return client.CreateContentFiltering(ctx, site, cf)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					cf, _ := model.(*unifi.ContentFiltering)
					return client.UpdateContentFiltering(ctx, site, cf)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteContentFiltering(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *contentFilteringResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	emptyStringSet := setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{}))
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_content_filtering` resource manages a content filtering rule in the " +
			"UniFi controller.\n\n" +
			"A content filtering rule enforces DNS-based allow/block lists, content categories and safe-search " +
			"for a set of clients or networks, optionally on a schedule.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the content filtering rule.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"allow_list": schema.SetAttribute{
				MarkdownDescription: "Domains that are always allowed.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet,
			},
			"block_list": schema.SetAttribute{
				MarkdownDescription: "Domains that are always blocked.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet,
			},
			"categories": schema.SetAttribute{
				MarkdownDescription: "Content categories to filter. Each of `FAMILY`, `ADVERTISEMENT`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf("FAMILY", "ADVERTISEMENT")),
				},
			},
			"client_macs": schema.SetAttribute{
				MarkdownDescription: "MAC addresses of the clients the rule applies to. MAC addresses are " +
					"case-insensitive and may use `:` or `-` separators.",
				Optional:    true,
				Computed:    true,
				ElementType: ut.MACType{},
				Default:     setdefault.StaticValue(types.SetValueMust(ut.MACType{}, []attr.Value{})),
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(validators.Mac),
					validators.UniqueMACs(),
				},
			},
			"network_ids": schema.SetAttribute{
				MarkdownDescription: "IDs of the networks the rule applies to.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet,
			},
			"safe_search": schema.SetAttribute{
				MarkdownDescription: "Search providers to enforce safe search on. Each of `GOOGLE`, `YOUTUBE`, `BING`.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             emptyStringSet,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf("GOOGLE", "YOUTUBE", "BING")),
				},
			},
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "An optional schedule that constrains when the rule is active.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						MarkdownDescription: "The schedule mode. One of `ALWAYS`, `EVERY_DAY`, `EVERY_WEEK`, " +
							"`ONE_TIME_ONLY`, `CUSTOM`.",
						Required: true,
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
						MarkdownDescription: "Days of the week the schedule repeats on. Each of `mon`, `tue`, " +
							"`wed`, `thu`, `fri`, `sat`, `sun`.",
						Optional:    true,
						ElementType: types.StringType,
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
			},
		},
	}
}
