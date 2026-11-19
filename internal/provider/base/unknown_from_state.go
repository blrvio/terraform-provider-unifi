package base

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// planWithUnknownsFromState returns a copy of plan whose unknown values are replaced
// by the prior state value at the same path. At apply time the only unknowns left
// are Optional+Computed attributes the configuration does not set; without this a
// model's AsUnifiModel reads them as zero values and the update overwrites them on
// the controller. Paths absent from state stay unknown.
func planWithUnknownsFromState(plan tfsdk.Plan, state tfsdk.State) (tfsdk.Plan, diag.Diagnostics) {
	var diags diag.Diagnostics
	if state.Raw.IsNull() || !state.Raw.IsKnown() {
		return plan, diags
	}
	filled, err := tftypes.Transform(plan.Raw, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if v.IsKnown() {
			return v, nil
		}
		prior, _, err := tftypes.WalkAttributePath(state.Raw, p)
		if err != nil {
			return v, nil //nolint:nilerr // not in state: leave unknown
		}
		if pv, ok := prior.(tftypes.Value); ok && pv.Type().Equal(v.Type()) {
			return pv, nil
		}
		return v, nil
	})
	if err != nil {
		diags.AddError("Error resolving unknown plan values from state", err.Error())
		return plan, diags
	}
	return tfsdk.Plan{Schema: plan.Schema, Raw: filled}, diags
}

// getPlanResolvingUnknowns decodes the plan into target after filling its unknown
// values from state (see planWithUnknownsFromState).
func getPlanResolvingUnknowns(ctx context.Context, plan tfsdk.Plan, state tfsdk.State, target any) diag.Diagnostics {
	filled, diags := planWithUnknownsFromState(plan, state)
	if diags.HasError() {
		return diags
	}
	diags.Append(filled.Get(ctx, target)...)
	return diags
}
