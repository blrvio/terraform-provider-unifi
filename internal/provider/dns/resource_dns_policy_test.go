package dns

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// marshalBody runs a model through AsUnifiModel and marshals the resulting
// Official-API body to JSON, exercising the union From*/MarshalJSON path.
func marshalBody(t *testing.T, m *dnsPolicyModel) map[string]any {
	t.Helper()
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestDNSPolicyAsUnifiModel_ARecord(t *testing.T) {
	t.Parallel()
	m := &dnsPolicyModel{
		Type:        types.StringValue(dnsPolicyTypeA),
		Domain:      types.StringValue("foo.example.com"),
		IPv4Address: types.StringValue("10.0.0.5"),
		TTLSeconds:  types.Int32Value(300),
		// Enabled left null -> should default to true.
	}
	out := marshalBody(t, m)
	assert.Equal(t, "A_RECORD", out["type"])
	assert.Equal(t, true, out["enabled"], "enabled must survive the top-level MarshalJSON overlay")
	assert.Equal(t, "foo.example.com", out["domain"])
	assert.Equal(t, "10.0.0.5", out["ipv4Address"])
	assert.EqualValues(t, 300, out["ttlSeconds"])
}

func TestDNSPolicyAsUnifiModel_DisabledMxRecord(t *testing.T) {
	t.Parallel()
	m := &dnsPolicyModel{
		Type:             types.StringValue(dnsPolicyTypeMX),
		Domain:           types.StringValue("example.com"),
		MailServerDomain: types.StringValue("mail.example.com"),
		Priority:         types.Int32Value(10),
		Enabled:          types.BoolValue(false),
	}
	out := marshalBody(t, m)
	assert.Equal(t, "MX_RECORD", out["type"])
	assert.Equal(t, false, out["enabled"])
	assert.Equal(t, "mail.example.com", out["mailServerDomain"])
	assert.EqualValues(t, 10, out["priority"])
}

func TestDNSPolicyMerge_ARecordRoundTrip(t *testing.T) {
	t.Parallel()
	// Simulate a controller read response (top-level + union both populated).
	raw := `{"id":"11111111-2222-3333-4444-555555555555","type":"A_RECORD","enabled":true,` +
		`"domain":"foo.example.com","ipv4Address":"10.0.0.5","ttlSeconds":300}`
	var policy official.DNSPolicy
	require.NoError(t, json.Unmarshal([]byte(raw), &policy))

	m := &dnsPolicyModel{}
	diags := m.Merge(context.Background(), &policy)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "11111111-2222-3333-4444-555555555555", m.GetID())
	assert.Equal(t, "A_RECORD", m.Type.ValueString())
	assert.True(t, m.Enabled.ValueBool())
	assert.Equal(t, "foo.example.com", m.Domain.ValueString())
	assert.Equal(t, "10.0.0.5", m.IPv4Address.ValueString())
	assert.Equal(t, int32(300), m.TTLSeconds.ValueInt32())
	// Fields belonging to other variants stay null.
	assert.True(t, m.Text.IsNull())
	assert.True(t, m.MailServerDomain.IsNull())
}

func TestDNSPolicyMerge_ForwardDomain(t *testing.T) {
	t.Parallel()
	raw := `{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","type":"FORWARD_DOMAIN","enabled":false,` +
		`"domain":"internal.example.com","ipAddress":"192.168.1.53"}`
	var policy official.DNSPolicy
	require.NoError(t, json.Unmarshal([]byte(raw), &policy))

	m := &dnsPolicyModel{}
	diags := m.Merge(context.Background(), &policy)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "FORWARD_DOMAIN", m.Type.ValueString())
	assert.False(t, m.Enabled.ValueBool())
	assert.Equal(t, "internal.example.com", m.Domain.ValueString())
	assert.Equal(t, "192.168.1.53", m.IPAddress.ValueString())
}

func TestDNSPolicyAsUnifiModel_UnknownType(t *testing.T) {
	t.Parallel()
	m := &dnsPolicyModel{Type: types.StringValue("BOGUS")}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError())
}
