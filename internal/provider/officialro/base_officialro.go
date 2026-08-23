// Package officialro contains read-only Terraform Plugin Framework data sources
// backed by the UniFi Official API (integration/v1). The entities exposed here
// (WANs, VPN servers, site-to-site tunnels, switch LAGs, device tags) are
// list-only on the Official API and are not writable, so they are surfaced as
// data sources rather than resources.
package officialro

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

// officialBase carries the shared Framework wiring (client + validators) for the
// Official-API read-only data sources and implements base.Resource so that
// base.ConfigureDatasource can inject them.
type officialBase struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func (d *officialBase) SetClient(client *base.Client) { d.client = client }

func (d *officialBase) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *officialBase) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *officialBase) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

// prepare runs the common Read preamble: it verifies the provider is configured,
// gates on the minimum controller version that exposes the Official API, and
// resolves the requested site name (falling back to the provider default) to its
// Official-API UUID. It returns the resolved UUID and the site name that was
// actually used so the caller can echo it back into state.
func (d *officialBase) prepare(ctx context.Context, siteName string) (siteID uuid.UUID, resolvedName string, diags diag.Diagnostics) {
	diags.Append(base.CheckConfigured(d.client)...)
	if diags.HasError() {
		return uuid.UUID{}, "", diags
	}
	diags.Append(d.RequireMinVersion(base.ControllerVersionOfficialAPI.String())...)
	if diags.HasError() {
		return uuid.UUID{}, "", diags
	}
	resolvedName = siteName
	if resolvedName == "" {
		resolvedName = d.client.Site
	}
	id, resolveDiags := d.client.ResolveSiteUUID(ctx, siteName)
	diags.Append(resolveDiags...)
	return id, resolvedName, diags
}

// siteAttribute is the standard Optional+Computed site attribute for the
// read-only data sources. It cannot reuse types.SiteAttribute because that
// returns a resource/schema attribute, not a datasource/schema one.
func siteAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: "The name of the UniFi site to query. If not specified, the provider's default site is used.",
		Optional:            true,
		Computed:            true,
	}
}

// findByID returns the first element of items whose id (as reported by idOf)
// equals the requested id, and whether such an element was found. It is the
// list-and-match selection primitive shared by the data sources that have no
// dedicated Get endpoint.
func findByID[T any](items []T, id string, idOf func(T) string) (T, bool) {
	for _, item := range items {
		if idOf(item) == id {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// uuidsToStrings converts a slice of Official-API UUIDs to their canonical
// string form, preserving order. A nil input yields a nil slice.
func uuidsToStrings(ids []uuid.UUID) []string {
	if ids == nil {
		return nil
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}
