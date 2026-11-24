package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeepStringOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		prior, fresh []string
		want         []string
	}{
		{"same set keeps prior order", []string{"a", "b", "c"}, []string{"c", "a", "b"}, []string{"a", "b", "c"}},
		{"identical", []string{"a", "b"}, []string{"a", "b"}, []string{"a", "b"}},
		{"different member takes fresh", []string{"a", "b"}, []string{"a", "x"}, []string{"a", "x"}},
		{"added member takes fresh", []string{"a"}, []string{"b", "a"}, []string{"b", "a"}},
		{"duplicates must match counts", []string{"a", "a", "b"}, []string{"a", "b", "b"}, []string{"a", "b", "b"}},
		{"no prior takes fresh", nil, []string{"b", "a"}, []string{"b", "a"}},
		{"both empty", []string{}, []string{}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, KeepStringOrder(tt.prior, tt.fresh))
		})
	}
}
