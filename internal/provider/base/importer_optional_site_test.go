package base

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestImportIDWithOptionalSite(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		id, defaultSite  string
		wantID, wantSite string
		wantErr          bool
	}{
		"site and id":             {id: "lab:abc", defaultSite: "default", wantID: "abc", wantSite: "lab"},
		"bare id uses default":    {id: "abc", defaultSite: "default", wantID: "abc", wantSite: "default"},
		"bare id without default": {id: "abc", wantErr: true},
		"empty":                   {id: "", defaultSite: "default", wantErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := &resource.ImportStateResponse{}
			id, site := ImportIDWithOptionalSite(resource.ImportStateRequest{ID: tc.id}, resp, tc.defaultSite)
			assert.Equal(t, tc.wantErr, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			if !tc.wantErr {
				assert.Equal(t, tc.wantID, id)
				assert.Equal(t, tc.wantSite, site)
			}
		})
	}
}
