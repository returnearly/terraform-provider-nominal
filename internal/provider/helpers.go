package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func optionalString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() {
		return "", false
	}

	return value.ValueString(), true
}

func stringOrNull(value *string) types.String {
	if value == nil || *value == "" {
		return types.StringNull()
	}

	return types.StringValue(*value)
}

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}

	return types.StringValue(value)
}

func floatOrNull(value *float64) types.Float64 {
	if value == nil {
		return types.Float64Null()
	}

	return types.Float64Value(*value)
}

func stringList(ctx context.Context, list types.List) ([]string, bool) {
	if list.IsNull() || list.IsUnknown() {
		return nil, false
	}

	var values []string
	_ = list.ElementsAs(ctx, &values, false)
	if values == nil {
		values = []string{}
	}

	return values, true
}

func stringListValue(values []string) types.List {
	if values == nil {
		values = []string{}
	}

	elems := make([]attr.Value, 0, len(values))
	for _, value := range values {
		elems = append(elems, types.StringValue(value))
	}

	return types.ListValueMust(types.StringType, elems)
}

func idsFromNodes(nodes []gqlID) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.ID != "" {
			ids = append(ids, node.ID)
		}
	}

	return ids
}

func keyValuesInput(items []keyValueModel) []map[string]string {
	if len(items) == 0 {
		return nil
	}

	out := make([]map[string]string, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]string{
			"key":   item.Key.ValueString(),
			"value": item.Value.ValueString(),
		})
	}

	return out
}

func keyValuesModel(items []gqlKeyValue) []keyValueModel {
	if len(items) == 0 {
		return nil
	}

	out := make([]keyValueModel, 0, len(items))
	for _, item := range items {
		out = append(out, keyValueModel{
			Key:   types.StringValue(item.Key),
			Value: types.StringValue(item.Value),
		})
	}

	return out
}

func expressionsFromConditions(conditions []gqlCondition) []string {
	out := make([]string, 0, len(conditions))
	for _, condition := range conditions {
		if condition.Expression != "" {
			out = append(out, condition.Expression)
		}
	}

	return out
}
