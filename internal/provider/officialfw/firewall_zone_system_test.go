package officialfw

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func zoneFromConsole(t *testing.T, raw string) *officialFirewallZoneModel {
	t.Helper()
	var z official.FirewallZone
	require.NoError(t, json.Unmarshal([]byte(raw), &z))
	m := &officialFirewallZoneModel{}
	require.False(t, m.Merge(context.Background(), &z).HasError())
	m.SetSite("default")
	return m
}

// Shapes read from a UniFi Network 11.0.99 console.
const (
	vpnZone      = `{"id":"d1bebd02-0000-4000-8000-000000000001","name":"Vpn","networkIds":["663bc70c-85bc-415f-9316-19dbdc2353c0"],"metadata":{"origin":"SYSTEM_DEFINED","configurable":false}}`
	internalZone = `{"id":"ad72fd2d-0000-4000-8000-000000000002","name":"Internal","networkIds":[],"metadata":{"origin":"SYSTEM_DEFINED","configurable":true}}`
	userZone     = `{"id":"ad72fd2d-0000-4000-8000-000000000003","name":"Lab","networkIds":[],"metadata":{"origin":"USER_DEFINED"}}`
)

func TestFirewallZoneOriginAndConfigurable(t *testing.T) {
	t.Parallel()
	vpn := zoneFromConsole(t, vpnZone)
	assert.Equal(t, "SYSTEM_DEFINED", vpn.Origin.ValueString())
	assert.False(t, vpn.Configurable.ValueBool())
	assert.True(t, zoneFromConsole(t, internalZone).Configurable.ValueBool())
	user := zoneFromConsole(t, userZone)
	assert.Equal(t, "USER_DEFINED", user.Origin.ValueString())
	assert.True(t, user.Configurable.ValueBool())
}

func TestCheckSystemZoneChange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var sresp resource.SchemaResponse
	NewOfficialFirewallZoneResource().Schema(ctx, resource.SchemaRequest{}, &sresp)
	s := sresp.Schema
	typ := s.Type().TerraformType(ctx)

	toState := func(m *officialFirewallZoneModel) tfsdk.State {
		st := tfsdk.State{Schema: s, Raw: tftypes.NewValue(typ, nil)}
		require.False(t, st.Set(ctx, m).HasError())
		return st
	}
	toPlan := func(m *officialFirewallZoneModel) tfsdk.Plan {
		st := toState(m)
		return tfsdk.Plan{Schema: s, Raw: st.Raw}
	}
	nullPlan := tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(typ, nil)}

	renamed := zoneFromConsole(t, internalZone)
	renamed.Name = types.StringValue("Inside")
	vpnMoreNets := zoneFromConsole(t, vpnZone)
	vpnMoreNets.NetworkIDs = types.ListValueMust(types.StringType, nil)
	internalNets := zoneFromConsole(t, internalZone)
	internalNets.NetworkIDs, _ = types.ListValueFrom(ctx, types.StringType, []string{"d837e7a9-0000-4000-8000-000000000009"})

	cases := map[string]struct {
		state   *officialFirewallZoneModel
		plan    *tfsdk.Plan
		wantErr bool
	}{
		"system zone no change":           {state: zoneFromConsole(t, vpnZone), plan: ptr(toPlan(zoneFromConsole(t, vpnZone)))},
		"destroy system zone":             {state: zoneFromConsole(t, vpnZone), plan: &nullPlan, wantErr: true},
		"rename system zone":              {state: zoneFromConsole(t, internalZone), plan: ptr(toPlan(renamed)), wantErr: true},
		"networks of non-configurable":    {state: zoneFromConsole(t, vpnZone), plan: ptr(toPlan(vpnMoreNets)), wantErr: true},
		"networks of configurable system": {state: zoneFromConsole(t, internalZone), plan: ptr(toPlan(internalNets))},
		"destroy user zone":               {state: zoneFromConsole(t, userZone), plan: &nullPlan},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := resource.ModifyPlanRequest{State: toState(tc.state), Plan: *tc.plan}
			diags := checkSystemZoneChange(ctx, req)
			assert.Equal(t, tc.wantErr, diags.HasError(), "%v", diags)
		})
	}
}

func ptr[T any](v T) *T { return &v }
