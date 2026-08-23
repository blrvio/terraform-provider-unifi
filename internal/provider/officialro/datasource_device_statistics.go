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
	_ datasource.DataSource              = &deviceStatisticsDataSource{}
	_ datasource.DataSourceWithConfigure = &deviceStatisticsDataSource{}
	_ base.Resource                      = &deviceStatisticsDataSource{}
)

type deviceStatisticsDataSource struct {
	officialBase
}

type deviceStatisticsModel struct {
	Site                 types.String  `tfsdk:"site"`
	DeviceID             types.String  `tfsdk:"device_id"`
	CPUUtilizationPct    types.Float64 `tfsdk:"cpu_utilization_pct"`
	MemoryUtilizationPct types.Float64 `tfsdk:"memory_utilization_pct"`
	LoadAverage1Min      types.Float64 `tfsdk:"load_average_1_min"`
	LoadAverage5Min      types.Float64 `tfsdk:"load_average_5_min"`
	LoadAverage15Min     types.Float64 `tfsdk:"load_average_15_min"`
	UptimeSec            types.Int64   `tfsdk:"uptime_sec"`
	LastHeartbeatAt      types.String  `tfsdk:"last_heartbeat_at"`
	NextHeartbeatAt      types.String  `tfsdk:"next_heartbeat_at"`
	UplinkRxRateBps      types.Int64   `tfsdk:"uplink_rx_rate_bps"`
	UplinkTxRateBps      types.Int64   `tfsdk:"uplink_tx_rate_bps"`
	JSON                 types.String  `tfsdk:"json"`
}

// mapDeviceStatistics converts an Official-API latest-statistics record into its
// Framework model, serializing the full record (including per-radio interface
// statistics) into the json attribute.
func mapDeviceStatistics(s official.LatestStatisticsForADevice) (deviceStatisticsModel, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return deviceStatisticsModel{}, err
	}
	model := deviceStatisticsModel{
		CPUUtilizationPct:    float64PtrValue(s.CpuUtilizationPct),
		MemoryUtilizationPct: float64PtrValue(s.MemoryUtilizationPct),
		LoadAverage1Min:      float64PtrValue(s.LoadAverage1Min),
		LoadAverage5Min:      float64PtrValue(s.LoadAverage5Min),
		LoadAverage15Min:     float64PtrValue(s.LoadAverage15Min),
		UptimeSec:            int64PtrValue(s.UptimeSec),
		LastHeartbeatAt:      timePtrValue(s.LastHeartbeatAt),
		NextHeartbeatAt:      timePtrValue(s.NextHeartbeatAt),
		UplinkRxRateBps:      types.Int64Null(),
		UplinkTxRateBps:      types.Int64Null(),
		JSON:                 types.StringValue(string(raw)),
	}
	if s.Uplink != nil {
		model.UplinkRxRateBps = int64PtrValue(s.Uplink.RxRateBps)
		model.UplinkTxRateBps = int64PtrValue(s.Uplink.TxRateBps)
	}
	return model, nil
}

// NewDeviceStatisticsDataSource returns the unifi_device_statistics data source,
// which fetches the latest statistics for an adopted device via the Official API.
func NewDeviceStatisticsDataSource() datasource.DataSource {
	return &deviceStatisticsDataSource{}
}

func (d *deviceStatisticsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_statistics"
}

func (d *deviceStatisticsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the latest statistics for an adopted device via the UniFi **Official API** " +
			"(`integration/v1`). Read-only. Requires a controller running version 10.1.78 or later with API-key " +
			"authentication.",
		Attributes: map[string]schema.Attribute{
			"site": siteAttribute(),
			"device_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the adopted device whose statistics to fetch.",
				Required:            true,
			},
			"cpu_utilization_pct": schema.Float64Attribute{
				MarkdownDescription: "CPU utilization percentage, if reported.",
				Computed:            true,
			},
			"memory_utilization_pct": schema.Float64Attribute{
				MarkdownDescription: "Memory utilization percentage, if reported.",
				Computed:            true,
			},
			"load_average_1_min": schema.Float64Attribute{
				MarkdownDescription: "1-minute load average, if reported.",
				Computed:            true,
			},
			"load_average_5_min": schema.Float64Attribute{
				MarkdownDescription: "5-minute load average, if reported.",
				Computed:            true,
			},
			"load_average_15_min": schema.Float64Attribute{
				MarkdownDescription: "15-minute load average, if reported.",
				Computed:            true,
			},
			"uptime_sec": schema.Int64Attribute{
				MarkdownDescription: "Device uptime in seconds, if reported.",
				Computed:            true,
			},
			"last_heartbeat_at": schema.StringAttribute{
				MarkdownDescription: "When the last heartbeat was received, in RFC 3339 format.",
				Computed:            true,
			},
			"next_heartbeat_at": schema.StringAttribute{
				MarkdownDescription: "When the next heartbeat is expected, in RFC 3339 format.",
				Computed:            true,
			},
			"uplink_rx_rate_bps": schema.Int64Attribute{
				MarkdownDescription: "Uplink receive rate in bits per second, if reported.",
				Computed:            true,
			},
			"uplink_tx_rate_bps": schema.Int64Attribute{
				MarkdownDescription: "Uplink transmit rate in bits per second, if reported.",
				Computed:            true,
			},
			"json": schema.StringAttribute{
				MarkdownDescription: "The full statistics record, including per-radio interface statistics, serialized as a JSON string.",
				Computed:            true,
			},
		},
	}
}

func (d *deviceStatisticsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceStatisticsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteID, siteName, diags := d.prepare(ctx, state.Site.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entityID, err := uuid.Parse(state.DeviceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid device id",
			fmt.Sprintf("Could not parse %q as a UUID: %s", state.DeviceID.ValueString(), err),
		)
		return
	}

	stats, err := d.client.Official().Devices().GetAdoptedLatestStatistics(ctx, siteID, entityID)
	if err != nil {
		if errors.Is(err, unifi.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Device statistics not found",
				fmt.Sprintf("No statistics for device %q were found on site %q.", state.DeviceID.ValueString(), siteName),
			)
			return
		}
		resp.Diagnostics.Append(base.OfficialAPIErrorDiagnostics("Error reading device statistics", err)...)
		return
	}

	model, err := mapDeviceStatistics(*stats)
	if err != nil {
		resp.Diagnostics.AddError("Error serializing device statistics", err.Error())
		return
	}
	model.Site = types.StringValue(siteName)
	model.DeviceID = types.StringValue(state.DeviceID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
