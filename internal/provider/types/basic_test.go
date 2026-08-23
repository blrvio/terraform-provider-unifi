package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64OrNull(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		wantNull bool
		wantVal  float64
	}{
		{name: "zero is null", input: 0, wantNull: true},
		{name: "positive value", input: 12.5, wantNull: false, wantVal: 12.5},
		{name: "negative value", input: -3.25, wantNull: false, wantVal: -3.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Float64OrNull(tt.input)
			assert.Equal(t, tt.wantNull, got.IsNull())
			if !tt.wantNull {
				assert.Equal(t, tt.wantVal, got.ValueFloat64())
			}
		})
	}
}

func TestStringOrNull(t *testing.T) {
	t.Parallel()

	assert.True(t, StringOrNull("").IsNull())
	got := StringOrNull("hello")
	assert.False(t, got.IsNull())
	assert.Equal(t, "hello", got.ValueString())
}

func TestInt32OrNull(t *testing.T) {
	t.Parallel()

	assert.True(t, Int32OrNull(0).IsNull())
	got := Int32OrNull(42)
	assert.False(t, got.IsNull())
	assert.Equal(t, int32(42), got.ValueInt32())
}

func TestInt64OrNull(t *testing.T) {
	t.Parallel()

	assert.True(t, Int64OrNull(0).IsNull())
	got := Int64OrNull(42)
	assert.False(t, got.IsNull())
	assert.Equal(t, int64(42), got.ValueInt64())
}
