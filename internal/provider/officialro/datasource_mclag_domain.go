package officialro

import (
	"context"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &mcLagDomainDataSource{}
	_ datasource.DataSourceWithConfigure = &mcLagDomainDataSource{}
	_ base.Resource                      = &mcLagDomainDataSource{}
)

type mcLagDomainDataSource struct {
	officialBase
}

type mcLagDomainModel struct {
	Site           types.String `tfsdk:"site"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	MetadataOrigin types.String `tfsdk:"metadata_origin"`
	JSON           types.String `tfsdk:"json"`
}

// NewMcLagDomainDataSource returns the unifi_mclag_domain data source, which
// fetches a single MC-LAG domain by id via the Official API.
func NewMcLagDomainDataSource() datasource.DataSource {
	return &mcLagDomainDataSource{}
}

func (d *mcLagDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mclag_domain"
}

func (d *mcLagDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single MC-LAG domain by id via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the MC-LAG domain to look up.",
				Required:            true,
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
		},
	}
}

func (d *mcLagDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state mcLagDomainModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid MC-LAG domain id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.ID.ValueString(), err),
		)
		return
	}

	domain, err := d.client.Official().Switching().GetMcLagDomain(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"MC-LAG domain not found",
				fmt.Sprintf("No MC-LAG domain with id %q was found on site %q.", state.ID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading MC-LAG domain", err)...)
		return
	}

	item, err := mapMcLagDomain(*domain)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing MC-LAG domain", err.Error())
		return
	}
	state.Site = types.StringValue(siteName)
	state.ID = item.ID
	state.Name = item.Name
	state.MetadataOrigin = item.MetadataOrigin
	state.JSON = item.JSON
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
