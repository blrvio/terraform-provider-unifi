package hotspotop

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotspotOperatorModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := HotspotOperatorModel{
		Name:      types.StringValue("front-desk"),
		Note:      types.StringValue("lobby operator"),
		XPassword: types.StringValue("s3cr3t"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.HotspotOp)
	require.True(t, ok, "Expected model to be *unifi.HotspotOp")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "front-desk", typed.Name)
	assert.Equal(t, "lobby operator", typed.Note)
	assert.Equal(t, "s3cr3t", typed.XPassword)
}

func TestHotspotOperatorModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.HotspotOp{
		ID:        "merge-id",
		Name:      "front-desk",
		Note:      "lobby operator",
		XPassword: "s3cr3t",
	}

	var d HotspotOperatorModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "front-desk", d.Name.ValueString())
	assert.Equal(t, "lobby operator", d.Note.ValueString())
	assert.Equal(t, "s3cr3t", d.XPassword.ValueString())
}

func TestHotspotOperatorModel_Merge_EmptyOptionals(t *testing.T) {
	t.Parallel()

	var d HotspotOperatorModel
	diags := d.Merge(context.Background(), &unifi.HotspotOp{ID: "id", Name: "front-desk"})
	assert.False(t, diags.HasError())
	assert.True(t, d.Note.IsNull())
	assert.True(t, d.XPassword.IsNull())
}

func TestHotspotOperatorModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d HotspotOperatorModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
