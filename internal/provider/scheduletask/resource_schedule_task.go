package scheduletask

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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// ScheduleTaskModel represents the data model for a UniFi scheduled task.
//
// The API's upgrade_targets is a list of {mac} objects; it is modeled here as a
// flat set of MAC addresses (an isomorphic, friendlier representation) using the
// semantic-equality MACType so differently-formatted-but-equal MACs do not
// produce a perpetual diff.
type ScheduleTaskModel struct {
	base.Model
	Action          types.String `tfsdk:"action"`
	CronExpr        types.String `tfsdk:"cron_expr"`
	ExecuteOnlyOnce types.Bool   `tfsdk:"execute_only_once"`
	Name            types.String `tfsdk:"name"`
	UpgradeTargets  types.Set    `tfsdk:"upgrade_targets"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *ScheduleTaskModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var macs []string
	if ut.IsDefined(m.UpgradeTargets) {
		diags.Append(m.UpgradeTargets.ElementsAs(ctx, &macs, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	targets := make([]unifi.ScheduleTaskUpgradeTargets, 0, len(macs))
	for _, mac := range macs {
		targets = append(targets, unifi.ScheduleTaskUpgradeTargets{MAC: utils.CleanMAC(mac)})
	}

	return &unifi.ScheduleTask{
		ID:              m.ID.ValueString(),
		Action:          m.Action.ValueString(),
		CronExpr:        m.CronExpr.ValueString(),
		ExecuteOnlyOnce: m.ExecuteOnlyOnce.ValueBool(),
		Name:            m.Name.ValueString(),
		UpgradeTargets:  targets,
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *ScheduleTaskModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.ScheduleTask)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.ScheduleTask, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Action = types.StringValue(model.Action)
	m.CronExpr = types.StringValue(model.CronExpr)
	m.ExecuteOnlyOnce = types.BoolValue(model.ExecuteOnlyOnce)
	m.Name = types.StringValue(model.Name)

	macs := make([]string, 0, len(model.UpgradeTargets))
	for _, t := range model.UpgradeTargets {
		macs = append(macs, t.MAC)
	}
	set, diags := types.SetValueFrom(ctx, ut.MACType{}, macs)
	m.UpgradeTargets = set
	return diags
}

var (
	_ resource.Resource                = &scheduleTaskResource{}
	_ resource.ResourceWithConfigure   = &scheduleTaskResource{}
	_ resource.ResourceWithImportState = &scheduleTaskResource{}
	_ base.Resource                    = &scheduleTaskResource{}
	_ base.ResourceModel               = &ScheduleTaskModel{}
)

type scheduleTaskResource struct {
	*base.GenericResource[*ScheduleTaskModel]
}

// NewScheduleTaskResource creates a new instance of the schedule task resource.
func NewScheduleTaskResource() resource.Resource {
	return &scheduleTaskResource{
		GenericResource: base.NewGenericResource(
			"unifi_schedule_task",
			func() *ScheduleTaskModel { return &ScheduleTaskModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetScheduleTask(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					task, _ := model.(*unifi.ScheduleTask)
					return client.CreateScheduleTask(ctx, site, task)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					task, _ := model.(*unifi.ScheduleTask)
					return client.UpdateScheduleTask(ctx, site, task)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteScheduleTask(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *scheduleTaskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_schedule_task` resource manages a scheduled task in the UniFi " +
			"controller.\n\n" +
			"Currently the controller supports scheduled firmware upgrades: a named, cron-scheduled `upgrade` " +
			"action applied to a set of target devices.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the scheduled task.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The action to perform on the schedule. Currently only `upgrade` " +
					"(firmware upgrade) is supported. Defaults to `upgrade`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("upgrade"),
				Validators: []validator.String{
					stringvalidator.OneOf("upgrade"),
				},
			},
			"cron_expr": schema.StringAttribute{
				MarkdownDescription: "The cron expression that determines when the task runs " +
					"(for example `0 4 * * 0` for 04:00 every Sunday).",
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"execute_only_once": schema.BoolAttribute{
				MarkdownDescription: "Whether the task runs only once and is then removed. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"upgrade_targets": schema.SetAttribute{
				MarkdownDescription: "The set of device MAC addresses the task targets. MAC addresses are " +
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
		},
	}
}
