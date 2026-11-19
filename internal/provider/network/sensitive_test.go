package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// G2: the PPPoE password must never be rendered in plans or generated config.
func TestWANPasswordSensitive(t *testing.T) {
	t.Parallel()
	assert.True(t, ResourceNetwork().Schema["x_wan_password"].Sensitive, "resource")
	assert.True(t, DataNetwork().Schema["x_wan_password"].Sensitive, "data source")
}

func TestValidateWANPasswordAnchored(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		value   string
		wantErr bool
	}{
		"plain":            {value: "s3cr3t-P4ss", wantErr: false},
		"contains space":   {value: "two words", wantErr: true},
		"contains quote":   {value: `pa"ss`, wantErr: true},
		"contains apostr.": {value: "pa'ss", wantErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, errs := validateWANPassword(tc.value, "x_wan_password")
			assert.Equal(t, tc.wantErr, len(errs) > 0, "errors: %v", errs)
		})
	}
}
