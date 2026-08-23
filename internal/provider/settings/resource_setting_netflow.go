package settings

import (
	"context"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
	"github.com/blrvio/terraform-provider-unifi/internal/provider/validators"
)

// netflowModel represents the NetFlow export settings for a UniFi site, used to
// stream flow records to an external NetFlow collector.
type netflowModel struct {
	base.Model
	Enabled             types.Bool   `tfsdk:"enabled"`
	AutoEngineIDEnabled types.Bool   `tfsdk:"auto_engine_id_enabled"`
	EngineID            types.Int32  `tfsdk:"engine_id"`
	ExportFrequency     types.Int32  `tfsdk:"export_frequency"`
	NetworkIDs          types.List   `tfsdk:"network_ids"`
	Port                types.Int32  `tfsdk:"port"`
	RefreshRate         types.Int32  `tfsdk:"refresh_rate"`
	SamplingMode        types.String `tfsdk:"sampling_mode"`
	SamplingRate        types.Int32  `tfsdk:"sampling_rate"`
	Server              types.String `tfsdk:"server"`
	Version             types.Int32  `tfsdk:"version"`
}

func (d *netflowModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model := &unifi.SettingNetflow{
		ID:         d.ID.ValueString(),
		Enabled:    d.Enabled.ValueBool(),
		NetworkIDs: []string{},
	}

	// Only set optional fields if NetFlow is enabled
	if d.Enabled.ValueBool() {
		if !d.AutoEngineIDEnabled.IsNull() {
			model.AutoEngineIDEnabled = d.AutoEngineIDEnabled.ValueBool()
		}
		if !d.EngineID.IsNull() {
			model.EngineID = int(d.EngineID.ValueInt32())
		}
		if !d.ExportFrequency.IsNull() {
			model.ExportFrequency = int(d.ExportFrequency.ValueInt32())
		}
		if !d.Port.IsNull() {
			model.Port = int(d.Port.ValueInt32())
		}
		if !d.RefreshRate.IsNull() {
			model.RefreshRate = int(d.RefreshRate.ValueInt32())
		}
		if !ut.IsEmptyString(d.SamplingMode) {
			model.SamplingMode = d.SamplingMode.ValueString()
		}
		if !d.SamplingRate.IsNull() {
			model.SamplingRate = int(d.SamplingRate.ValueInt32())
		}
		if !ut.IsEmptyString(d.Server) {
			model.Server = d.Server.ValueString()
		}
		if !d.Version.IsNull() {
			model.Version = int(d.Version.ValueInt32())
		}

		if !d.NetworkIDs.IsNull() {
			var networkIDs []string
			diags.Append(ut.ListElementsAs(ctx, d.NetworkIDs, &networkIDs)...)
			if diags.HasError() {
				return nil, diags
			}
			model.NetworkIDs = networkIDs
		}
	}

	return model, diags
}

func (d *netflowModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model, ok := other.(*unifi.SettingNetflow)
	if !ok {
		diags.AddError("Cannot merge", "Cannot merge type that is not *unifi.SettingNetflow")
		return diags
	}

	d.ID = types.StringValue(model.ID)
	d.Enabled = types.BoolValue(model.Enabled)

	// Only set optional fields if NetFlow is enabled
	if model.Enabled {
		d.AutoEngineIDEnabled = types.BoolValue(model.AutoEngineIDEnabled)
		d.EngineID = ut.Int32OrNull(model.EngineID)
		d.ExportFrequency = ut.Int32OrNull(model.ExportFrequency)
		d.Port = ut.Int32OrNull(model.Port)
		d.RefreshRate = ut.Int32OrNull(model.RefreshRate)
		d.SamplingMode = ut.StringOrNull(model.SamplingMode)
		d.SamplingRate = ut.Int32OrNull(model.SamplingRate)
		d.Server = ut.StringOrNull(model.Server)
		d.Version = ut.Int32OrNull(model.Version)

		list, ld := types.ListValueFrom(ctx, types.StringType, model.NetworkIDs)
		diags.Append(ld...)
		if diags.HasError() {
			return diags
		}
		d.NetworkIDs = list
	} else {
		d.AutoEngineIDEnabled = types.BoolNull()
		d.EngineID = types.Int32Null()
		d.ExportFrequency = types.Int32Null()
		d.Port = types.Int32Null()
		d.RefreshRate = types.Int32Null()
		d.SamplingMode = types.StringNull()
		d.SamplingRate = types.Int32Null()
		d.Server = types.StringNull()
		d.Version = types.Int32Null()
		d.NetworkIDs = ut.EmptyList(types.StringType)
	}

	return diags
}

