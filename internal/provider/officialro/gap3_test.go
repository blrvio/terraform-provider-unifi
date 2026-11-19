package officialro

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/blrvio/go-unifi/v11/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// schemaOf builds the schema for a data source so tests can round-trip a model
// through it, which validates that every model tfsdk tag matches an attribute.
func schemaOf(t *testing.T, ds datasource.DataSource) dschema.Schema {
	t.Helper()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	return resp.Schema
}

// assertRoundTrip sets model into a fresh state built from schema and fails on any
// diagnostic. It catches schema/model (tfsdk tag) mismatches without a controller.
func assertRoundTrip(t *testing.T, s dschema.Schema, model any) {
	t.Helper()
	state := tfsdk.State{Schema: s}
	diags := state.Set(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("state round-trip diagnostics: %v", diags)
	}
}

func strPtr(s string) *string        { return &s }
func f64Ptr(f float64) *float64      { return &f }
func i64Ptr(i int64) *int64          { return &i }
func timePtr(t time.Time) *time.Time { return &t }

func TestMapClientOverview(t *testing.T) {
	id := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	m, err := mapClientOverview(official.ClientOverview{
		Id:          id,
		Name:        "laptop",
		Type:        "WIRELESS",
		IpAddress:   strPtr("10.0.0.5"),
		ConnectedAt: timePtr(ts),
	})
	if err != nil {
		t.Fatalf("mapClientOverview: %v", err)
	}
	if m.ID.ValueString() != id.String() || m.Name.ValueString() != "laptop" || m.Type.ValueString() != "WIRELESS" {
		t.Fatalf("unexpected scalar mapping: %+v", m)
	}
	if m.IPAddress.ValueString() != "10.0.0.5" {
		t.Fatalf("ip = %q", m.IPAddress.ValueString())
	}
	if m.ConnectedAt.ValueString() != "2026-01-02T03:04:05Z" {
		t.Fatalf("connected_at = %q", m.ConnectedAt.ValueString())
	}
	if m.JSON.IsNull() || m.JSON.ValueString() == "" {
		t.Fatalf("expected json to be populated")
	}
}

func TestMapClientOverview_nullOptionals(t *testing.T) {
	m, err := mapClientOverview(official.ClientOverview{
		Id:   mustUUID(t, "11111111-1111-1111-1111-111111111111"),
		Name: "x", Type: "VPN",
	})
	if err != nil {
		t.Fatalf("mapClientOverview: %v", err)
	}
	if !m.IPAddress.IsNull() || !m.ConnectedAt.IsNull() {
		t.Fatalf("expected null optionals, got %+v", m)
	}
}

func TestMapAdoptedDeviceOverview_features(t *testing.T) {
	m, err := mapAdoptedDeviceOverview(official.AdoptedDeviceOverview{
		Id:         mustUUID(t, "22222222-2222-2222-2222-222222222222"),
		Name:       "usw",
		Model:      "USW-24",
		MacAddress: "aa:bb:cc:dd:ee:ff",
		IpAddress:  "10.0.0.2",
		State:      official.AdoptedDeviceOverviewState("ONLINE"),
		Supported:  true,
		Features:   []official.AdoptedDeviceOverviewFeatures{"switching", "gateway"},
	})
	if err != nil {
		t.Fatalf("mapAdoptedDeviceOverview: %v", err)
	}
	if len(m.Features) != 2 || m.Features[0] != "switching" || m.Features[1] != "gateway" {
		t.Fatalf("features = %v", m.Features)
	}
	if m.State.ValueString() != "ONLINE" || !m.Supported.ValueBool() {
		t.Fatalf("unexpected mapping: %+v", m)
	}
}

