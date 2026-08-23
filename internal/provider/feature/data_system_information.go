package feature

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// systemInformationDatasourceModel represents controller-wide system information.
type systemInformationDatasourceModel struct {
	Timezone        types.String `tfsdk:"timezone"`
	Version         types.String `tfsdk:"version"`
	PreviousVersion types.String `tfsdk:"previous_version"`
	Build           types.String `tfsdk:"build"`
	Name            types.String `tfsdk:"name"`
	Hostname        types.String `tfsdk:"hostname"`
	IPAddrs         types.List   `tfsdk:"ip_addrs"`
	Uptime          types.Int64  `tfsdk:"uptime"`
	UBNTDeviceType  types.String `tfsdk:"ubnt_device_type"`
	UDMVersion      types.String `tfsdk:"udm_version"`
}

var (
	_ datasource.DataSource              = &systemInformationDatasource{}
	_ datasource.DataSourceWithConfigure = &systemInformationDatasource{}
	_ base.Resource                      = &systemInformationDatasource{}
)

type systemInformationDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewSystemInformationDatasource() datasource.DataSource {
	return &systemInformationDatasource{}
}

func (d *systemInformationDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *systemInformationDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *systemInformationDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *systemInformationDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *systemInformationDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_system_information"
}

func (d *systemInformationDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_system_information` data source exposes controller-wide system " +
			"information (version, build, hostname, uptime and device type).",
		Attributes: map[string]schema.Attribute{
			"timezone":         schema.StringAttribute{Computed: true, MarkdownDescription: "The controller timezone."},
			"version":          schema.StringAttribute{Computed: true, MarkdownDescription: "The controller software version."},
			"previous_version": schema.StringAttribute{Computed: true, MarkdownDescription: "The previously installed controller version."},
			"build":            schema.StringAttribute{Computed: true, MarkdownDescription: "The controller build identifier."},
			"name":             schema.StringAttribute{Computed: true, MarkdownDescription: "The controller name."},
			"hostname":         schema.StringAttribute{Computed: true, MarkdownDescription: "The controller hostname."},
			"ip_addrs": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The controller IP addresses.",
			},
			"uptime":           schema.Int64Attribute{Computed: true, MarkdownDescription: "The controller uptime in seconds."},
			"ubnt_device_type": schema.StringAttribute{Computed: true, MarkdownDescription: "The Ubiquiti device type."},
			"udm_version":      schema.StringAttribute{Computed: true, MarkdownDescription: "The UDM firmware version (if applicable)."},
		},
	}
}

func (d *systemInformationDatasource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if diags := base.CheckConfigured(d.client); diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	info, err := d.client.GetSystemInformationContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get system information", err.Error())
		return
	}

	ipAddrs := info.IPAddrs
	if ipAddrs == nil {
		ipAddrs = []string{}
	}
	ipList, listDiags := types.ListValueFrom(ctx, types.StringType, ipAddrs)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := systemInformationDatasourceModel{
		Timezone:        ut.StringOrNull(info.Timezone),
		Version:         ut.StringOrNull(info.Version),
		PreviousVersion: ut.StringOrNull(info.PreviousVersion),
		Build:           ut.StringOrNull(info.Build),
		Name:            ut.StringOrNull(info.Name),
		Hostname:        ut.StringOrNull(info.Hostname),
		IPAddrs:         ipList,
		Uptime:          types.Int64Value(info.Uptime),
		UBNTDeviceType:  ut.StringOrNull(info.UBNTDeviceType),
		UDMVersion:      ut.StringOrNull(info.UDMVersion),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
