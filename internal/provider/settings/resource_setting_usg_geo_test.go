package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUsgGeoSchemaValid ensures the resource schema builds without diagnostics.
func TestUsgGeoSchemaValid(t *testing.T) {
	var resp resource.SchemaResponse
	NewUsgGeoResource().(interface {
		Schema(context.Context, resource.SchemaRequest, *resource.SchemaResponse)
	}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "schema build returned diagnostics: %v", resp.Diagnostics)

	_, ok := resp.Schema.Attributes["ip_filtering"]
	assert.True(t, ok, "expected an ip_filtering attribute")
}

// TestUsgGeoMergeRoundTrip verifies Merge maps the nested ip_filtering object
// from the SDK model back into state (countries string -> list), which is what
// makes a second apply converge to an empty plan.
func TestUsgGeoMergeRoundTrip(t *testing.T) {
	ctx := context.Background()

	m := &usgGeoModel{}
	diags := m.Merge(ctx, &unifi.SettingUsgGeo{
		ID: "abc",
		IPFiltering: unifi.SettingUsgGeoIPFiltering{
			Action:           "block",
			Countries:        "RU,CN",
			Enabled:          true,
			TrafficDirection: "both",
		},
	})
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)
	require.False(t, m.IPFiltering.IsNull(), "ip_filtering must be populated")

	var ipf usgGeoIPFilteringModel
	require.False(t, m.IPFiltering.As(ctx, &ipf, basetypes.ObjectAsOptions{}).HasError())
	assert.Equal(t, "block", ipf.Action.ValueString())
	assert.True(t, ipf.Enabled.ValueBool())
	assert.Equal(t, "both", ipf.TrafficDirection.ValueString())
	assert.Len(t, ipf.Countries.Elements(), 2)

	// And AsUnifiModel converts the list back to the comma-joined string.
	body, aDiags := m.AsUnifiModel(ctx)
	require.False(t, aDiags.HasError(), "AsUnifiModel diagnostics: %v", aDiags)
	geo, ok := body.(*unifi.SettingUsgGeo)
	require.True(t, ok)
	assert.Equal(t, "RU,CN", geo.IPFiltering.Countries)
	assert.Equal(t, types.StringValue("abc").ValueString(), geo.ID)
}
