package feature

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func typeName(t *testing.T, ds datasource.DataSource) string {
	t.Helper()
	var resp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{}, &resp)
	return resp.TypeName
}

func TestFeatureDatasource_Metadata(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "unifi_feature", typeName(t, NewFeatureDatasource()))
	assert.Equal(t, "unifi_features", typeName(t, NewFeaturesDatasource()))
	assert.Equal(t, "unifi_system_information", typeName(t, NewSystemInformationDatasource()))
}

func TestFeatureDatasourceModel_Site(t *testing.T) {
	t.Parallel()
	var m featureDatasourceModel
	m.SetSite("default")
	assert.Equal(t, "default", m.GetSite())
	assert.Equal(t, "default", m.GetRawSite().ValueString())
}
