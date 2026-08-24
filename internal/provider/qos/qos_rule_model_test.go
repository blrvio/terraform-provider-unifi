package qos

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQOSRuleModelRoundTrip mirrors the live "Critical Apps Prioritization" rule
// through Merge (API -> state) and AsUnifiModel (state -> API), asserting the
// nested destination app_ids ([]int) and scalar fields survive intact.
func TestQOSRuleModelRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	in := &unifi.QOSRule{
		ID:              "q1",
		Name:            "Critical Apps Prioritization",
		Enabled:         true,
		Index:           10001,
		Objective:       "PRIORITIZE",
		DownloadBurst:   "OFF",
		UploadBurst:     "OFF",
		WANOrVPNNetwork: "net1",
		Schedule:        unifi.QOSRuleSchedule{Mode: "ALWAYS"},
		Source:          unifi.QOSRuleSource{MatchingTarget: "ANY"},
		Destination: unifi.QOSRuleDestination{
			MatchingTarget:   "APP",
			PortMatchingType: "ANY",
			AppIDs:           []int{393220, 1114124},
		},
	}

	m := &qosRuleModel{}
	require.False(t, m.Merge(ctx, in).HasError())

	out, diags := m.AsUnifiModel(ctx)
	require.False(t, diags.HasError())
	rule, ok := out.(*unifi.QOSRule)
	require.True(t, ok)

	assert.Equal(t, in.ID, rule.ID)
	assert.Equal(t, in.Name, rule.Name)
	assert.Equal(t, in.Enabled, rule.Enabled)
	assert.Equal(t, in.Index, rule.Index)
	assert.Equal(t, in.Objective, rule.Objective)
	assert.Equal(t, in.DownloadBurst, rule.DownloadBurst)
	assert.Equal(t, in.UploadBurst, rule.UploadBurst)
	assert.Equal(t, in.WANOrVPNNetwork, rule.WANOrVPNNetwork)
	assert.Equal(t, "ALWAYS", rule.Schedule.Mode)
	assert.Equal(t, "ANY", rule.Source.MatchingTarget)
	assert.Equal(t, "APP", rule.Destination.MatchingTarget)
	assert.Equal(t, "ANY", rule.Destination.PortMatchingType)
	assert.Equal(t, []int{393220, 1114124}, rule.Destination.AppIDs)
}
