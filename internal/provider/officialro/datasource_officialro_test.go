package officialro

import (
	"context"
	"errors"
	"iter"
	"testing"

	"github.com/blrvio/go-unifi/v10/unifi"
	"github.com/blrvio/go-unifi/v10/unifi/official"
	"github.com/google/uuid"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

// seqOf builds an iter.Seq2[T, error] over the given items, matching the shape of
// the Official-API ListXxxAll iterators so it can back a *ClientMock.
func seqOf[T any](items []T) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
	}
}

// seqErr builds an iterator that yields a single error, exercising the
// short-circuit path of base.CollectAll.
func seqErr[T any](err error) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		yield(zero, err)
	}
}

func mustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}

func TestFindByID(t *testing.T) {
	type item struct {
		id   string
		name string
	}
	items := []item{{"a", "alpha"}, {"b", "bravo"}, {"c", "charlie"}}
	idOf := func(i item) string { return i.id }

	tests := []struct {
		name      string
		id        string
		wantFound bool
		wantName  string
	}{
		{"first", "a", true, "alpha"},
		{"middle", "b", true, "bravo"},
		{"last", "c", true, "charlie"},
		{"missing", "z", false, ""},
		{"empty", "", false, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, found := findByID(items, tc.id, idOf)
			if found != tc.wantFound {
				t.Fatalf("found = %v, want %v", found, tc.wantFound)
			}
			if found && got.name != tc.wantName {
				t.Fatalf("name = %q, want %q", got.name, tc.wantName)
			}
			if !found && got != (item{}) {
				t.Fatalf("expected zero value on miss, got %+v", got)
			}
		})
	}
}

func TestFindByID_emptySlice(t *testing.T) {
	got, found := findByID(nil, "a", func(s string) string { return s })
	if found {
		t.Fatalf("expected not found for empty slice, got %q", got)
	}
}

func TestUUIDsToStrings(t *testing.T) {
	a := "11111111-1111-1111-1111-111111111111"
	b := "22222222-2222-2222-2222-222222222222"

	t.Run("nil", func(t *testing.T) {
		if got := uuidsToStrings(nil); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		got := uuidsToStrings([]uuid.UUID{})
		if len(got) != 0 {
			t.Fatalf("expected empty slice, got %v", got)
		}
	})

	t.Run("order preserved", func(t *testing.T) {
		in := []uuid.UUID{mustUUID(t, a), mustUUID(t, b)}
		got := uuidsToStrings(in)
		want := []string{a, b}
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}

// TestCollectAndSelect exercises the real list-and-match pipeline used by the
// WAN/VPN/tunnel/device-tag Read handlers: drain a mock ListXxxAll iterator with
// base.CollectAll, then select by id with findByID.
func TestCollectAndSelect(t *testing.T) {
	wantID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	otherID := mustUUID(t, "44444444-4444-4444-4444-444444444444")

	mock := &official.SupportingClientMock{
		ListWansAllFunc: func(_ context.Context, _ uuid.UUID, _ string) iter.Seq2[official.WANOverview, error] {
			return seqOf([]official.WANOverview{
				{Id: otherID, Name: "wan-other"},
				{Id: wantID, Name: "wan-wanted"},
			})
		},
	}

	wans, err := base.CollectAll(mock.ListWansAll(context.Background(), uuid.UUID{}, ""))
	if err != nil {
		t.Fatalf("CollectAll: %v", err)
	}
	if len(wans) != 2 {
		t.Fatalf("expected 2 wans, got %d", len(wans))
	}

	wan, found := findByID(wans, wantID.String(), func(w official.WANOverview) string { return w.Id.String() })
	if !found {
		t.Fatalf("expected to find wan %s", wantID)
	}
	if wan.Name != "wan-wanted" {
		t.Fatalf("name = %q, want %q", wan.Name, "wan-wanted")
	}

	if _, found := findByID(wans, "no-such-id", func(w official.WANOverview) string { return w.Id.String() }); found {
		t.Fatalf("did not expect to find a wan for a bogus id")
	}
}

func TestCollectAll_errorShortCircuits(t *testing.T) {
	sentinel := errors.New("boom")
	mock := &official.SupportingClientMock{
		ListVpnServersAllFunc: func(_ context.Context, _ uuid.UUID, _ string) iter.Seq2[official.VPNServerOverview, error] {
			return seqErr[official.VPNServerOverview](sentinel)
		},
	}
	_, err := base.CollectAll(mock.ListVpnServersAll(context.Background(), uuid.UUID{}, ""))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

// TestGetLag_notFound verifies the switch_lag Read handler's not-found branch:
// GetLag returning unifi.ErrNotFound must be recognizable via errors.Is.
func TestGetLag_notFound(t *testing.T) {
	mock := &official.SwitchingClientMock{
		GetLagFunc: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*official.LAGDetails, error) {
			return nil, unifi.ErrNotFound
		},
	}
	_, err := mock.GetLag(context.Background(), uuid.UUID{}, uuid.UUID{})
	if !errors.Is(err, unifi.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetLag_found(t *testing.T) {
	lagID := mustUUID(t, "55555555-5555-5555-5555-555555555555")
	mock := &official.SwitchingClientMock{
		GetLagFunc: func(_ context.Context, _ uuid.UUID, id uuid.UUID) (*official.LAGDetails, error) {
			return &official.LAGDetails{Id: id, Type: "LOCAL"}, nil
		},
	}
	lag, err := mock.GetLag(context.Background(), uuid.UUID{}, lagID)
	if err != nil {
		t.Fatalf("GetLag: %v", err)
	}
	if lag.Id != lagID || lag.Type != "LOCAL" {
		t.Fatalf("unexpected lag %+v", lag)
	}
}
