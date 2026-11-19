package nat

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

func TestNatRuleRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// A DNAT rule in the shape the UniFi Network 11.0.99 UI builds.
	const raw = `{"_id":"r1","description":"https to server","enabled":true,"type":"DNAT","ip_version":"IPV4",` +
		`"in_interface":"wan1","protocol":"tcp","port":"443","ip_address":"10.1.160.217","rule_index":1,` +
		`"logging":false,"exclude":false,"pppoe_use_base_interface":false,"setting_preference":"manual","is_predefined":false,` +
		`"source_filter":{"filter_type":"NONE","invert_address":false,"invert_port":false},` +
		`"destination_filter":{"filter_type":"ADDRESS_AND_PORT","port":"8443","invert_address":false,"invert_port":false}}`
	var api unifi.NatRule
	require.NoError(t, json.Unmarshal([]byte(raw), &api))

	m := &natRuleModel{}
	require.False(t, m.Merge(ctx, &api).HasError())

	var sresp resource.SchemaResponse
	NewNatRuleResource().Schema(ctx, resource.SchemaRequest{}, &sresp)
	m.SetSite("default")
	state := tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}
	require.False(t, state.Set(ctx, m).HasError())

	body, diags := m.AsUnifiModel(ctx)
	require.False(t, diags.HasError())
	out, err := json.Marshal(body)
	require.NoError(t, err)
	assert.JSONEq(t, raw, string(out))
}
