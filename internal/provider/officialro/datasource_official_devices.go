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
	_ datasource.DataSource              = &officialDevicesDataSource{}
	_ datasource.DataSourceWithConfigure = &officialDevicesDataSource{}
	_ base.Resource                      = &officialDevicesDataSource{}
)

type officialDevicesDataSource struct {
	officialBase
}

type officialDevicesModel struct {
	Site    types.String                 `tfsdk:"site"`
	Filter  types.String                 `tfsdk:"filter"`
	Devices []adoptedDeviceOverviewModel `tfsdk:"devices"`
}

type adoptedDeviceOverviewModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Model             types.String `tfsdk:"model"`
	MACAddress        types.String `tfsdk:"mac_address"`
	IPAddress         types.String `tfsdk:"ip_address"`
	State             types.String `tfsdk:"state"`
	Supported         types.Bool   `tfsdk:"supported"`
	FirmwareUpdatable types.Bool   `tfsdk:"firmware_updatable"`
	FirmwareVersion   types.String `tfsdk:"firmware_version"`
	Features          []string     `tfsdk:"features"`
	JSON              types.String `tfsdk:"json"`
}

// mapAdoptedDeviceOverview converts an Official-API adopted-device overview into
// its Framework model, serializing the full record (including the nested
// interfaces) into the json attribute.
func mapAdoptedDeviceOverview(d official.AdoptedDeviceOverview) (adoptedDeviceOverviewModel, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return adoptedDeviceOverviewModel{}, err
	}
	features := make([]string, 0, len(d.Features))
	for _, f := range d.Features {
		features = append(features, string(f))
	}
	return adoptedDeviceOverviewModel{
		ID:                types.StringValue(d.Id.String()),
		Name:              types.StringValue(d.Name),
		Model:             types.StringValue(d.Model),
		MACAddress:        types.StringValue(d.MacAddress),
		IPAddress:         types.StringValue(d.IpAddress),
		State:             types.StringValue(string(d.State)),
		Supported:         types.BoolValue(d.Supported),
		FirmwareUpdatable: types.BoolValue(d.FirmwareUpdatable),
		FirmwareVersion:   stringPtrValue(d.FirmwareVersion),
		Features:          features,
		JSON:              types.StringValue(string(raw)),
	}, nil
}

// NewOfficialDevicesDataSource returns the unifi_official_devices data source,
// which lists the devices adopted on a site via the Official API. It is named
// distinctly from the internal unifi_device data source.
func NewOfficialDevicesDataSource() datasource.DataSource {
	return &officialDevicesDataSource{}
}

func (d *officialDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_official_devices"
}

func adoptedDeviceOverviewAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The UUID of the adopted device.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The device name.",
			Computed:            true,
		},
		"model": schema.StringAttribute{
			MarkdownDescription: "The device model.",
			Computed:            true,
		},
		"mac_address": schema.StringAttribute{
			MarkdownDescription: "The device MAC address.",
			Computed:            true,
		},
		"ip_address": schema.StringAttribute{
			MarkdownDescription: "The device IP address.",
			Computed:            true,
		},
		"state": schema.StringAttribute{
			MarkdownDescription: "The device state (e.g. `ONLINE`, `OFFLINE`, `UPDATING`).",
			Computed:            true,
		},
		"supported": schema.BoolAttribute{
			MarkdownDescription: "Whether the device is supported by the controller.",
			Computed:            true,
		},
		"firmware_updatable": schema.BoolAttribute{
			MarkdownDescription: "Whether a firmware update is available for the device.",
			Computed:            true,
		},
		"firmware_version": schema.StringAttribute{
			MarkdownDescription: "The device firmware version, if known.",
			Computed:            true,
		},
		"features": schema.ListAttribute{
			MarkdownDescription: "The device feature roles (e.g. `switching`, `accessPoint`, `gateway`).",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"json": schema.StringAttribute{
			MarkdownDescription: "The full device overview, including nested interfaces, serialized as a JSON string.",
			Computed:            true,
		},
	}
}

func (d *officialDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the devices adopted on a site via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Requires a controller running version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site":   siteAttribute(),
			"filter": filterAttribute("adopted devices"),
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "The adopted devices.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: adoptedDeviceOverviewAttributes(),
				},
			},
		},
	}
}

func (d *officialDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state officialDevicesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	devices, err := base.CollectAll(d.client.Official().Devices().ListAdoptedAll(ctx, siteID, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing adopted devices", err)...)
		return
	}

	state.Site = types.StringValue(siteName)
	state.Devices = make([]adoptedDeviceOverviewModel, 0, len(devices))
	for _, device := range devices {
		model, err := mapAdoptedDeviceOverview(device)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing adopted device", err.Error())
			return
		}
		state.Devices = append(state.Devices, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
