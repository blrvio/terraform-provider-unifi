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
	_ datasource.DataSource              = &switchStacksDataSource{}
	_ datasource.DataSourceWithConfigure = &switchStacksDataSource{}
	_ base.Resource                      = &switchStacksDataSource{}
)

type switchStacksDataSource struct {
	officialBase
}

type switchStacksModel struct {
	Site   types.String      `tfsdk:"site"`
	Filter types.String      `tfsdk:"filter"`
	Stacks []switchStackItem `tfsdk:"stacks"`
}

// switchStackItem is the shared model for the switch-stack list and single-get
// data sources. The nested lags are preserved in the json attribute.
type switchStackItem struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	MetadataOrigin  types.String `tfsdk:"metadata_origin"`
	MemberDeviceIDs []string     `tfsdk:"member_device_ids"`
	JSON            types.String `tfsdk:"json"`
}

// mapSwitchStack converts an Official-API switch stack into its Framework model,
// flattening the member device ids and serializing the full record (including
// nested lags) into the json attribute.
func mapSwitchStack(s official.SwitchStack) (switchStackItem, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return switchStackItem{}, err
	}
	members := make([]string, 0, len(s.Members))
	for _, m := range s.Members {
		members = append(members, m.DeviceId.String())
	}
	return switchStackItem{
		ID:              types.StringValue(s.Id.String()),
		Name:            types.StringValue(s.Name),
		MetadataOrigin:  types.StringValue(s.Metadata.Origin),
		MemberDeviceIDs: members,
		JSON:            types.StringValue(string(raw)),
	}, nil
}

func switchStackItemAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The UUID of the switch stack.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The switch stack name.",
			Computed:            true,
		},
		"metadata_origin": schema.StringAttribute{
			MarkdownDescription: "The origin of the switch stack entity (e.g. `USER_DEFINED`).",
			Computed:            true,
		},
		"member_device_ids": schema.ListAttribute{
			MarkdownDescription: "The UUIDs of the devices that are members of the stack.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"json": schema.StringAttribute{
			MarkdownDescription: "The full switch stack record, including nested lags, serialized as a JSON string.",
			Computed:            true,
		},
	}
}

// NewSwitchStacksDataSource returns the unifi_switch_stacks data source, which
// lists switch stacks on a site via the Official API.
func NewSwitchStacksDataSource() datasource.DataSource {
	return &switchStacksDataSource{}
}

func (d *switchStacksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_switch_stacks"
}

func (d *switchStacksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists switch stacks on a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("switch stacks"),
			"stacks": schema.ListNestedAttribute{
				MarkdownDescription: "The switch stacks defined on the site.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: switchStackItemAttributes(),
				},
			},
		},
	}
}

func (d *switchStacksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state switchStacksModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stacks, err := base.CollectAll(d.client.Official().Switching().ListSwitchStacksAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing switch stacks", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.Stacks = make([]switchStackItem, 0, len(stacks))
	for _, s := range stacks {
		model, err := mapSwitchStack(s)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing switch stack", err.Error())
			return
		}
		state.Stacks = append(state.Stacks, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