func TestMapDeviceStatistics_uplink(t *testing.T) {
	withUplink, err := mapDeviceStatistics(official.LatestStatisticsForADevice{
		CpuUtilizationPct: f64Ptr(12.5),
		UptimeSec:         i64Ptr(3600),
		Uplink: &official.LatestStatisticsForADeviceUplinkInterface{
			RxRateBps: i64Ptr(1000),
			TxRateBps: i64Ptr(2000),
		},
	})
	if err != nil {
		t.Fatalf("mapDeviceStatistics: %v", err)
	}
	if withUplink.CPUUtilizationPct.ValueFloat64() != 12.5 || withUplink.UptimeSec.ValueInt64() != 3600 {
		t.Fatalf("scalar mapping: %+v", withUplink)
	}
	if withUplink.UplinkRxRateBps.ValueInt64() != 1000 || withUplink.UplinkTxRateBps.ValueInt64() != 2000 {
		t.Fatalf("uplink mapping: %+v", withUplink)
	}

	noUplink, err := mapDeviceStatistics(official.LatestStatisticsForADevice{})
	if err != nil {
		t.Fatalf("mapDeviceStatistics: %v", err)
	}
	if !noUplink.UplinkRxRateBps.IsNull() || !noUplink.CPUUtilizationPct.IsNull() {
		t.Fatalf("expected nulls when uplink/stats absent: %+v", noUplink)
	}
}

func TestMapNetworkReferences(t *testing.T) {
	refID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	m, err := mapNetworkReferences(official.NetworkReferences{
		ReferenceResources: []official.NetworkReferenceResource{
			{
				ResourceType:   official.NetworkReferenceResourceResourceType("CLIENT"),
				ReferenceCount: 1,
				References:     &[]official.NetworkReferenceDetail{{ReferenceId: refID}},
			},
		},
	})
	if err != nil {
		t.Fatalf("mapNetworkReferences: %v", err)
	}
	if len(m.ReferenceResources) != 1 {
		t.Fatalf("resources = %d", len(m.ReferenceResources))
	}
	r := m.ReferenceResources[0]
	if r.ResourceType.ValueString() != "CLIENT" || r.ReferenceCount.ValueInt64() != 1 {
		t.Fatalf("resource mapping: %+v", r)
	}
	if len(r.ReferenceIDs) != 1 || r.ReferenceIDs[0] != refID.String() {
		t.Fatalf("reference ids = %v", r.ReferenceIDs)
	}
}

// mustLAG decodes an Official-API LAG payload the way the SDK does on a real
// read, so the union variant (and its members) is populated exactly as in
// production rather than through the SDK's From* setters.
func mustLAG(t *testing.T, payload string) official.LAGDetails {
	t.Helper()
	var l official.LAGDetails
	if err := json.Unmarshal([]byte(payload), &l); err != nil {
		t.Fatalf("decode LAG payload: %v", err)
	}
	return l
}

const (
	lagID    = "55555555-5555-5555-5555-555555555555"
	lagDevID = "44444444-4444-4444-4444-444444444444"
	lagMeta  = `"metadata":{"origin":"USER_DEFINED"}`
)

func TestMapLagDetails_local(t *testing.T) {
	m, diags := mapLagDetails(mustLAG(t, `{"id":"`+lagID+`","type":"LOCAL",`+lagMeta+
		`,"members":[{"deviceId":"`+lagDevID+`","portIdxs":[1,2,3]}]}`))
	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if m.Type.ValueString() != "LOCAL" || m.MetadataOrigin.ValueString() != "USER_DEFINED" {
		t.Fatalf("scalar mapping: %+v", m)
	}
	if len(m.Members) != 1 || m.Members[0].DeviceID.ValueString() != lagDevID {
		t.Fatalf("members = %+v", m.Members)
	}
	if !m.Members[0].UnitID.IsNull() || !m.Members[0].UnitMacAddress.IsNull() {
		t.Fatalf("unit fields must be null outside a switch stack: %+v", m.Members[0])
	}
	if len(m.Members[0].PortIdxs) != 3 || m.Members[0].PortIdxs[2] != 3 {
		t.Fatalf("port idxs = %v", m.Members[0].PortIdxs)
	}
}

