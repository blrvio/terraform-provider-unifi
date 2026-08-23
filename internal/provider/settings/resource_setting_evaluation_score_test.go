package settings

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEvaluationScoreModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := evaluationScoreModel{
		DismissedIDs: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("ab12"),
			types.StringValue("cd345"),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError())

	typed, ok := unifiModel.(*unifi.SettingEvaluationScore)
	assert.True(t, ok, "Expected model to be *unifi.SettingEvaluationScore")
	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, []string{"ab12", "cd345"}, typed.DismissedIDs)
}

func TestEvaluationScoreModel_Merge(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingEvaluationScore{
		ID:           "test-id",
		DismissedIDs: []string{"ab12"},
	}

	var model evaluationScoreModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.Equal(t, "test-id", model.ID.ValueString())

	var dismissed []string
	model.DismissedIDs.ElementsAs(context.Background(), &dismissed, false)
	assert.Equal(t, []string{"ab12"}, dismissed)
}

func TestEvaluationScoreModel_Merge_EmptyList(t *testing.T) {
	t.Parallel()

	src := &unifi.SettingEvaluationScore{ID: "test-id"}

	var model evaluationScoreModel
	diags := model.Merge(context.Background(), src)
	assert.False(t, diags.HasError())

	assert.False(t, model.DismissedIDs.IsNull())
	assert.Equal(t, 0, len(model.DismissedIDs.Elements()))
}
