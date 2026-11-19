package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cases := map[string]struct {
		in       ScheduleValues
		wantNull bool
	}{
		"zero is null": {in: ScheduleValues{}, wantNull: true},
		// Shape a UniFi Network 11.0.99 console returns for "Always".
		"always":      {in: ScheduleValues{Mode: "ALWAYS", RepeatOnDays: []string{}}},
		"weekly":      {in: ScheduleValues{Mode: "EVERY_WEEK", RepeatOnDays: []string{"sat", "sun"}, TimeRangeStart: "08:00", TimeRangeEnd: "23:59"}},
		"all day one": {in: ScheduleValues{Mode: "ONE_TIME_ONLY", Date: "2026-12-24", TimeAllDay: true, RepeatOnDays: []string{}}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			obj, diags := ScheduleToObject(ctx, tc.in)
			require.False(t, diags.HasError())
			assert.Equal(t, tc.wantNull, obj.IsNull())
			back, diags := ScheduleFromObject(ctx, obj)
			require.False(t, diags.HasError())
			if tc.wantNull {
				assert.True(t, back.IsZero())
				return
			}
			assert.ElementsMatch(t, tc.in.RepeatOnDays, back.RepeatOnDays)
			back.RepeatOnDays, tc.in.RepeatOnDays = nil, nil
			assert.Equal(t, tc.in, back)
		})
	}
}
