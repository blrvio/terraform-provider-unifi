package contentfiltering

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

func strSet(t *testing.T, values ...string) types.Set {
	t.Helper()
	set, diags := types.SetValueFrom(context.Background(), types.StringType, values)
	require.False(t, diags.HasError(), "SetValueFrom diagnostics: %v", diags)
	return set
}

func macSet(t *testing.T, values ...string) types.Set {
	t.Helper()
	set, diags := types.SetValueFrom(context.Background(), ut.MACType{}, values)
	require.False(t, diags.HasError(), "SetValueFrom diagnostics: %v", diags)
	return set
}

func scheduleObject(t *testing.T, sm scheduleModel) types.Object {
	t.Helper()
	obj, diags := types.ObjectValueFrom(context.Background(), sm.AttributeTypes(), sm)
	require.False(t, diags.HasError(), "ObjectValueFrom diagnostics: %v", diags)
	return obj
}

func TestContentFilteringModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := ContentFilteringModel{
		Name:       types.StringValue("kids"),
		Enabled:    types.BoolValue(true),
		AllowList:  strSet(t, "example.com"),
		BlockList:  strSet(t, "bad.example"),
		Categories: strSet(t, "FAMILY"),
		ClientMACs: macSet(t, "AA:BB:CC:DD:EE:FF"),
		NetworkIDs: strSet(t, "net-1"),
		SafeSearch: strSet(t, "GOOGLE"),
		Schedule: scheduleObject(t, scheduleModel{
			Mode:           types.StringValue("EVERY_WEEK"),
			Date:           types.StringNull(),
			DateStart:      types.StringNull(),
			DateEnd:        types.StringNull(),
			RepeatOnDays:   strSet(t, "mon", "tue"),
			TimeAllDay:     types.BoolValue(false),
			TimeRangeStart: types.StringValue("08:00"),
			TimeRangeEnd:   types.StringValue("20:00"),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.ContentFiltering)
	require.True(t, ok, "Expected model to be *unifi.ContentFiltering")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "kids", typed.Name)
	assert.True(t, typed.Enabled)
	assert.Equal(t, []string{"example.com"}, typed.AllowList)
	assert.Equal(t, []string{"bad.example"}, typed.BlockList)
	assert.Equal(t, []string{"FAMILY"}, typed.Categories)
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, typed.ClientMACs, "MACs are normalized")
	assert.Equal(t, []string{"net-1"}, typed.NetworkIDs)
	assert.Equal(t, []string{"GOOGLE"}, typed.SafeSearch)
	assert.Equal(t, "EVERY_WEEK", typed.Schedule.Mode)
	assert.ElementsMatch(t, []string{"mon", "tue"}, typed.Schedule.RepeatOnDays)
	assert.Equal(t, "08:00", typed.Schedule.TimeRangeStart)
	assert.Equal(t, "20:00", typed.Schedule.TimeRangeEnd)
}

func TestContentFilteringModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.ContentFiltering{
		ID:         "merge-id",
		Name:       "kids",
		Enabled:    true,
		AllowList:  []string{"example.com"},
		BlockList:  []string{"bad.example"},
		Categories: []string{"FAMILY"},
		ClientMACs: []string{"aa:bb:cc:dd:ee:ff"},
		NetworkIDs: []string{"net-1"},
		SafeSearch: []string{"GOOGLE"},
		Schedule: unifi.ContentFilteringSchedule{
			Mode:           "EVERY_WEEK",
			RepeatOnDays:   []string{"mon", "tue"},
			TimeRangeStart: "08:00",
			TimeRangeEnd:   "20:00",
		},
	}

	var d ContentFilteringModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "kids", d.Name.ValueString())
	assert.True(t, d.Enabled.ValueBool())
	assert.False(t, d.Schedule.IsNull())

	var sm scheduleModel
	d.Schedule.As(context.Background(), &sm, basetypes.ObjectAsOptions{})
	assert.Equal(t, "EVERY_WEEK", sm.Mode.ValueString())
	assert.Equal(t, "08:00", sm.TimeRangeStart.ValueString())

	var macs []string
	d.ClientMACs.ElementsAs(context.Background(), &macs, false)
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, macs)
}

func TestContentFilteringModel_Merge_ZeroScheduleIsNull(t *testing.T) {
	t.Parallel()

	var d ContentFilteringModel
	diags := d.Merge(context.Background(), &unifi.ContentFiltering{ID: "id", Name: "n"})
	assert.False(t, diags.HasError())
	assert.True(t, d.Schedule.IsNull(), "an unset schedule round-trips to a null object")
	// Empty collections round-trip to known empty sets (matching schema defaults).
	assert.False(t, d.AllowList.IsNull())
	assert.Equal(t, 0, len(d.AllowList.Elements()))
}

func TestContentFilteringModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d ContentFilteringModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
