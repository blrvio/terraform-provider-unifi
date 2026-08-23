package officialro

import (
	"context"
	"encoding/json"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &mcLagDomainsDataSource{}
	_ datasource.DataSourceWithConfigure = &mcLagDomainsDataSource{}
	_ base.Resource                      = &mcLagDomainsDataSource{}
)

type mcLagDomainsDataSource struct {
	officialBase
}

type mcLagDomainsModel struct {
	Site    types.String      `tfsdk:"site"`
	Filter  types.String      `tfsdk:"filter"`
	Domains []mcLagDomainItem `tfsdk:"domains"`
}

// mcLagDomainItem is the shared {id, name, metadata_origin, json} model for the
// MC-LAG domain list and single-get data sources. The nested lags and peers are
// preserved in the json attribute.
type mcLagDomainItem struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	JSON           types.String `tfsdk:"json"`
}

// mapMcLagDomain converts an Official-API MC-LAG domain into its Framework model,
// serializing the full record (including nested lags and peers) into the json
// attribute.
func mapMcLagDomain(d official.McLagDomain) (mcLagDomainItem, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return mcLagDomainItem{}, err
	}
	return mcLagDomainItem{
		ID:             types.StringValue(d.Id.String()),
		Name:           types.StringValue(d.Name),
		MetadataOrigin: types.StringValue(d.Metadata.Origin),
		JSON:           types.StringValue(string(raw)),
	}, nil
}

// mcLagDomainItemAttributes returns the nested-object attributes for the MC-LAG
// domain list (which has no per-item site attribute).
func mcLagDomainItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The UUID of the MC-LAG domain.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The MC-LAG domain name.",
			Computed:            true,
		},
		"metadata_origin": schema.StringAttribute{
			MarkdownDescription: "The origin of the MC-LAG domain entity (e.g. `USER_DEFINED`).",
			Computed:            true,
		},
		"json": schema.StringAttribute{
			MarkdownDescription: "The full MC-LAG domain record, including nested lags and peers, serialized as a JSON string.",
			Computed:            true,
		},
	}
}

// NewMcLagDomainsDataSource returns the unifi_mclag_domains data source, which
// lists MC-LAG domains on a site via the Official API.
func NewMcLagDomainsDataSource() datasource.DataSource {
	return &mcLagDomainsDataSource{}
}

func (d *mcLagDomainsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mclag_domains"
}

func (d *mcLagDomainsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists MC-LAG domains on a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("MC-LAG domains"),
			"domains": schema.ListNestedAttribute{
				MarkdownDescription: "The MC-LAG domains defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mcLagDomainItemAttributes(),
				},
			},
		},
	}
}

func (d *mcLagDomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state mcLagDomainsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domains, err := base.CollectAll(d.client.Official().Switching().ListMcLagDomainsAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing MC-LAG domains", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.Domains = make([]mcLagDomainItem, 0, len(domains))
	for _, dom := range domains {
		model, err := mapMcLagDomain(dom)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing MC-LAG domain", err.Error())
			return
		}
		state.Domains = append(state.Domains, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
