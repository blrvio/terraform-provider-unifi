package mediafile

import (
	"context"
	"fmt"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
	ut "github.com/blrvio/terraform-provider-unifi/internal/provider/types"
)

// MediaFileModel represents the data model for a UniFi media file.
type MediaFileModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

// AsUnifiModel converts the Terraform model to the UniFi API model.
func (m *MediaFileModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	return &unifi.MediaFile{
		ID:   m.ID.ValueString(),
		Name: m.Name.ValueString(),
	}, diags
}

// Merge updates the Terraform model with values from the UniFi API model.
func (m *MediaFileModel) Merge(_ context.Context, other interface{}) diag.Diagnostics {
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
	_ resource.Resource                = &mediaFileResource{}
	_ resource.ResourceWithConfigure   = &mediaFileResource{}
	_ resource.ResourceWithImportState = &mediaFileResource{}
	_ base.Resource                    = &mediaFileResource{}
	_ base.ResourceModel               = &MediaFileModel{}
)

type mediaFileResource struct {
	*base.GenericResource[*MediaFileModel]
}

// NewMediaFileResource creates a new instance of the media file resource.
func NewMediaFileResource() resource.Resource {
	return &mediaFileResource{
		GenericResource: base.NewGenericResource(
			"unifi_media_file",
			func() *MediaFileModel { return &MediaFileModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetMediaFile(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					mediaFile, _ := model.(*unifi.MediaFile)
					return client.CreateMediaFile(ctx, site, mediaFile)
				},
				Update: func(ctx context.Context, client *base.Client, site string, model interface{}) (interface{}, error) {
					mediaFile, _ := model.(*unifi.MediaFile)
					return client.UpdateMediaFile(ctx, site, mediaFile)
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteMediaFile(ctx, site, id)
				},
			},
		),
	}
}

// Schema defines the schema for the resource.
func (r *mediaFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_media_file` resource manages a media file in the UniFi controller.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the media file.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}
