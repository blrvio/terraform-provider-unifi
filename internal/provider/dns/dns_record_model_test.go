package dns

import (
	"context"
	"testing"

	"github.com/blrvio/go-unifi/v11/unifi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// G14: the controller returns port/priority/weight = 0 on every static-dns record
// (shape read from a UniFi Network 11.0.99 console); only SRV/MX use them.
func TestDNSRecordMergeByType(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		in                                 unifi.DNSRecord
		portNull, priorityNull, weightNull bool
	}{
		"A":     {in: unifi.DNSRecord{RecordType: "A", Key: "argocd.wnst", Value: "10.1.160.217"}, portNull: true, priorityNull: true, weightNull: true},
		"CNAME": {in: unifi.DNSRecord{RecordType: "CNAME", Key: "ollama.local", Value: "x"}, portNull: true, priorityNull: true, weightNull: true},
		"MX":    {in: unifi.DNSRecord{RecordType: "MX", Priority: 10}, portNull: true, weightNull: true},
		"SRV":   {in: unifi.DNSRecord{RecordType: "SRV", Port: 5060, Priority: 10, Weight: 0}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := &dnsRecordModel{}
			require.False(t, m.Merge(context.Background(), &tc.in).HasError())
			assert.Equal(t, tc.portNull, m.Port.IsNull(), "port")
			assert.Equal(t, tc.priorityNull, m.Priority.IsNull(), "priority")
			assert.Equal(t, tc.weightNull, m.Weight.IsNull(), "weight")
		})
	}
}
