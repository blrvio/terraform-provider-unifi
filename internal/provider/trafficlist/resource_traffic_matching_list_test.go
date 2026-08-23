package trafficlist

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// itemsList builds a Framework list of item objects from the given item models.
func itemsList(t *testing.T, items ...tmlItemModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), tmlItemObjectType(), items)
	require.False(t, diags.HasError(), "ListValueFrom diagnostics: %v", diags)
	return list
}

// marshalBody runs a model through AsUnifiModel and marshals the resulting
// Official-API body to JSON, exercising the union From*/MarshalJSON path.
func marshalBody(t *testing.T, m *trafficMatchingListModel) map[string]any {
	t.Helper()
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestTMLAsUnifiModel_IPv4(t *testing.T) {
	t.Parallel()
	m := &trafficMatchingListModel{
		Type: types.StringValue(tmlTypeIPv4),
		Name: types.StringValue("blocklist"),
		Items: itemsList(t,
			tmlItemModel{
				MatchType: types.StringValue(matchIPAddress),
				Value:     types.StringValue("10.0.0.5"),
			},
			tmlItemModel{
				MatchType: types.StringValue(matchSubnet),
				Value:     types.StringValue("192.168.1.0/24"),
			},
			tmlItemModel{
				MatchType: types.StringValue(matchIPAddressRange),
				Start:     types.StringValue("10.1.0.1"),
				Stop:      types.StringValue("10.1.0.100"),
			},
		),
	}
	out := marshalBody(t, m)
	assert.Equal(t, "IPV4_ADDRESSES", out["type"])
	assert.Equal(t, "blocklist", out["name"], "name must survive the top-level MarshalJSON overlay")

	items, ok := out["items"].([]any)
	require.True(t, ok, "items should be a JSON array")
	require.Len(t, items, 3)

	first, ok := items[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "IP_ADDRESS", first["type"])
	assert.Equal(t, "10.0.0.5", first["value"])

	second, ok := items[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "SUBNET", second["type"])
	assert.Equal(t, "192.168.1.0/24", second["value"])

	third, ok := items[2].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "IP_ADDRESS_RANGE", third["type"])
	assert.Equal(t, "10.1.0.1", third["start"])
	assert.Equal(t, "10.1.0.100", third["stop"])
}

func TestTMLAsUnifiModel_Ports(t *testing.T) {
	t.Parallel()
	m := &trafficMatchingListModel{
		Type: types.StringValue(tmlTypePort),
		Name: types.StringValue("web-ports"),
		Items: itemsList(t,
			tmlItemModel{
				MatchType: types.StringValue(matchPortNumber),
				Value:     types.StringValue("443"),
			},
			tmlItemModel{
				MatchType: types.StringValue(matchPortRange),
				Start:     types.StringValue("8000"),
				Stop:      types.StringValue("8100"),
			},
		),
	}
	out := marshalBody(t, m)
	assert.Equal(t, "PORTS", out["type"])
	assert.Equal(t, "web-ports", out["name"])

	items := asSlice(t, out["items"])
	require.Len(t, items, 2)

	first := asMap(t, items[0])
	assert.Equal(t, "PORT_NUMBER", first["type"])
	assert.EqualValues(t, 443, first["value"], "port value must be numeric on the wire")

	second := asMap(t, items[1])
	assert.Equal(t, "PORT_NUMBER_RANGE", second["type"])
	assert.EqualValues(t, 8000, second["start"])
	assert.EqualValues(t, 8100, second["stop"])
}

func TestTMLAsUnifiModel_IPv6(t *testing.T) {
	t.Parallel()
	m := &trafficMatchingListModel{
		Type: types.StringValue(tmlTypeIPv6),
		Name: types.StringValue("v6-list"),
		Items: itemsList(t,
			tmlItemModel{
				MatchType: types.StringValue(matchIPAddress),
				Value:     types.StringValue("2001:db8::1"),
			},
			tmlItemModel{
				MatchType: types.StringValue(matchSubnet),
				Value:     types.StringValue("2001:db8::/32"),
			},
		),
	}
	out := marshalBody(t, m)
	assert.Equal(t, "IPV6_ADDRESSES", out["type"])
	assert.Equal(t, "v6-list", out["name"])

	items := asSlice(t, out["items"])
	require.Len(t, items, 2)
	assert.Equal(t, "IP_ADDRESS", asMap(t, items[0])["type"])
	assert.Equal(t, "2001:db8::1", asMap(t, items[0])["value"])
	assert.Equal(t, "SUBNET", asMap(t, items[1])["type"])
	assert.Equal(t, "2001:db8::/32", asMap(t, items[1])["value"])
}

