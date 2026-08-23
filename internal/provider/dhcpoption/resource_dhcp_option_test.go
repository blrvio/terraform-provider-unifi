package dhcpoption

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDHCPOptionModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := DHCPOptionModel{
		Code:   types.StringValue("114"),
		Name:   types.StringValue("captive-portal"),
		Signed: types.BoolValue(false),
		Type:   types.StringValue("text"),
		Width:  types.Int32Value(32),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.DHCPOption)
	require.True(t, ok, "Expected model to be *unifi.DHCPOption")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "114", typed.Code)
	assert.Equal(t, "captive-portal", typed.Name)
	assert.False(t, typed.Signed)
	assert.Equal(t, "text", typed.Type)
	assert.Equal(t, 32, typed.Width)
}

func TestDHCPOptionModel_AsUnifiModel_NullWidth(t *testing.T) {
	t.Parallel()

	model := DHCPOptionModel{
		Code:   types.StringValue("114"),
		Name:   types.StringValue("opt"),
		Signed: types.BoolValue(false),
		Type:   types.StringValue("text"),
		Width:  types.Int32Null(),
	}

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())
	typed, ok := unifiModel.(*unifi.DHCPOption)
	require.True(t, ok, "Expected model to be *unifi.DHCPOption")
	assert.Equal(t, 0, typed.Width, "null width must become the omitempty zero value")
}

func TestDHCPOptionModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.DHCPOption{
		ID:     "merge-id",
		Code:   "114",
		Name:   "captive-portal",
		Signed: true,
		Type:   "integer",
		Width:  16,
	}

	var d DHCPOptionModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "114", d.Code.ValueString())
	assert.Equal(t, "captive-portal", d.Name.ValueString())
	assert.True(t, d.Signed.ValueBool())
	assert.Equal(t, "integer", d.Type.ValueString())
	assert.Equal(t, int32(16), d.Width.ValueInt32())
}

func TestDHCPOptionModel_Merge_ZeroWidthIsNull(t *testing.T) {
	t.Parallel()

	var d DHCPOptionModel
	diags := d.Merge(context.Background(), &unifi.DHCPOption{ID: "id", Code: "1", Name: "n", Type: "text"})
	assert.False(t, diags.HasError())
	assert.True(t, d.Width.IsNull(), "a zero width round-trips to null")
}

func TestDHCPOptionModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d DHCPOptionModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
