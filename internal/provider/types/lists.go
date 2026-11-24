package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/blrvio/terraform-provider-unifi/internal/provider/utils"
)

func DefaultEmptyList(elementType attr.Type) defaults.List {
	return listdefault.StaticValue(EmptyList(elementType))
}

func EmptyList(elementType attr.Type) types.List {
	return types.ListValueMust(elementType, []attr.Value{})
}

func ListElementsAs(ctx context.Context, list types.List, target interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	if !IsDefined(list) {
		return diags
	}
	if diagErr := list.ElementsAs(ctx, target, false); diagErr != nil {
		diags = append(diags, diagErr...)
	}
	return diags
}

func ListElementsToString(ctx context.Context, list types.List) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !IsDefined(list) {
		return "", diags
	}
	if list.ElementType(ctx) == types.StringType {
		var target []string
		diags.Append(ListElementsAs(ctx, list, &target)...)
		if diags.HasError() {
			return "", diags
		}
		return utils.JoinNonEmpty(target, ","), diags
	}
	diags.AddError("List is not a list of types.StringType", "List is not a list of strings")
	return "", diags
}

func StringToListElements(ctx context.Context, value string) (types.List, diag.Diagnostics) {
	countries := utils.SplitAndTrim(value, ",")
	if len(countries) == 0 {
		return types.ListNull(types.StringType), diag.Diagnostics{}
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, countries)
	if diags.HasError() {
		return types.ListNull(types.StringType), diags
	}
	return list, diags
}

// KeepStringOrder returns fresh in the order of prior when both hold the same strings (as multisets), so a
// controller that returns an unordered collection in its own order does not churn a list attribute.
// Otherwise fresh is returned unchanged.
func KeepStringOrder(prior, fresh []string) []string {
	if len(prior) != len(fresh) {
		return fresh
	}
	counts := make(map[string]int, len(fresh))
	for _, s := range fresh {
		counts[s]++
	}
	for _, s := range prior {
		if counts[s] == 0 {
			return fresh
		}
		counts[s]--
	}
	return prior
}
