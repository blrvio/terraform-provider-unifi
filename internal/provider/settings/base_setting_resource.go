package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/blrvio/go-unifi/v11/unifi"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/base"
)

type settingWriter = func(context.Context, *base.Client, string, interface{}) (interface{}, error)

// SettingOption customizes NewSettingResource.
type SettingOption func(*settingOptions)

type settingOptions struct {
	mapWriteError func(error) error
}

// WithWriteErrorMapper rewrites errors from the setting's PUT (e.g. into a clearer
// "not supported on this controller" message).
func WithWriteErrorMapper(fn func(error) error) SettingOption {
	return func(o *settingOptions) { o.mapWriteError = fn }
}

// NewSettingResource creates a singleton setting resource for the controller setting
// stored under key (a unifi.Setting*Key constant).
//
// Writes are read-modify-write. The set/setting PUT replaces the whole document, and
// a model's body carries zero values for every non-omitempty SDK field it does not
// model, so the body is not sent as is: it is compared with the body the same model
// builds from the current setting, and only the fields that differ are patched over
// the raw current document, which is then PUT in full. Fields the model (or the SDK
// struct) does not know about are kept. Unset Optional+Computed attributes are taken
// from state rather than sent as zero values (ResolveUnknownsFromState).
func NewSettingResource[T base.ResourceModel](
	typeName string,
	key string,
	modelFactory func() T,
	getter func(context.Context, *base.Client, string) (interface{}, error),
	updater settingWriter,
	opts ...SettingOption,
) *base.GenericResource[T] {
	var o settingOptions
	for _, opt := range opts {
		opt(&o)
	}
	baseline := func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
		current, err := getter(ctx, client, site)
		if err != nil {
			return nil, err
		}
		m := modelFactory()
		if diags := m.Merge(ctx, current); diags.HasError() {
			return nil, fmt.Errorf("building the current %s body: %v", key, diags)
		}
		body, diags := m.AsUnifiModel(ctx)
		if diags.HasError() {
			return nil, fmt.Errorf("building the current %s body: %v", key, diags)
		}
		return body, nil
	}
	write := readModifyWriteSetting(key, getter, baseline, updater, o)
	return base.NewGenericResource(
		typeName,
		modelFactory,
		base.ResourceFunctions{
			Read: func(ctx context.Context, client *base.Client, site, _ string) (interface{}, error) {
				res, err := getter(ctx, client, site)
				if errors.Is(err, unifi.ErrNotFound) {
					return nil, fmt.Errorf("%w: the controller has no stored %q setting yet (it keeps defaults until the setting is "+
						"first saved), or this controller version does not support it; there is nothing to import", err, key)
				}
				return res, err
			},
			Create: write,
			Update: write,
			// Settings PUT responses (the write echo) are eventually
			// consistent on the controller and can intermittently omit
			// collection fields, causing "inconsistent result after apply".
			// Re-read from the persisted datastore (GetSetting*) so post-apply
			// state is reliable.
			ReadAfterWrite:           true,
			ResolveUnknownsFromState: true,
		})
}

type settingReader = func(context.Context, *base.Client, string) (interface{}, error)

func readModifyWriteSetting(key string, getter, baseline settingReader, updater settingWriter, o settingOptions) settingWriter {
	mapErr := func(err error) error {
		if err != nil && o.mapWriteError != nil {
			return o.mapWriteError(err)
		}
		return err
	}
	return func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
		current, err := currentSettingDocument(ctx, client.Client, site, key)
		if err != nil {
			return nil, err
		}
		if current == nil {
			// No stored document yet: nothing to preserve, write the typed body.
			res, err := updater(ctx, client, site, body)
			return res, mapErr(err)
		}
		before, err := baseline(ctx, client, site)
		if err != nil {
			return nil, err
		}
		merged, err := patchSettingDocument(current, before, body, key)
		if err != nil {
			return nil, err
		}
		res, err := client.SetSetting(ctx, site, key, merged)
		if errors.Is(err, unifi.ErrNotFound) || (err == nil && res == nil) {
			// Some controllers answer the PUT without echoing the document.
			return getter(ctx, client, site)
		}
		return res, mapErr(err)
	}
}

// currentSettingDocument returns the raw stored document for key, or nil when the
// controller has none.
func currentSettingDocument(ctx context.Context, client unifi.Client, site, key string) (map[string]any, error) {
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := client.Get(ctx, fmt.Sprintf("s/%s/get/setting", site), nil, &resp); err != nil {
		return nil, fmt.Errorf("reading setting %s before update: %w", key, err)
	}
	for _, d := range resp.Data {
		if k, _ := d["key"].(string); k == key {
			return d, nil
		}
	}
	return nil, nil
}

// overlaySettingDocument deep-merges the JSON form of body over the current document.
// Objects merge key by key; arrays and scalars from body replace the current value.
func overlaySettingDocument(current map[string]any, body interface{}, key string) (map[string]any, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encoding setting %s: %w", key, err)
	}
	var overlay map[string]any
	if err := json.Unmarshal(raw, &overlay); err != nil {
		return nil, fmt.Errorf("decoding setting %s: %w", key, err)
	}
	merged := deepMergeJSON(current, overlay)
	merged["key"] = key
	return merged, nil
}

// patchSettingDocument applies to current only the fields where body differs from
// before (the body the same model builds from the current setting).
func patchSettingDocument(current map[string]any, before, body interface{}, key string) (map[string]any, error) {
	beforeMap, err := toJSONMap(before, key)
	if err != nil {
		return nil, err
	}
	afterMap, err := toJSONMap(body, key)
	if err != nil {
		return nil, err
	}
	merged := deepMergeJSON(current, diffJSON(beforeMap, afterMap))
	merged["key"] = key
	return merged, nil
}

func toJSONMap(v interface{}, key string) (map[string]any, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encoding setting %s: %w", key, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("decoding setting %s: %w", key, err)
	}
	return m, nil
}

// diffJSON returns the keys of after whose value differs from before; nested objects
// are diffed recursively, other values (arrays included) compare as a whole.
func diffJSON(before, after map[string]any) map[string]any {
	out := map[string]any{}
	for k, av := range after {
		bv, ok := before[k]
		if am, isMap := av.(map[string]any); isMap {
			if bm, bIsMap := bv.(map[string]any); ok && bIsMap {
				if sub := diffJSON(bm, am); len(sub) > 0 {
					out[k] = sub
				}
				continue
			}
		}
		if !ok || !reflect.DeepEqual(bv, av) {
			out[k] = av
		}
	}
	return out
}

func deepMergeJSON(dst, src map[string]any) map[string]any {
	out := make(map[string]any, len(dst)+len(src))
	for k, v := range dst {
		out[k] = v
	}
	for k, sv := range src {
		if sm, ok := sv.(map[string]any); ok {
			if dm, ok := out[k].(map[string]any); ok {
				out[k] = deepMergeJSON(dm, sm)
				continue
			}
		}
		out[k] = sv
	}
	return out
}
