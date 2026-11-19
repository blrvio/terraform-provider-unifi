package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// G12: configuring only api_key must not trigger the username/password deprecation warning.
func TestProviderNoUserPassWarningWithAPIKeyOnly(t *testing.T) {
	t.Setenv("UNIFI_USERNAME", "")
	t.Setenv("UNIFI_PASSWORD", "")
	p := New("test")()
	diags := p.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
		"api_key": "k",
		"api_url": "https://10.0.0.1",
	}))
	for _, d := range diags {
		assert.NotContains(t, d.Summary+d.Detail, "deprecated", "unexpected diagnostic: %+v", d)
	}

	diags = p.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
		"api_key":  "k",
		"username": "admin",
	}))
	found := false
	for _, d := range diags {
		if d.Summary == "Argument is deprecated" {
			found = true
		}
	}
	assert.True(t, found, "setting username must still warn: %+v", diags)
}
