package device

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// G16: config generated from an imported port profile carries the controller's
// "unset" values (0 / ""); it must validate, and level/rate only conflict when both set.
func TestPortProfileGeneratedConfigValidates(t *testing.T) {
	t.Parallel()
	r := ResourcePortProfile()
	require.NoError(t, r.InternalValidate(nil, true))

	generated := map[string]interface{}{
		"name": "pp-disabled", "speed": 0, "egress_rate_limit_kbps": 0,
		"stormctrl_bcast_level": 0, "stormctrl_bcast_rate": 0,
		"stormctrl_mcast_level": 0, "stormctrl_mcast_rate": 0,
		"stormctrl_ucast_level": 0, "stormctrl_ucast_rate": 0,
		"stormctrl_type": "", "poe_mode": "", "tagged_vlan_mgmt": "", "forward": "",
	}
	diags := r.Validate(terraform.NewResourceConfigRaw(generated))
	assert.False(t, diags.HasError(), "%v", diags)

	_, err := r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(generated), nil)
	assert.NoError(t, err)

	both := map[string]interface{}{"name": "pp", "stormctrl_bcast_level": 10, "stormctrl_bcast_rate": 100}
	_, err = r.Diff(context.Background(), nil, terraform.NewResourceConfigRaw(both), nil)
	assert.ErrorContains(t, err, "cannot both be set")

	assert.True(t, r.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{"name": "pp", "speed": 33})).HasError(), "invalid speed still rejected")
}

func TestDevicePortOverrideGeneratedConfigValidates(t *testing.T) {
	t.Parallel()
	r := ResourceDevice()
	cfg := map[string]interface{}{
		"mac": "aa:bb:cc:dd:ee:ff",
		"port_override": []interface{}{map[string]interface{}{
			"number": 1, "aggregate_num_ports": 0, "op_mode": "", "poe_mode": "",
		}},
	}
	diags := r.Validate(terraform.NewResourceConfigRaw(cfg))
	assert.False(t, diags.HasError(), "%v", diags)
}
