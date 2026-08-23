// Package trafficflow contains a read-only Terraform Plugin Framework data source
// for UniFi Traffic Flows. Unlike the officialro data sources, Traffic Flows is
// served by the go-unifi *internal* API (API v2), not the Official API
// (integration/v1), so it takes a site name string directly and is not gated on
// the Official-API controller version.
package trafficflow

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

// trafficFlowBase carries the shared Framework wiring (client + validators) and
// implements base.Resource so that base.ConfigureDatasource can inject them.
type trafficFlowBase struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func (d *trafficFlowBase) SetClient(client *base.Client) { d.client = client }

func (d *trafficFlowBase) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *trafficFlowBase) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *trafficFlowBase) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}
