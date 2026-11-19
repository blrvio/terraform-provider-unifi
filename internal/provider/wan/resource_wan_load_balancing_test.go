package wan

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

// Shape read from a UniFi Network 11.0.99 console in failover mode.
const consoleWANLB = `{"mode":"FAILOVER_ONLY","wan_interfaces":[` +
	`{"mode":"DISTRIBUTED","name":"Internet Principal 1","priority":1,"wan_networkgroup":"WAN","weight":50},` +
	`{"mode":"FAILOVER_ONLY","name":"Internet 2","priority":2,"wan_networkgroup":"WAN2","weight":50}]}`

func TestWANLoadBalancingRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var lb unifi.WANLoadBalancing
	require.NoError(t, json.Unmarshal([]byte(consoleWANLB), &lb))

	m := &wanLoadBalancingModel{}
	require.False(t, m.Merge(ctx, &lb).HasError())
	assert.Equal(t, "wan_load_balancing", m.ID.ValueString())
	assert.Len(t, m.WANInterfaces.Elements(), 2)

	var sresp resource.SchemaResponse
	NewWANLoadBalancingResource().Schema(ctx, resource.SchemaRequest{}, &sresp)
	m.SetSite("default")
	state := tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}
	require.False(t, state.Set(ctx, m).HasError())

	body, diags := m.AsUnifiModel(ctx)
	require.False(t, diags.HasError())
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, consoleWANLB, string(raw), "sorted by priority, same as the console")
}

func TestOverlayWANLoadBalancing(t *testing.T) {
	t.Parallel()
	var current unifi.WANLoadBalancing
	require.NoError(t, json.Unmarshal([]byte(consoleWANLB), &current))

	planned := &unifi.WANLoadBalancing{Mode: "DISTRIBUTED", WANInterfaces: []unifi.WANLoadBalancingInterface{
		{WANNetworkgroup: "WAN2", Mode: "DISTRIBUTED", Priority: 1, Weight: 30},
	}}
	got, err := overlayWANLoadBalancing(&current, planned)
	require.NoError(t, err)
	assert.Equal(t, "DISTRIBUTED", got.Mode)
	assert.Equal(t, "Internet 2", got.WANInterfaces[1].Name, "name kept from the controller")
	assert.Equal(t, 30, got.WANInterfaces[1].Weight)
	assert.Equal(t, 50, got.WANInterfaces[0].Weight, "WAN not in the plan is kept")
	assert.Equal(t, 50, current.WANInterfaces[1].Weight, "current is not mutated")

	_, err = overlayWANLoadBalancing(&current, &unifi.WANLoadBalancing{WANInterfaces: []unifi.WANLoadBalancingInterface{{WANNetworkgroup: "WAN9"}}})
	assert.ErrorContains(t, err, "WAN9")
}
