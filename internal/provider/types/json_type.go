package types

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// JSONType is a custom string type for attributes that hold a JSON document. Two
// values are semantically equal when they decode to the same document, ignoring
// whitespace, object key order and keys whose value is null. That lets a config
// built with jsonencode() (alphabetical keys) match the controller's echo (Go struct
// field order, explicit nulls) without a perpetual diff. The Terraform type is still
// a plain string, so existing state needs no upgrade.
type JSONType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = JSONType{}

func (t JSONType) Equal(o attr.Type) bool {
	other, ok := o.(JSONType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t JSONType) String() string {
	return "types.JSONType"
}

func (t JSONType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return JSONValue{StringValue: in}, nil
}

func (t JSONType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}
	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}
	return stringValuable, nil
}

func (t JSONType) ValueType(_ context.Context) attr.Value {
	return JSONValue{}
}

// JSONValue is the value type produced by JSONType.
type JSONValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuableWithSemanticEquals = JSONValue{}

// JSONStringValue, JSONNull and JSONUnknown build JSONValues.
func JSONStringValue(s string) JSONValue { return JSONValue{StringValue: basetypes.NewStringValue(s)} }
func JSONNull() JSONValue                { return JSONValue{StringValue: basetypes.NewStringNull()} }
func JSONUnknown() JSONValue             { return JSONValue{StringValue: basetypes.NewStringUnknown()} }

// JSONStringOrNull returns a null value for the empty string.
func JSONStringOrNull(s string) JSONValue {
	if s == "" {
		return JSONNull()
	}
	return JSONStringValue(s)
}

func (v JSONValue) Type(_ context.Context) attr.Type {
	return JSONType{}
}

func (v JSONValue) Equal(o attr.Value) bool {
	other, ok := o.(JSONValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals reports whether both values decode to the same JSON document
// (see JSONType). Null/unknown or undecodable values fall back to strict equality.
func (v JSONValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	newValue, ok := newValuable.(JSONValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("expected value type %T but got %T. Please report this to the provider developers.", v, newValuable),
		)
		return false, diags
	}
	if v.IsNull() || v.IsUnknown() || newValue.IsNull() || newValue.IsUnknown() {
		return v.StringValue.Equal(newValue.StringValue), diags
	}
	return JSONSemanticallyEqual(v.ValueString(), newValue.ValueString()), diags
}

// JSONSemanticallyEqual reports whether a and b decode to the same JSON document,
// ignoring whitespace, key order and null-valued object keys. Invalid JSON compares
// byte for byte.
func JSONSemanticallyEqual(a, b string) bool {
	if a == b {
		return true
	}
	var da, db any
	if json.Unmarshal([]byte(a), &da) != nil || json.Unmarshal([]byte(b), &db) != nil {
		return false
	}
	return reflect.DeepEqual(dropJSONNulls(da), dropJSONNulls(db))
}

func dropJSONNulls(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if val == nil {
				continue
			}
			out[k] = dropJSONNulls(val)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = dropJSONNulls(val)
		}
		return out
	default:
		return v
	}
}

// MarshalJSONValue encodes v as a compact JSON attribute value, reporting failures
// against the attribute name. A nil pointer/interface encodes as a null value.
func MarshalJSONValue(attrName string, v any) (JSONValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Pointer && reflect.ValueOf(v).IsNil()) {
		return JSONNull(), diags
	}
	raw, err := json.Marshal(v)
	if err != nil {
		diags.AddError("Failed to encode "+attrName, err.Error())
		return JSONNull(), diags
	}
	return JSONStringValue(string(raw)), diags
}

// UnmarshalJSONValue decodes a defined JSON attribute value into target; null or
// unknown values leave target untouched and report defined=false.
func UnmarshalJSONValue(attrName string, v JSONValue, target any) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v.IsNull() || v.IsUnknown() {
		return false, diags
	}
	if err := json.Unmarshal([]byte(v.ValueString()), target); err != nil {
		diags.AddError("Invalid "+attrName, fmt.Sprintf("`%s` is not valid JSON for this attribute: %s", attrName, err))
		return false, diags
	}
	return true, diags
}
