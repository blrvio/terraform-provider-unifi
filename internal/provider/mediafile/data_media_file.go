package mediafile

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

// MediaFileDatasourceModel represents the data model for a UniFi media file data source.
type MediaFileDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *MediaFileDatasourceModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return nil, diag.Diagnostics{}
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *MediaFileDatasourceModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	model, ok := other.(*unifi.MediaFile)
	if !ok {
		diags.AddError("Invalid model type", fmt.Sprintf("Expected *unifi.MediaFile, got %T", other))
		return diags
	}

	m.ID = types.StringValue(model.ID)
	m.Name = types.StringValue(model.Name)
	return diags
}

var (
	_ datasource.DataSource              = &mediaFileDatasource{}
	_ datasource.DataSourceWithConfigure = &mediaFileDatasource{}
	_ base.Resource                      = &mediaFileDatasource{}
)

type mediaFileDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

func NewMediaFileDatasource() datasource.DataSource {
	return &mediaFileDatasource{}
}

func (d *mediaFileDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *mediaFileDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *mediaFileDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *mediaFileDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *mediaFileDatasource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "unifi_media_file"
}

func (d *mediaFileDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_media_file` data source retrieves an existing media file by name or id.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the media file to look up. Either `name` or `id` must be set.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (d *mediaFileDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MediaFileDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	mediaFiles, err := d.client.ListMediaFile(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list media files", err.Error())
		return
	}

	id := state.GetID()
	name := state.Name.ValueString()
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "Either `id` or `name` must be set to look up a media file.")
		return
	}

	var found *unifi.MediaFile
	for i := range mediaFiles {
		f := mediaFiles[i]
		if (id != "" && f.ID == id) || (name != "" && f.Name == name) {
			found = &f
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Media file not found", fmt.Sprintf("No media file matching id=%q name=%q was found", id, name))
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, found)...)
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
