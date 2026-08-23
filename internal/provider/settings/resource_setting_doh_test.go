package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestDohModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	serverNames, diags := types.ListValueFrom(ctx, types.StringType, []string{"cloudflare", "google"})
	assert.False(t, diags.HasError())

	customServers, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: (&dohCustomServerModel{}).AttributeTypes()}, []dohCustomServerModel{
		{
			Enabled:    types.BoolValue(true),
			SdnsStamp:  types.StringValue("sdns://AgcAAAAAAAAAAAA"),
			ServerName: types.StringValue("my-resolver"),
		},
	})
	assert.False(t, diags.HasError())

	model := dohModel{
		State:         types.StringValue("custom"),
		ServerNames:   serverNames,
		CustomServers: customServers,
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(ctx)
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingDoh)
	assert.True(t, ok, "Expected model to be *unifi.SettingDoh")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "custom", typed.State)
	assert.Equal(t, []string{"cloudflare", "google"}, typed.ServerNames)

	assert.Len(t, typed.CustomServers, 1)
	assert.True(t, typed.CustomServers[0].Enabled)
	assert.Equal(t, "sdns://AgcAAAAAAAAAAAA", typed.CustomServers[0].SdnsStamp)
	assert.Equal(t, "my-resolver", typed.CustomServers[0].ServerName)
}

func TestDohModel_Merge(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	source := &unifi.SettingDoh{
		ID:          "test-id",
		State:       "custom",
		ServerNames: []string{"cloudflare"},
		CustomServers: []unifi.SettingDohCustomServers{
			{Enabled: true, SdnsStamp: "sdns://AgcAAAAAAAAAAAA", ServerName: "my-resolver"},
		},
	}

	var model dohModel
	diags := model.Merge(ctx, source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "custom", model.State.ValueString())

	var serverNames []string
	diags = model.ServerNames.ElementsAs(ctx, &serverNames, false)
	assert.False(t, diags.HasError())
	assert.Equal(t, []string{"cloudflare"}, serverNames)

	var customServers []dohCustomServerModel
	diags = model.CustomServers.ElementsAs(ctx, &customServers, false)
	assert.False(t, diags.HasError())
	assert.Len(t, customServers, 1)
	assert.True(t, customServers[0].Enabled.ValueBool())
	assert.Equal(t, "sdns://AgcAAAAAAAAAAAA", customServers[0].SdnsStamp.ValueString())
	assert.Equal(t, "my-resolver", customServers[0].ServerName.ValueString())
}
