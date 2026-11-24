package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONValueSemanticEquals(t *testing.T) {
	t.Parallel()
	cases := map[string]struct {
		old, new JSONValue
		want     bool
	}{
		"identical": {JSONStringValue(`{"a":1}`), JSONStringValue(`{"a":1}`), true},
		"key order (jsonencode vs Go struct order)": {
			JSONStringValue(`{"trafficFilter":{"networkFilter":{"matchOpposite":false,"networkIds":["n"]},"type":"NETWORK"},"zoneId":"z"}`),
			JSONStringValue(`{"trafficFilter":{"networkFilter":{"networkIds":["n"],"matchOpposite":false},"type":"NETWORK"},"zoneId":"z"}`),
			true,
		},
		"whitespace":              {JSONStringValue(`{ "a" : [1, 2] }`), JSONStringValue(`{"a":[1,2]}`), true},
		"null key ignored":        {JSONStringValue(`{"type":"WPA2","passphrase":"x"}`), JSONStringValue(`{"passphrase":"x","radiusConfiguration":null,"type":"WPA2"}`), true},
		"nested null key ignored": {JSONStringValue(`{"a":{"b":1}}`), JSONStringValue(`{"a":{"b":1,"c":null}}`), true},
		"different value":         {JSONStringValue(`{"a":1}`), JSONStringValue(`{"a":2}`), false},
		"array order matters":     {JSONStringValue(`{"a":[1,2]}`), JSONStringValue(`{"a":[2,1]}`), false},
		"extra non-null key":      {JSONStringValue(`{"a":1}`), JSONStringValue(`{"a":1,"b":false}`), false},
		"empty array is not null": {JSONStringValue(`{"a":[]}`), JSONStringValue(`{}`), false},
		"invalid json strict":     {JSONStringValue(`{bad`), JSONStringValue(`{bad }`), false},
		"null vs null":            {JSONNull(), JSONNull(), true},
		"null vs value":           {JSONNull(), JSONStringValue(`{}`), false},
		"unknown vs value":        {JSONUnknown(), JSONStringValue(`{}`), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, diags := tc.old.StringSemanticEquals(context.Background(), tc.new)
			require.False(t, diags.HasError())
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestJSONStringOrNull(t *testing.T) {
	t.Parallel()
	assert.True(t, JSONStringOrNull("").IsNull())
	assert.Equal(t, `{}`, JSONStringOrNull(`{}`).ValueString())
}

func TestJSONTypeRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tfv, err := JSONStringValue(`{"a":1}`).ToTerraformValue(ctx)
	require.NoError(t, err)
	v, err := JSONType{}.ValueFromTerraform(ctx, tfv)
	require.NoError(t, err)
	jv, ok := v.(JSONValue)
	require.True(t, ok)
	assert.Equal(t, `{"a":1}`, jv.ValueString())
	assert.True(t, JSONType{}.Equal(jv.Type(ctx)))
}

func TestMarshalUnmarshalJSONValue(t *testing.T) {
	t.Parallel()
	type obj struct {
		A int `json:"a"`
	}
	v, diags := MarshalJSONValue("x", obj{A: 1})
	require.False(t, diags.HasError())
	assert.Equal(t, `{"a":1}`, v.ValueString())

	var nilPtr *obj
	v, diags = MarshalJSONValue("x", nilPtr)
	require.False(t, diags.HasError())
	assert.True(t, v.IsNull())

	var got obj
	ok, diags := UnmarshalJSONValue("x", JSONStringValue(`{"a":2}`), &got)
	require.False(t, diags.HasError())
	assert.True(t, ok)
	assert.Equal(t, 2, got.A)

	ok, diags = UnmarshalJSONValue("x", JSONNull(), &got)
	assert.False(t, ok)
	assert.False(t, diags.HasError())

	_, diags = UnmarshalJSONValue("x", JSONStringValue(`{bad`), &got)
	assert.True(t, diags.HasError())
}

// Values read from the controller are stored as Terraform's jsonencode() would write
// them, so an imported object matches a jsonencode()d configuration byte for byte.
func TestCanonicalJSON(t *testing.T) {
	t.Parallel()
	cases := map[string]struct{ in, want string }{
		// Go struct order (networkIds before matchOpposite) and an explicit null.
		"policy source": {
			`{"trafficFilter":{"networkFilter":{"networkIds":["n1"],"matchOpposite":false},"type":"NETWORK"},"zoneId":"z","ipsecFilter":null}`,
			`{"trafficFilter":{"networkFilter":{"matchOpposite":false,"networkIds":["n1"]},"type":"NETWORK"},"zoneId":"z"}`,
		},
		"large int keeps precision":    {`{"b":9007199254740993,"a":1.5}`, `{"a":1.5,"b":9007199254740993}`},
		"html escaped like jsonencode": {`{"a":"<x>&"}`, `{"a":"\u003cx\u003e\u0026"}`},
		"array order kept":             {`[3,1,2]`, `[3,1,2]`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := CanonicalJSON([]byte(tc.in))
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
	_, err := CanonicalJSON([]byte(`{bad`))
	assert.Error(t, err)
}

// G21: the controller returns networkIds/macAddresses in its own order.
func TestJSONSemanticallyEqual_UnorderedIDArrays(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{
			"networkIds order ignored",
			`{"trafficFilter":{"networkFilter":{"matchOpposite":false,"networkIds":["a","b","c"]},"type":"NETWORK"},"zoneId":"z"}`,
			`{"zoneId":"z","trafficFilter":{"type":"NETWORK","networkFilter":{"networkIds":["c","a","b"],"matchOpposite":false}}}`,
			true,
		},
		{"macAddresses order ignored", `{"macAddresses":["x","y"]}`, `{"macAddresses":["y","x"]}`, true},
		{"different members differ", `{"networkIds":["a","b"]}`, `{"networkIds":["a","c"]}`, false},
		{"missing member differs", `{"networkIds":["a","b"]}`, `{"networkIds":["a"]}`, false},
		{"other arrays stay ordered", `{"items":["a","b"]}`, `{"items":["b","a"]}`, false},
		{"non-string arrays stay ordered", `{"networkIds":[1,2]}`, `{"networkIds":[2,1]}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, JSONSemanticallyEqual(tt.a, tt.b))
		})
	}
}
