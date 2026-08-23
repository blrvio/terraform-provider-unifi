package dashboard

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func moduleList(t *testing.T, modules ...dashboardModuleModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), dashboardModuleObjectType(), modules)
	require.False(t, diags.HasError(), "ListValueFrom diagnostics: %v", diags)
	return list
}

func TestDashboardModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := DashboardModel{
		Name:              types.StringValue("ops"),
		Desc:              types.StringValue("operations"),
		ControllerVersion: types.StringValue("10.5.67"),
		IsPublic:          types.BoolValue(true),
		Modules: moduleList(t, dashboardModuleModel{
			ID:           types.StringValue("m1"),
			ModuleID:     types.StringValue("traffic"),
			Config:       types.StringValue("{}"),
			Restrictions: types.StringValue(""),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.Dashboard)
	require.True(t, ok, "Expected model to be *unifi.Dashboard")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "ops", typed.Name)
	assert.Equal(t, "operations", typed.Desc)
	assert.Equal(t, "10.5.67", typed.ControllerVersion)
	assert.True(t, typed.IsPublic)
	require.Len(t, typed.Modules, 1)
	assert.Equal(t, "m1", typed.Modules[0].ID)
	assert.Equal(t, "traffic", typed.Modules[0].ModuleID)
}

func TestDashboardModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.Dashboard{
		ID:       "merge-id",
		Name:     "ops",
		IsPublic: true,
		Modules: []unifi.DashboardModules{
			{ID: "m1", ModuleID: "traffic", Config: "{}"},
		},
	}

	var d DashboardModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "ops", d.Name.ValueString())
	assert.True(t, d.IsPublic.ValueBool())

	var modules []dashboardModuleModel
	d.Modules.ElementsAs(context.Background(), &modules, false)
	require.Len(t, modules, 1)
	assert.Equal(t, "m1", modules[0].ID.ValueString())
	assert.Equal(t, "traffic", modules[0].ModuleID.ValueString())
}

func TestDashboardModel_Merge_EmptyModules(t *testing.T) {
	t.Parallel()

	var d DashboardModel
	diags := d.Merge(context.Background(), &unifi.Dashboard{ID: "id", Name: "empty"})
	assert.False(t, diags.HasError())
	assert.False(t, d.Modules.IsNull())
	assert.Equal(t, 0, len(d.Modules.Elements()))
}

func TestDashboardModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d DashboardModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
