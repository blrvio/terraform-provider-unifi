package scheduletask

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

func macSet(t *testing.T, macs ...string) types.Set {
	t.Helper()
	set, diags := types.SetValueFrom(context.Background(), ut.MACType{}, macs)
	require.False(t, diags.HasError(), "SetValueFrom diagnostics: %v", diags)
	return set
}

func TestScheduleTaskModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := ScheduleTaskModel{
		Action:          types.StringValue("upgrade"),
		CronExpr:        types.StringValue("0 4 * * 0"),
		ExecuteOnlyOnce: types.BoolValue(true),
		Name:            types.StringValue("weekly-upgrade"),
		UpgradeTargets:  macSet(t, "AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.ScheduleTask)
	require.True(t, ok, "Expected model to be *unifi.ScheduleTask")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "upgrade", typed.Action)
	assert.Equal(t, "0 4 * * 0", typed.CronExpr)
	assert.True(t, typed.ExecuteOnlyOnce)
	assert.Equal(t, "weekly-upgrade", typed.Name)

	macs := make([]string, 0, len(typed.UpgradeTargets))
	for _, tt := range typed.UpgradeTargets {
		macs = append(macs, tt.MAC)
	}
	// AsUnifiModel normalizes MACs to the controller's canonical form.
	assert.ElementsMatch(t, []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"}, macs)
}

func TestScheduleTaskModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.ScheduleTask{
		ID:              "merge-id",
		Action:          "upgrade",
		CronExpr:        "0 4 * * 0",
		ExecuteOnlyOnce: false,
		Name:            "weekly-upgrade",
		UpgradeTargets: []unifi.ScheduleTaskUpgradeTargets{
			{MAC: "aa:bb:cc:dd:ee:ff"},
		},
	}

	var d ScheduleTaskModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "upgrade", d.Action.ValueString())
	assert.Equal(t, "0 4 * * 0", d.CronExpr.ValueString())
	assert.False(t, d.ExecuteOnlyOnce.ValueBool())
	assert.Equal(t, "weekly-upgrade", d.Name.ValueString())

	var macs []string
	d.UpgradeTargets.ElementsAs(context.Background(), &macs, false)
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, macs)
}

func TestScheduleTaskModel_Merge_EmptyTargets(t *testing.T) {
	t.Parallel()

	var d ScheduleTaskModel
	diags := d.Merge(context.Background(), &unifi.ScheduleTask{ID: "id", Name: "n", Action: "upgrade", CronExpr: "0 0 * * *"})
	assert.False(t, diags.HasError())
	assert.False(t, d.UpgradeTargets.IsNull())
	assert.Equal(t, 0, len(d.UpgradeTargets.Elements()))
}

func TestScheduleTaskModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d ScheduleTaskModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
