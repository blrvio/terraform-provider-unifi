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
	_ datasource.DataSource              = &officialDeviceDataSource{}
	_ datasource.DataSourceWithConfigure = &officialDeviceDataSource{}
	_ base.Resource                      = &officialDeviceDataSource{}
)

type officialDeviceDataSource struct {
	officialBase
}

type adoptedDeviceModel struct {
	Site              types.String `tfsdk:"site"`
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Model             types.String `tfsdk:"model"`
	MACAddress        types.String `tfsdk:"mac_address"`
	IPAddress         types.String `tfsdk:"ip_address"`
	State             types.String `tfsdk:"state"`
	Supported         types.Bool   `tfsdk:"supported"`
	FirmwareUpdatable types.Bool   `tfsdk:"firmware_updatable"`
	FirmwareVersion   types.String `tfsdk:"firmware_version"`
	ConfigurationID   types.String `tfsdk:"configuration_id"`
	AdoptedAt         types.String `tfsdk:"adopted_at"`
	ProvisionedAt     types.String `tfsdk:"provisioned_at"`
	JSON              types.String `tfsdk:"json"`
}

// mapAdoptedDeviceDetails converts an Official-API adopted-device detail record
// into its Framework model, serializing the full record (including nested
// features, interfaces and uplink) into the json attribute.
func mapAdoptedDeviceDetails(d official.AdoptedDeviceDetails) (adoptedDeviceModel, error) {
	raw, err := json.Marshal(d)
	if err != nil {
		return adoptedDeviceModel{}, err
	}
	return adoptedDeviceModel{
		ID:                types.StringValue(d.Id.String()),
		Name:              types.StringValue(d.Name),
		Model:             types.StringValue(d.Model),
		MACAddress:        types.StringValue(d.MacAddress),
		IPAddress:         types.StringValue(d.IpAddress),
		State:             types.StringValue(string(d.State)),
		Supported:         types.BoolValue(d.Supported),
		FirmwareUpdatable: types.BoolValue(d.FirmwareUpdatable),
		FirmwareVersion:   stringPtrValue(d.FirmwareVersion),
		ConfigurationID:   types.StringValue(d.ConfigurationId),
		AdoptedAt:         timePtrValue(d.AdoptedAt),
		ProvisionedAt:     timePtrValue(d.ProvisionedAt),
		JSON:              types.StringValue(string(raw)),
	}, nil
}

// NewOfficialDeviceDataSource returns the unifi_official_device data source, which
// fetches a single adopted device by id via the Official API.
func NewOfficialDeviceDataSource() datasource.DataSource {
	return &officialDeviceDataSource{}
}

func (d *officialDeviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_official_device"
}

func (d *officialDeviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single adopted device by id via the UniFi **Official API** (`integration/v1`). " +
			"Read-only. Named distinctly from the internal `unifi_device` data source. Requires a controller running " +
			"version 10.1.78 or later with API-key authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the adopted device to look up.",
				Required:            true,
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
			"configuration_id": schema.StringAttribute{
				MarkdownDescription: "The device configuration id.",
				Computed:            true,
			},
			"adopted_at": schema.StringAttribute{
				MarkdownDescription: "When the device was adopted, in RFC 3339 format.",
				Computed:            true,
			},
			"provisioned_at": schema.StringAttribute{
				MarkdownDescription: "When the device was last provisioned, in RFC 3339 format.",
				Computed:            true,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full device details, including nested features, interfaces and uplink, serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *officialDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state adoptedDeviceModel
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
			"Invalid device id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.ID.ValueString(), err),
		)
		return
	}

	device, err := d.client.Official().Devices().GetAdopted(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Device not found",
				fmt.Sprintf("No adopted device with id %q was found on site %q.", state.ID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading adopted device", err)...)
		return
	}

	model, err := mapAdoptedDeviceDetails(*device)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing adopted device", err.Error())
		return
	}
	model.Site = types.StringValue(siteName)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
