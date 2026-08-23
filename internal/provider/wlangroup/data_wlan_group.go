package wlangroup

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// WLANGroupDatasourceModel represents the data model for a UniFi WLAN group data source.
type WLANGroupDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *WLANGroupDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *WLANGroupDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.WLANGroup)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.WLANGroup, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	return diags
}

var (
	_ datasource.DataSource              = &wlanGroupDatasource{}
	_ datasource.DataSourceWithConfigure = &wlanGroupDatasource{}
	_ base.Resource                      = &wlanGroupDatasource{}
)

type wlanGroupDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewWLANGroupDatasource() datasource.DataSource {
	return &wlanGroupDatasource{}
}

func (d *wlanGroupDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *wlanGroupDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *wlanGroupDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *wlanGroupDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *wlanGroupDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_wlan_group"
}

func (d *wlanGroupDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_wlan_group` data source retrieves the ID of a WLAN group by name. " +
			"Leave `name` blank to look up the default WLAN group.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the WLAN group to look up. Leave blank to look up the default group.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *wlanGroupDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state WLANGroupDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	groups, err := d.client.ListWLANGroup(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list WLAN groups", err.Error())
		return
	}
	if len(groups) == 0 {
		resp.Diagnostics.AddError("WLAN group not found", "No WLAN groups found")
		return
	}

	name := state.Name.ValueString()
	var found *unifi.WLANGroup
	for i := range groups {
		g := groups[i]
		if (name == "" && g.HiddenID == "default") || g.Name == name {
			found = &g
			break
		}
	}
	if found == nil {
		resp.Diagnostics.AddError("WLAN group not found", fmt.Sprintf("No WLAN group with name %q found", name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
