package wifibroadcast

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// weekdayOrder is the order blackout schedule days are stored in state (SUN..SAT).
var weekdayOrder = map[string]int{"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6}

// The controller returns some collections in arbitrary (hash) order that changes
// between writes: the client filter MAC list (a set) and the blackout schedule
// days. Storing them in a fixed order keeps a configuration written in that order
// free of perpetual diffs and of "inconsistent result after apply" errors.

// sortClientFilteringPolicy sorts macAddressFilter ascending.
func sortClientFilteringPolicy(v ut.JSONValue) (ut.JSONValue, diag.Diagnostics) {
	return reorderJSON("client_filtering_policy", v, func(doc map[string]any) {
		macs, ok := doc["macAddressFilter"].([]any)
		if !ok {
			return
		}
		sort.SliceStable(macs, func(i, j int) bool {
			a, _ := macs[i].(string)
			b, _ := macs[j].(string)
			return strings.ToLower(a) < strings.ToLower(b)
		})
	})
}

// sortBlackoutSchedule sorts days SUN..SAT and each day's timeRanges by startTime.
func sortBlackoutSchedule(v ut.JSONValue) (ut.JSONValue, diag.Diagnostics) {
	return reorderJSON("blackout_schedule", v, func(doc map[string]any) {
		days, ok := doc["days"].([]any)
		if !ok {
			return
		}
		sort.SliceStable(days, func(i, j int) bool {
			return dayIndex(days[i]) < dayIndex(days[j])
		})
		for _, d := range days {
			day, ok := d.(map[string]any)
			if !ok {
				continue
			}
			ranges, ok := day["timeRanges"].([]any)
			if !ok {
				continue
			}
			sort.SliceStable(ranges, func(i, j int) bool {
				return startTime(ranges[i]) < startTime(ranges[j])
			})
		}
	})
}

func dayIndex(v any) int {
	if m, ok := v.(map[string]any); ok {
		if s, ok := m["day"].(string); ok {
			if i, ok := weekdayOrder[s]; ok {
				return i
			}
		}
	}
	return len(weekdayOrder)
}

func startTime(v any) string {
	if m, ok := v.(map[string]any); ok {
		s, _ := m["startTime"].(string)
		return s
	}
	return ""
}

// reorderJSON decodes a JSON object attribute, lets fn reorder it in place and
// re-encodes it canonically. Null, unknown and non-object values pass through.
func reorderJSON(attr string, v ut.JSONValue, fn func(map[string]any)) (ut.JSONValue, diag.Diagnostics) {
	if v.IsNull() || v.IsUnknown() {
		return v, nil
	}
	dec := json.NewDecoder(bytes.NewReader([]byte(v.ValueString())))
	dec.UseNumber()
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil || doc == nil {
		return v, nil // not an object: leave the value untouched
	}
	fn(doc)
	return ut.MarshalJSONValue(attr, doc)
}
