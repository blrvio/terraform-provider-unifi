package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperMailModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := superMailModel{
		Provider: types.StringValue("smtp"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperMail)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperMail")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "smtp", typed.Provider)

	// Null provider is omitted.
	empty := superMailModel{Provider: types.StringNull()}
	empty.ID = types.StringValue("id")
	unifiModel, diags = empty.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())
	mailTyped, ok := unifiModel.(*unifi.SettingSuperMail)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperMail")
	assert.Equal(t, "", mailTyped.Provider)
}

func TestSuperMailModel_Merge(t *testing.T) {
	t.Parallel()

	source := &unifi.SettingSuperMail{
		ID:       "test-id",
		Provider: "cloud",
	}

	model := superMailModel{}
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "cloud", model.Provider.ValueString())

	// Empty string round-trips to null.
	empty := superMailModel{}
	diags = empty.Merge(context.Background(), &unifi.SettingSuperMail{ID: "id"})
	assert.False(t, diags.HasError())
	assert.True(t, empty.Provider.IsNull())
}