var (
	_ base.ResourceModel                    = &netflowModel{}
	_ resource.Resource                     = &netflowResource{}
	_ resource.ResourceWithConfigure        = &netflowResource{}
	_ resource.ResourceWithImportState      = &netflowResource{}
	_ resource.ResourceWithConfigValidators = &netflowResource{}
)

type netflowResource struct {
	*base.GenericResource[*netflowModel]
}

func (r *netflowResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		validators.RequiredNoneIf(path.MatchRoot("enabled"), types.BoolValue(false),
			path.MatchRoot("auto_engine_id_enabled"),
			path.MatchRoot("engine_id"),
			path.MatchRoot("export_frequency"),
			path.MatchRoot("network_ids"),
			path.MatchRoot("port"),
			path.MatchRoot("refresh_rate"),
			path.MatchRoot("sampling_mode"),
			path.MatchRoot("sampling_rate"),
			path.MatchRoot("server"),
			path.MatchRoot("version"),
		),
	}
}

func (r *netflowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_netflow` resource manages NetFlow export settings for a UniFi site, streaming " +
			"flow records to an external NetFlow collector. When `enabled` is `false`, all other attributes must be omitted.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether NetFlow export is enabled.",
				Required:            true,
			},
			"auto_engine_id_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the NetFlow engine ID is assigned automatically.",
				Optional:            true,
				Computed:            true,
			},
			"engine_id": schema.Int32Attribute{
				MarkdownDescription: "The NetFlow engine ID used to identify the exporter.",
				Optional:            true,
				Computed:            true,
			},
			"export_frequency": schema.Int32Attribute{
				MarkdownDescription: "How often (in seconds) flow records are exported to the collector.",
				Optional:            true,
				Computed:            true,
			},
			"network_ids": schema.ListAttribute{
				MarkdownDescription: "List of network IDs for which flow records are exported.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"port": schema.Int32Attribute{
				MarkdownDescription: "The UDP port of the NetFlow collector. Valid values: 1-65535.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.Between(1, 65535),
				},
			},
			"refresh_rate": schema.Int32Attribute{
				MarkdownDescription: "How often (in seconds) the NetFlow template is refreshed.",
				Optional:            true,
				Computed:            true,
			},
			"sampling_mode": schema.StringAttribute{
				MarkdownDescription: "The flow sampling mode. Valid values: `off`, `hash`, `random`, `deterministic`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("off", "hash", "random", "deterministic"),
				},
			},
			"sampling_rate": schema.Int32Attribute{
				MarkdownDescription: "The flow sampling rate (one out of every N packets).",
				Optional:            true,
				Computed:            true,
			},
			"server": schema.StringAttribute{
				MarkdownDescription: "Hostname or IP address of the NetFlow collector.",
				Optional:            true,
				Computed:            true,
			},
			"version": schema.Int32Attribute{
				MarkdownDescription: "The NetFlow protocol version. Valid values: 5, 9, 10.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int32{
					int32validator.OneOf(5, 9, 10),
				},
			},
		},
	}
}

// NewNetflowResource creates a new instance of the NetFlow setting resource.
func NewNetflowResource() resource.Resource {
	r := &netflowResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_netflow",
		func() *netflowModel { return &netflowModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingNetflow(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			b, _ := body.(*unifi.SettingNetflow)
			return client.UpdateSettingNetflow(ctx, site, b)
		},
	)
	return r
}