func TestMapLagDetails_multiChassisNilMembers(t *testing.T) {
	l := mustLAG(t, `{"id":"`+lagID+`","type":"MULTI_CHASSIS",`+lagMeta+`}`)
	m, diags := mapLagDetails(l)
	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if m.Members == nil || len(m.Members) != 0 {
		t.Fatalf("expected an empty (non-nil) member list, got %#v", m.Members)
	}
	_, membersJSON, diags := lagMembers(l)
	if diags.HasError() || membersJSON != "[]" {
		t.Fatalf("members_json = %q (diags %v)", membersJSON, diags)
	}
}

func TestMapLagDetails_switchStack(t *testing.T) {
	l := mustLAG(t, `{"id":"`+lagID+`","type":"SWITCH_STACK",`+lagMeta+
		`,"switchStackId":"88888888-8888-8888-8888-888888888888"`+
		`,"members":[{"unitId":2,"unitMacAddress":"aa:bb:cc:dd:ee:02","portIdxs":[25,26]}]}`)
	m, diags := mapLagDetails(l)
	if diags.HasError() || diags.WarningsCount() > 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(m.Members) != 1 {
		t.Fatalf("members = %+v", m.Members)
	}
	got := m.Members[0]
	if !got.DeviceID.IsNull() {
		t.Fatalf("device_id must be null for a switch-stack member (never the zero UUID), got %q", got.DeviceID.ValueString())
	}
	if got.UnitID.ValueInt64() != 2 || got.UnitMacAddress.ValueString() != "aa:bb:cc:dd:ee:02" {
		t.Fatalf("unit mapping: %+v", got)
	}
	if len(got.PortIdxs) != 2 || got.PortIdxs[0] != 25 {
		t.Fatalf("port idxs = %v", got.PortIdxs)
	}
	_, membersJSON, _ := lagMembers(l)
	if membersJSON != `[{"portIdxs":[25,26],"unitId":2,"unitMacAddress":"aa:bb:cc:dd:ee:02"}]` {
		t.Fatalf("members_json = %s", membersJSON)
	}
	if !strings.Contains(m.JSON.ValueString(), `"switchStackId":"88888888-8888-8888-8888-888888888888"`) {
		t.Fatalf("json must keep the variant fields: %s", m.JSON.ValueString())
	}
}

func TestMapLagDetails_unknownTypeWarns(t *testing.T) {
	l := mustLAG(t, `{"id":"`+lagID+`","type":"FUTURE_KIND",`+lagMeta+`,"members":[{"x":1}]}`)
	m, diags := mapLagDetails(l)
	if diags.HasError() {
		t.Fatalf("an unknown LAG type must not fail the read: %v", diags)
	}
	if diags.WarningsCount() != 1 {
		t.Fatalf("expected exactly one warning, got %v", diags)
	}
	if m.Members == nil || len(m.Members) != 0 || m.Type.ValueString() != "FUTURE_KIND" {
		t.Fatalf("unexpected mapping: %+v", m)
	}
	if _, membersJSON, _ := lagMembers(l); membersJSON != "[]" {
		t.Fatalf("members_json = %q", membersJSON)
	}
}

