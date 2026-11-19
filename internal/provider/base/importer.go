package base

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ImportSiteAndIDWithDefaults returns an SDKv2 importer that splits `site:id` like
// ImportSiteAndID and writes defaults for Terraform-only attributes the controller
// cannot report (SDKv2 does not apply schema defaults on import, so they would
// otherwise read as null and plan an update right after the import).
func ImportSiteAndIDWithDefaults(defaults map[string]interface{}) schema.StateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
		out, err := ImportSiteAndID(ctx, d, meta)
		if err != nil {
			return nil, err
		}
		if err := SetImportDefaults(d, defaults); err != nil {
			return nil, err
		}
		return out, nil
	}
}

// SetImportDefaults writes Terraform-only attribute defaults during an SDKv2 import.
func SetImportDefaults(d *schema.ResourceData, defaults map[string]interface{}) error {
	for k, v := range defaults {
		if err := d.Set(k, v); err != nil {
			return fmt.Errorf("setting import default %s: %w", k, err)
		}
	}
	return nil
}

// for Terraform Plugin SDK v2.
func ImportSiteAndID(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	if id := d.Id(); strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		d.SetId(importParts[1])
		if err := d.Set("site", importParts[0]); err != nil {
			return nil, err
		}
	}
	return []*schema.ResourceData{d}, nil
}

// for Terraform Plugin Framework.
func ImportIDWithSite(req resource.ImportStateRequest, resp *resource.ImportStateResponse) (string, string) {
	id := req.ID
	if id == "" {
		resp.Diagnostics.AddError("Invalid ID", "ID is required")
		return "", ""
	}

	if strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		if len(importParts) == 2 {
			return importParts[1], importParts[0]
		}
		resp.Diagnostics.AddError("Invalid ID", "ID contains too many colon-separated parts. Format should be 'site:id'")
		return "", ""
	}
	resp.Diagnostics.AddError("Invalid ID", "ID does not contain site part. Format should be 'site:id'")
	return id, ""
}

// ImportIDWithOptionalSite accepts `site:id` or a bare `id`; a bare id imports into
// defaultSite (the provider's configured site).
func ImportIDWithOptionalSite(req resource.ImportStateRequest, resp *resource.ImportStateResponse, defaultSite string) (string, string) {
	if req.ID != "" && !strings.Contains(req.ID, ":") {
		if defaultSite == "" {
			resp.Diagnostics.AddError("Invalid ID", "ID does not contain a site part and the provider has no default site. Format should be 'site:id'")
			return "", ""
		}
		return req.ID, defaultSite
	}
	return ImportIDWithSite(req, resp)
}
