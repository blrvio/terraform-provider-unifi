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

func TestPredefinedPolicyPatchBody(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		logging, report types.Bool
		wantJSON        string
	}{
		"both set":     {types.BoolValue(true), types.BoolValue(false), `{"loggingEnabled":true,"reportFlowEnabled":false}`},
		"only logging": {types.BoolValue(false), types.BoolUnknown(), `{"loggingEnabled":false}`},
		"nothing set":  {types.BoolNull(), types.BoolNull(), `{}`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := &officialFirewallPredefinedPolicyModel{LoggingEnabled: tc.logging, ReportFlowEnabled: tc.report}
			raw, err := json.Marshal(m.patchBody())
			require.NoError(t, err)
			assert.JSONEq(t, tc.wantJSON, string(raw))
		})
	}
}

func TestPredefinedPolicyMergeAndSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	// Trimmed shape of the predefined "Block Invalid Traffic" policy on UniFi Network 11.0.99.
	var p official.FirewallPolicy
	require.NoError(t, json.Unmarshal([]byte(`{"id":"0b6a8c0e-0000-4000-8000-000000000001","name":"Block Invalid Traffic",`+
		`"enabled":true,"loggingEnabled":false,"reportFlowEnabled":true,"metadata":{"origin":"SYSTEM_DEFINED"},`+
		`"action":{"type":"BLOCK"},"source":{"zoneId":"ad72fd2d-0000-4000-8000-000000000002"},`+
		`"destination":{"zoneId":"53afaeb1-0000-4000-8000-000000000003"},"ipProtocolScope":{"ipVersion":"IPV4_AND_IPV6"}}`), &p))

	m := &officialFirewallPredefinedPolicyModel{}
	require.False(t, m.Merge(ctx, &p).HasError())
	assert.Equal(t, "SYSTEM_DEFINED", m.Origin.ValueString())
	assert.Equal(t, m.ID.ValueString(), m.PolicyID.ValueString())
	assert.True(t, m.ReportFlowEnabled.ValueBool())

	var sresp resource.SchemaResponse
	NewOfficialFirewallPredefinedPolicyResource().Schema(ctx, resource.SchemaRequest{}, &sresp)
	m.SetSite("default")
	state := tfsdk.State{Schema: sresp.Schema, Raw: tftypes.NewValue(sresp.Schema.Type().TerraformType(ctx), nil)}
	require.False(t, state.Set(ctx, m).HasError())
}
