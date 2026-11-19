package wifibroadcast

import (
	"encoding/json"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi/official"
)

// wifiBroadcastReadOnlyKeys are present in WifiBroadcastDetails but not accepted
// in the create/update body.
var wifiBroadcastReadOnlyKeys = []string{"id", "metadata"}

// wifiBroadcastManagedOptionalKeys are the optional JSON objects this resource
// owns. They are dropped from the current document before the plan is overlaid,
// so a null attribute in the plan removes the setting instead of silently
// keeping the controller's value.
var wifiBroadcastManagedOptionalKeys = []string{
	"network",
	"basicDataRateKbpsByFrequencyGHz",
	"blackoutScheduleConfiguration",
	"broadcastingDeviceFilter",
	"clientFilteringPolicy",
	"mdnsProxyConfiguration",
	"multicastFilteringPolicy",
}

// overlayWifiBroadcast builds the PUT body for an update as a read-modify-write:
// the controller's current document (minus read-only and managed-optional keys)
// carries every field this resource does not model — e.g. the STANDARD variant's
// arpProxyEnabled, bssTransitionEnabled, advertiseDeviceName and
// broadcastingFrequenciesGHz, which the controller rejects when absent — and the
// plan's modeled fields are laid on top.
func overlayWifiBroadcast(current *official.WifiBroadcastDetails, plan *official.WifiBroadcastCreateOrUpdate) (*official.WifiBroadcastCreateOrUpdate, error) {
	raw, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("marshaling current WiFi broadcast: %w", err)
	}
	doc := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("decoding current WiFi broadcast: %w", err)
	}
	for _, k := range wifiBroadcastReadOnlyKeys {
		delete(doc, k)
	}
	for _, k := range wifiBroadcastManagedOptionalKeys {
		delete(doc, k)
	}
	if raw, err = json.Marshal(doc); err != nil {
		return nil, fmt.Errorf("re-encoding current WiFi broadcast: %w", err)
	}

	// UnmarshalJSON keeps the whole document in the body's union, so fields
	// without a top-level struct field survive the round trip.
	merged := &official.WifiBroadcastCreateOrUpdate{}
	if err := json.Unmarshal(raw, merged); err != nil {
		return nil, fmt.Errorf("building WiFi broadcast update body: %w", err)
	}

	merged.Name = plan.Name
	merged.Type = plan.Type
	merged.Enabled = plan.Enabled
	merged.HideName = plan.HideName
	merged.ClientIsolationEnabled = plan.ClientIsolationEnabled
	merged.UapsdEnabled = plan.UapsdEnabled
	merged.MulticastToUnicastConversionEnabled = plan.MulticastToUnicastConversionEnabled
	merged.Channel2gLockedTo6 = plan.Channel2gLockedTo6
	merged.DtimPeriod2gLockedTo3 = plan.DtimPeriod2gLockedTo3
	merged.SecurityConfiguration = plan.SecurityConfiguration
	merged.Network = plan.Network
	merged.BasicDataRateKbpsByFrequencyGHz = plan.BasicDataRateKbpsByFrequencyGHz
	merged.BlackoutScheduleConfiguration = plan.BlackoutScheduleConfiguration
	merged.BroadcastingDeviceFilter = plan.BroadcastingDeviceFilter
	merged.ClientFilteringPolicy = plan.ClientFilteringPolicy
	merged.MdnsProxyConfiguration = plan.MdnsProxyConfiguration
	merged.MulticastFilteringPolicy = plan.MulticastFilteringPolicy
	return merged, nil
}
