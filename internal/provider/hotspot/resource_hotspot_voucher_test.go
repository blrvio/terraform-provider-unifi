package hotspot

import (
	"context"
	"testing"
	"time"

	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func i64ptr(v int64) *int64 { return &v }

func TestHotspotVoucherAsUnifiModel_AllFields(t *testing.T) {
	t.Parallel()
	m := &hotspotVoucherModel{
		Name:                 types.StringValue("front-desk"),
		TimeLimitMinutes:     types.Int64Value(1440),
		AuthorizedGuestLimit: types.Int64Value(5),
		DataUsageLimitMBytes: types.Int64Value(2048),
		RxRateLimitKbps:      types.Int64Value(5000),
		TxRateLimitKbps:      types.Int64Value(1000),
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	req, ok := body.(*official.HotspotVoucherCreationRequest)
	require.True(t, ok, "expected *official.HotspotVoucherCreationRequest, got %T", body)

	require.NotNil(t, req.Count)
	assert.Equal(t, int32(1), *req.Count, "Count must always be 1")
	assert.Equal(t, "front-desk", req.Name)
	assert.Equal(t, int64(1440), req.TimeLimitMinutes)
	require.NotNil(t, req.AuthorizedGuestLimit)
	assert.Equal(t, int64(5), *req.AuthorizedGuestLimit)
	require.NotNil(t, req.DataUsageLimitMBytes)
	assert.Equal(t, int64(2048), *req.DataUsageLimitMBytes)
	require.NotNil(t, req.RxRateLimitKbps)
	assert.Equal(t, int64(5000), *req.RxRateLimitKbps)
	require.NotNil(t, req.TxRateLimitKbps)
	assert.Equal(t, int64(1000), *req.TxRateLimitKbps)
}

func TestHotspotVoucherAsUnifiModel_OptionalsNil(t *testing.T) {
	t.Parallel()
	m := &hotspotVoucherModel{
		Name:             types.StringValue("minimal"),
		TimeLimitMinutes: types.Int64Value(60),
		// All optional int64s left null -> nil pointers in the request.
	}
	body, diags := m.AsUnifiModel(context.Background())
	require.False(t, diags.HasError(), "AsUnifiModel diagnostics: %v", diags)

	req, ok := body.(*official.HotspotVoucherCreationRequest)
	require.True(t, ok)

	require.NotNil(t, req.Count)
	assert.Equal(t, int32(1), *req.Count)
	assert.Equal(t, "minimal", req.Name)
	assert.Equal(t, int64(60), req.TimeLimitMinutes)
	assert.Nil(t, req.AuthorizedGuestLimit)
	assert.Nil(t, req.DataUsageLimitMBytes)
	assert.Nil(t, req.RxRateLimitKbps)
	assert.Nil(t, req.TxRateLimitKbps)
}

func TestHotspotVoucherMerge_FullDetails(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	created := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	activated := time.Date(2026, 8, 23, 11, 0, 0, 0, time.UTC)
	expires := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)

	v := &official.HotspotVoucherDetails{
		Id:                   id,
		Code:                 "1234-5678",
		Name:                 "front-desk",
		CreatedAt:            created,
		ActivatedAt:          &activated,
		ExpiresAt:            &expires,
		Expired:              false,
		AuthorizedGuestCount: 2,
		AuthorizedGuestLimit: i64ptr(5),
		DataUsageLimitMBytes: i64ptr(2048),
		TimeLimitMinutes:     1440,
		RxRateLimitKbps:      i64ptr(5000),
		TxRateLimitKbps:      i64ptr(1000),
	}

	m := &hotspotVoucherModel{}
	diags := m.Merge(context.Background(), v)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, id.String(), m.GetID())
	assert.Equal(t, "front-desk", m.Name.ValueString())
	assert.Equal(t, int64(1440), m.TimeLimitMinutes.ValueInt64())
	assert.Equal(t, int64(5), m.AuthorizedGuestLimit.ValueInt64())
	assert.Equal(t, int64(2048), m.DataUsageLimitMBytes.ValueInt64())
	assert.Equal(t, int64(5000), m.RxRateLimitKbps.ValueInt64())
	assert.Equal(t, int64(1000), m.TxRateLimitKbps.ValueInt64())

	assert.Equal(t, "1234-5678", m.Code.ValueString())
	assert.Equal(t, created.Format(time.RFC3339), m.CreatedAt.ValueString())
	assert.Equal(t, activated.Format(time.RFC3339), m.ActivatedAt.ValueString())
	assert.Equal(t, expires.Format(time.RFC3339), m.ExpiresAt.ValueString())
	assert.False(t, m.Expired.ValueBool())
	assert.Equal(t, int64(2), m.AuthorizedGuestCount.ValueInt64())
}

func TestHotspotVoucherMerge_NullPointersAndTimes(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")

	v := &official.HotspotVoucherDetails{
		Id:                   id,
		Code:                 "abcd-efgh",
		Name:                 "minimal",
		CreatedAt:            time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC),
		ActivatedAt:          nil,
		ExpiresAt:            nil,
		Expired:              true,
		AuthorizedGuestCount: 0,
		AuthorizedGuestLimit: nil,
		DataUsageLimitMBytes: nil,
		TimeLimitMinutes:     60,
		RxRateLimitKbps:      nil,
		TxRateLimitKbps:      nil,
	}

	m := &hotspotVoucherModel{}
	diags := m.Merge(context.Background(), v)
	require.False(t, diags.HasError(), "Merge diagnostics: %v", diags)

	assert.Equal(t, id.String(), m.GetID())
	assert.Equal(t, int64(60), m.TimeLimitMinutes.ValueInt64())
	// nil *int64 -> null.
	assert.True(t, m.AuthorizedGuestLimit.IsNull())
	assert.True(t, m.DataUsageLimitMBytes.IsNull())
	assert.True(t, m.RxRateLimitKbps.IsNull())
	assert.True(t, m.TxRateLimitKbps.IsNull())
	// nil *time.Time -> null string.
	assert.True(t, m.ActivatedAt.IsNull())
	assert.True(t, m.ExpiresAt.IsNull())
	// Zero createdAt would be null too, but here it is set.
	assert.False(t, m.CreatedAt.IsNull())
	assert.True(t, m.Expired.ValueBool())
	assert.Equal(t, int64(0), m.AuthorizedGuestCount.ValueInt64())
}

func TestHotspotVoucherMerge_WrongType(t *testing.T) {
	t.Parallel()
	m := &hotspotVoucherModel{}
	diags := m.Merge(context.Background(), &official.VoucherCreationResult{})
	assert.True(t, diags.HasError())
}

func TestCreatedVoucherID(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	result := &official.VoucherCreationResult{
		Vouchers: &[]official.HotspotVoucherDetails{{Id: id, Code: "1234"}},
	}
	got, diags := createdVoucherID(result)
	require.False(t, diags.HasError(), "createdVoucherID diagnostics: %v", diags)
	assert.Equal(t, id, got)

	// Empty result -> error.
	_, diags = createdVoucherID(&official.VoucherCreationResult{})
	assert.True(t, diags.HasError())

	// Nil result -> error.
	_, diags = createdVoucherID(nil)
	assert.True(t, diags.HasError())
}
