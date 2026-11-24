package user

import (
	"reflect"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/hashicorp/go-cty/cty"
)

func ptr[T any](v T) *T { return &v }

func TestApplyUserOverrides(t *testing.T) {
	t.Parallel()

	console := &unifi.User{
		FixedApEnabled:                true,
		FixedApMAC:                    "aa:bb:cc:dd:ee:ff",
		VirtualNetworkOverrideEnabled: true,
		VirtualNetworkOverrideID:      "net-cam",
	}

	tests := []struct {
		name    string
		o       userOverrides
		current *unifi.User
		want    unifi.User
	}{
		{
			name:    "unconfigured keeps the console values",
			current: console,
			want:    *console,
		},
		{
			name: "unconfigured on a fresh create sends zero values",
			want: unifi.User{},
		},
		{
			name:    "network id alone enables the override",
			o:       userOverrides{VirtualNetworkOverrideID: ptr("net-iot")},
			current: &unifi.User{},
			want:    unifi.User{VirtualNetworkOverrideEnabled: true, VirtualNetworkOverrideID: "net-iot"},
		},
		{
			name:    "explicit false disables the override and keeps the AP lock",
			o:       userOverrides{VirtualNetworkOverrideEnabled: ptr(false)},
			current: console,
			want: unifi.User{
				FixedApEnabled: true, FixedApMAC: "aa:bb:cc:dd:ee:ff",
				VirtualNetworkOverrideEnabled: false, VirtualNetworkOverrideID: "net-cam",
			},
		},
		{
			name:    "empty MAC removes the AP lock",
			o:       userOverrides{FixedApMAC: ptr("")},
			current: console,
			want: unifi.User{
				FixedApEnabled: false, FixedApMAC: "",
				VirtualNetworkOverrideEnabled: true, VirtualNetworkOverrideID: "net-cam",
			},
		},
		{
			name:    "explicit flag wins over the derived one",
			o:       userOverrides{FixedApMAC: ptr("11:22:33:44:55:66"), FixedApEnabled: ptr(false)},
			current: &unifi.User{},
			want:    unifi.User{FixedApEnabled: false, FixedApMAC: "11:22:33:44:55:66"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := &unifi.User{Name: "client"}
			applyUserOverrides(req, tt.o, tt.current)
			tt.want.Name = "client"
			if !reflect.DeepEqual(*req, tt.want) {
				t.Errorf("got %+v, want %+v", *req, tt.want)
			}
		})
	}
}

func TestRawConfigNonNull(t *testing.T) {
	t.Parallel()

	raw := cty.ObjectVal(map[string]cty.Value{
		"set_false": cty.False,
		"set_empty": cty.StringVal(""),
		"null":      cty.NullVal(cty.String),
	})
	for name, want := range map[string]bool{"set_false": true, "set_empty": true, "null": false, "missing": false} {
		if got := rawConfigNonNull(raw, name); got != want {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
	if rawConfigNonNull(cty.NullVal(cty.EmptyObject), "x") {
		t.Error("null config must report false")
	}
	if rawConfigNonNull(cty.UnknownVal(cty.EmptyObject), "x") {
		t.Error("unknown config must report false")
	}
}
