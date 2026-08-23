package acl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// asBody runs a model through AsUnifiModel and returns the concrete body.
func asBody(t *testing.T, m *aclRuleModel) *official.ACLRuleUpdate {
	t.Helper()
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
	b, ok := body.(*official.ACLRuleUpdate)
	require.True(t, ok, "expected *official.ACLRuleUpdate, got %T", body)
	return b
}

func TestACLRuleAsUnifiModel_SourceFilterRoundTrip(t *testing.T) {
	t.Parallel()
	m := &aclRuleModel{
		Name:         types.StringValue("block-guests"),
		Action:       types.StringValue(aclRuleActionBlock),
		Type:         types.StringValue(aclRuleTypeIPv4),
		SourceFilter: types.StringValue(`{"type":"NETWORK","networkId":"abc"}`),
		// Enabled left null -> should default to true.
	}
	body := asBody(t, m)

	assert.Equal(t, official.ACLRuleAction("BLOCK"), body.Action)
	assert.Equal(t, "block-guests", body.Name)
	assert.Equal(t, "IPV4", body.Type)
	assert.True(t, body.Enabled, "unset enabled must default to true")

	// SourceFilter must round-trip through the JSON string into a decoded value.
	require.NotNil(t, body.SourceFilter)
	raw, err := json.Marshal(body.SourceFilter)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"NETWORK","networkId":"abc"}`, string(raw))

	// Unset filters stay nil (omitted from the request).
	assert.Nil(t, body.DestinationFilter)
	assert.Nil(t, body.EnforcingDeviceFilter)
	// Deprecated Index must never be sent.
	assert.Nil(t, body.Index)
}

func TestACLRuleAsUnifiModel_Disabled(t *testing.T) {
	t.Parallel()
	m := &aclRuleModel{
		Name:    types.StringValue("allow-mgmt"),
		Action:  types.StringValue(aclRuleActionAllow),
		Type:    types.StringValue(aclRuleTypeMAC),
		Enabled: types.BoolValue(false),
	}
	body := asBody(t, m)
	assert.Equal(t, official.ACLRuleAction("ALLOW"), body.Action)
	assert.False(t, body.Enabled)
	assert.Equal(t, "MAC", body.Type)
}

func TestACLRuleAsUnifiModel_InvalidJSON(t *testing.T) {
	t.Parallel()
	m := &aclRuleModel{
		Name:         types.StringValue("bad"),
		Action:       types.StringValue(aclRuleActionBlock),
		Type:         types.StringValue(aclRuleTypeIPv4),
		SourceFilter: types.StringValue(`{not json`),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError())
}

func TestACLRuleAsUnifiModel_MarshalOverlay(t *testing.T) {
	t.Parallel()
	// The union MarshalJSON overlays top-level fields; verify enabled survives.
	m := &aclRuleModel{
		Name:   types.StringValue("r"),
		Action: types.StringValue(aclRuleActionAllow),
		Type:   types.StringValue(aclRuleTypeIPv4),
	}
	body := asBody(t, m)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, true, out["enabled"], "enabled must survive the top-level MarshalJSON overlay")
	assert.Equal(t, "ALLOW", out["action"])
	assert.Equal(t, "IPV4", out["type"])
}

func TestACLRuleMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":"11111111-2222-3333-4444-555555555555","name":"block-guests","action":"BLOCK",` +
		`"type":"IPV4","enabled":true,"index":3,"description":"no guests",` +
		`"sourceFilter":{"type":"NETWORK","networkId":"abc"},"metadata":{}}`
	var rule official.ACLRule
	require.NoError(t, json.Unmarshal([]byte(raw), &rule))

	m := &aclRuleModel{}
	diags := m.Merge(context.Background(), &rule)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "11111111-2222-3333-4444-555555555555", m.GetID())
	assert.Equal(t, "block-guests", m.Name.ValueString())
	assert.Equal(t, "BLOCK", m.Action.ValueString())
	assert.Equal(t, "IPV4", m.Type.ValueString())
	assert.True(t, m.Enabled.ValueBool())
	assert.Equal(t, int32(3), m.Index.ValueInt32())
	assert.Equal(t, "no guests", m.Description.ValueString())
	assert.JSONEq(t, `{"type":"NETWORK","networkId":"abc"}`, m.SourceFilter.ValueString())
	// Unset filters merge to null.
	assert.True(t, m.DestinationFilter.IsNull())
	assert.True(t, m.EnforcingDeviceFilter.IsNull())
}

func TestACLRuleMerge_NoDescription(t *testing.T) {
	t.Parallel()
	raw := `{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","name":"r","action":"ALLOW",` +
		`"type":"MAC","enabled":false,"index":0,"metadata":{}}`
	var rule official.ACLRule
	require.NoError(t, json.Unmarshal([]byte(raw), &rule))

	m := &aclRuleModel{}
	diags := m.Merge(context.Background(), &rule)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "MAC", m.Type.ValueString())
	assert.False(t, m.Enabled.ValueBool())
	assert.True(t, m.Description.IsNull())
	assert.True(t, m.SourceFilter.IsNull())
}

func TestRuleIDsUUIDRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	want := []string{
		"11111111-2222-3333-4444-555555555555",
		"66666666-7777-8888-9999-aaaaaaaaaaaa",
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, want)
	require.False(t, diags.HasError())

	ids, diags := ruleIDsToUUIDs(ctx, list)
	require.False(t, diags.HasError(), "ruleIDsToUUIDs diagnostics: %v", diags)
	require.Len(t, ids, 2)
	assert.Equal(t, want[0], ids[0].String())
	assert.Equal(t, want[1], ids[1].String())

	// Convert back to a Framework list.
	uuids := make([]uuid.UUID, len(ids))
	copy(uuids, ids)
	back, diags := uuidsToRuleIDs(ctx, uuids)
	require.False(t, diags.HasError())
	var got []string
	require.False(t, back.ElementsAs(ctx, &got, false).HasError())
	assert.Equal(t, want, got)
}

func TestRuleIDsToUUIDs_Invalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	list, diags := types.ListValueFrom(ctx, types.StringType, []string{"not-a-uuid"})
	require.False(t, diags.HasError())
	_, diags = ruleIDsToUUIDs(ctx, list)
	assert.True(t, diags.HasError())
}

func TestOrderID(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "default", orderID("default"))
	assert.Equal(t, "ordering", orderID(""))
}
