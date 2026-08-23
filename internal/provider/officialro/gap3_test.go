package officialro

import (
	"context"
	"testing"
	"time"

	"github.com/blrvio/go-unifi/v10/unifi/official"
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

func TestMapLagDetails_members(t *testing.T) {
	dev := mustUUID(t, "44444444-4444-4444-4444-444444444444")
	m, err := mapLagDetails(official.LAGDetails{
		Id:      mustUUID(t, "55555555-5555-5555-5555-555555555555"),
		Type:    "LOCAL",
		Members: []official.LagMember{{DeviceId: dev, PortIdxs: []int32{1, 2, 3}}},
	})
	if err != nil {
		t.Fatalf("mapLagDetails: %v", err)
	}
	if len(m.Members) != 1 || m.Members[0].DeviceID.ValueString() != dev.String() {
		t.Fatalf("members = %+v", m.Members)
	}
	if len(m.Members[0].PortIdxs) != 3 || m.Members[0].PortIdxs[2] != 3 {
		t.Fatalf("port idxs = %v", m.Members[0].PortIdxs)
	}
}

func TestMapSwitchStack_members(t *testing.T) {
	a := mustUUID(t, "66666666-6666-6666-6666-666666666666")
	b := mustUUID(t, "77777777-7777-7777-7777-777777777777")
	m, err := mapSwitchStack(official.SwitchStack{
		Id:      mustUUID(t, "88888888-8888-8888-8888-888888888888"),
		Name:    "stack-1",
		Members: []official.SwitchStackMember{{DeviceId: a}, {DeviceId: b}},
	})
	if err != nil {
		t.Fatalf("mapSwitchStack: %v", err)
	}
	if len(m.MemberDeviceIDs) != 2 || m.MemberDeviceIDs[0] != a.String() || m.MemberDeviceIDs[1] != b.String() {
		t.Fatalf("member ids = %v", m.MemberDeviceIDs)
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
	lag, _ := mapLagDetails(official.LAGDetails{Id: id, Type: "LOCAL", Members: []official.LagMember{{DeviceId: id, PortIdxs: []int32{1}}}})
	dom, _ := mapMcLagDomain(official.McLagDomain{Id: id, Name: "d"})
	stack, _ := mapSwitchStack(official.SwitchStack{Id: id, Name: "s", Members: []official.SwitchStackMember{{DeviceId: id}}})

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
		{"switch_lags", NewSwitchLAGsDataSource(), &switchLAGsModel{LAGs: []lagDetailsModel{lag}}},
		{"mclag_domains", NewMcLagDomainsDataSource(), &mcLagDomainsModel{Domains: []mcLagDomainItem{dom}}},
		{"mclag_domain", NewMcLagDomainDataSource(), &mcLagDomainModel{ID: dom.ID, Name: dom.Name, MetadataOrigin: dom.MetadataOrigin, JSON: dom.JSON}},
		{"switch_stacks", NewSwitchStacksDataSource(), &switchStacksModel{Stacks: []switchStackItem{stack}}},
		{"switch_stack", NewSwitchStackDataSource(), &switchStackModel{ID: stack.ID, Name: stack.Name, MetadataOrigin: stack.MetadataOrigin, MemberDeviceIDs: stack.MemberDeviceIDs, JSON: stack.JSON}},
		{"controller_info", NewControllerInfoDataSource(), &controllerInfoModel{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertRoundTrip(t, schemaOf(t, tc.ds), tc.model)
		})
	}
}
