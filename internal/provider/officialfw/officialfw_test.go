package officialfw

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// listOfStrings builds a Framework list value from the given strings for tests.
func listOfStrings(t *testing.T, in ...string) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), types.StringType, in)
	require.False(t, diags.HasError(), "ListValueFrom diagnostics: %v", diags)
	return list
}

// ---- Zone ----

func TestFirewallZoneAsUnifiModel_NetworkIDs(t *testing.T) {
	t.Parallel()
	m := &officialFirewallZoneModel{
		Name: types.StringValue("LAN"),
		NetworkIDs: listOfStrings(t,
			"11111111-1111-1111-1111-111111111111",
			"22222222-2222-2222-2222-222222222222",
		),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	zoneBody, ok := body.(*official.FirewallZoneCreateOrUpdate)
	require.True(t, ok, "expected *official.FirewallZoneCreateOrUpdate, got %T", body)
	assert.Equal(t, "LAN", zoneBody.Name)
	require.Len(t, zoneBody.NetworkIds, 2)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", zoneBody.NetworkIds[0].String())
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", zoneBody.NetworkIds[1].String())

	// Verify the wire representation carries the camelCase key.
	raw, err := json.Marshal(zoneBody)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, "LAN", out["name"])
	assert.Len(t, out["networkIds"], 2)
}

func TestFirewallZoneAsUnifiModel_InvalidUUID(t *testing.T) {
	t.Parallel()
	m := &officialFirewallZoneModel{
		Name:       types.StringValue("LAN"),
		NetworkIDs: listOfStrings(t, "not-a-uuid"),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError(), "expected error for invalid UUID")
}

func TestFirewallZoneMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":"33333333-3333-3333-3333-333333333333","name":"IoT",` +
		`"networkIds":["44444444-4444-4444-4444-444444444444","55555555-5555-5555-5555-555555555555"]}`
	var zone official.FirewallZone
	require.NoError(t, json.Unmarshal([]byte(raw), &zone))

	m := &officialFirewallZoneModel{}
	diags := m.Merge(context.Background(), &zone)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "33333333-3333-3333-3333-333333333333", m.GetID())
	assert.Equal(t, "IoT", m.Name.ValueString())

	var got []string
	m.NetworkIDs.ElementsAs(context.Background(), &got, false)
	assert.Equal(t, []string{
		"44444444-4444-4444-4444-444444444444",
		"55555555-5555-5555-5555-555555555555",
	}, got)
}

// ---- Policy ----

