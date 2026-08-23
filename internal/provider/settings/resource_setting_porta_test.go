package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestPortaModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := portaModel{
		Ugw3WAN2Enabled: types.BoolValue(true),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingPorta)
	assert.True(t, ok, "Expected model to be *unifi.SettingPorta")

	assert.Equal(t, "test-id", typed.ID)
	assert.True(t, typed.Ugw3WAN2Enabled)
}

func TestPortaModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingPorta{
		ID:              "merge-id",
		Ugw3WAN2Enabled: true,
	}

	var d portaModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.True(t, d.Ugw3WAN2Enabled.ValueBool())
}
