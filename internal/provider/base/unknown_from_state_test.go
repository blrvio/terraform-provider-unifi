package base

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanWithUnknownsFromState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := schema.Schema{Attributes: map[string]schema.Attribute{
		"name":     schema.StringAttribute{Required: true},
		"computed": schema.StringAttribute{Optional: true, Computed: true},
		"nested": schema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: map[string]schema.Attribute{
			"flag": schema.BoolAttribute{Optional: true, Computed: true},
		}},
	}}
	typ := s.Type().TerraformType(ctx)
	nestedTyp := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"flag": tftypes.Bool}}
	plan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(typ, map[string]tftypes.Value{
		"name":     tftypes.NewValue(tftypes.String, "new"),
		"computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"nested":   tftypes.NewValue(nestedTyp, map[string]tftypes.Value{"flag": tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue)}),
	})}
	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(typ, map[string]tftypes.Value{
		"name":     tftypes.NewValue(tftypes.String, "old"),
		"computed": tftypes.NewValue(tftypes.String, "from-state"),
		"nested":   tftypes.NewValue(nestedTyp, map[string]tftypes.Value{"flag": tftypes.NewValue(tftypes.Bool, true)}),
	})}

	var got struct {
		Name     string `tfsdk:"name"`
		Computed string `tfsdk:"computed"`
		Nested   struct {
			Flag bool `tfsdk:"flag"`
		} `tfsdk:"nested"`
	}
	require.False(t, getPlanResolvingUnknowns(ctx, plan, state, &got).HasError())
	assert.Equal(t, "new", got.Name, "known plan values win")
	assert.Equal(t, "from-state", got.Computed, "unknown takes the prior state value")
	assert.True(t, got.Nested.Flag, "nested unknowns resolve too")

	// No prior state (create): unknowns stay unknown.
	created, diags := planWithUnknownsFromState(plan, tfsdk.State{Schema: s, Raw: tftypes.NewValue(typ, nil)})
	require.False(t, diags.HasError())
	assert.True(t, created.Raw.Equal(plan.Raw))
}
