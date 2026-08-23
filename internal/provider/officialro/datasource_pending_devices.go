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
	_ datasource.DataSource              = &pendingDevicesDataSource{}
	_ datasource.DataSourceWithConfigure = &pendingDevicesDataSource{}
	_ base.Resource                      = &pendingDevicesDataSource{}
)

type pendingDevicesDataSource struct {
	officialBase
}

type pendingDevicesModel struct {
	Filter  types.String                 `tfsdk:"filter"`
	Devices []pendingDeviceOverviewModel `tfsdk:"devices"`
}

type pendingDeviceOverviewModel struct {
	MACAddress            types.String `tfsdk:"mac_address"`
	IPAddress             types.String `tfsdk:"ip_address"`
	Model                 types.String `tfsdk:"model"`
	State                 types.String `tfsdk:"state"`
	Supported             types.Bool   `tfsdk:"supported"`
	FirmwareUpdatable     types.Bool   `tfsdk:"firmware_updatable"`
	FirmwareVersion       types.String `tfsdk:"firmware_version"`
	Features              []string     `tfsdk:"features"`
	AdoptionTargetSiteIDs []string     `tfsdk:"adoption_target_site_ids"`
	JSON                  types.String `tfsdk:"json"`
}

// mapPendingDevice converts an Official-API pending-adoption device into its
// Framework model, serializing the full record into the json attribute.
func mapPendingDevice(d official.DevicePendingAdoption) (pendingDeviceOverviewModel, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return pendingDeviceOverviewModel{}, err
	}
	features := make([]string, 0, len(d.Features))
	for _, f := range d.Features {
		features = append(features, string(f))
	}
	return pendingDeviceOverviewModel{
		MACAddress:            types.StringValue(d.MacAddress),
		IPAddress:             types.StringValue(d.IpAddress),
		Model:                 types.StringValue(d.Model),
		State:                 types.StringValue(string(d.State)),
		Supported:             types.BoolValue(d.Supported),
		FirmwareUpdatable:     types.BoolValue(d.FirmwareUpdatable),
		FirmwareVersion:       stringPtrValue(d.FirmwareVersion),
		Features:              features,
		AdoptionTargetSiteIDs: uuidsToStrings(d.AdoptionTargetSiteIds),
		JSON:                  types.StringValue(string(raw)),
	}, nil
}

// NewPendingDevicesDataSource returns the unifi_pending_devices data source, which
// lists devices awaiting adoption via the Official API. This endpoint is
// site-independent.
func NewPendingDevicesDataSource() datasource.DataSource {
	return &pendingDevicesDataSource{}
}

func (d *pendingDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pending_devices"
}

func (d *pendingDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists devices awaiting adoption via the UniFi **Official API** (`integration/v1`). " +
			"Read-only and site-independent. Requires a controller running version 10.1.78 or later with API-key " +
			"authentication.",
		Attributes: map[string]schema.Attribute{
			"filter": filterAttribute("pending devices"),
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "The devices awaiting adoption.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mac_address": schema.StringAttribute{
							MarkdownDescription: "The device MAC address.",
							Computed:            true,
						},
						"ip_address": schema.StringAttribute{
							MarkdownDescription: "The device IP address.",
							Computed:            true,
						},
						"model": schema.StringAttribute{
							MarkdownDescription: "The device model.",
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "The device state (e.g. `PENDING_ADOPTION`).",
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
						"adoption_target_site_ids": schema.ListAttribute{
							MarkdownDescription: "The UUIDs of the sites the device can be adopted into.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"json": schema.StringAttribute{
							MarkdownDescription: "The full pending-device record serialized as a JSON string.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *pendingDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state pendingDevicesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(d.prepareVersionOnly(ctx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	devices, err := base.CollectAll(d.client.Official().Devices().ListPendingAll(ctx, state.Filter.ValueString()))
	if err != nil {
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error listing pending devices", err)...)
		return
	}

	state.Devices = make([]pendingDeviceOverviewModel, 0, len(devices))
	for _, device := range devices {
		model, err := mapPendingDevice(device)
		if err != nil {
			resp.Diagnostics.AddError("Error serializing pending device", err.Error())
			return
		}
		state.Devices = append(state.Devices, model)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
