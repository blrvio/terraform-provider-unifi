package officialro

import (
	"context"
	"encoding/json"

	"github.com/blrvio/go-unifi/v11/unifi/official"
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
	ID             types.String           `tfsdk:"id"`
	Name           types.String           `tfsdk:"name"`
	MetadataOrigin types.String           `tfsdk:"metadata_origin"`
	DeviceID       types.String           `tfsdk:"device_id"`
	Units          []switchStackUnitModel `tfsdk:"units"`
	JSON           types.String           `tfsdk:"json"`
}

// switchStackUnitModel is one physical switch (unit) of a switch stack.
type switchStackUnitModel struct {
	ID         types.Int64  `tfsdk:"id"`
	MacAddress types.String `tfsdk:"mac_address"`
	Role       types.String `tfsdk:"role"`
	Order      types.Int64  `tfsdk:"order"`
}

// mapSwitchStackUnits converts the Official-API stack units into their
// Framework models; an absent role or order maps to null.
func mapSwitchStackUnits(units []official.SwitchStackUnit) []switchStackUnitModel {
	out := make([]switchStackUnitModel, 0, len(units))
	for _, u := range units {
		unit := switchStackUnitModel{
			ID:         types.Int64Value(int64(u.Id)),
			MacAddress: types.StringValue(u.MacAddress),
			Role:       types.StringNull(),
			Order:      types.Int64Null(),
		}
		if u.Role != nil {
			unit.Role = types.StringValue(string(*u.Role))
		}
		if u.Order != nil {
			unit.Order = types.Int64Value(int64(*u.Order))
		}
		out = append(out, unit)
	}
	return out
}

// mapSwitchStack converts an Official-API switch stack into its Framework model,
// flattening the stack units and serializing the full record (including nested
// lags) into the json attribute.
func mapSwitchStack(s official.SwitchStack) (switchStackItem, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return switchStackItem{}, err
	}
	deviceID := types.StringNull()
	if s.DeviceId != nil {
		deviceID = types.StringValue(s.DeviceId.String())
	}
	return switchStackItem{
		ID:             types.StringValue(s.Id.String()),
		Name:           types.StringValue(s.Name),
		MetadataOrigin: types.StringValue(s.Metadata.Origin),
		DeviceID:       deviceID,
		Units:          mapSwitchStackUnits(s.Units),
		JSON:           types.StringValue(string(raw)),
	}, nil
}

// switchStackItemAttributes returns the computed attributes shared by the
// unifi_switch_stacks list items and the unifi_switch_stack data source.
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
		"device_id": schema.StringAttribute{
			MarkdownDescription: "The UUID of the device that represents the stack as a whole. Null when the controller " +
				"does not report it (controllers before 10.6.109).",
			Computed: true,
		},
		"units": schema.ListNestedAttribute{
			MarkdownDescription: "The physical switches (units) that make up the stack. Requires a controller running " +
				"10.6.109 or later; older controllers report no units.",
			Computed: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						MarkdownDescription: "The unit id within the stack (referenced by `unit_id` on switch-stack LAG members).",
						Computed:            true,
					},
					"mac_address": schema.StringAttribute{
						MarkdownDescription: "The MAC address of the unit.",
						Computed:            true,
					},
					"role": schema.StringAttribute{
						MarkdownDescription: "The unit role: `ACTIVE_CONTROLLER`, `BACKUP_CONTROLLER` or `MEMBER`. Null when not reported.",
						Computed:            true,
					},
					"order": schema.Int64Attribute{
						MarkdownDescription: "The position of the unit in the stack. Null when not reported.",
						Computed:            true,
					},
				},
			},
		},
		"json": schema.StringAttribute{
			MarkdownDescription: "The full switch stack record, including units and nested lags, serialized as a JSON string.",
			Computed:            true,
		},
	}
}

// switchStacksFilterAttribute is the generic filter attribute, documenting the
// unit-based properties that replaced members.deviceId in controller 10.6.109.
func switchStacksFilterAttribute() schema.StringAttribute {
	attr := filterAttribute("switch stacks")
	attr.MarkdownDescription += " Stacks can be filtered by unit with `units.id` and `units.macAddress` " +
		"(the former `members.deviceId` property was removed in controller 10.6.109)."
	return attr
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
			"filter": switchStacksFilterAttribute(),
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
