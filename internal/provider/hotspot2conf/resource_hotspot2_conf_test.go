package hotspot2conf

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func capabList(t *testing.T, items ...capabModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), capabObjectType(), items)
	require.False(t, diags.HasError(), "capab ListValueFrom diagnostics: %v", diags)
	return list
}

func languageTextList(t *testing.T, items ...languageTextModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), languageTextObjectType(), items)
	require.False(t, diags.HasError(), "languageText ListValueFrom diagnostics: %v", diags)
	return list
}

func osuList(t *testing.T, items ...osuModel) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), osuObjectType(), items)
	require.False(t, diags.HasError(), "osu ListValueFrom diagnostics: %v", diags)
	return list
}

func emptyLanguageTextList(t *testing.T) types.List {
	t.Helper()
	return languageTextList(t)
}

func emptyOsuIconList(t *testing.T) types.List {
	t.Helper()
	list, diags := types.ListValueFrom(context.Background(), osuIconObjectType(), []osuIconModel{})
	require.False(t, diags.HasError(), "osuIcon ListValueFrom diagnostics: %v", diags)
	return list
}

func TestHotspot2ConfModel_AsUnifiModel(t *testing.T) {
	t.Parallel()

	model := Hotspot2ConfModel{
		Name:          types.StringValue("passpoint-1"),
		NetworkType:   types.Int32Value(2),
		VenueGroup:    types.Int32Value(2),
		VenueType:     types.Int32Value(8),
		DisableDgaf:   types.BoolValue(true),
		MetricsStatus: types.BoolValue(true),
		Capab: capabList(t, capabModel{
			Port:     types.Int32Value(443),
			Protocol: types.StringValue("tcp"),
			Status:   types.StringValue("open"),
		}),
		Osu: osuList(t, osuModel{
			Description: languageTextList(t, languageTextModel{
				Language: types.StringValue("eng"),
				Text:     types.StringValue("Example OSU"),
			}),
			FriendlyName:     emptyLanguageTextList(t),
			Icon:             emptyOsuIconList(t),
			MethodOmaDm:      types.BoolValue(true),
			MethodSoapXMLSpp: types.BoolValue(false),
			Nai:              types.StringValue("anonymous@example.com"),
			ServerURI:        types.StringValue("https://osu.example.com"),
		}),
	}
	model.ID = types.StringValue("test-id")

	unifiModel, diags := model.AsUnifiModel(context.Background())
	assert.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	typed, ok := unifiModel.(*unifi.Hotspot2Conf)
	require.True(t, ok, "Expected model to be *unifi.Hotspot2Conf")

	assert.Equal(t, "test-id", typed.ID)
	assert.Equal(t, "passpoint-1", typed.Name)
	assert.Equal(t, 2, typed.NetworkType)
	assert.Equal(t, 2, typed.VenueGroup)
	assert.Equal(t, 8, typed.VenueType)
	assert.True(t, typed.DisableDgaf)
	assert.True(t, typed.MetricsStatus)

	require.Len(t, typed.Capab, 1)
	assert.Equal(t, 443, typed.Capab[0].Port)
	assert.Equal(t, "tcp", typed.Capab[0].Protocol)
	assert.Equal(t, "open", typed.Capab[0].Status)

	require.Len(t, typed.Osu, 1)
	assert.True(t, typed.Osu[0].MethodOmaDm)
	assert.Equal(t, "anonymous@example.com", typed.Osu[0].Nai)
	assert.Equal(t, "https://osu.example.com", typed.Osu[0].ServerUri)
	require.Len(t, typed.Osu[0].Description, 1)
	assert.Equal(t, "eng", typed.Osu[0].Description[0].Language)
	assert.Equal(t, "Example OSU", typed.Osu[0].Description[0].Text)
}

func TestHotspot2ConfModel_Merge(t *testing.T) {
	t.Parallel()

	model := &unifi.Hotspot2Conf{
		ID:            "merge-id",
		Name:          "passpoint-1",
		NetworkType:   2,
		VenueGroup:    2,
		VenueType:     8,
		DisableDgaf:   true,
		MetricsStatus: true,
		Capab: []unifi.Hotspot2ConfCapab{
			{Port: 443, Protocol: "tcp", Status: "open"},
		},
		Osu: []unifi.Hotspot2ConfOsu{
			{
				Description: []unifi.Hotspot2ConfDescription{
					{Language: "eng", Text: "Example OSU"},
				},
				MethodOmaDm: true,
				Nai:         "anonymous@example.com",
				ServerUri:   "https://osu.example.com",
			},
		},
	}

	var m Hotspot2ConfModel
	diags := m.Merge(context.Background(), model)
	assert.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, "merge-id", m.ID.ValueString())
	assert.Equal(t, "passpoint-1", m.Name.ValueString())
	assert.Equal(t, int32(2), m.NetworkType.ValueInt32())
	assert.Equal(t, int32(2), m.VenueGroup.ValueInt32())
	assert.Equal(t, int32(8), m.VenueType.ValueInt32())
	assert.True(t, m.DisableDgaf.ValueBool())
	assert.True(t, m.MetricsStatus.ValueBool())

	var capab []capabModel
	m.Capab.ElementsAs(context.Background(), &capab, false)
	require.Len(t, capab, 1)
	assert.Equal(t, int32(443), capab[0].Port.ValueInt32())
	assert.Equal(t, "tcp", capab[0].Protocol.ValueString())
	assert.Equal(t, "open", capab[0].Status.ValueString())

	var osu []osuModel
	m.Osu.ElementsAs(context.Background(), &osu, false)
	require.Len(t, osu, 1)
	assert.True(t, osu[0].MethodOmaDm.ValueBool())
	assert.Equal(t, "anonymous@example.com", osu[0].Nai.ValueString())

	var desc []languageTextModel
	osu[0].Description.ElementsAs(context.Background(), &desc, false)
	require.Len(t, desc, 1)
	assert.Equal(t, "eng", desc[0].Language.ValueString())
	assert.Equal(t, "Example OSU", desc[0].Text.ValueString())
}

func TestHotspot2ConfModel_Merge_EmptyLists(t *testing.T) {
	t.Parallel()

	var m Hotspot2ConfModel
	diags := m.Merge(context.Background(), &unifi.Hotspot2Conf{ID: "id", Name: "empty"})
	assert.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	// nil SDK slices must coalesce to non-null empty lists to avoid perpetual diffs.
	assert.False(t, m.Capab.IsNull(), "capab should be non-null")
	assert.Equal(t, 0, len(m.Capab.Elements()))

	assert.False(t, m.DomainNameList.IsNull(), "domain_name_list should be non-null")
	assert.Equal(t, 0, len(m.DomainNameList.Elements()))

	assert.False(t, m.Osu.IsNull(), "osu should be non-null")
	assert.Equal(t, 0, len(m.Osu.Elements()))
}

func TestHotspot2ConfModel_Merge_InvalidType(t *testing.T) {
	t.Parallel()

	var m Hotspot2ConfModel
	diags := m.Merge(context.Background(), &unifi.Tag{})
	assert.True(t, diags.HasError())
}
