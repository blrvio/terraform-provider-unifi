package wlangroup

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWLANGroupModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := WLANGroupModel{
		Name: types.StringValue("guest-group"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.WLANGroup)
	require.True(t, ok, "Expected model to be *unifi.WLANGroup")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "guest-group", typed.Name)
}

func TestWLANGroupModel_Merge(t *testing.T) {
	t.Parallel()

	var d WLANGroupModel
	diags := d.Merge(context.Background(), &unifi.WLANGroup{ID: "merge-id", Name: "guest-group"})
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "guest-group", d.Name.ValueString())
}

func TestWLANGroupModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d WLANGroupModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
