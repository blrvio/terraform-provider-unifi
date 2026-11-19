package radius

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

// defaultProfile is the built-in "Default" profile as a UniFi Network 11.0.99 console
// returns it: gateway as server (no IP), no interim interval, system attributes (G10).
func defaultProfile() *unifi.RADIUSProfile {
	return &unifi.RADIUSProfile{
		ID: "p1", SiteID: "s1", Name: "Default", HiddenID: "Default", NoDelete: true,
		UseUsgAuthServer: true,
		AuthServers:      []unifi.RADIUSProfileAuthServers{{Port: 1812, XSecret: "s"}},
		AcctServers:      []unifi.RADIUSProfileAcctServers{{Port: 1813, XSecret: "s"}},
	}
}

func TestRadiusDefaultProfileReadAndValidate(t *testing.T) {
	t.Parallel()
	r := ResourceRadiusProfile()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{})
	require.False(t, resourceRadiusProfileSetResourceData(defaultProfile(), d, "default").HasError())
	assert.Equal(t, 3600, d.Get("interim_update_interval"))
	assert.Equal(t, "", d.Get("auth_server.0.ip"))

	// The matching configuration validates: no IP, empty vlan_wlan_mode.
	authElem, ok := r.Schema["auth_server"].Elem.(*schema.Resource)
	require.True(t, ok)
	ipSchema := authElem.Schema["ip"]
	assert.False(t, ipSchema.Required)
	_, errs := ipSchema.ValidateFunc("", "ip")
	assert.Empty(t, errs)
	_, errs = ipSchema.ValidateFunc("not-an-ip", "ip")
	assert.NotEmpty(t, errs)
	_, errs = r.Schema["vlan_wlan_mode"].ValidateFunc("", "vlan_wlan_mode")
	assert.Empty(t, errs)
}

func TestRadiusProfileUpdateKeepsSystemAttributes(t *testing.T) {
	t.Parallel()
	managed := &unifi.RADIUSProfile{
		Name: "Default", UseUsgAuthServer: true, InterimUpdateInterval: 3600,
		AuthServers: []unifi.RADIUSProfileAuthServers{{Port: 1812, XSecret: "new"}},
	}
	got := overlayManagedRadiusProfile(defaultProfile(), managed)
	assert.True(t, got.NoDelete)
	assert.Equal(t, "Default", got.HiddenID)
	assert.Equal(t, "s1", got.SiteID)
	assert.Equal(t, 0, got.InterimUpdateInterval, "an unset interval stays unset while config keeps the default")
	assert.Equal(t, "new", got.AuthServers[0].XSecret)
}

type fakeRadiusClient struct {
	unifi.Client
	deleted bool
}

func (f *fakeRadiusClient) GetRADIUSProfile(context.Context, string, string) (*unifi.RADIUSProfile, error) {
	return defaultProfile(), nil
}

func (f *fakeRadiusClient) DeleteRADIUSProfile(context.Context, string, string) error {
	f.deleted = true
	return nil
}

func TestRadiusProfileDeleteRefusesSystemProfile(t *testing.T) {
	t.Parallel()
	fake := &fakeRadiusClient{}
	d := ResourceRadiusProfile().TestResourceData()
	d.SetId("p1")
	diags := resourceRadiusProfileDelete(context.Background(), d, &base.Client{Client: fake, Site: "default"})
	assert.True(t, diags.HasError())
	assert.False(t, fake.deleted, "must not call DELETE on a no-delete profile")
}
