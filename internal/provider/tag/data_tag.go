package tag

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

// TagDatasourceModel represents the data model for a UniFi tag data source.
type TagDatasourceModel struct {
	base.Model
	Name        types.String `tfsdk:"name"`
	MemberTable types.Set    `tfsdk:"member_table"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *TagDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *TagDatasourceModel) Merge(ctx context.Context, other interface{}) diag.Diagnostics {
	model, ok := other.(*unifi.Tag)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.Tag, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	members := model.MemberTable
	if members == nil {
		members = []string{}
	}
	set, diags := types.SetValueFrom(ctx, types.StringType, members)
	m.MemberTable = set
	return diags
}

var (
	_ datasource.DataSource              = &tagDatasource{}
	_ datasource.DataSourceWithConfigure = &tagDatasource{}
	_ base.Resource                      = &tagDatasource{}
)

type tagDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewTagDatasource() datasource.DataSource {
	return &tagDatasource{}
}

func (d *tagDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *tagDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *tagDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *tagDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *tagDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_tag"
}

func (d *tagDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_tag` data source retrieves an existing tag by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the tag to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
			"member_table": schema.SetAttribute{
				MarkdownDescription: "The set of member identifiers that belong to this tag.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *tagDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TagDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	tags, err := d.client.ListTag(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list tags", err.Error())
		return
	}

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a tag.")
		return
	}

	var found *unifi.Tag
	for i := range tags {
		t := tags[i]
		if (id != "" && t.ID == id) || (name != "" && t.Name == name) {
			found = &t
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Tag not found", fmt.Sprintf("No tag matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
