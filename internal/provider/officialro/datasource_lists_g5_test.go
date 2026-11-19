package officialro

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// The list data sources' Go models must line up with their schemas (G5).
func TestWANAndVPNServerListsModelMatchesSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := map[string]struct {
		ds    datasource.DataSource
		model any
	}{
		"unifi_wans": {NewWANsDataSource(), &wansModel{
			Site: types.StringValue("default"), Filter: types.StringNull(),
			WANs: []wanItem{{ID: types.StringValue("09c7bb9a-39e1-47c1-ba1c-df4c83a6aba9"), Name: types.StringValue("Internet Principal 1")}},
		}},
		"unifi_vpn_servers": {NewVPNServersDataSource(), &vpnServersModel{
			Site: types.StringValue("default"), Filter: types.StringNull(),
			Servers: []vpnServerItem{{
				ID: types.StringValue("663bc70c-85bc-415f-9316-19dbdc2353c0"), Name: types.StringValue("WireGuard Server 1"),
				Type: types.StringValue("WIREGUARD"), Enabled: types.BoolValue(true), MetadataOrigin: types.StringValue("USER_DEFINED"),
			}},
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var resp datasource.SchemaResponse
			tc.ds.Schema(ctx, datasource.SchemaRequest{}, &resp)
			require.False(t, resp.Diagnostics.HasError())
			state := tfsdk.State{Schema: resp.Schema, Raw: tftypes.NewValue(resp.Schema.Type().TerraformType(ctx), nil)}
			require.False(t, state.Set(ctx, tc.model).HasError())
		})
	}
}
