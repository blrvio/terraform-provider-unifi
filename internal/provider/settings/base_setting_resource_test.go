package settings

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

// fakeSettingsClient serves one raw setting document and records the PUT body.
type fakeSettingsClient struct {
	unifi.Client
	doc     map[string]any
	putKey  string
	putBody map[string]any
}

func (f *fakeSettingsClient) Get(_ context.Context, _ string, _ any, respBody any) error {
	raw, err := json.Marshal(map[string]any{"meta": map[string]any{"rc": "ok"}, "data": []any{f.doc}})
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, respBody)
}

func (f *fakeSettingsClient) GetSettingConnectivity(_ context.Context, _ string) (*unifi.SettingConnectivity, error) {
	raw, _ := json.Marshal(f.doc)
	var s unifi.SettingConnectivity
	return &s, json.Unmarshal(raw, &s)
}

func (f *fakeSettingsClient) SetSetting(_ context.Context, _ string, key string, body any) (any, error) {
	f.putKey = key
	raw, _ := json.Marshal(body)
	f.putBody = map[string]any{}
	_ = json.Unmarshal(raw, &f.putBody)
	var s unifi.SettingConnectivity
	return &s, json.Unmarshal(raw, &s)
}

// R1: a field the model does not manage, and one the SDK struct does not even know,
// survive an update; only the changed modeled field is written.
func TestSettingUpdateKeepsUnmodeledFields(t *testing.T) {
	t.Parallel()
	fake := &fakeSettingsClient{doc: map[string]any{
		"_id": "c1", "site_id": "s1", "key": "connectivity",
		"enabled": true, "enable_isolated_wlan": true, "uplink_type": "gateway",
		"x_mesh_essid": "mesh", "x_mesh_psk": "psk", "future_field": map[string]any{"a": 1.0},
	}}
	client := &base.Client{Client: fake, Site: "default"}
	r, ok := NewConnectivityResource().(*connectivityResource)
	require.True(t, ok)

	_, err := r.Handlers.Update(context.Background(), client, "default", &unifi.SettingConnectivity{ID: "c1", Enabled: false})
	require.NoError(t, err)

	assert.Equal(t, unifi.SettingConnectivityKey, fake.putKey)
	assert.Equal(t, false, fake.putBody["enabled"], "the changed modeled field is written")
	assert.Equal(t, true, fake.putBody["enable_isolated_wlan"], "non-omitempty field the model does not manage must not be zeroed")
	assert.Equal(t, "gateway", fake.putBody["uplink_type"])
	assert.Equal(t, "psk", fake.putBody["x_mesh_psk"])
	assert.Equal(t, map[string]any{"a": 1.0}, fake.putBody["future_field"], "field unknown to the SDK must survive")
	assert.Equal(t, "connectivity", fake.putBody["key"])
	assert.Equal(t, "c1", fake.putBody["_id"])
}

func TestDiffJSON(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		before, after, want map[string]any
	}{
		"unchanged":        {map[string]any{"a": 1.0}, map[string]any{"a": 1.0}, map[string]any{}},
		"changed scalar":   {map[string]any{"a": 1.0}, map[string]any{"a": 2.0}, map[string]any{"a": 2.0}},
		"new key":          {map[string]any{}, map[string]any{"a": true}, map[string]any{"a": true}},
		"nested partial":   {map[string]any{"o": map[string]any{"x": 1.0, "y": 1.0}}, map[string]any{"o": map[string]any{"x": 1.0, "y": 2.0}}, map[string]any{"o": map[string]any{"y": 2.0}}},
		"array as a whole": {map[string]any{"l": []any{1.0, 2.0}}, map[string]any{"l": []any{1.0}}, map[string]any{"l": []any{1.0}}},
		"absent in after":  {map[string]any{"a": 1.0}, map[string]any{}, map[string]any{}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, diffJSON(tc.before, tc.after))
		})
	}
}

func TestDeepMergeJSON(t *testing.T) {
	t.Parallel()
	got := deepMergeJSON(
		map[string]any{"a": 1.0, "o": map[string]any{"x": 1.0, "keep": true}, "l": []any{1.0}},
		map[string]any{"o": map[string]any{"x": 2.0}, "l": []any{}},
	)
	assert.Equal(t, map[string]any{"a": 1.0, "o": map[string]any{"x": 2.0, "keep": true}, "l": []any{}}, got)
}
