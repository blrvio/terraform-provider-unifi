package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestRadioAiModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	channelsNa, diags := types.ListValueFrom(ctx, types.Int32Type, []int32{36, 40})
	assert.False(t, diags.HasError())

	radios, diags := types.ListValueFrom(ctx, types.StringType, []string{"na", "ng"})
	assert.False(t, diags.HasError())

	optimize, diags := types.ListValueFrom(ctx, types.StringType, []string{"channel", "power"})
	assert.False(t, diags.HasError())

	blacklist, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&radioAiChannelsBlacklistModel{}).AttributeTypes()}, []radioAiChannelsBlacklistModel{
		{
			Channel:      types.Int32Value(149),
			ChannelWidth: types.Int32Value(80),
			Radio:        types.StringValue("na"),
		},
	})
	assert.False(t, diags.HasError())

	radiosConfig, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&radioAiRadiosConfigurationModel{}).AttributeTypes()}, []radioAiRadiosConfigurationModel{
		{
			ChannelWidth: types.Int32Value(40),
			Dfs:          types.BoolValue(true),
			Radio:        types.StringValue("ng"),
		},
	})
	assert.False(t, diags.HasError())

	model := radioAiModel{
		AutoAdjustChannelsToCountry: types.BoolValue(true),
		AutoChannelPresetsType:      types.StringValue("conservative"),
		ChannelsNa:                  channelsNa,
		CronExpr:                    types.StringValue("0 3 * * *"),
		Default:                     types.BoolValue(false),
		Enabled:                     types.BoolValue(true),
		Optimize:                    optimize,
		Radios:                      radios,
		RadiosConfiguration:         radiosConfig,
		ChannelsBlacklist:           blacklist,
		SettingPreference:           types.StringValue("manual"),
		UseXy:                       types.BoolValue(true),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(ctx)
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingRadioAi)
	assert.True(t, ok, "Expected model to be *unifi.SettingRadioAi")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.AutoAdjustChannelsToCountry)
	assert.Equal(t, "conservative", typed.AutoChannelPresetsType)
	assert.Equal(t, []int{36, 40}, typed.ChannelsNa)
	assert.Equal(t, "0 3 * * *", typed.CronExpr)
	assert.True(t, typed.Enabled)
	assert.Equal(t, []string{"channel", "power"}, typed.Optimize)
	assert.Equal(t, []string{"na", "ng"}, typed.Radios)
	assert.Equal(t, "manual", typed.SettingPreference)
	assert.True(t, typed.UseXy)

	assert.Len(t, typed.ChannelsBlacklist, 1)
	assert.Equal(t, 149, typed.ChannelsBlacklist[0].Channel)
	assert.Equal(t, 80, typed.ChannelsBlacklist[0].ChannelWidth)
	assert.Equal(t, "na", typed.ChannelsBlacklist[0].Radio)

	assert.Len(t, typed.RadiosConfiguration, 1)
	assert.Equal(t, 40, typed.RadiosConfiguration[0].ChannelWidth)
	assert.True(t, typed.RadiosConfiguration[0].Dfs)
	assert.Equal(t, "ng", typed.RadiosConfiguration[0].Radio)
}

func TestRadioAiModel_Merge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	source := &unifi.SettingRadioAi{
		ID:                          "test-id",
		AutoAdjustChannelsToCountry: true,
		AutoChannelPresetsType:      "custom",
		ChannelsNa:                  []int{36, 149},
		CronExpr:                    "0 4 * * *",
		Default:                     true,
		Enabled:                     true,
		ExcludeDevices:              []string{"aa:bb:cc:dd:ee:ff"},
		Optimize:                    []string{"channel"},
		Radios:                      []string{"na"},
		SettingPreference:           "manual",
		UseXy:                       true,
		ChannelsBlacklist: []unifi.SettingRadioAiChannelsBlacklist{
			{Channel: 161, ChannelWidth: 160, Radio: "na"},
		},
		RadiosConfiguration: []unifi.SettingRadioAiRadiosConfiguration{
			{ChannelWidth: 80, Dfs: false, Radio: "na"},
		},
	}

	var model radioAiModel
	diags := model.Merge(ctx, source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.True(t, model.AutoAdjustChannelsToCountry.ValueBool())
	assert.Equal(t, "custom", model.AutoChannelPresetsType.ValueString())
	assert.Equal(t, "0 4 * * *", model.CronExpr.ValueString())
	assert.True(t, model.Default.ValueBool())
	assert.True(t, model.Enabled.ValueBool())
	assert.Equal(t, "manual", model.SettingPreference.ValueString())
	assert.True(t, model.UseXy.ValueBool())

	var channelsNa []int32
	diags = model.ChannelsNa.ElementsAs(ctx, &channelsNa, false)
	assert.False(t, diags.HasError())
	assert.Equal(t, []int32{36, 149}, channelsNa)

	var excludeDevices []string
	diags = model.ExcludeDevices.ElementsAs(ctx, &excludeDevices, false)
	assert.False(t, diags.HasError())
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, excludeDevices)

	var blacklist []radioAiChannelsBlacklistModel
	diags = model.ChannelsBlacklist.ElementsAs(ctx, &blacklist, false)
	assert.False(t, diags.HasError())
	assert.Len(t, blacklist, 1)
	assert.Equal(t, int32(161), blacklist[0].Channel.ValueInt32())
	assert.Equal(t, int32(160), blacklist[0].ChannelWidth.ValueInt32())
	assert.Equal(t, "na", blacklist[0].Radio.ValueString())

	var radiosConfig []radioAiRadiosConfigurationModel
	diags = model.RadiosConfiguration.ElementsAs(ctx, &radiosConfig, false)
	assert.False(t, diags.HasError())
	assert.Len(t, radiosConfig, 1)
	assert.Equal(t, int32(80), radiosConfig[0].ChannelWidth.ValueInt32())
	assert.False(t, radiosConfig[0].Dfs.ValueBool())
	assert.Equal(t, "na", radiosConfig[0].Radio.ValueString())
}