func TestFirewallPolicyAsUnifiModel_Body(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyModel{
		Name:            types.StringValue("allow-lan-to-wan"),
		Action:          types.StringValue(firewallPolicyActionAllow),
		Source:          ut.JSONStringValue(`{"zoneId":"11111111-1111-1111-1111-111111111111"}`),
		Destination:     ut.JSONStringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope: ut.JSONStringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
		// Enabled null -> defaults true; LoggingEnabled null -> defaults false.
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	policyBody, ok := body.(*official.FirewallPolicyCreateOrUpdate)
	require.True(t, ok, "expected *official.FirewallPolicyCreateOrUpdate, got %T", body)
	assert.Equal(t, "allow-lan-to-wan", policyBody.Name)
	assert.Equal(t, firewallPolicyActionAllow, policyBody.Action.Type)
	assert.True(t, policyBody.Enabled)
	assert.False(t, policyBody.LoggingEnabled)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", policyBody.Source.ZoneId.String())
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", policyBody.Destination.ZoneId.String())

	// Verify the wire keys on the marshaled body.
	raw, err := json.Marshal(policyBody)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Equal(t, "allow-lan-to-wan", out["name"])
	assert.Equal(t, true, out["enabled"])
	action, ok := out["action"].(map[string]any)
	require.True(t, ok, "action should be an object")
	assert.Equal(t, "ALLOW", action["type"])
	source, ok := out["source"].(map[string]any)
	require.True(t, ok, "source should be an object")
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", source["zoneId"])
	dest, ok := out["destination"].(map[string]any)
	require.True(t, ok, "destination should be an object")
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", dest["zoneId"])
}

func TestFirewallPolicyAsUnifiModel_WithFilters(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyModel{
		Name:                  types.StringValue("block-invalid"),
		Action:                types.StringValue(firewallPolicyActionBlock),
		Enabled:               types.BoolValue(false),
		LoggingEnabled:        types.BoolValue(true),
		Description:           types.StringValue("drop invalid"),
		IpsecFilter:           types.StringValue(firewallPolicyIpsecMatchEncrypted),
		ConnectionStateFilter: listOfStrings(t, firewallPolicyConnStateNew, firewallPolicyConnStateInvalid),
		Source:                ut.JSONStringValue(`{"zoneId":"11111111-1111-1111-1111-111111111111"}`),
		Destination:           ut.JSONStringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope:       ut.JSONStringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	policyBody, ok := body.(*official.FirewallPolicyCreateOrUpdate)
	require.True(t, ok, "expected *official.FirewallPolicyCreateOrUpdate, got %T", body)
	assert.False(t, policyBody.Enabled)
	assert.True(t, policyBody.LoggingEnabled)
	require.NotNil(t, policyBody.Description)
	assert.Equal(t, "drop invalid", *policyBody.Description)
	require.NotNil(t, policyBody.IpsecFilter)
	assert.Equal(t, official.FirewallPolicyIpsecFilter(firewallPolicyIpsecMatchEncrypted), *policyBody.IpsecFilter)
	require.NotNil(t, policyBody.ConnectionStateFilter)
	assert.Equal(t, []official.FirewallPolicyConnectionStateFilter{
		official.FirewallPolicyConnectionStateFilter(firewallPolicyConnStateNew),
		official.FirewallPolicyConnectionStateFilter(firewallPolicyConnStateInvalid),
	}, *policyBody.ConnectionStateFilter)
}

func TestFirewallPolicyAsUnifiModel_InvalidSourceJSON(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyModel{
		Name:            types.StringValue("bad"),
		Action:          types.StringValue(firewallPolicyActionAllow),
		Source:          ut.JSONStringValue(`{not json`),
		Destination:     ut.JSONStringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope: ut.JSONStringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
	}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError(), "expected error for invalid source JSON")
}

func TestFirewallPolicyMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":"66666666-6666-6666-6666-666666666666","name":"allow-lan-to-wan",` +
		`"action":{"type":"ALLOW"},"enabled":true,"loggingEnabled":false,"index":3,` +
		`"source":{"zoneId":"11111111-1111-1111-1111-111111111111"},` +
		`"destination":{"zoneId":"22222222-2222-2222-2222-222222222222"},` +
		`"ipProtocolScope":{"ipVersion":"BOTH","type":"ALL"},` +
		`"metadata":{}}`
	var policy official.FirewallPolicy
	require.NoError(t, json.Unmarshal([]byte(raw), &policy))

	m := &officialFirewallPolicyModel{}
	diags := m.Merge(context.Background(), &policy)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "66666666-6666-6666-6666-666666666666", m.GetID())
	assert.Equal(t, "allow-lan-to-wan", m.Name.ValueString())
	assert.Equal(t, "ALLOW", m.Action.ValueString())
	assert.True(t, m.Enabled.ValueBool())
	assert.False(t, m.LoggingEnabled.ValueBool())
	assert.Equal(t, int32(3), m.Index.ValueInt32())
	assert.True(t, m.IpsecFilter.IsNull())
	assert.True(t, m.ConnectionStateFilter.IsNull())
	assert.True(t, m.Schedule.IsNull())

	// The complex objects come back as compact JSON strings carrying the zoneId.
	assert.JSONEq(t, `{"zoneId":"11111111-1111-1111-1111-111111111111"}`, m.Source.ValueString())
	assert.JSONEq(t, `{"zoneId":"22222222-2222-2222-2222-222222222222"}`, m.Destination.ValueString())
	assert.JSONEq(t, `{"ipVersion":"BOTH","type":"ALL"}`, m.IPProtocolScope.ValueString())
}

