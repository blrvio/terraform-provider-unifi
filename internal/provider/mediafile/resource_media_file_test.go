package mediafile

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaFileModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := MediaFileModel{
		Name: types.StringValue("intro.mp4"),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.MediaFile)
	require.True(t, ok, "Expected model to be *unifi.MediaFile")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "intro.mp4", typed.Name)
}

func TestMediaFileModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.MediaFile{
		ID:   "merge-id",
		Name: "intro.mp4",
	}

	var d MediaFileModel
	diags := d.Merge(context.Background(), model)
	assert.False(t, diags.HasError())

	assert.Equal(t, "merge-id", d.ID.ValueString())
	assert.Equal(t, "intro.mp4", d.Name.ValueString())
}

func TestMediaFileModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var d MediaFileModel
	diags := d.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
