package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperEventsModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := superEventsModel{
		Ignored: types.StringValue("opaque-value"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperEvents)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperEvents")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "opaque-value", typed.Ignored)
}

func TestSuperEventsModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingSuperEvents{
		ID:      "merge-id",
		Ignored: "opaque-value",
	}

	var d superEventsModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "opaque-value", d.Ignored.ValueString())
}

func TestSuperEventsModel_Merge_Empty(t *testing.T) {
	t.Parallel()

	model := &unifi.SettingSuperEvents{
		ID:      "merge-id",
		Ignored: "",
	}

	var d superEventsModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	// Empty string round-trips to null.
	assert.True(t, d.Ignored.IsNull())
}