// ---- Policy order ----

func TestFirewallPolicyOrderAsUnifiModel_Conversion(t *testing.T) {
	t.Parallel()
	const a, b = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	cases := map[string]struct {
		before, after types.List
		wantJSON      string
	}{
		"before only (UI default)": {
			before: listOfStrings(t, a, b), after: types.ListNull(types.StringType),
			wantJSON: `{"orderedFirewallPolicyIds":{"beforeSystemDefined":["` + a + `","` + b + `"],"afterSystemDefined":[]}}`,
		},
		"after only": {
			before: ut.EmptyList(types.StringType), after: listOfStrings(t, a),
			wantJSON: `{"orderedFirewallPolicyIds":{"beforeSystemDefined":[],"afterSystemDefined":["` + a + `"]}}`,
		},
		"both empty never sends null": {
			before: types.ListUnknown(types.StringType), after: types.ListNull(types.StringType),
			wantJSON: `{"orderedFirewallPolicyIds":{"beforeSystemDefined":[],"afterSystemDefined":[]}}`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := &officialFirewallPolicyOrderModel{
				SourceZoneID:      types.StringValue("11111111-1111-1111-1111-111111111111"),
				DestinationZoneID: types.StringValue("22222222-2222-2222-2222-222222222222"),
				BeforeSystemIDs:   tc.before,
				AfterSystemIDs:    tc.after,
			}
			body, diags := m.AsUnifiModel(context.Background())
			require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
			ordering, ok := body.(*official.FirewallPolicyOrdering)
			require.True(t, ok, "expected *official.FirewallPolicyOrdering, got %T", body)
			raw, err := json.Marshal(ordering)
			require.NoError(t, err)
			assert.JSONEq(t, tc.wantJSON, string(raw))
			assert.NotContains(t, string(raw), "null")
		})
	}
}

func TestFirewallPolicyOrderAsUnifiModel_InvalidUUID(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyOrderModel{BeforeSystemIDs: listOfStrings(t, "not-a-uuid"), AfterSystemIDs: ut.EmptyList(types.StringType)}
	_, diags := m.AsUnifiModel(context.Background())
	assert.True(t, diags.HasError())
}

func TestFirewallPolicyOrderMerge_RoundTrip(t *testing.T) {
	t.Parallel()
	// Shape read from a UniFi Network 11.0.99 console (Internal -> Internal).
	raw := `{"orderedFirewallPolicyIds":{"beforeSystemDefined":[` +
		`"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"],` +
		`"afterSystemDefined":[]}}`
	var ordering official.FirewallPolicyOrdering
	require.NoError(t, json.Unmarshal([]byte(raw), &ordering))

	m := &officialFirewallPolicyOrderModel{}
	diags := m.Merge(context.Background(), &ordering)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	var before, after []string
	m.BeforeSystemIDs.ElementsAs(context.Background(), &before, false)
	m.AfterSystemIDs.ElementsAs(context.Background(), &after, false)
	assert.Equal(t, []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}, before)
	assert.Empty(t, after)
	assert.False(t, m.AfterSystemIDs.IsNull(), "an empty list must be [] in state, matching the schema default")
}

