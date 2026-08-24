package qos

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

var (
	_ resource.Resource                = &qosRuleResource{}
	_ resource.ResourceWithConfigure   = &qosRuleResource{}
	_ resource.ResourceWithImportState = &qosRuleResource{}
	_ base.Resource                    = &qosRuleResource{}
)

type qosRuleResource struct {
	*base.GenericResource[*qosRuleModel]
}

// NewQOSRuleResource creates the unifi_qos_rule resource.
func NewQOSRuleResource() resource.Resource {
	return &qosRuleResource{
		GenericResource: base.NewGenericResource(
			"unifi_qos_rule",
			func() *qosRuleModel { return &qosRuleModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetQOSRule(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					rule, ok := model.(*unifi.QOSRule)
					if !ok {
						return nil, fmt.Errorf("unexpected model type %T, expected *unifi.QOSRule", model)
					}
					return client.CreateQOSRule(ctx, site, rule)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					rule, ok := model.(*unifi.QOSRule)
					if !ok {
						return nil, fmt.Errorf("unexpected model type %T, expected *unifi.QOSRule", model)
					}
					return client.UpdateQOSRule(ctx, site, rule)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteQOSRule(ctx, site, id)
				},
			},
		),
	}
}

func matchingListAttribute(desc string) schema.ListAttribute {
	return schema.ListAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		ElementType:         types.StringType,
	}
}

func (r *qosRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_qos_rule` resource manages a UniFi QoS rule (Quality of Service). " +
			"QoS rules prioritize or rate-limit matched traffic on a WAN/VPN network — for example the " +
			"built-in *Critical Apps Prioritization* rule.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the QoS rule.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the rule is enabled. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "The evaluation order of the rule. Lower is evaluated first.",
				Optional:            true,
				Computed:            true,
			},
			"objective": schema.StringAttribute{
				MarkdownDescription: "The QoS objective for matched traffic, e.g. `PRIORITIZE`.",
				Optional:            true,
				Computed:            true,
			},
			"download_burst": schema.StringAttribute{
				MarkdownDescription: "Whether download burst is enabled. One of `ON`, `OFF`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.OneOf("ON", "OFF")},
			},
			"upload_burst": schema.StringAttribute{
				MarkdownDescription: "Whether upload burst is enabled. One of `ON`, `OFF`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.OneOf("ON", "OFF")},
			},
			"wan_or_vpn_network": schema.StringAttribute{
				MarkdownDescription: "The id of the WAN or VPN network the rule applies to.",
				Optional:            true,
			},
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "An optional schedule constraining when the rule is active.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						MarkdownDescription: "The schedule mode. One of `ALWAYS`, `EVERY_DAY`, `EVERY_WEEK`, `ONE_TIME_ONLY`, `CUSTOM`.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("ALWAYS", "EVERY_DAY", "EVERY_WEEK", "ONE_TIME_ONLY", "CUSTOM"),
						},
					},
					"date":             schema.StringAttribute{MarkdownDescription: "A single date (`YYYY-MM-DD`) for one-time schedules.", Optional: true},
					"date_start":       schema.StringAttribute{MarkdownDescription: "The start date (`YYYY-MM-DD`) of the schedule window.", Optional: true},
					"date_end":         schema.StringAttribute{MarkdownDescription: "The end date (`YYYY-MM-DD`) of the schedule window.", Optional: true},
					"repeat_on_days":   schema.SetAttribute{MarkdownDescription: "Days of the week the schedule repeats on (`mon`..`sun`).", Optional: true, ElementType: types.StringType},
					"time_all_day":     schema.BoolAttribute{MarkdownDescription: "Whether the rule is active the entire day.", Optional: true},
					"time_range_start": schema.StringAttribute{MarkdownDescription: "The start time (`HH:MM`) of the active window.", Optional: true},
					"time_range_end":   schema.StringAttribute{MarkdownDescription: "The end time (`HH:MM`) of the active window.", Optional: true},
				},
			},
			"source": schema.SingleNestedAttribute{
				MarkdownDescription: "The traffic source matched by the rule.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"matching_target": schema.StringAttribute{
						MarkdownDescription: "What the source matches on. One of `ANY`, `CLIENT`, `NETWORK`, `IP`, `MAC`, `REGION`.",
						Optional:            true,
						Computed:            true,
						Validators:          []validator.String{stringvalidator.OneOf("ANY", "CLIENT", "NETWORK", "IP", "MAC", "REGION")},
					},
					"port_matching_type": schema.StringAttribute{
						MarkdownDescription: "How ports are matched. One of `ANY`, `SPECIFIC`, `OBJECT`.",
						Optional:            true,
						Validators:          []validator.String{stringvalidator.OneOf("ANY", "SPECIFIC", "OBJECT")},
					},
					"network_ids": matchingListAttribute("Network ids matched when `matching_target` is `NETWORK`."),
					"client_macs": matchingListAttribute("Client MAC addresses matched when `matching_target` is `CLIENT` or `MAC`."),
					"ips":         matchingListAttribute("IP addresses matched when `matching_target` is `IP`."),
					"regions":     matchingListAttribute("Regions matched when `matching_target` is `REGION`."),
				},
			},
			"destination": schema.SingleNestedAttribute{
				MarkdownDescription: "The traffic destination matched by the rule.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"matching_target": schema.StringAttribute{
						MarkdownDescription: "What the destination matches on. One of `ANY`, `APP`, `APP_CATEGORY`, `IP`, `NETWORK`, `REGION`, `WEB`.",
						Optional:            true,
						Computed:            true,
						Validators:          []validator.String{stringvalidator.OneOf("ANY", "APP", "APP_CATEGORY", "IP", "NETWORK", "REGION", "WEB")},
					},
					"port_matching_type": schema.StringAttribute{
						MarkdownDescription: "How ports are matched. One of `ANY`, `SPECIFIC`, `OBJECT`.",
						Optional:            true,
						Validators:          []validator.String{stringvalidator.OneOf("ANY", "SPECIFIC", "OBJECT")},
					},
					"app_ids": schema.ListAttribute{
						MarkdownDescription: "Application ids matched when `matching_target` is `APP`.",
						Optional:            true,
						ElementType:         types.Int64Type,
					},
					"app_category_ids": schema.ListAttribute{
						MarkdownDescription: "Application category ids matched when `matching_target` is `APP_CATEGORY`.",
						Optional:            true,
						ElementType:         types.Int64Type,
					},
					"network_ids": matchingListAttribute("Network ids matched when `matching_target` is `NETWORK`."),
					"ips":         matchingListAttribute("IP addresses matched when `matching_target` is `IP`."),
					"regions":     matchingListAttribute("Regions matched when `matching_target` is `REGION`."),
				},
			},
		},
	}
}
