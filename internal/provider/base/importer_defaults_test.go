package base

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A1: Terraform-only attributes get their defaults on import instead of null.
func TestImportSiteAndIDWithDefaults(t *testing.T) {
	t.Parallel()
	r := &schema.Resource{Schema: map[string]*schema.Schema{
		"site":           {Type: schema.TypeString, Optional: true, Computed: true},
		"allow_existing": {Type: schema.TypeBool, Optional: true, Default: true},
		"destructive":    {Type: schema.TypeBool, Optional: true, Default: true},
	}}
	d := r.TestResourceData()
	d.SetId("lab:abc")

	out, err := ImportSiteAndIDWithDefaults(map[string]interface{}{"allow_existing": true})(context.Background(), d, nil)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "abc", out[0].Id())
	assert.Equal(t, "lab", out[0].Get("site"))
	assert.Equal(t, true, out[0].Get("allow_existing"))

	raw := out[0].State().Attributes
	assert.Equal(t, "true", raw["allow_existing"])
	_, set := raw["destructive"]
	assert.False(t, set, "attributes not listed are left unset")
}
