package hotspotpackage

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotspotPackageModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := HotspotPackageModel{
		Amount:                     types.Float64Value(9.99),
		Hours:                      types.Int32Value(24),
		Name:                       types.StringValue("day-pass"),
		Currency:                   types.StringValue("USD"),
		LimitOverwrite:             types.BoolValue(true),
		PaymentFieldsEmailEnabled:  types.BoolValue(true),
		CustomPaymentFieldsEnabled: types.BoolValue(false),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.HotspotPackage)
	require.True(t, ok, "Expected model to be *unifi.HotspotPackage")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, 9.99, typed.Amount)
	assert.Equal(t, 24, typed.Hours)
	assert.Equal(t, "day-pass", typed.Name)
	assert.Equal(t, "USD", typed.Currency)
	assert.True(t, typed.LimitOverwrite)
	assert.True(t, typed.PaymentFieldsEmailEnabled)
	assert.False(t, typed.CustomPaymentFieldsEnabled)
}

func TestHotspotPackageModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.HotspotPackage{
		ID:                        "merge-id",
		Amount:                    19.5,
		TrialReset:                12,
		Hours:                     48,
		LimitDown:                 1000,
		Name:                      "week-pass",
		Currency:                  "EUR",
		LimitOverwrite:            true,
		PaymentFieldsEmailEnabled: true,
	}

	var d HotspotPackageModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, 19.5, d.Amount.ValueFloat64())
	assert.Equal(t, 12.0, d.TrialReset.ValueFloat64())
	assert.Equal(t, int32(48), d.Hours.ValueInt32())
	assert.Equal(t, int32(1000), d.LimitDown.ValueInt32())
	assert.Equal(t, "week-pass", d.Name.ValueString())
	assert.Equal(t, "EUR", d.Currency.ValueString())
	assert.True(t, d.LimitOverwrite.ValueBool())
	assert.True(t, d.PaymentFieldsEmailEnabled.ValueBool())
}

func TestHotspotPackageModel_Merge_EmptyOptionals(t *testing.T) {
	t.Parallel()

	var d HotspotPackageModel
	diags := d.Merge(context.Background(), &unifi.HotspotPackage{ID: "id"})
	assert.False(t, diags.HasError())
	assert.True(t, d.Amount.IsNull())
	assert.True(t, d.Hours.IsNull())
	assert.True(t, d.Name.IsNull())
	assert.True(t, d.Currency.IsNull())
	assert.False(t, d.LimitOverwrite.ValueBool())
}

func TestHotspotPackageModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d HotspotPackageModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
