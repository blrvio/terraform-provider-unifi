package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestGlobalNatModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := globalNatModel{
		Mode: types.StringValue("custom"),
		ExcludedNetworkIDs: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("net-1"),
			types.StringValue("net-2"),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingGlobalNat)
	assert.True(t, ok, "Expected model to be *unifi.SettingGlobalNat")
	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "custom", typed.Mode)
	assert.Equal(t, []string{"net-1", "net-2"}, typed.ExcludedNetworkIDs)
}

func TestGlobalNatModel_Merge(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingGlobalNat{
		ID:                 "test-id",
		Mode:               "auto",
		ExcludedNetworkIDs: []string{"net-1"},
	}

	var model globalNatModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "auto", model.Mode.ValueString())

	var excluded []string
	model.ExcludedNetworkIDs.ElementsAs(context.Background(), &excluded, false)
	assert.Equal(t, []string{"net-1"}, excluded)
}

func TestGlobalNatModel_Merge_EmptyList(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingGlobalNat{ID: "test-id", Mode: "off"}

	var model globalNatModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.False(t, model.ExcludedNetworkIDs.IsNull())
	assert.Equal(t, 0, len(model.ExcludedNetworkIDs.Elements()))
	assert.Equal(t, "off", model.Mode.ValueString())
}
