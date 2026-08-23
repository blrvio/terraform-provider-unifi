package channelplan

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func radioList(t *testing.T, items ...radioTableItemModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), radioTableObjectType(), items)
	require.False(t, diags.HasError(), "ListValueFrom diagnostics: %v", diags)
	return list
}

func TestChannelPlanModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := ChannelPlanModel{
		Date: types.StringValue("2026-08-23T04:00:00Z"),
		RadioTable: radioList(t, radioTableItemModel{
			Channel:     types.StringValue("36"),
			DeviceMAC:   types.StringValue("AA:BB:CC:DD:EE:FF"),
			Name:        types.StringValue("rai0"),
			TxPower:     types.StringValue("auto"),
			TxPowerMode: types.StringValue("auto"),
			Width:       types.Int32Value(80),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.ChannelPlan)
	require.True(t, ok, "Expected model to be *unifi.ChannelPlan")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "2026-08-23T04:00:00Z", typed.Date)
	require.Len(t, typed.RadioTable, 1)
	assert.Equal(t, "36", typed.RadioTable[0].Channel)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", typed.RadioTable[0].DeviceMAC, "device_mac is normalized")
	assert.Equal(t, "rai0", typed.RadioTable[0].Name)
	assert.Equal(t, "auto", typed.RadioTable[0].TxPower)
	assert.Equal(t, "auto", typed.RadioTable[0].TxPowerMode)
	assert.Equal(t, 80, typed.RadioTable[0].Width)
}

func TestChannelPlanModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.ChannelPlan{
		ID:   "merge-id",
		Date: "2026-08-23T04:00:00Z",
		RadioTable: []unifi.ChannelPlanRadioTable{
			{
				Channel:     "36",
				DeviceMAC:   "aa:bb:cc:dd:ee:ff",
				Name:        "rai0",
				TxPower:     "auto",
				TxPowerMode: "auto",
				Width:       80,
			},
		},
	}

	var d ChannelPlanModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "2026-08-23T04:00:00Z", d.Date.ValueString())

	var items []radioTableItemModel
	d.RadioTable.ElementsAs(context.Background(), &items, false)
	require.Len(t, items, 1)
	assert.Equal(t, "36", items[0].Channel.ValueString())
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", items[0].DeviceMAC.ValueString())
	assert.Equal(t, "rai0", items[0].Name.ValueString())
	assert.Equal(t, int32(80), items[0].Width.ValueInt32())
}

func TestChannelPlanModel_Merge_EmptyRadioTable(t *testing.T) {
	t.Parallel()

	var d ChannelPlanModel
	diags := d.Merge(context.Background(), &unifi.ChannelPlan{ID: "id", Date: ""})
	assert.False(t, diags.HasError())
	assert.False(t, d.RadioTable.IsNull())
	assert.Equal(t, 0, len(d.RadioTable.Elements()))
}

func TestChannelPlanModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d ChannelPlanModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
