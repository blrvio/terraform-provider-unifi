package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// evaluationScoreModel represents the evaluation score settings for a UniFi site,
// tracking which evaluation score recommendations have been dismissed.
type evaluationScoreModel struct {
	base.Model
	DismissedIDs types.List `tfsdk:"dismissed_ids"`
}

func (d *evaluationScoreModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingEvaluationScore{
		ID: d.ID.ValueString(),
	}

	if !d.DismissedIDs.IsNull() {
		var dismissed []string
		diags.Append(ut.ListElementsAs(ctx, d.DismissedIDs, &dismissed)...)
		if diags.HasError() {
			return nil, diags
		}
		model.DismissedIDs = dismissed
	}

	return model, diags
}

func (d *evaluationScoreModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingEvaluationScore)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingEvaluationScore")
		return diags
	}

	d.ID = types.StringValue(model.ID)

	if len(model.DismissedIDs) > 0 {
		list, ld := types.ListValueFrom(ctx, types.StringType, model.DismissedIDs)
		diags.Append(ld...)
		if diags.HasError() {
			return diags
		}
		d.DismissedIDs = list
	} else {
		d.DismissedIDs = ut.EmptyList(types.StringType)
	}

	return diags
}

var (
	_ base.ResourceModel               = &evaluationScoreModel{}
	_ resource.Resource                = &evaluationScoreResource{}
	_ resource.ResourceWithConfigure   = &evaluationScoreResource{}
	_ resource.ResourceWithImportState = &evaluationScoreResource{}
)

type evaluationScoreResource struct {
	*base.GenericResource[*evaluationScoreModel]
}

func (r *evaluationScoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_evaluation_score` resource manages the evaluation score settings for a UniFi " +
			"site, tracking which evaluation score recommendations have been dismissed.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"dismissed_ids": schema.ListAttribute{
				MarkdownDescription: "List of evaluation score recommendation IDs that have been dismissed.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// NewEvaluationScoreResource creates a new instance of the evaluation score setting resource.
func NewEvaluationScoreResource() resource.Resource {
	r := &evaluationScoreResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_evaluation_score",
		func() *evaluationScoreModel { return &evaluationScoreModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingEvaluationScore(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingEvaluationScore)
			return client.UpdateSettingEvaluationScore(ctx, site, b)
		},
	)
	return r
}
