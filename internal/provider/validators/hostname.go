package validators

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// A regex pattern for validating hostnames without protocol schemes
// This matches hostnames according to RFC 1035 with some limitations.
var hostnameRegex = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$`)

// hostLabelOrFQDNRegex also accepts a single host label (RFC 1123), e.g. "UDM-Pro-Central".
var hostLabelOrFQDNRegex = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$`)

// Hostname returns a validator which ensures that the string value is a fully
// qualified hostname (at least one dot) without protocol schemes. Use it for domains.
func Hostname() validator.String {
	return hostnameValidator{re: hostnameRegex, desc: "must be a valid hostname (without protocol scheme)"}
}

// HostLabelOrFQDN returns a validator that accepts a single host label (a device's own
// name, or a server reachable by short name) or a fully qualified hostname.
func HostLabelOrFQDN() validator.String {
	return hostnameValidator{re: hostLabelOrFQDNRegex, desc: "must be a valid host name or fully qualified hostname (without protocol scheme)"}
}

type hostnameValidator struct {
	re   *regexp.Regexp
	desc string
}

func (v hostnameValidator) Description(_ context.Context) string {
	return v.desc
}

func (v hostnameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v hostnameValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	value := req.ConfigValue
	if !types.IsDefined(value) {
		return
	}

	val := value.ValueString()

	// Check if the hostname has a scheme (which it shouldn't)
	if strings.Contains(val, "://") {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q should not include a protocol scheme (e.g., http:// or https://)", val),
			),
		)
		return
	}

	// Convert to lowercase for validation
	hostname := strings.ToLower(val)

	if !v.re.MatchString(hostname) {
		resp.Diagnostics.Append(
			validatordiag.InvalidAttributeValueDiagnostic(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("%q is not a valid hostname", val),
			),
		)
	}
}
