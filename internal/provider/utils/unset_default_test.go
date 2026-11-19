package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrDefaultAndKeepUnset(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		current, planned, def int
		wantRead, wantWrite   int
	}{
		"unset, default planned": {current: 0, planned: 3600, def: 3600, wantRead: 3600, wantWrite: 0},
		"unset, value planned":   {current: 0, planned: 60, def: 3600, wantRead: 3600, wantWrite: 60},
		"set, default planned":   {current: 3600, planned: 3600, def: 3600, wantRead: 3600, wantWrite: 3600},
		"set, other planned":     {current: 120, planned: 60, def: 3600, wantRead: 120, wantWrite: 60},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantRead, OrDefault(tc.current, tc.def))
			assert.Equal(t, tc.wantWrite, KeepUnset(tc.current, tc.planned, tc.def))
		})
	}
	assert.Equal(t, "none", OrDefault("", "none"))
	assert.Equal(t, "", KeepUnset("", "none", "none"))
}