func TestFirewallPolicyOrderUpgradeStateV0(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r, ok := NewOfficialFirewallPolicyOrderResource().(*officialFirewallPolicyOrderResource)
	require.True(t, ok)
	up := r.UpgradeState(ctx)[0]

	priorType := up.PriorSchema.Type().TerraformType(ctx)
	priorRaw := tftypes.NewValue(priorType, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, "z1:z2"),
		"site":                tftypes.NewValue(tftypes.String, "default"),
		"source_zone_id":      tftypes.NewValue(tftypes.String, "z1"),
		"destination_zone_id": tftypes.NewValue(tftypes.String, "z2"),
		"policy_ids": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		}),
	})
	var sresp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sresp)
	req := resource.UpgradeStateRequest{State: &tfsdk.State{Raw: priorRaw, Schema: *up.PriorSchema}}
	resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: sresp.Schema}}
	up.StateUpgrader(ctx, req, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var got officialFirewallPolicyOrderModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	var before, after []string
	got.BeforeSystemIDs.ElementsAs(ctx, &before, false)
	got.AfterSystemIDs.ElementsAs(ctx, &after, false)
	assert.Empty(t, before)
	assert.Equal(t, []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}, after)
	assert.Equal(t, "z1:z2", got.ID.ValueString())
}

func TestFirewallPolicyOrderMerge_InvalidType(t *testing.T) {
	t.Parallel()
	m := &officialFirewallPolicyOrderModel{}
	diags := m.Merge(context.Background(), "not an ordering")
	assert.True(t, diags.HasError(), "expected error for wrong type")
}

// ---- Policy: report_flow_enabled ----

// basePolicyModel is a minimal valid policy model for body-building tests.
func basePolicyModel() *officialFirewallPolicyModel {
	return &officialFirewallPolicyModel{
		Name:            types.StringValue("p"),
		Action:          types.StringValue(firewallPolicyActionAllow),
		Source:          ut.JSONStringValue(`{"zoneId":"11111111-1111-1111-1111-111111111111"}`),
		Destination:     ut.JSONStringValue(`{"zoneId":"22222222-2222-2222-2222-222222222222"}`),
		IPProtocolScope: ut.JSONStringValue(`{"ipVersion":"BOTH","type":"ALL"}`),
	}
}

// policyWireBody builds the request body for m and returns it decoded as a map,
// i.e. exactly the JSON keys the controller receives.
func policyWireBody(t *testing.T, m *officialFirewallPolicyModel) (*official.FirewallPolicyCreateOrUpdate, map[string]any) {
	t.Helper()
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)
	b, diags := asFirewallPolicyBody(body)
	require.False(t, diags.HasError(), "asFirewallPolicyBody diagnostics: %v", diags)
	raw, err := json.Marshal(b)
	require.NoError(t, err)
	out := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &out))
	return b, out
}

func TestFirewallPolicyAsUnifiModel_ReportFlowOmitted(t *testing.T) {
	t.Parallel()
	for name, v := range map[string]types.Bool{
		"unknown (create without the attribute)": types.BoolUnknown(),
		"null (older controller state)":          types.BoolNull(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := basePolicyModel()
			m.ReportFlowEnabled = v
			b, wire := policyWireBody(t, m)
			assert.Nil(t, b.ReportFlowEnabled)
			_, present := wire["reportFlowEnabled"]
			assert.False(t, present, "reportFlowEnabled must be omitted so the controller applies its default")
		})
	}
}

// reportFlowAttr returns the report_flow_enabled schema attribute.
func reportFlowAttr(t *testing.T) schema.BoolAttribute {
	t.Helper()
	r, ok := NewOfficialFirewallPolicyResource().(*officialFirewallPolicyResource)
	require.True(t, ok)
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "schema diagnostics: %v", resp.Diagnostics)
	attr, ok := resp.Schema.Attributes["report_flow_enabled"].(schema.BoolAttribute)
	require.True(t, ok, "report_flow_enabled must be a bool attribute")
	return attr
}

func TestFirewallPolicySchema_ReportFlowOptionalComputed(t *testing.T) {
	t.Parallel()
	attr := reportFlowAttr(t)
	assert.True(t, attr.Optional)
	assert.True(t, attr.Computed)
	assert.Nil(t, attr.Default, "no static default: the controller default applies on create")
	assert.Len(t, attr.PlanModifiers, 1)
}

