package hotspotop

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

// HotspotOperatorDatasourceModel represents the data model for a UniFi hotspot operator data source.
type HotspotOperatorDatasourceModel struct {
	base.Model
	Name      types.String `tfsdk:"name"`
	Note      types.String `tfsdk:"note"`
	XPassword types.String `tfsdk:"x_password"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *HotspotOperatorDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *HotspotOperatorDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.HotspotOp)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.HotspotOp, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	m.Note = ut.StringOrNull(model.Note)
	m.XPassword = ut.StringOrNull(model.XPassword)
	return diags
}

var (
	_ datasource.DataSource              = &hotspotOperatorDatasource{}
	_ datasource.DataSourceWithConfigure = &hotspotOperatorDatasource{}
	_ base.Resource                      = &hotspotOperatorDatasource{}
)

type hotspotOperatorDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewHotspotOperatorDatasource() datasource.DataSource {
	return &hotspotOperatorDatasource{}
}

func (d *hotspotOperatorDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *hotspotOperatorDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *hotspotOperatorDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *hotspotOperatorDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *hotspotOperatorDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_hotspot_operator"
}

func (d *hotspotOperatorDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_hotspot_operator` data source retrieves an existing hotspot operator by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the hotspot operator to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "A free-form note for the hotspot operator.",
				Computed:            true,
			},
			"x_password": schema.StringAttribute{
				MarkdownDescription: "The password for the hotspot operator account.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *hotspotOperatorDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state HotspotOperatorDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	hotspotOps, err := d.client.ListHotspotOp(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list hotspot operators", err.Error())
		return
	}

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a hotspot operator.")
		return
	}

	var found *unifi.HotspotOp
	for i := range hotspotOps {
		h := hotspotOps[i]
		if (id != "" && h.ID == id) || (name != "" && h.Name == name) {
			found = &h
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Hotspot operator not found", fmt.Sprintf("No hotspot operator matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
