package base

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Official returns the Official UniFi API (integration/v1) surface of the
// underlying go-unifi client. Availability is gated by go-unifi: every operation
// returns unifi.ErrOfficialAPIUnavailable (controller < 10.1.78, old-style
// controller, or non API-key auth) or unifi.ErrOfficialAPIDisabled. Turn those
// into user-facing diagnostics with OfficialAPIErrorDiagnostics, and gate at plan
// time with RequireOfficialAPI.
func (c *Client) Official() official.Client {
	return c.Client.Official()
}

// RequireOfficialAPI reports a friendly diagnostic when the controller is too old
// to expose the Official API. It is a cheap, version-only pre-check (using the
// capability resolved once at configuration time) intended for ModifyPlan, so
// users get an actionable error before any API call. The authoritative gate still
// runs inside every Official operation and is surfaced via
// OfficialAPIErrorDiagnostics.
func (c *Client) RequireOfficialAPI() diag.Diagnostics {
	var diags diag.Diagnostics
	if !c.OfficialAvailable {
		current := "unknown"
		if c.Version != nil {
			current = c.Version.String()
		}
		diags.AddError(
			"UniFi Official API not available",
			fmt.Sprintf("This resource or data source uses the UniFi Official API (integration/v1), which requires a "+
				"new-style controller running version %s or later, reachable with API-key authentication. "+
				"The configured controller reports version %s.", ControllerVersionOfficialAPI, current),
		)
	}
	return diags
}

// OfficialAPIErrorDiagnostics translates an error returned by an Official API
// operation into user-facing diagnostics, giving the go-unifi gate sentinels
// (ErrOfficialAPIUnavailable / ErrOfficialAPIDisabled) actionable messages and
// passing anything else through as a generic error. It returns no diagnostics for
// a nil error. Callers that treat 404 as "resource gone" must check
// errors.Is(err, unifi.ErrNotFound) before calling this.
func OfficialAPIErrorDiagnostics(summary string, err error) diag.Diagnostics {
	var diags diag.Diagnostics
	switch {
	case err == nil:
		return diags
	case errors.Is(err, unifi.ErrOfficialAPIUnavailable):
		diags.AddError(
			"UniFi Official API unavailable",
			"This resource or data source uses the UniFi Official API (integration/v1), which requires a new-style "+
				"controller running version "+ControllerVersionOfficialAPI.String()+" or later, reachable with API-key "+
				"authentication. The controller reported it as unavailable: "+err.Error(),
		)
	case errors.Is(err, unifi.ErrOfficialAPIDisabled):
		diags.AddError(
			"UniFi Official API disabled",
			"The UniFi Official API (integration/v1) is disabled for this provider client and cannot be used by this "+
				"resource or data source: "+err.Error(),
		)
	default:
		diags.AddError(summary, err.Error())
	}
	return diags
}

// ResolveSiteUUID maps a UniFi site name to its Official-API site UUID, falling
// back to the provider's configured site when name is empty. go-unifi caches the
// name→UUID lookup. Errors are returned as diagnostics: a friendly "site not
// found" for official.ErrSiteNotFound and the gate translation otherwise.
func (c *Client) ResolveSiteUUID(ctx context.Context, name string) (uuid.UUID, diag.Diagnostics) {
	site := name
	if site == "" {
		site = c.Site
	}
	return resolveSiteUUID(ctx, c.Client.Official(), site)
}

// ResolveSiteUUIDFromConfig resolves the site UUID for the site named by the
// config's "site" attribute (or the provider default), mirroring
// ResolveSiteFromConfig for Official-API resources that need a UUID.
func (c *Client) ResolveSiteUUIDFromConfig(ctx context.Context, config tfsdk.Config) (uuid.UUID, diag.Diagnostics) {
	site, diags := c.ResolveSiteFromConfig(ctx, config)
	if diags.HasError() {
		return uuid.UUID{}, diags
	}
	id, resolveDiags := c.ResolveSiteUUID(ctx, site)
	diags.Append(resolveDiags...)
	return id, diags
}

// resolveSiteUUID holds the testable core of ResolveSiteUUID: it takes the
// Official client directly so it can be exercised with official.ClientMock.
func resolveSiteUUID(ctx context.Context, oc official.Client, site string) (uuid.UUID, diag.Diagnostics) {
	var diags diag.Diagnostics
	id, err := oc.Sites().ResolveID(ctx, site)
	if err != nil {
		if errors.Is(err, official.ErrSiteNotFound) {
			diags.AddError(
				"UniFi site not found",
				fmt.Sprintf("Could not resolve the UniFi site %q to an Official-API site ID. Verify the site name "+
					"(the provider `site` attribute or the `UNIFI_SITE` environment variable).", site),
			)
			return uuid.UUID{}, diags
		}
		diags.Append(OfficialAPIErrorDiagnostics("Failed to resolve UniFi site", err)...)
		return uuid.UUID{}, diags
	}
	return id, diags
}

// CheckConfigured reports a diagnostic when the provider client was not wired up
// (Configure not run). Exported for the custom-CRUD Official-API resources.
func CheckConfigured(c *Client) diag.Diagnostics {
	return checkClientConfigured(c)
}

// ResolveSiteAndID resolves a site name to its Official-API UUID and parses an
// entity id string into a UUID, returning both. It is the common preamble for
// Official-API resource Read/Update/Delete handlers.
func (c *Client) ResolveSiteAndID(ctx context.Context, siteName, id string) (siteID uuid.UUID, entityID uuid.UUID, diags diag.Diagnostics) {
	siteID, diags = c.ResolveSiteUUID(ctx, siteName)
	if diags.HasError() {
		return uuid.UUID{}, uuid.UUID{}, diags
	}
	entityID, err := uuid.Parse(id)
	if err != nil {
		diags.AddError("Invalid resource ID", fmt.Sprintf("Could not parse %q as a UUID: %s", id, err))
		return uuid.UUID{}, uuid.UUID{}, diags
	}
	return siteID, entityID, diags
}

// CollectAll drains an Official-API paginated iterator (a ListXxxAll result) into
// a slice, short-circuiting on the first error. It is a thin wrapper over
// official.Collect kept here so resources and data sources have a single import
// for paginated reads. For a single page, call the corresponding ListXxxPage.
func CollectAll[T any](seq iter.Seq2[T, error]) ([]T, error) {
	return official.Collect(seq)
}