func TestMapSwitchStack_units(t *testing.T) {
	dev := mustUUID(t, "66666666-6666-6666-6666-666666666666")
	role := official.SwitchStackUnitRole("ACTIVE_CONTROLLER")
	order := int32(1)
	m, err := mapSwitchStack(official.SwitchStack{
		Id:       mustUUID(t, "88888888-8888-8888-8888-888888888888"),
		Name:     "stack-1",
		DeviceId: &dev,
		Units: []official.SwitchStackUnit{
			{Id: 1, MacAddress: "aa:bb:cc:dd:ee:01", Role: &role, Order: &order},
			{Id: 2, MacAddress: "aa:bb:cc:dd:ee:02"},
		},
	})
	if err != nil {
		t.Fatalf("mapSwitchStack: %v", err)
	}
	if m.DeviceID.ValueString() != dev.String() {
		t.Fatalf("device_id = %q", m.DeviceID.ValueString())
	}
	if len(m.Units) != 2 {
		t.Fatalf("units = %+v", m.Units)
	}
	first, second := m.Units[0], m.Units[1]
	if first.ID.ValueInt64() != 1 || first.MacAddress.ValueString() != "aa:bb:cc:dd:ee:01" ||
		first.Role.ValueString() != "ACTIVE_CONTROLLER" || first.Order.ValueInt64() != 1 {
		t.Fatalf("first unit = %+v", first)
	}
	if second.ID.ValueInt64() != 2 || !second.Role.IsNull() || !second.Order.IsNull() {
		t.Fatalf("absent role/order must be null: %+v", second)
	}
	if !strings.Contains(m.JSON.ValueString(), `"units":[`) || !strings.Contains(m.JSON.ValueString(), `"deviceId":"`+dev.String()) {
		t.Fatalf("json must include units and deviceId: %s", m.JSON.ValueString())
	}
}

func TestMapSwitchStack_noDeviceNoUnits(t *testing.T) {
	m, err := mapSwitchStack(official.SwitchStack{Id: mustUUID(t, "88888888-8888-8888-8888-888888888888"), Name: "s"})
	if err != nil {
		t.Fatalf("mapSwitchStack: %v", err)
	}
	if !m.DeviceID.IsNull() {
		t.Fatalf("absent deviceId must be null, got %q", m.DeviceID.ValueString())
	}
	if m.Units == nil || len(m.Units) != 0 {
		t.Fatalf("expected an empty (non-nil) unit list, got %#v", m.Units)
	}
}

func TestMapDPIAndCountry(t *testing.T) {
	if got := mapDPIApplication(official.DPIApplication{Id: 7, Name: "netflix"}); got.ID.ValueInt64() != 7 || got.Name.ValueString() != "netflix" {
		t.Fatalf("dpi app = %+v", got)
	}
	if got := mapDPICategory(official.DPICategory{Id: 3, Name: "streaming"}); got.ID.ValueInt64() != 3 || got.Name.ValueString() != "streaming" {
		t.Fatalf("dpi cat = %+v", got)
	}
	if got := mapCountry(official.CountryDefinition{Code: "BR", Name: "Brazil"}); got.Code.ValueString() != "BR" || got.Name.ValueString() != "Brazil" {
		t.Fatalf("country = %+v", got)
	}
}

