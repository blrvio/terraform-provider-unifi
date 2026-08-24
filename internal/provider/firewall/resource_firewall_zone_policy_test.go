package firewall

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFirewallZonePolicyIndexReadOnly is the primary, controller-free regression
// lock for issue #122. The `index` attribute MUST be Computed-only (no Optional,
// no Default) so the Plugin Framework marks it unknown on update; re-adding a
// default (as in v1.0.0, which set StaticInt64(10000)) reintroduces the
// "Provider produced inconsistent result after apply: .index" error when the
// controller renumbers policies. This test fails immediately if that regresses.
func TestFirewallZonePolicyIndexReadOnly(t *testing.T) {
	var resp resource.SchemaResponse
	NewFirewallZonePolicyResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "schema build returned diagnostics: %v", resp.Diagnostics)

	attr, ok := resp.Schema.Attributes["index"]
	require.True(t, ok, "expected an `index` attribute in the schema")

	indexAttr, ok := attr.(schema.Int64Attribute)
	require.True(t, ok, "expected `index` to be an Int64Attribute, got %T", attr)

	assert.True(t, indexAttr.Computed, "`index` must be Computed")
	assert.False(t, indexAttr.Optional, "`index` must not be Optional (it is controller-assigned)")
	assert.False(t, indexAttr.Required, "`index` must not be Required")
	assert.Nil(t, indexAttr.Default, "`index` must not declare a Default; a default suppresses unknown-marking on update and reintroduces issue #122")
}

// TestFirewallZonePolicyMergeConnectionStatesCustom is a controller-free
// regression lock for the "inconsistent result after apply" on
// connection_states. The controller returns connection_state_type in UPPERCASE
// ("CUSTOM"), but Merge previously gated the connection_states copy-back on the
// lowercase literal "custom", so a CUSTOM policy's states were nulled on read —
// the planned ["NEW"] never matched the applied null and the plan never
// converged. Merge must repopulate connection_states whenever the type is CUSTOM
// (case-insensitively).
func TestFirewallZonePolicyMergeConnectionStatesCustom(t *testing.T) {
	m := &FirewallZonePolicyModel{}
	diags := m.Merge(context.Background(), &unifi.FirewallZonePolicy{
		ConnectionStateType: "CUSTOM",
		ConnectionStates:    []string{"NEW"},
	})
	require.False(t, diags.HasError(), "Merge returned diagnostics: %v", diags)

	require.False(t, m.ConnectionStates.IsNull(), "connection_states must be populated when connection_state_type is CUSTOM")
	require.Len(t, m.ConnectionStates.Elements(), 1, "expected exactly one connection state")
	assert.Equal(t, `"NEW"`, m.ConnectionStates.Elements()[0].String())
}

func TestNewFirewallPolicyTargetModelPortParsing(t *testing.T) {
	t.Run("NumericPortParses", func(t *testing.T) {
		m := NewFirewallPolicyTargetModel(t.Context(), "", nil, nil, false, false, false, "8443", "", "zone1")
		assert.False(t, m.Port.IsNull())
		assert.Equal(t, int32(8443), m.Port.ValueInt32())
	})

	t.Run("EmptyPortIsNull", func(t *testing.T) {
		m := NewFirewallPolicyTargetModel(t.Context(), "", nil, nil, false, false, false, "", "", "zone1")
		assert.True(t, m.Port.IsNull())
	})

	t.Run("RangeValueIsNull", func(t *testing.T) {
		// The controller accepts ranges/lists ("8000-8010,9443") but the
		// schema attribute is an Int32 — unrepresentable values read as null.
		m := NewFirewallPolicyTargetModel(t.Context(), "", nil, nil, false, false, false, "8000-8010", "", "zone1")
		assert.True(t, m.Port.IsNull())
	})
}