// TestFirewallPolicyUpdate_ReportFlowFalsePreserved is the regression guard for
// the silent re-enable: with report_flow_enabled unset in config and false in
// state, the plan must carry false (UseStateForUnknown) and the full-replace PUT
// must send reportFlowEnabled=false instead of omitting it (which the controller
// treats as true).
func TestFirewallPolicyUpdate_ReportFlowFalsePreserved(t *testing.T) {
	t.Parallel()
	attr := reportFlowAttr(t)

	// A non-null prior state marks this as an update (UseStateForUnknown is a
	// no-op on create); only its nullness matters to the modifier.
	req := planmodifier.BoolRequest{
		State:       tfsdk.State{Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{})},
		Path:        path.Root("report_flow_enabled"),
		ConfigValue: types.BoolNull(),
		PlanValue:   types.BoolUnknown(),
		StateValue:  types.BoolValue(false),
	}
	resp := &planmodifier.BoolResponse{PlanValue: req.PlanValue}
	for _, pm := range attr.PlanModifiers {
		pm.PlanModifyBool(context.Background(), req, resp)
	}
	require.False(t, resp.Diagnostics.HasError())
	require.True(t, resp.PlanValue.Equal(types.BoolValue(false)), "plan = %s", resp.PlanValue)

	m := basePolicyModel()
	m.ReportFlowEnabled = resp.PlanValue
	b, wire := policyWireBody(t, m)
	require.NotNil(t, b.ReportFlowEnabled)
	assert.False(t, *b.ReportFlowEnabled)
	assert.Equal(t, false, wire["reportFlowEnabled"])

	// An explicit true in config is sent as true.
	m.ReportFlowEnabled = types.BoolValue(true)
	_, wire = policyWireBody(t, m)
	assert.Equal(t, true, wire["reportFlowEnabled"])
}

func TestFirewallPolicyMerge_ReportFlow(t *testing.T) {
	t.Parallel()
	for _, want := range []bool{true, false} {
		raw := `{"id":"66666666-6666-6666-6666-666666666666","name":"p","action":{"type":"ALLOW"},` +
			`"enabled":true,"loggingEnabled":false,"index":1,` +
			`"reportFlowEnabled":` + map[bool]string{true: "true", false: "false"}[want] + `,` +
			`"source":{"zoneId":"11111111-1111-1111-1111-111111111111"},` +
			`"destination":{"zoneId":"22222222-2222-2222-2222-222222222222"},` +
			`"ipProtocolScope":{"ipVersion":"BOTH","type":"ALL"},"metadata":{}}`
		var policy official.FirewallPolicy
		require.NoError(t, json.Unmarshal([]byte(raw), &policy))

		m := &officialFirewallPolicyModel{}
		diags := m.Merge(context.Background(), &policy)
		require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)
		assert.True(t, m.ReportFlowEnabled.Equal(types.BoolValue(want)), "report_flow_enabled = %s, want %v", m.ReportFlowEnabled, want)
	}
}

// TestFirewallPolicy_ReportFlowGatedByControllerVersion covers controllers that
// predate reportFlowEnabled (< 11.0.99): the field is never sent and the
// attribute stays null rather than a meaningless false.
func TestFirewallPolicy_ReportFlowGatedByControllerVersion(t *testing.T) {
	t.Parallel()
	for version, supported := range map[string]bool{"10.6.109": false, "11.0.99": true} {
		t.Run(version, func(t *testing.T) {
			t.Parallel()
			r, ok := NewOfficialFirewallPolicyResource().(*officialFirewallPolicyResource)
			require.True(t, ok)
			r.SetClient(&base.Client{Version: base.AsVersion(version)})

			v := false
			body := &official.FirewallPolicyCreateOrUpdate{ReportFlowEnabled: &v}
			r.dropUnsupportedReportFlow(body)
			assert.Equal(t, supported, body.ReportFlowEnabled != nil)

			m := &officialFirewallPolicyModel{ReportFlowEnabled: types.BoolValue(false)}
			r.nullUnsupportedReportFlow(m)
			assert.Equal(t, !supported, m.ReportFlowEnabled.IsNull())
		})
	}
}
