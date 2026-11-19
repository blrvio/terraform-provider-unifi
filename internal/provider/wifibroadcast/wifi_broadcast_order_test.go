package wifibroadcast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

func TestSortClientFilteringPolicy(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		in, want ut.JSONValue
	}{
		"null":    {ut.JSONNull(), ut.JSONNull()},
		"unknown": {ut.JSONUnknown(), ut.JSONUnknown()},
		"sorted from controller order": {
			ut.JSONStringValue(`{"action":"ALLOW","macAddressFilter":["44:3e:07:59:0d:5c","00:ae:f7:ac:56:ac","30:48:7d:93:8c:20"]}`),
			ut.JSONStringValue(`{"action":"ALLOW","macAddressFilter":["00:ae:f7:ac:56:ac","30:48:7d:93:8c:20","44:3e:07:59:0d:5c"]}`),
		},
		"no list": {
			ut.JSONStringValue(`{"action":"BLOCK"}`),
			ut.JSONStringValue(`{"action":"BLOCK"}`),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, diags := sortClientFilteringPolicy(tc.in)
			require.False(t, diags.HasError())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSortBlackoutSchedule(t *testing.T) {
	t.Parallel()
	// Order as returned by UniFi Network 11.0.99 after a PUT in SUN..SAT order.
	in := ut.JSONStringValue(`{"days":[` +
		`{"day":"MON","type":"ALL_DAY"},{"day":"WED","type":"ALL_DAY"},` +
		`{"day":"FRI","type":"TIME_RANGE","timeRanges":[{"startTime":"00:00","endTime":"16:00"}]},` +
		`{"day":"SUN","type":"TIME_RANGE","timeRanges":[{"startTime":"22:00","endTime":"23:59"},{"startTime":"02:00","endTime":"09:00"}]},` +
		`{"day":"SAT","type":"TIME_RANGE","timeRanges":[{"startTime":"03:00","endTime":"09:00"}]},` +
		`{"day":"THU","type":"ALL_DAY"},{"day":"TUE","type":"ALL_DAY"}]}`)
	want := ut.JSONStringValue(`{"days":[` +
		`{"day":"SUN","timeRanges":[{"endTime":"09:00","startTime":"02:00"},{"endTime":"23:59","startTime":"22:00"}],"type":"TIME_RANGE"},` +
		`{"day":"MON","type":"ALL_DAY"},{"day":"TUE","type":"ALL_DAY"},{"day":"WED","type":"ALL_DAY"},{"day":"THU","type":"ALL_DAY"},` +
		`{"day":"FRI","timeRanges":[{"endTime":"16:00","startTime":"00:00"}],"type":"TIME_RANGE"},` +
		`{"day":"SAT","timeRanges":[{"endTime":"09:00","startTime":"03:00"}],"type":"TIME_RANGE"}]}`)
	got, diags := sortBlackoutSchedule(in)
	require.False(t, diags.HasError())
	assert.Equal(t, want, got)

	nullGot, diags := sortBlackoutSchedule(ut.JSONNull())
	require.False(t, diags.HasError())
	assert.True(t, nullGot.IsNull())
}
