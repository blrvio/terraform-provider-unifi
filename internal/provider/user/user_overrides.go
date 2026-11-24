package user

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// userOverrides holds the configured AP-lock and network-override values; nil means "not in the configuration".
// The controller's user PUT is a full replace, so each nil one is copied from the controller instead of reset.
type userOverrides struct {
	FixedApEnabled                *bool
	FixedApMAC                    *string
	VirtualNetworkOverrideEnabled *bool
	VirtualNetworkOverrideID      *string
}

// rawConfigNonNull reports whether name is present in the raw configuration, including false and "",
// which utils.IsRawConfigSet treats as unset.
func rawConfigNonNull(raw cty.Value, name string) bool {
	if raw.IsNull() || !raw.IsKnown() || !raw.Type().HasAttribute(name) {
		return false
	}
	return !raw.GetAttr(name).IsNull()
}

func userOverridesFromConfig(d *schema.ResourceData) userOverrides {
	raw := d.GetRawConfig()
	var o userOverrides
	if rawConfigNonNull(raw, "fixed_ap_enabled") {
		v, _ := d.Get("fixed_ap_enabled").(bool)
		o.FixedApEnabled = &v
	}
	if rawConfigNonNull(raw, "fixed_ap_mac") {
		v, _ := d.Get("fixed_ap_mac").(string)
		o.FixedApMAC = &v
	}
	if rawConfigNonNull(raw, "virtual_network_override_enabled") {
		v, _ := d.Get("virtual_network_override_enabled").(bool)
		o.VirtualNetworkOverrideEnabled = &v
	}
	if rawConfigNonNull(raw, "virtual_network_override_id") {
		v, _ := d.Get("virtual_network_override_id").(string)
		o.VirtualNetworkOverrideID = &v
	}
	return o
}

// applyUserOverrides sets the override fields of req: configured values win, everything else comes from
// current (nil on a fresh create). Setting only the MAC or network ID turns the matching flag on.
func applyUserOverrides(req *unifi.User, o userOverrides, current *unifi.User) {
	if current != nil {
		req.FixedApEnabled = current.FixedApEnabled
		req.FixedApMAC = current.FixedApMAC
		req.VirtualNetworkOverrideEnabled = current.VirtualNetworkOverrideEnabled
		req.VirtualNetworkOverrideID = current.VirtualNetworkOverrideID
	}

	if o.FixedApMAC != nil {
		req.FixedApMAC = *o.FixedApMAC
		req.FixedApEnabled = *o.FixedApMAC != ""
	}
	if o.FixedApEnabled != nil {
		req.FixedApEnabled = *o.FixedApEnabled
	}

	if o.VirtualNetworkOverrideID != nil {
		req.VirtualNetworkOverrideID = *o.VirtualNetworkOverrideID
		req.VirtualNetworkOverrideEnabled = *o.VirtualNetworkOverrideID != ""
	}
	if o.VirtualNetworkOverrideEnabled != nil {
		req.VirtualNetworkOverrideEnabled = *o.VirtualNetworkOverrideEnabled
	}
}

// resourceUserCustomizeDiff marks an unconfigured flag as unknown when its MAC/network ID changes, because
// applyUserOverrides derives the flag from it.
func resourceUserCustomizeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	raw := d.GetRawConfig()
	for flag, value := range map[string]string{
		"fixed_ap_enabled":                 "fixed_ap_mac",
		"virtual_network_override_enabled": "virtual_network_override_id",
	} {
		if !rawConfigNonNull(raw, flag) && rawConfigNonNull(raw, value) && d.HasChange(value) {
			if err := d.SetNewComputed(flag); err != nil {
				return fmt.Errorf("marking %s as computed: %w", flag, err)
			}
		}
	}
	return nil
}

func userOverridesToState(resp *unifi.User) map[string]interface{} {
	return map[string]interface{}{
		"fixed_ap_enabled":                 resp.FixedApEnabled,
		"fixed_ap_mac":                     resp.FixedApMAC,
		"virtual_network_override_enabled": resp.VirtualNetworkOverrideEnabled,
		"virtual_network_override_id":      resp.VirtualNetworkOverrideID,
	}
}
