package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestSuperIdentityModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := superIdentityModel{
		Hostname: types.StringValue("udm-pro"),
		Name:     types.StringValue("Head Office"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingSuperIdentity)
	assert.True(t, ok, "Expected model to be *unifi.SettingSuperIdentity")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "udm-pro", typed.Hostname)
	assert.Equal(t, "Head Office", typed.Name)
}

func TestSuperIdentityModel_Merge(t *testing.T) {
	t.Parallel()

	source := &unifi.SettingSuperIdentity{
		ID:       "test-id",
		Hostname: "udm-pro",
		Name:     "Head Office",
	}

	model := superIdentityModel{}
	diags := model.Merge(context.Background(), source)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "udm-pro", model.Hostname.ValueString())
	assert.Equal(t, "Head Office", model.Name.ValueString())

	// Empty strings round-trip to null.
	empty := superIdentityModel{}
	diags = empty.Merge(context.Background(), &unifi.SettingSuperIdentity{ID: "id"})
	assert.False(t, diags.HasError())
	assert.True(t, empty.Hostname.IsNull())
	assert.True(t, empty.Name.IsNull())
}
