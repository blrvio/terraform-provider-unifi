package broadcastgroup

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

func TestBroadcastGroupModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := BroadcastGroupModel{
		Name:        types.StringValue("floor-1-aps"),
		MemberTable: memberSet(t, "ap-id-1", "ap-id-2"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.BroadcastGroup)
	require.True(t, ok, "Expected model to be *unifi.BroadcastGroup")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "floor-1-aps", typed.Name)
	assert.ElementsMatch(t, []string{"ap-id-1", "ap-id-2"}, typed.MemberTable)
}

func TestBroadcastGroupModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.BroadcastGroup{
		ID:          "merge-id",
		Name:        "floor-1-aps",
		MemberTable: []string{"ap-id-1"},
	}

	var d BroadcastGroupModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "floor-1-aps", d.Name.ValueString())

	var members []string
	d.MemberTable.ElementsAs(context.Background(), &members, false)
	assert.Equal(t, []string{"ap-id-1"}, members)
}

func TestBroadcastGroupModel_Merge_EmptyMembers(t *testing.T) {
	t.Parallel()

	var d BroadcastGroupModel
	diags := d.Merge(context.Background(), &unifi.BroadcastGroup{ID: "id", Name: "empty"})
	assert.False(t, diags.HasError())
	assert.False(t, d.MemberTable.IsNull())
	assert.Equal(t, 0, len(d.MemberTable.Elements()))
}

func TestBroadcastGroupModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d BroadcastGroupModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
