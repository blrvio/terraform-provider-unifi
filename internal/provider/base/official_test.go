package base

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOfficialAPIErrorDiagnostics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		wantErrors  int
		wantSummary string
	}{
		{name: "nil error yields no diagnostics", err: nil, wantErrors: 0},
		{
			name:        "unavailable sentinel",
			err:         unifi.ErrOfficialAPIUnavailable,
			wantErrors:  1,
			wantSummary: "UniFi Official API unavailable",
		},
		{
			name:        "wrapped unavailable sentinel",
			err:         fmt.Errorf("probe failed: %w", unifi.ErrOfficialAPIUnavailable),
			wantErrors:  1,
			wantSummary: "UniFi Official API unavailable",
		},
		{
			name:        "disabled sentinel",
			err:         unifi.ErrOfficialAPIDisabled,
			wantErrors:  1,
			wantSummary: "UniFi Official API disabled",
		},
		{
			name:        "generic error uses caller summary",
			err:         errors.New("boom"),
			wantErrors:  1,
			wantSummary: "custom summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diags := OfficialAPIErrorDiagnostics("custom summary", tt.err)
			assert.Equal(t, tt.wantErrors, diags.ErrorsCount())
			if tt.wantErrors > 0 {
				assert.Equal(t, tt.wantSummary, diags.Errors()[0].Summary())
			}
		})
	}
}

// mockOfficial builds an official.Client whose Sites().ResolveID returns the
// provided (id, err). Only the Sites accessor is wired; other groups panic if
// touched, which keeps the test honest about what it exercises.
func mockOfficial(id uuid.UUID, err error) official.Client {
	return &official.ClientMock{
		SitesFunc: func() official.SitesClient {
			return &official.SitesClientMock{
				ResolveIDFunc: func(_ context.Context, _ string) (uuid.UUID, error) {
					return id, err
				},
			}
		},
	}
}

func TestResolveSiteUUID(t *testing.T) {
	t.Parallel()

	want := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	tests := []struct {
		name        string
		id          uuid.UUID
		err         error
		wantErr     bool
		wantSummary string
		wantID      uuid.UUID
	}{
		{name: "resolves name to uuid", id: want, err: nil, wantErr: false, wantID: want},
		{
			name:        "site not found",
			err:         official.ErrSiteNotFound,
			wantErr:     true,
			wantSummary: "UniFi site not found",
		},
		{
			name:        "gate error is translated",
			err:         fmt.Errorf("%w: requires a new-style controller", unifi.ErrOfficialAPIUnavailable),
			wantErr:     true,
			wantSummary: "UniFi Official API unavailable",
		},
		{
			name:        "generic error surfaces resolve summary",
			err:         errors.New("network down"),
			wantErr:     true,
			wantSummary: "Failed to resolve UniFi site",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, diags := resolveSiteUUID(context.Background(), mockOfficial(tt.id, tt.err), "default")
			if tt.wantErr {
				assert.True(t, diags.HasError())
				assert.Equal(t, tt.wantSummary, diags.Errors()[0].Summary())
				return
			}
			assert.False(t, diags.HasError())
			assert.Equal(t, tt.wantID, id)
		})
	}
}

func TestRequireOfficialAPI(t *testing.T) {
	t.Parallel()

	t.Run("available controller passes", func(t *testing.T) {
		t.Parallel()
		c := &Client{Version: AsVersion("10.2.0"), OfficialAvailable: true}
		assert.False(t, c.RequireOfficialAPI().HasError())
	})

	t.Run("old controller is rejected with actionable message", func(t *testing.T) {
		t.Parallel()
		c := &Client{Version: AsVersion("9.5.0"), OfficialAvailable: false}
		diags := c.RequireOfficialAPI()
		assert.True(t, diags.HasError())
		assert.Equal(t, "UniFi Official API not available", diags.Errors()[0].Summary())
	})
}

func TestSupportsOfficialAPI(t *testing.T) {
	t.Parallel()

	assert.True(t, (&Client{Version: ControllerVersionOfficialAPI}).SupportsOfficialAPI())
	assert.True(t, (&Client{Version: AsVersion("10.5.67")}).SupportsOfficialAPI())
	assert.False(t, (&Client{Version: AsVersion("10.1.77")}).SupportsOfficialAPI())
	assert.False(t, (&Client{Version: AsVersion("9.0.108")}).SupportsOfficialAPI())
}
