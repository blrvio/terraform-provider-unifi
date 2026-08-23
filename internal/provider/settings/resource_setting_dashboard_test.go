package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDashboardModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	widgets, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&dashboardWidgetModel{}).AttributeTypes()}, []dashboardWidgetModel{
		{
			Enabled: types.BoolValue(true),
			Name:    types.StringValue("wan_activity"),
		},
	})
	assert.False(t, diags.HasError())

	model := dashboardModel{
		LayoutPreference: types.StringValue("manual"),
		Widgets:          widgets,
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(ctx)
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingDashboard)
	assert.True(t, ok, "Expected model to be *unifi.SettingDashboard")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "manual", typed.LayoutPreference)

	assert.Len(t, typed.Widgets, 1)
	assert.True(t, typed.Widgets[0].Enabled)
	assert.Equal(t, "wan_activity", typed.Widgets[0].Name)
}

func TestDashboardModel_Merge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	source := &unifi.SettingDashboard{
		ID:               "test-id",
		LayoutPreference: "manual",
		Widgets: []unifi.SettingDashboardWidgets{
			{Enabled: false, Name: "most_active_apps"},
		},
	}

	var model dashboardModel
	diags := model.Merge(ctx, source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "manual", model.LayoutPreference.ValueString())

	var widgets []dashboardWidgetModel
	diags = model.Widgets.ElementsAs(ctx, &widgets, false)
	assert.False(t, diags.HasError())
	assert.Len(t, widgets, 1)
	assert.False(t, widgets[0].Enabled.ValueBool())
	assert.Equal(t, "most_active_apps", widgets[0].Name.ValueString())
}
