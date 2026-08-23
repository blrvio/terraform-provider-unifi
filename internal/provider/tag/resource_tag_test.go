package tag

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func memberSet(t *testing.T, members ...string) types.Set {
	t.Helper()
	set, diags := types.SetValueFrom(context.Background(), types.StringType, members)
	require.False(t, diags.HasError(), "SetValueFrom diagnostics: %v", diags)
	return set
}

func TestTagModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := TagModel{
		Name:        types.StringValue("iot-devices"),
		MemberTable: memberSet(t, "aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.Tag)
	require.True(t, ok, "Expected model to be *unifi.Tag")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "iot-devices", typed.Name)
	assert.ElementsMatch(t, []string{"aa:bb:cc:dd:ee:ff", "11:22:33:44:55:66"}, typed.MemberTable)
}

func TestTagModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.Tag{
		ID:          "merge-id",
		Name:        "iot-devices",
		MemberTable: []string{"aa:bb:cc:dd:ee:ff"},
	}

	var d TagModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "iot-devices", d.Name.ValueString())

	var members []string
	d.MemberTable.ElementsAs(context.Background(), &members, false)
	assert.Equal(t, []string{"aa:bb:cc:dd:ee:ff"}, members)
}

func TestTagModel_Merge_EmptyMembers(t *testing.T) {
	t.Parallel()

	// A tag with no members must round-trip to a known, empty set (not null),
	// matching the schema default so there is no perpetual diff.
	var d TagModel
	diags := d.Merge(context.Background(), &unifi.Tag{ID: "id", Name: "empty"})
	assert.False(t, diags.HasError())
	assert.False(t, d.MemberTable.IsNull())
	assert.Equal(t, 0, len(d.MemberTable.Elements()))
}

func TestTagModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d TagModel
	diags := d.Merge(context.Background(), &unifi.BroadcastGroup{})
	assert.True(t, diags.HasError())
}
