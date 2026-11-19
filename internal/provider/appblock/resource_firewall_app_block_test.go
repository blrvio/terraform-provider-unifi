package appblock

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Shape read from a UniFi Network 11.0.99 console (CyberSecure > Simple App Blocking).
const consoleAppBlock = `{"_id":"1c087508-9d46-4173-87fd-098f75913811","app_category_ids":[13,6,11,24,8,4,0,3,5,7,17,1],` +
	`"app_ids":[],"client_macs":["6c:3c:7c:a7:ab:32"],"name":"[Block:internet] Impressora","network_ids":[],` +
	`"schedule":{"mode":"ALWAYS","repeat_on_days":[],"time_all_day":false},"target_type":"APP_CATEGORY","type":"DEVICE"}`

func TestFirewallAppBlockRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	var api unifi.FirewallAppBlock
	require.NoError(t, json.Unmarshal([]byte(consoleAppBlock), &api))

	m := &firewallAppBlockModel{}
	require.False(t, m.Merge(ctx, &api).HasError())
	assert.Equal(t, "[Block:internet] Impressora", m.Name.ValueString())
	assert.Len(t, m.AppCategoryIDs.Elements(), 12)
	assert.False(t, m.Schedule.IsNull())

	// The model must fit the schema (import sets it as state).
	var sresp resource.SchemaResponse
	NewFirewallAppBlockResource().Schema(ctx, resource.SchemaRequest{}, &sresp)
	require.False(t, sresp.Diagnostics.HasError())
	m.SetSite("default")
	state := tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}
	require.False(t, state.Set(ctx, m).HasError())

	body, diags := m.AsUnifiModel(ctx)
	require.False(t, diags.HasError())
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	var want, got map[string]any
	require.NoError(t, json.Unmarshal([]byte(consoleAppBlock), &want))
	require.NoError(t, json.Unmarshal(raw, &got))
	for _, k := range []string{"name", "type", "target_type", "app_ids", "client_macs", "network_ids", "_id"} {
		assert.Equal(t, want[k], got[k], k)
	}
	// An empty repeat_on_days is omitted on write (SDK omitempty), as for content filtering.
	wantSchedule, _ := want["schedule"].(map[string]any)
	delete(wantSchedule, "repeat_on_days")
	assert.Equal(t, wantSchedule, got["schedule"])
	assert.ElementsMatch(t, want["app_category_ids"], got["app_category_ids"])
}

func TestFirewallAppBlockConfigValidators(t *testing.T) {
	t.Parallel()
	r, ok := NewFirewallAppBlockResource().(*firewallAppBlockResource)
	require.True(t, ok)
	assert.Len(t, r.ConfigValidators(context.Background()), 4)
}