func TestTMLMerge_IPv4RoundTrip(t *testing.T) {
	t.Parallel()
	// Simulate a controller read response (top-level + union both populated).
	raw := `{"id":"11111111-2222-3333-4444-555555555555","type":"IPV4_ADDRESSES","name":"blocklist",` +
		`"items":[{"type":"IP_ADDRESS","value":"10.0.0.5"},{"type":"SUBNET","value":"192.168.1.0/24"}]}`
	var tml official.TrafficMatchingList
	require.NoError(t, json.Unmarshal([]byte(raw), &tml))

	m := &trafficMatchingListModel{}
	diags := m.Merge(context.Background(), &tml)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "11111111-2222-3333-4444-555555555555", m.GetID())
	assert.Equal(t, "IPV4_ADDRESSES", m.Type.ValueString())
	assert.Equal(t, "blocklist", m.Name.ValueString())

	var items []tmlItemModel
	diags = m.Items.ElementsAs(context.Background(), &items, false)
	require.False(t, diags.HasError(), "ElementsAs diagnostics: %v", diags)
	require.Len(t, items, 2)

	assert.Equal(t, "IP_ADDRESS", items[0].MatchType.ValueString())
	assert.Equal(t, "10.0.0.5", items[0].Value.ValueString())
	assert.True(t, items[0].Start.IsNull())

	assert.Equal(t, "SUBNET", items[1].MatchType.ValueString())
	assert.Equal(t, "192.168.1.0/24", items[1].Value.ValueString())
}

func TestTMLMerge_PortsRoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","type":"PORTS","name":"web-ports",` +
		`"items":[{"type":"PORT_NUMBER_RANGE","start":8000,"stop":8100}]}`
	var tml official.TrafficMatchingList
	require.NoError(t, json.Unmarshal([]byte(raw), &tml))

	m := &trafficMatchingListModel{}
	diags := m.Merge(context.Background(), &tml)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "PORTS", m.Type.ValueString())
	assert.Equal(t, "web-ports", m.Name.ValueString())

	var items []tmlItemModel
	diags = m.Items.ElementsAs(context.Background(), &items, false)
	require.False(t, diags.HasError(), "ElementsAs diagnostics: %v", diags)
	require.Len(t, items, 1)

	assert.Equal(t, "PORT_NUMBER_RANGE", items[0].MatchType.ValueString())
	// Ports come back numeric and are formatted to strings in the model.
	assert.Equal(t, "8000", items[0].Start.ValueString())
	assert.Equal(t, "8100", items[0].Stop.ValueString())
	assert.True(t, items[0].Value.IsNull())
}

func TestTMLPortsRoundTripThroughWire(t *testing.T) {
	t.Parallel()
	// Build a body, marshal it, unmarshal as a read response, Merge, and verify
	// the port value survives the string<->int32 conversion in both directions.
	m := &trafficMatchingListModel{
		Type: types.StringValue(tmlTypePort),
		Name: types.StringValue("dns"),
		Items: itemsList(t,
			tmlItemModel{MatchType: types.StringValue(matchPortNumber), Value: types.StringValue("53")},
		),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError())
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	// Add an id so the read model is complete.
	var obj map[string]any
	require.NoError(t, json.Unmarshal(raw, &obj))
	obj["id"] = "99999999-8888-7777-6666-555555555555"
	raw, err = json.Marshal(obj)
	require.NoError(t, err)

	var tml official.TrafficMatchingList
	require.NoError(t, json.Unmarshal(raw, &tml))

	out := &trafficMatchingListModel{}
	require.False(t, out.Merge(context.Background(), &tml).HasError())

	var items []tmlItemModel
	require.False(t, out.Items.ElementsAs(context.Background(), &items, false).HasError())
	require.Len(t, items, 1)
	assert.Equal(t, "PORT_NUMBER", items[0].MatchType.ValueString())
	assert.Equal(t, "53", items[0].Value.ValueString())
}

func TestTMLAsUnifiModel_UnknownType(t *testing.T) {
	t.Parallel()
	m := &trafficMatchingListModel{
		Type:  types.StringValue("BOGUS"),
		Name:  types.StringValue("x"),
		Items: itemsList(t),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError())
}

func asSlice(t *testing.T, v any) []any {
	t.Helper()
	s, ok := v.([]any)
	require.True(t, ok, "expected []any, got %T", v)
	return s
}

func asMap(t *testing.T, v any) map[string]any {
	t.Helper()
	m, ok := v.(map[string]any)
	require.True(t, ok, "expected map[string]any, got %T", v)
	return m
}
