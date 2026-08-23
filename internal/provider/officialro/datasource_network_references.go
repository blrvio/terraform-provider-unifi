package officialro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

var (
	_ datasource.DataSource              = &networkReferencesDataSource{}
	_ datasource.DataSourceWithConfigure = &networkReferencesDataSource{}
	_ base.Resource                      = &networkReferencesDataSource{}
)

type networkReferencesDataSource struct {
	officialBase
}

type networkReferencesModel struct {
	Site               types.String               `tfsdk:"site"`
	NetworkID          types.String               `tfsdk:"network_id"`
	ReferenceResources []networkReferenceResModel `tfsdk:"reference_resources"`
	JSON               types.String               `tfsdk:"json"`
}

type networkReferenceResModel struct {
	ResourceType   types.String `tfsdk:"resource_type"`
	ReferenceCount types.Int64  `tfsdk:"reference_count"`
	ReferenceIDs   []string     `tfsdk:"reference_ids"`
}

// mapNetworkReferences converts an Official-API network-references record into its
// Framework model, flattening each resource's reference ids and serializing the
// full record into the json attribute.
func mapNetworkReferences(nr official.NetworkReferences) (networkReferencesModel, error) {
	raw, err := json.Marshal(nr)
	if err != nil {
		return networkReferencesModel{}, err
	}
	resources := make([]networkReferenceResModel, 0, len(nr.ReferenceResources))
	for _, r := range nr.ReferenceResources {
		var ids []string
		if r.References != nil {
			ids = make([]string, 0, len(*r.References))
			for _, ref := range *r.References {
				ids = append(ids, ref.ReferenceId.String())
			}
		}
		resources = append(resources, networkReferenceResModel{
			ResourceType:   types.StringValue(string(r.ResourceType)),
			ReferenceCount: types.Int64Value(int64(r.ReferenceCount)),
			ReferenceIDs:   ids,
		})
	}
	return networkReferencesModel{
		ReferenceResources: resources,
		JSON:               types.StringValue(string(raw)),
	}, nil
}

// NewNetworkReferencesDataSource returns the unifi_network_references data source,
// which reports the resources referencing a network (clients, devices, routes,
// WiFi, NAT rules, ...) via the Official API.
func NewNetworkReferencesDataSource() datasource.DataSource {
	return &networkReferencesDataSource{}
}

func (d *networkReferencesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_references"
}

func (d *networkReferencesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reports the resources referencing a network (clients, devices, routes, WiFi, NAT rules, ...) " +
			"via the UniFi **Official API** (`integration/v1`). Read-only. Requires a controller running version 10.1.78 " +
			"or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the network whose references to fetch.",
				Required:            true,
			},
			"reference_resources": schema.ListNestedAttribute{
				MarkdownDescription: "The resource types referencing the network, grouped by type.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_type": schema.StringAttribute{
							MarkdownDescription: "The referencing resource type (e.g. `CLIENT`, `DEVICE`, `WIFI`, `NAT_RULE`).",
							Computed:            true,
						},
						"reference_count": schema.Int64Attribute{
							MarkdownDescription: "The number of references of this resource type.",
							Computed:            true,
						},
						"reference_ids": schema.ListAttribute{
							MarkdownDescription: "The UUIDs of the referencing resources, when enumerated by the controller.",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full network-references record serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *networkReferencesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networkReferencesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityID, err := uuid.Parse(state.NetworkID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid network id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.NetworkID.ValueString(), err),
		)
		return
	}

	refs, err := d.client.Official().Networks().GetReferences(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Network not found",
				fmt.Sprintf("No network with id %q was found on site %q.", state.NetworkID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading network references", err)...)
		return
	}

	model, err := mapNetworkReferences(*refs)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing network references", err.Error())
		return
	}
	model.Site = types.StringValue(siteName)
	model.NetworkID = types.StringValue(state.NetworkID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