// TestStateRoundTrips validates every new data source's schema against a populated
// model, catching tfsdk tag / attribute mismatches without a live controller.
func TestStateRoundTrips(t *testing.T) {
	id := mustUUID(t, "99999999-9999-9999-9999-999999999999")

	clientItem, _ := mapClientOverview(official.ClientOverview{Id: id, Name: "c", Type: "WIRED"})
	deviceItem, _ := mapAdoptedDeviceOverview(official.AdoptedDeviceOverview{Id: id, Features: []official.AdoptedDeviceOverviewFeatures{"switching"}})
	deviceDetail, _ := mapAdoptedDeviceDetails(official.AdoptedDeviceDetails{Id: id})
	stats, _ := mapDeviceStatistics(official.LatestStatisticsForADevice{Uplink: &official.LatestStatisticsForADeviceUplinkInterface{RxRateBps: i64Ptr(1)}})
	clientDetail, _ := mapClientDetails(official.ClientDetails{Id: id, Name: "c", Type: "WIRED"})
	pending, _ := mapPendingDevice(official.DevicePendingAdoption{MacAddress: "a", Features: []official.DevicePendingAdoptionFeatures{"gateway"}})
	network, _ := mapNetworkOverview(official.NetworkOverview{Id: id, Name: "lan"})
	refs, _ := mapNetworkReferences(official.NetworkReferences{ReferenceResources: []official.NetworkReferenceResource{{ResourceType: "CLIENT", ReferenceCount: 1}}})
	lag, _ := mapLagDetails(mustLAG(t, `{"id":"`+id.String()+`","type":"LOCAL",`+lagMeta+`,"members":[{"deviceId":"`+id.String()+`","portIdxs":[1]}]}`))
	stackLag, _ := mapLagDetails(mustLAG(t, `{"id":"`+id.String()+`","type":"SWITCH_STACK",`+lagMeta+`,"members":[{"unitId":1,"unitMacAddress":"a","portIdxs":[1]}]}`))
	dom, _ := mapMcLagDomain(official.McLagDomain{Id: id, Name: "d"})
	role := official.SwitchStackUnitRole("MEMBER")
	stack, _ := mapSwitchStack(official.SwitchStack{Id: id, Name: "s", DeviceId: &id, Units: []official.SwitchStackUnit{{Id: 1, MacAddress: "a", Role: &role}, {Id: 2, MacAddress: "b"}}})

	cases := []struct {
		name  string
		ds    datasource.DataSource
		model any
	}{
		{"clients", NewClientsDataSource(), &clientsModel{Clients: []clientOverviewModel{clientItem}}},
		{"client", NewClientDataSource(), &clientDetail},
		{"official_devices", NewOfficialDevicesDataSource(), &officialDevicesModel{Devices: []adoptedDeviceOverviewModel{deviceItem}}},
		{"official_device", NewOfficialDeviceDataSource(), &deviceDetail},
		{"device_statistics", NewDeviceStatisticsDataSource(), &stats},
		{"pending_devices", NewPendingDevicesDataSource(), &pendingDevicesModel{Devices: []pendingDeviceOverviewModel{pending}}},
		{"networks", NewNetworksDataSource(), &networksModel{Networks: []networkOverviewModel{network}}},
		{"network_references", NewNetworkReferencesDataSource(), &refs},
		{"dpi_applications", NewDPIApplicationsDataSource(), &dpiApplicationsModel{Applications: []dpiCatalogEntry{mapDPIApplication(official.DPIApplication{Id: 1, Name: "a"})}}},
		{"dpi_categories", NewDPIApplicationCategoriesDataSource(), &dpiApplicationCategoriesModel{Categories: []dpiCatalogEntry{mapDPICategory(official.DPICategory{Id: 1, Name: "a"})}}},
		{"countries", NewCountriesDataSource(), &countriesModel{Countries: []countryEntry{mapCountry(official.CountryDefinition{Code: "BR", Name: "Brazil"})}}},
		{"radius_profiles", NewRadiusProfilesDataSource(), &radiusProfilesModel{Profiles: []radiusProfileOverviewModel{mapRadiusProfile(official.RadiusProfileOverview{Id: id, Name: "r"})}}},
		{"switch_lags", NewSwitchLAGsDataSource(), &switchLAGsModel{LAGs: []lagDetailsModel{lag, stackLag}}},
		{"switch_lag", NewSwitchLAGDataSource(), &switchLAGModel{ID: lag.ID, Type: lag.Type, MetadataOrigin: lag.MetadataOrigin, MembersJSON: lag.JSON, JSON: lag.JSON}},
		{"mclag_domains", NewMcLagDomainsDataSource(), &mcLagDomainsModel{Domains: []mcLagDomainItem{dom}}},
		{"mclag_domain", NewMcLagDomainDataSource(), &mcLagDomainModel{ID: dom.ID, Name: dom.Name, MetadataOrigin: dom.MetadataOrigin, JSON: dom.JSON}},
		{"switch_stacks", NewSwitchStacksDataSource(), &switchStacksModel{Stacks: []switchStackItem{stack}}},
		{"switch_stack", NewSwitchStackDataSource(), &switchStackModel{ID: stack.ID, Name: stack.Name, MetadataOrigin: stack.MetadataOrigin, DeviceID: stack.DeviceID, Units: stack.Units, JSON: stack.JSON}},
		{"controller_info", NewControllerInfoDataSource(), &controllerInfoModel{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertRoundTrip(t, schemaOf(t, tc.ds), tc.model)
		})
	}
}
